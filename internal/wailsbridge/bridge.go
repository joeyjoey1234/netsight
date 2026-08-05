package wailsbridge

import (
	"context"
	"fmt"
	"netsight/internal/model"
	"netsight/internal/storage"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Bridge struct {
	ctx   context.Context
	store storage.Store
}

func NewBridge(store storage.Store) *Bridge {
	return &Bridge{store: store}
}

func (b *Bridge) SetContext(ctx context.Context) {
	b.ctx = ctx
}

func (b *Bridge) Greet(name string) string {
	return fmt.Sprintf("Hello %s! NetSight is running.", name)
}

func (b *Bridge) GetNetworkInfo() (*model.InterfaceInfo, error) {
	return &model.InterfaceInfo{
		Name:   "placeholder",
		IPs:    []string{},
		MAC:    "",
		MTU:    1500,
	}, nil
}

type ScanInput struct {
	Subnet string `json:"subnet"`
	Preset string `json:"preset"`
}

func (b *Bridge) StartScan(input ScanInput) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (b *Bridge) StopScan(scanID string) error {
	return fmt.Errorf("not implemented")
}

func (b *Bridge) GetDevices() ([]*model.Device, error) {
	return b.store.ListDevices("")
}

func (b *Bridge) GetScanHistory() ([]*model.Scan, error) {
	return b.store.ListScans("")
}

func (b *Bridge) StartPacketCapture(iface string, filter string) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (b *Bridge) StopPacketCapture() error {
	return fmt.Errorf("not implemented")
}

type PingInput struct {
	Target string `json:"target"`
	Count  int    `json:"count"`
}

func (b *Bridge) RunPing(input PingInput) (*model.PingResult, error) {
	return nil, fmt.Errorf("not implemented")
}

type TracerouteInput struct {
	Target string `json:"target"`
	Mode   string `json:"mode"`
}

func (b *Bridge) RunTraceroute(input TracerouteInput) ([]*model.Hop, error) {
	return nil, fmt.Errorf("not implemented")
}

func (b *Bridge) StartServer(serverType string, config map[string]interface{}) error {
	return fmt.Errorf("not implemented")
}

func (b *Bridge) StopServer(serverType string) error {
	return fmt.Errorf("not implemented")
}

func (b *Bridge) ExportPDF(scanID string) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (b *Bridge) ExportDrawIO(scanID string) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (b *Bridge) CreateProject(name string) (*model.Project, error) {
	return nil, fmt.Errorf("not implemented")
}

func (b *Bridge) LoadProject(id string) (*model.Project, error) {
	return b.store.GetProject(id)
}

func (b *Bridge) WakeOnLAN(mac string) error {
	return fmt.Errorf("not implemented")
}

type IPerfInput struct {
	Target     string `json:"target"`
	ServerMode bool   `json:"serverMode"`
	Duration   int    `json:"duration"`
}

func (b *Bridge) RunIPerf(input IPerfInput) (*model.IPerfResult, error) {
	return nil, fmt.Errorf("not implemented")
}

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

func (b *Bridge) EmitCaptureStats(packetsPerSec, bytesPerSec int64) {
	runtime.EventsEmit(b.ctx, "capture:stats", packetsPerSec, bytesPerSec)
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
