package storage

import (
	"fmt"
	"netsight/internal/model"
	"sync"
)

type MemoryStore struct {
	mu       sync.RWMutex
	projects map[string]*model.Project
	scans    map[string]*model.Scan
	devices  map[string]*model.Device

	projectScans map[string][]string
	scanDevices  map[string][]string
	ports        map[string][]*model.Port
	links        []*model.Link
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		projects:     make(map[string]*model.Project),
		scans:        make(map[string]*model.Scan),
		devices:      make(map[string]*model.Device),
		projectScans: make(map[string][]string),
		scanDevices:  make(map[string][]string),
		ports:        make(map[string][]*model.Port),
		links:        make([]*model.Link, 0),
	}
}

func (s *MemoryStore) CreateProject(project *model.Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.projects[project.ID] = project
	s.projectScans[project.ID] = []string{}
	return nil
}

func (s *MemoryStore) GetProject(id string) (*model.Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.projects[id]
	if !ok {
		return nil, fmt.Errorf("project not found: %s", id)
	}
	return p, nil
}

func (s *MemoryStore) ListProjects() ([]*model.Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*model.Project, 0, len(s.projects))
	for _, p := range s.projects {
		result = append(result, p)
	}
	return result, nil
}

func (s *MemoryStore) DeleteProject(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.projects, id)
	delete(s.projectScans, id)
	return nil
}

func (s *MemoryStore) SaveScan(projectID string, scan *model.Scan) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scans[scan.ID] = scan
	s.projectScans[projectID] = append(s.projectScans[projectID], scan.ID)
	s.scanDevices[scan.ID] = []string{}
	return nil
}

func (s *MemoryStore) UpdateScan(scan *model.Scan) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.scans[scan.ID]; !ok {
		return fmt.Errorf("scan not found: %s", scan.ID)
	}
	s.scans[scan.ID] = scan
	return nil
}

func (s *MemoryStore) GetScan(id string) (*model.Scan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sc, ok := s.scans[id]
	if !ok {
		return nil, fmt.Errorf("scan not found: %s", id)
	}
	return sc, nil
}

func (s *MemoryStore) ListScans(projectID string) ([]*model.Scan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	scanIDs := s.projectScans[projectID]
	result := make([]*model.Scan, 0, len(scanIDs))
	for _, sid := range scanIDs {
		if sc, ok := s.scans[sid]; ok {
			result = append(result, sc)
		}
	}
	return result, nil
}

func (s *MemoryStore) GetLatestScan(projectID string) (*model.Scan, error) {
	scans, err := s.ListScans(projectID)
	if err != nil {
		return nil, err
	}
	if len(scans) == 0 {
		return nil, fmt.Errorf("no scans for project: %s", projectID)
	}
	latest := scans[0]
	for _, sc := range scans[1:] {
		if sc.Timestamp.After(latest.Timestamp) {
			latest = sc
		}
	}
	return latest, nil
}

func (s *MemoryStore) SaveDevice(projectID string, device *model.Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.devices[device.ID] = device
	return nil
}

func (s *MemoryStore) SaveScanDevice(scanID, deviceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, id := range s.scanDevices[scanID] {
		if id == deviceID {
			return nil
		}
	}
	s.scanDevices[scanID] = append(s.scanDevices[scanID], deviceID)
	return nil
}

func (s *MemoryStore) SaveDevices(projectID string, devices []*model.Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, d := range devices {
		s.devices[d.ID] = d
	}
	return nil
}

func (s *MemoryStore) GetDevice(id string) (*model.Device, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.devices[id]
	if !ok {
		return nil, fmt.Errorf("device not found: %s", id)
	}
	return d, nil
}

func (s *MemoryStore) ListDevices(projectID string) ([]*model.Device, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*model.Device, 0, len(s.devices))
	for _, d := range s.devices {
		result = append(result, d)
	}
	return result, nil
}

func (s *MemoryStore) GetDevicesByScan(scanID string) ([]*model.Device, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	deviceIDs := s.scanDevices[scanID]
	result := make([]*model.Device, 0, len(deviceIDs))
	for _, did := range deviceIDs {
		if d, ok := s.devices[did]; ok {
			result = append(result, d)
		}
	}
	return result, nil
}

func (s *MemoryStore) SavePorts(deviceID, scanID string, ports []*model.Port) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ports[deviceID] = ports
	return nil
}

func (s *MemoryStore) GetPorts(deviceID string) ([]*model.Port, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ports[deviceID], nil
}

func (s *MemoryStore) SaveLinks(links []*model.Link) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.links = append(s.links, links...)
	return nil
}

func (s *MemoryStore) ListLinks() ([]*model.Link, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*model.Link, len(s.links))
	copy(result, s.links)
	return result, nil
}

func (s *MemoryStore) SaveFinding(scanID string, finding *model.Finding) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	sc, ok := s.scans[scanID]
	if !ok {
		return fmt.Errorf("scan not found: %s", scanID)
	}
	sc.Findings = append(sc.Findings, *finding)
	return nil
}

func (s *MemoryStore) GetNewDevices(projectID, scanA, scanB string) ([]*model.Device, error) {
	devicesA, _ := s.GetDevicesByScan(scanA)
	devicesB, _ := s.GetDevicesByScan(scanB)
	seen := make(map[string]bool)
	for _, d := range devicesA {
		seen[d.ID] = true
	}
	var result []*model.Device
	for _, d := range devicesB {
		if !seen[d.ID] {
			result = append(result, d)
		}
	}
	return result, nil
}

func (s *MemoryStore) GetMissingDevices(projectID, scanA, scanB string) ([]*model.Device, error) {
	devicesA, _ := s.GetDevicesByScan(scanA)
	devicesB, _ := s.GetDevicesByScan(scanB)
	seen := make(map[string]bool)
	for _, d := range devicesB {
		seen[d.ID] = true
	}
	var result []*model.Device
	for _, d := range devicesA {
		if !seen[d.ID] {
			result = append(result, d)
		}
	}
	return result, nil
}

func (s *MemoryStore) GetChangedDevices(projectID, scanA, scanB string) ([]*model.Device, error) {
	devicesA, _ := s.GetDevicesByScan(scanA)
	devicesB, _ := s.GetDevicesByScan(scanB)
	devMapA := make(map[string]*model.Device)
	for _, d := range devicesA {
		devMapA[d.ID] = d
	}
	var result []*model.Device
	for _, d := range devicesB {
		if prev, ok := devMapA[d.ID]; ok {
			changed := prev.MAC != d.MAC
			if !changed && len(prev.IPs) == len(d.IPs) {
				for i, ip := range prev.IPs {
					if ip != d.IPs[i] {
						changed = true
						break
					}
				}
			} else {
				changed = true
			}
			if changed {
				result = append(result, d)
			}
		}
	}
	return result, nil
}

func (s *MemoryStore) Close() error {
	return nil
}
