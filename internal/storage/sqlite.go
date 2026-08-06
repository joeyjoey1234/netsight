package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"netsight/internal/model"
	"time"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite: %w", err)
	}

	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA foreign_keys=ON")

	store := &SQLiteStore{db: db}
	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return store, nil
}

func (s *SQLiteStore) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS projects (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		settings_json TEXT
	);

	CREATE TABLE IF NOT EXISTS scans (
		id TEXT PRIMARY KEY,
		project_id TEXT REFERENCES projects(id) ON DELETE CASCADE,
		timestamp DATETIME,
		subnet TEXT,
		duration_ms INTEGER,
		preset TEXT,
		status TEXT,
		devices_found INTEGER,
		findings_json TEXT
	);

	CREATE TABLE IF NOT EXISTS devices (
		id TEXT PRIMARY KEY,
		project_id TEXT REFERENCES projects(id) ON DELETE CASCADE,
		mac TEXT,
		ips_json TEXT,
		vendor TEXT,
		hostname TEXT,
		os TEXT,
		role TEXT,
		model TEXT,
		vlans_json TEXT,
		first_seen DATETIME,
		last_seen DATETIME,
		notes TEXT
	);

	CREATE TABLE IF NOT EXISTS ports (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		device_id TEXT REFERENCES devices(id) ON DELETE CASCADE,
		scan_id TEXT REFERENCES scans(id) ON DELETE CASCADE,
		number INTEGER,
		protocol TEXT,
		service TEXT,
		version TEXT,
		banner TEXT,
		state TEXT
	);

	CREATE TABLE IF NOT EXISTS links (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_id TEXT REFERENCES devices(id) ON DELETE CASCADE,
		target_id TEXT REFERENCES devices(id) ON DELETE CASCADE,
		type TEXT,
		src_port TEXT,
		dst_port TEXT,
		vlan INTEGER
	);

	CREATE TABLE IF NOT EXISTS findings (
		id TEXT PRIMARY KEY,
		scan_id TEXT REFERENCES scans(id) ON DELETE CASCADE,
		device_id TEXT REFERENCES devices(id) ON DELETE SET NULL,
		type TEXT,
		severity TEXT,
		title TEXT,
		description TEXT,
		recommendation TEXT,
		timestamp DATETIME
	);

	CREATE TABLE IF NOT EXISTS scan_devices (
		scan_id TEXT REFERENCES scans(id) ON DELETE CASCADE,
		device_id TEXT REFERENCES devices(id) ON DELETE CASCADE,
		PRIMARY KEY (scan_id, device_id)
	);

	CREATE INDEX IF NOT EXISTS idx_scans_project ON scans(project_id);
	CREATE INDEX IF NOT EXISTS idx_devices_project ON devices(project_id);
	CREATE INDEX IF NOT EXISTS idx_ports_device ON ports(device_id);
	CREATE INDEX IF NOT EXISTS idx_links_source ON links(source_id);
	CREATE INDEX IF NOT EXISTS idx_links_target ON links(target_id);
	CREATE INDEX IF NOT EXISTS idx_findings_scan ON findings(scan_id);
	`
	_, err := s.db.Exec(schema)
	return err
}

func marshalJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func (s *SQLiteStore) CreateProject(project *model.Project) error {
	_, err := s.db.Exec(
		"INSERT INTO projects (id, name, created_at, settings_json) VALUES (?, ?, ?, ?)",
		project.ID, project.Name, project.Created, marshalJSON(project.Settings),
	)
	return err
}

func (s *SQLiteStore) GetProject(id string) (*model.Project, error) {
	row := s.db.QueryRow("SELECT id, name, created_at, settings_json FROM projects WHERE id = ?", id)
	p := &model.Project{}
	var settingsJSON string
	err := row.Scan(&p.ID, &p.Name, &p.Created, &settingsJSON)
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(settingsJSON), &p.Settings)
	return p, nil
}

func (s *SQLiteStore) ListProjects() ([]*model.Project, error) {
	rows, err := s.db.Query("SELECT id, name, created_at, settings_json FROM projects ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*model.Project
	for rows.Next() {
		p := &model.Project{}
		var settingsJSON string
		if err := rows.Scan(&p.ID, &p.Name, &p.Created, &settingsJSON); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(settingsJSON), &p.Settings)
		projects = append(projects, p)
	}
	return projects, nil
}

func (s *SQLiteStore) DeleteProject(id string) error {
	_, err := s.db.Exec("DELETE FROM projects WHERE id = ?", id)
	return err
}

func (s *SQLiteStore) SaveScan(projectID string, scan *model.Scan) error {
	findingsJSON := marshalJSON(scan.Findings)
	_, err := s.db.Exec(
		`INSERT INTO scans (id, project_id, timestamp, subnet, duration_ms, preset, status, devices_found, findings_json)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		scan.ID, projectID, scan.Timestamp, scan.Subnet,
		scan.Duration.Milliseconds(), scan.Preset, scan.Status,
		scan.DevicesFound, findingsJSON,
	)
	return err
}

func (s *SQLiteStore) UpdateScan(scan *model.Scan) error {
	_, err := s.db.Exec(`UPDATE scans SET timestamp = ?, subnet = ?, duration_ms = ?, preset = ?, status = ?, devices_found = ?, findings_json = ? WHERE id = ?`,
		scan.Timestamp, scan.Subnet, scan.Duration.Milliseconds(), scan.Preset, scan.Status, scan.DevicesFound, marshalJSON(scan.Findings), scan.ID)
	return err
}

func (s *SQLiteStore) GetScan(id string) (*model.Scan, error) {
	row := s.db.QueryRow(
		"SELECT id, timestamp, subnet, duration_ms, preset, status, devices_found, findings_json FROM scans WHERE id = ?",
		id,
	)
	scan := &model.Scan{}
	var durationMs int64
	var findingsJSON string
	err := row.Scan(&scan.ID, &scan.Timestamp, &scan.Subnet, &durationMs,
		&scan.Preset, &scan.Status, &scan.DevicesFound, &findingsJSON)
	if err != nil {
		return nil, err
	}
	scan.Duration = time.Duration(durationMs) * time.Millisecond
	json.Unmarshal([]byte(findingsJSON), &scan.Findings)
	return scan, nil
}

func (s *SQLiteStore) ListScans(projectID string) ([]*model.Scan, error) {
	rows, err := s.db.Query(
		"SELECT id, timestamp, subnet, duration_ms, preset, status, devices_found, findings_json FROM scans WHERE project_id = ? ORDER BY timestamp DESC",
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scans []*model.Scan
	for rows.Next() {
		scan := &model.Scan{}
		var durationMs int64
		var findingsJSON string
		if err := rows.Scan(&scan.ID, &scan.Timestamp, &scan.Subnet, &durationMs,
			&scan.Preset, &scan.Status, &scan.DevicesFound, &findingsJSON); err != nil {
			return nil, err
		}
		scan.Duration = time.Duration(durationMs) * time.Millisecond
		json.Unmarshal([]byte(findingsJSON), &scan.Findings)
		scans = append(scans, scan)
	}
	return scans, nil
}

func (s *SQLiteStore) GetLatestScan(projectID string) (*model.Scan, error) {
	row := s.db.QueryRow(
		`SELECT id, timestamp, subnet, duration_ms, preset, status, devices_found, findings_json
		 FROM scans WHERE project_id = ? ORDER BY timestamp DESC LIMIT 1`,
		projectID,
	)
	scan := &model.Scan{}
	var durationMs int64
	var findingsJSON string
	err := row.Scan(&scan.ID, &scan.Timestamp, &scan.Subnet, &durationMs,
		&scan.Preset, &scan.Status, &scan.DevicesFound, &findingsJSON)
	if err != nil {
		return nil, err
	}
	scan.Duration = time.Duration(durationMs) * time.Millisecond
	json.Unmarshal([]byte(findingsJSON), &scan.Findings)
	return scan, nil
}

func (s *SQLiteStore) SaveDevice(projectID string, device *model.Device) error {
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO devices (id, project_id, mac, ips_json, vendor, hostname, os, role, model, vlans_json, first_seen, last_seen, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		device.ID, projectID, device.MAC, marshalJSON(device.IPs),
		device.Vendor, device.Hostname, device.OS, device.Role, device.Model,
		marshalJSON(device.VLANs), device.FirstSeen, device.LastSeen, device.Notes,
	)
	return err
}

func (s *SQLiteStore) SaveScanDevice(scanID, deviceID string) error {
	_, err := s.db.Exec("INSERT OR IGNORE INTO scan_devices (scan_id, device_id) VALUES (?, ?)", scanID, deviceID)
	return err
}

func (s *SQLiteStore) SaveDevices(projectID string, devices []*model.Device) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		`INSERT OR REPLACE INTO devices (id, project_id, mac, ips_json, vendor, hostname, os, role, model, vlans_json, first_seen, last_seen, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, device := range devices {
		_, err = stmt.Exec(
			device.ID, projectID, device.MAC, marshalJSON(device.IPs),
			device.Vendor, device.Hostname, device.OS, device.Role, device.Model,
			marshalJSON(device.VLANs), device.FirstSeen, device.LastSeen, device.Notes,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *SQLiteStore) GetDevice(id string) (*model.Device, error) {
	row := s.db.QueryRow(
		"SELECT id, mac, ips_json, vendor, hostname, os, role, model, vlans_json, first_seen, last_seen, notes FROM devices WHERE id = ?", id,
	)
	return scanDevice(row)
}

func (s *SQLiteStore) ListDevices(projectID string) ([]*model.Device, error) {
	rows, err := s.db.Query(
		"SELECT id, mac, ips_json, vendor, hostname, os, role, model, vlans_json, first_seen, last_seen, notes FROM devices WHERE project_id = ?",
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDevices(rows)
}

func (s *SQLiteStore) GetDevicesByScan(scanID string) ([]*model.Device, error) {
	rows, err := s.db.Query(
		`SELECT d.id, d.mac, d.ips_json, d.vendor, d.hostname, d.os, d.role, d.model, d.vlans_json, d.first_seen, d.last_seen, d.notes
		 FROM devices d
		 JOIN scan_devices sd ON d.id = sd.device_id
		 WHERE sd.scan_id = ?`, scanID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDevices(rows)
}

func scanDevice(row *sql.Row) (*model.Device, error) {
	d := &model.Device{}
	var ipsJSON, vlansJSON string
	err := row.Scan(&d.ID, &d.MAC, &ipsJSON, &d.Vendor, &d.Hostname, &d.OS, &d.Role, &d.Model, &vlansJSON, &d.FirstSeen, &d.LastSeen, &d.Notes)
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(ipsJSON), &d.IPs)
	json.Unmarshal([]byte(vlansJSON), &d.VLANs)
	return d, nil
}

func scanDevices(rows *sql.Rows) ([]*model.Device, error) {
	var devices []*model.Device
	for rows.Next() {
		d := &model.Device{}
		var ipsJSON, vlansJSON string
		if err := rows.Scan(&d.ID, &d.MAC, &ipsJSON, &d.Vendor, &d.Hostname, &d.OS, &d.Role, &d.Model, &vlansJSON, &d.FirstSeen, &d.LastSeen, &d.Notes); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(ipsJSON), &d.IPs)
		json.Unmarshal([]byte(vlansJSON), &d.VLANs)
		devices = append(devices, d)
	}
	return devices, nil
}

func (s *SQLiteStore) SavePorts(deviceID, scanID string, ports []*model.Port) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM ports WHERE device_id = ? AND scan_id = ?", deviceID, scanID)
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(
		"INSERT INTO ports (device_id, scan_id, number, protocol, service, version, banner, state) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, p := range ports {
		_, err = stmt.Exec(deviceID, scanID, p.Number, p.Protocol, p.Service, p.Version, p.Banner, p.State)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *SQLiteStore) GetPorts(deviceID string) ([]*model.Port, error) {
	rows, err := s.db.Query(
		"SELECT device_id, number, protocol, service, version, banner, state FROM ports WHERE device_id = ?", deviceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ports []*model.Port
	for rows.Next() {
		p := &model.Port{}
		if err := rows.Scan(&p.DeviceID, &p.Number, &p.Protocol, &p.Service, &p.Version, &p.Banner, &p.State); err != nil {
			return nil, err
		}
		ports = append(ports, p)
	}
	return ports, nil
}

func (s *SQLiteStore) SaveLinks(links []*model.Link) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM links")
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(
		"INSERT INTO links (source_id, target_id, type, src_port, dst_port, vlan) VALUES (?, ?, ?, ?, ?, ?)",
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, l := range links {
		_, err = stmt.Exec(l.SourceID, l.TargetID, l.Type, l.SrcPort, l.DstPort, l.VLAN)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *SQLiteStore) ListLinks() ([]*model.Link, error) {
	rows, err := s.db.Query("SELECT source_id, target_id, type, src_port, dst_port, vlan FROM links")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []*model.Link
	for rows.Next() {
		l := &model.Link{}
		if err := rows.Scan(&l.SourceID, &l.TargetID, &l.Type, &l.SrcPort, &l.DstPort, &l.VLAN); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	return links, nil
}

func (s *SQLiteStore) SaveFinding(scanID string, finding *model.Finding) error {
	_, err := s.db.Exec(
		`INSERT INTO findings (id, scan_id, device_id, type, severity, title, description, recommendation, timestamp)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		finding.ID, scanID, finding.DeviceID, finding.Type, finding.Severity,
		finding.Title, finding.Description, finding.Recommendation, finding.Timestamp,
	)
	return err
}

func (s *SQLiteStore) GetNewDevices(projectID, scanA, scanB string) ([]*model.Device, error) {
	rows, err := s.db.Query(
		`SELECT d.id, d.mac, d.ips_json, d.vendor, d.hostname, d.os, d.role, d.model, d.vlans_json, d.first_seen, d.last_seen, d.notes
		 FROM devices d
		 JOIN scan_devices sdB ON d.id = sdB.device_id AND sdB.scan_id = ?
		 WHERE d.project_id = ?
		 AND d.id NOT IN (SELECT device_id FROM scan_devices WHERE scan_id = ?)`,
		scanB, projectID, scanA,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDevices(rows)
}

func (s *SQLiteStore) GetMissingDevices(projectID, scanA, scanB string) ([]*model.Device, error) {
	rows, err := s.db.Query(
		`SELECT d.id, d.mac, d.ips_json, d.vendor, d.hostname, d.os, d.role, d.model, d.vlans_json, d.first_seen, d.last_seen, d.notes
		 FROM devices d
		 JOIN scan_devices sdA ON d.id = sdA.device_id AND sdA.scan_id = ?
		 WHERE d.project_id = ?
		 AND d.id NOT IN (SELECT device_id FROM scan_devices WHERE scan_id = ?)`,
		scanA, projectID, scanB,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDevices(rows)
}

func (s *SQLiteStore) GetChangedDevices(projectID, scanA, scanB string) ([]*model.Device, error) {
	rows, err := s.db.Query(
		`SELECT d2.id, d2.mac, d2.ips_json, d2.vendor, d2.hostname, d2.os, d2.role, d2.model, d2.vlans_json, d2.first_seen, d2.last_seen, d2.notes
		 FROM devices d1
		 JOIN scan_devices sd1 ON d1.id = sd1.device_id AND sd1.scan_id = ?
		 JOIN devices d2 ON d1.id = d2.id
		 WHERE d2.project_id = ?
		 AND (d1.mac != d2.mac OR d1.ips_json != d2.ips_json)`,
		scanA, projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDevices(rows)
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
