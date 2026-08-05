package wailsbridge

import (
	"context"
	"fmt"
	"net"
	"netsight/internal/capture"
	"netsight/internal/export"
	"netsight/internal/model"
	"netsight/internal/scan"
	"netsight/internal/server"
	"netsight/internal/storage"
	"netsight/internal/tools"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Bridge struct {
	ctx           context.Context
	store         storage.Store
	scanner       *scan.ScanOrchestrator
	captureEngine *capture.Engine
	serverManager *server.Manager
	netInfo       []*model.InterfaceInfo
	mu            sync.Mutex
	pingCancel    context.CancelFunc
	traceCancel   context.CancelFunc
}

func NewBridge(store storage.Store) *Bridge {
	b := &Bridge{store: store}
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
	Subnet string `json:"subnet"`
	Preset string `json:"preset"`
}

func (b *Bridge) StartScan(input ScanInput) (string, error) {
	b.scanner = scan.NewScanOrchestrator()
	b.scanner.OnProgress = func(scanID string, progress int) {
		runtime.EventsEmit(b.ctx, "scan:progress", scanID, progress)
	}
	b.scanner.OnDeviceFound = func(device *model.Device) {
		runtime.EventsEmit(b.ctx, "scan:device-found", device)
		b.store.SaveDevice("", device)
	}
	b.scanner.OnComplete = func(scanID string, status string) {
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
	return b.store.ListDevices("")
}

func (b *Bridge) GetScanHistory() ([]*model.Scan, error) {
	return b.store.ListScans("")
}

// === Packet Capture ===
func (b *Bridge) StartPacketCapture(iface string, filter string) (string, error) {
	b.captureEngine = capture.NewEngine()
	err := b.captureEngine.Start(iface, filter,
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

	err := tools.Ping(ctx, input.Target, input.Count, 56, 64, func(result *model.PingResult) {
		runtime.EventsEmit(b.ctx, "tool:ping-result", result)
	})
	return nil, err
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
	iperfPath := "embedded/iperf3.exe"
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
		cfg.ReadOnly = v.(bool)
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

	time.AfterFunc(1*time.Second, func() {
		addr := fmt.Sprintf("127.0.0.1:%d", cfg.Port)
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err != nil {
			runtime.EventsEmit(b.ctx, "server:status", &model.ServerState{
				Type:   serverType,
				Port:   cfg.Port,
				Status: "error",
				Error:  fmt.Sprintf("Self-test failed: %v", err),
			})
			return
		}
		conn.Close()
	})
	return nil
}

func (b *Bridge) StopServer(serverType string) error {
	return b.serverManager.StopServer(serverType)
}

// === Export ===
func (b *Bridge) ExportPDF(scanID string) (string, error) {
	devices, _ := b.store.ListDevices("")
	var findings []*model.Finding
	return export.GeneratePDF(scanID, "", devices, findings, "")

func (b *Bridge) ExportDrawIO(scanID string) (string, error) {
	devices, _ := b.store.ListDevices("")
	links, _ := b.store.ListLinks()
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
	return project, nil
}

func (b *Bridge) LoadProject(id string) (*model.Project, error) {
	return b.store.GetProject(id)
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
