package wailsbridge

import (
	"context"
	"fmt"
	"netsight/internal/capture"
	"netsight/internal/export"
	"netsight/internal/model"
	"netsight/internal/scan"
	"netsight/internal/server"
	"netsight/internal/storage"
	"netsight/internal/tools"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Bridge struct {
	ctx             context.Context
	store           storage.Store
	scanner         *scan.ScanOrchestrator
	captureEngine   *capture.Engine
	serverManager   *server.Manager
	netInfo         []*model.InterfaceInfo
	activeProjectID string
	mu              sync.Mutex
	pingCancel      context.CancelFunc
	traceCancel     context.CancelFunc
	scanStates      map[string]*model.Scan
}

func NewBridge(store storage.Store) *Bridge {
	b := &Bridge{store: store, scanStates: make(map[string]*model.Scan)}
	b.serverManager = server.NewManager(func(state *model.ServerState) {
		b.EmitServerStatus(state)
	})
	return b
}

func (b *Bridge) SetContext(ctx context.Context) {
	b.ctx = ctx
	b.refreshNetworkInfo()
}

func (b *Bridge) refreshNetworkInfo() {
	info, err := tools.GetNetworkInfo()
	if err == nil {
		b.mu.Lock()
		b.netInfo = []*model.InterfaceInfo{info}
		b.mu.Unlock()
	}
}

func (b *Bridge) Greet(name string) string {
	return fmt.Sprintf("Hello %s! NetSight is running.", name)
}

// === Network Info ===
func (b *Bridge) GetNetworkInfo() (*model.InterfaceInfo, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.netInfo) > 0 {
		return b.netInfo[0], nil
	}
	return tools.GetNetworkInfo()
}

func (b *Bridge) GetAllNetworkInfo() ([]*model.InterfaceInfo, error) {
	return tools.GetAllInterfaces(), nil
}

func (b *Bridge) GetAvailableSubnets() ([]string, error) {
	return tools.GetAvailableSubnets()
}

// === Scan ===
type ScanInput struct {
	Subnet    string `json:"subnet"`
	Preset    string `json:"preset"`
	ProjectID string `json:"projectId"`
}

func (b *Bridge) StartScan(input ScanInput) (string, error) {
	projectID, err := b.ensureProject(input.ProjectID)
	if err != nil {
		return "", err
	}
	b.mu.Lock()
	b.activeProjectID = projectID
	b.mu.Unlock()
	b.scanner = scan.NewScanOrchestrator()
	b.scanner.OnProgress = func(scanID string, progress int) {
		runtime.EventsEmit(b.ctx, "scan:progress", scanID, progress)
	}
	b.scanner.OnDeviceFound = func(scanID string, device *model.Device) {
		runtime.EventsEmit(b.ctx, "scan:device-found", device)
		if err := b.store.SaveDevice(projectID, device); err != nil {
			runtime.EventsEmit(b.ctx, "scan:error", err.Error())
			return
		}
		if err := b.store.SaveScanDevice(scanID, device.ID); err != nil {
			runtime.EventsEmit(b.ctx, "scan:error", err.Error())
		}
	}
	b.scanner.OnStarted = func(sc *model.Scan) {
		b.mu.Lock()
		b.scanStates[sc.ID] = sc
		b.mu.Unlock()
		if err := b.store.SaveScan(projectID, sc); err != nil {
			runtime.EventsEmit(b.ctx, "scan:error", sc.ID, err.Error())
		}
	}
	b.scanner.OnPortsFound = func(scanID, deviceID string, ports []*model.Port) {
		if err := b.store.SavePorts(deviceID, scanID, ports); err != nil {
			runtime.EventsEmit(b.ctx, "scan:error", err.Error())
		}
	}
	b.scanner.OnError = func(scanID string, scanErr error) { runtime.EventsEmit(b.ctx, "scan:error", scanID, scanErr.Error()) }
	b.scanner.OnComplete = func(scanID string, status string) {
		b.mu.Lock()
		if sc := b.scanStates[scanID]; sc != nil {
			sc.Status = status
			sc.Duration = time.Since(sc.Timestamp)
			sc.DevicesFound = len(mustDevices(b.store, scanID))
			if err := b.store.UpdateScan(sc); err != nil {
				runtime.EventsEmit(b.ctx, "scan:error", scanID, err.Error())
			}
			delete(b.scanStates, scanID)
		}
		b.mu.Unlock()
		runtime.EventsEmit(b.ctx, "scan:complete", scanID, status)
	}
	return b.scanner.StartScan(input.Subnet, input.Preset)
}

func (b *Bridge) StopScan(scanID string) error {
	if b.scanner != nil {
		return b.scanner.StopScan(scanID)
	}
	return nil
}

func (b *Bridge) GetDevices() ([]*model.Device, error) {
	b.mu.Lock()
	projectID := b.activeProjectID
	b.mu.Unlock()
	return b.store.ListDevices(projectID)
}

func (b *Bridge) GetScanHistory() ([]*model.Scan, error) {
	b.mu.Lock()
	projectID := b.activeProjectID
	b.mu.Unlock()
	return b.store.ListScans(projectID)
}

// === Packet Capture ===
func (b *Bridge) StartPacketCapture(iface string, filter string) (string, error) {
	deviceName, err := capture.ResolveInterfaceName(iface)
	if err != nil {
		return "", err
	}
	b.captureEngine = capture.NewEngine()
	err = b.captureEngine.Start(deviceName, filter,
		func(packet *model.PacketSummary) {
			runtime.EventsEmit(b.ctx, "capture:packet", packet)
		},
		func(pps, bps int64) {
			runtime.EventsEmit(b.ctx, "capture:stats", pps, bps)
		},
	)
	if err != nil {
		return "", err
	}
	return "capture-1", nil
}

func (b *Bridge) StopPacketCapture() error {
	if b.captureEngine != nil {
		b.captureEngine.Stop()
	}
	return nil
}

// === Tools ===
type PingInput struct {
	Target string `json:"target"`
	Count  int    `json:"count"`
}

func (b *Bridge) RunPing(input PingInput) (*model.PingResult, error) {
	ctx, cancel := context.WithCancel(b.ctx)
	b.mu.Lock()
	if b.pingCancel != nil {
		b.pingCancel()
	}
	b.pingCancel = cancel
	b.mu.Unlock()

	if input.Count == 0 {
		input.Count = 4
	}

	var last *model.PingResult
	err := tools.Ping(ctx, input.Target, input.Count, 56, 64, func(result *model.PingResult) {
		last = result
		runtime.EventsEmit(b.ctx, "tool:ping-result", result)
	})
	return last, err
}

type TracerouteInput struct {
	Target string `json:"target"`
	Mode   string `json:"mode"`
}

func (b *Bridge) RunTraceroute(input TracerouteInput) ([]*model.Hop, error) {
	ctx, cancel := context.WithCancel(b.ctx)
	b.mu.Lock()
	if b.traceCancel != nil {
		b.traceCancel()
	}
	b.traceCancel = cancel
	b.mu.Unlock()

	var hops []*model.Hop
	err := tools.Traceroute(ctx, input.Target, input.Mode, 30, func(hop *model.Hop) {
		hops = append(hops, hop)
		runtime.EventsEmit(b.ctx, "tool:traceroute-hop", hop)
	})
	return hops, err
}

func (b *Bridge) RunNSLookup(query string, types []string) ([]*tools.DNSResult, error) {
	return tools.NSLookup(context.Background(), query, types)
}

func (b *Bridge) WakeOnLAN(mac string) error {
	return tools.WakeOnLAN(mac)
}

type IPerfInput struct {
	Target     string `json:"target"`
	ServerMode bool   `json:"serverMode"`
	Duration   int    `json:"duration"`
}

func (b *Bridge) RunIPerf(input IPerfInput) (*model.IPerfResult, error) {
	if input.Target == "" {
		input.Target = "iperf.he.net"
	}
	if input.Duration == 0 {
		input.Duration = 10
	}
	// On Windows, iperf3.exe is embedded at the same level as the app binary.
	// Also try the embedded/ directory relative to the working directory.
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	iperfPath := filepath.Join(filepath.Dir(exe), "embedded", "iperf3.exe")
	return nil, tools.RunIPerf(context.Background(), input.Target, input.ServerMode, input.Duration, iperfPath, func(result *model.IPerfResult) {
		runtime.EventsEmit(b.ctx, "iperf:result", result)
	})
}

// === Servers ===
func (b *Bridge) StartServer(serverType string, config map[string]interface{}) error {
	cfg := &model.ServerConfig{}
	if v, ok := config["port"]; ok {
		cfg.Port = int(toFloat64(v))
	}
	if v, ok := config["interface"]; ok {
		cfg.Interface = toString(v)
	}
	if v, ok := config["rootDir"]; ok {
		cfg.RootDir = toString(v)
	}
	if v, ok := config["readOnly"]; ok {
		if value, ok := v.(bool); ok {
			cfg.ReadOnly = value
		} else {
			return fmt.Errorf("readOnly must be a boolean")
		}
	}
	if v, ok := config["poolStart"]; ok {
		cfg.PoolStart = toString(v)
	}
	if v, ok := config["poolEnd"]; ok {
		cfg.PoolEnd = toString(v)
	}
	if v, ok := config["gateway"]; ok {
		cfg.Gateway = toString(v)
	}
	if v, ok := config["dns"]; ok {
		cfg.DNS = toString(v)
	}

	err := b.serverManager.StartServer(serverType, cfg)
	if err != nil {
		return err
	}

	return nil
}

func (b *Bridge) Shutdown() {
	if b.captureEngine != nil {
		b.captureEngine.Stop()
	}
	if b.scanner != nil {
		_ = b.scanner.StopScan("")
	}
	b.serverManager.Shutdown()
	b.mu.Lock()
	if b.pingCancel != nil {
		b.pingCancel()
	}
	if b.traceCancel != nil {
		b.traceCancel()
	}
	b.mu.Unlock()
}

func (b *Bridge) ensureProject(id string) (string, error) {
	if id != "" {
		if _, err := b.store.GetProject(id); err != nil {
			return "", err
		}
		return id, nil
	}
	projects, err := b.store.ListProjects()
	if err != nil {
		return "", err
	}
	if len(projects) > 0 {
		return projects[0].ID, nil
	}
	p := &model.Project{ID: "default", Name: "Default", Created: time.Now()}
	return p.ID, b.store.CreateProject(p)
}
func mustDevices(store storage.Store, scanID string) []*model.Device {
	devices, _ := store.GetDevicesByScan(scanID)
	return devices
}

func (b *Bridge) StopServer(serverType string) error {
	return b.serverManager.StopServer(serverType)
}

// === Export ===
func (b *Bridge) ExportPDF(scanID string) (string, error) {
	devices, err := b.store.GetDevicesByScan(scanID)
	if err != nil {
		return "", err
	}
	sc, err := b.store.GetScan(scanID)
	if err != nil {
		return "", err
	}
	findings := make([]*model.Finding, 0, len(sc.Findings))
	for i := range sc.Findings {
		findings = append(findings, &sc.Findings[i])
	}
	return export.GeneratePDF(scanID, "", devices, findings, "")
}

func (b *Bridge) ExportDrawIO(scanID string) (string, error) {
	devices, err := b.store.GetDevicesByScan(scanID)
	if err != nil {
		return "", err
	}
	links, err := b.store.ListLinks()
	if err != nil {
		return "", err
	}
	return export.ExportDrawIO(devices, links)
}

// === Projects ===
func (b *Bridge) CreateProject(name string) (*model.Project, error) {
	project := &model.Project{
		ID:      fmt.Sprintf("proj-%d", time.Now().UnixNano()),
		Name:    name,
		Created: time.Now(),
	}
	err := b.store.CreateProject(project)
	if err != nil {
		return nil, err
	}
	b.mu.Lock()
	b.activeProjectID = project.ID
	b.mu.Unlock()
	return project, nil
}

func (b *Bridge) LoadProject(id string) (*model.Project, error) {
	project, err := b.store.GetProject(id)
	if err != nil {
		return nil, err
	}
	b.mu.Lock()
	b.activeProjectID = id
	b.mu.Unlock()
	project.Devices, err = b.store.ListDevices(id)
	if err != nil {
		return nil, err
	}
	project.Scans, err = b.store.ListScans(id)
	if err != nil {
		return nil, err
	}
	for _, scan := range project.Scans {
		project.Findings = append(project.Findings, scan.Findings...)
	}
	return project, nil
}

func (b *Bridge) ListProjects() ([]*model.Project, error) {
	return b.store.ListProjects()
}

// === Event Emitters ===
func (b *Bridge) EmitScanProgress(scanID string, progress int) {
	runtime.EventsEmit(b.ctx, "scan:progress", scanID, progress)
}
func (b *Bridge) EmitDeviceFound(device *model.Device) {
	runtime.EventsEmit(b.ctx, "scan:device-found", device)
}
func (b *Bridge) EmitScanComplete(scanID string) {
	runtime.EventsEmit(b.ctx, "scan:complete", scanID)
}
func (b *Bridge) EmitPacket(packet *model.PacketSummary) {
	runtime.EventsEmit(b.ctx, "capture:packet", packet)
}
func (b *Bridge) EmitCaptureStats(pps, bps int64) {
	runtime.EventsEmit(b.ctx, "capture:stats", pps, bps)
}
func (b *Bridge) EmitServerStatus(state *model.ServerState) {
	runtime.EventsEmit(b.ctx, "server:status", state)
}
func (b *Bridge) EmitPingResult(result *model.PingResult) {
	runtime.EventsEmit(b.ctx, "tool:ping-result", result)
}
func (b *Bridge) EmitTracerouteHop(hop *model.Hop) {
	runtime.EventsEmit(b.ctx, "tool:traceroute-hop", hop)
}
func (b *Bridge) EmitFinding(finding *model.Finding) {
	runtime.EventsEmit(b.ctx, "survey:finding", finding)
}
func (b *Bridge) EmitIPerfResult(result *model.IPerfResult) {
	runtime.EventsEmit(b.ctx, "iperf:result", result)
}

func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case float32:
		return float64(val)
	}
	return 0
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
