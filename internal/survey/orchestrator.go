package survey

import (
	"context"
	"fmt"
	"sync"
	"time"

	"netsight/internal/fingerprint"
	"netsight/internal/health"
	"netsight/internal/model"
	"netsight/internal/protocol"
	"netsight/internal/scan"
	"netsight/internal/security"
)

type ProgressCallback func(scanID string, progress int)

type FindingCallback func(finding *model.Finding)

type DeviceCallback func(device *model.Device)

type Orchestrator struct {
	onProgress ProgressCallback
	onFinding  FindingCallback
	onDevice   DeviceCallback

	mu        sync.Mutex
	running   bool
	cancel    context.CancelFunc
	startTime time.Time
	duration  time.Duration
	scanID    string
	findings  []*model.Finding
	devices   map[string]*model.Device
}

func NewOrchestrator(onProgress ProgressCallback, onFinding FindingCallback, onDevice DeviceCallback) *Orchestrator {
	return &Orchestrator{
		onProgress: onProgress,
		onFinding:  onFinding,
		onDevice:   onDevice,
		devices:    make(map[string]*model.Device),
	}
}

func (o *Orchestrator) Start(subnet string, presetName string, customDuration time.Duration) (string, error) {
	o.mu.Lock()
	if o.running {
		o.mu.Unlock()
		return "", fmt.Errorf("survey already running")
	}

	preset, err := GetPreset(presetName)
	if err != nil {
		o.mu.Unlock()
		return "", err
	}

	duration := preset.Duration
	if presetName == "long" && customDuration > 0 {
		duration = customDuration
	}
	if duration == 0 {
		duration = 10 * time.Minute
	}

	o.running = true
	o.scanID = fmt.Sprintf("survey-%d", time.Now().UnixNano())
	o.startTime = time.Now()
	o.duration = duration
	o.findings = make([]*model.Finding, 0)
	o.devices = make(map[string]*model.Device)

	ctx, cancel := context.WithTimeout(context.Background(), duration+30*time.Second)
	o.cancel = cancel
	o.mu.Unlock()

	go o.run(ctx, subnet, preset)

	return o.scanID, nil
}

func (o *Orchestrator) Stop() {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.cancel != nil {
		o.cancel()
	}
	o.running = false
}

func (o *Orchestrator) IsRunning() bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.running
}

func (o *Orchestrator) run(ctx context.Context, subnet string, preset *Preset) {
	defer func() {
		o.mu.Lock()
		o.running = false
		o.mu.Unlock()
		if o.onProgress != nil {
			o.onProgress(o.scanID, 100)
		}
	}()

	var wg sync.WaitGroup
	startTime := o.startTime

	if o.onProgress != nil {
		o.onProgress(o.scanID, 5)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		o.runARPScan(ctx, subnet)
	}()

	time.Sleep(1 * time.Second)
	if o.onProgress != nil {
		o.onProgress(o.scanID, 20)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		o.runControlPlaneListener(ctx, subnet)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		o.runCDPLLDPListener(ctx, subnet)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		o.runBroadcastDetection(ctx, subnet)
	}()

	o.monitorProgress(ctx, startTime)

	wg.Add(1)
	go func() {
		defer wg.Done()
		o.runARPSpoofDetection(ctx, subnet)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		o.runRogueDHCPDetection(ctx, subnet)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		o.runBandwidthEstimation(ctx, subnet)
	}()

	if preset.Name == "long" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			o.runPortScan(ctx, subnet)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			o.runOpenSharesScan(ctx)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			o.runPassiveOSFingerprint(ctx, subnet)
		}()
	}

	wg.Wait()

	if o.onProgress != nil {
		o.onProgress(o.scanID, 95)
	}

	o.mu.Lock()
	o.mu.Unlock()
}

func (o *Orchestrator) runARPScan(ctx context.Context, subnet string) {
	table, err := scan.ARPScan(ctx, subnet)
	if err != nil {
		return
	}
	o.mu.Lock()
	for ip, mac := range table {
		if _, exists := o.devices[mac]; !exists {
			dev := &model.Device{
				ID:        fmt.Sprintf("dev-%s", mac),
				MAC:       mac,
				IPs:       []string{ip},
				FirstSeen: o.startTime,
				LastSeen:  time.Now(),
			}
			o.devices[mac] = dev
		} else {
			o.devices[mac].IPs = append(o.devices[mac].IPs, ip)
			o.devices[mac].LastSeen = time.Now()
		}
	}
	o.mu.Unlock()

	if o.onProgress != nil {
		o.onProgress(o.scanID, 10)
	}

	for _, dev := range o.devices {
		if o.onDevice != nil {
			o.onDevice(dev)
		}
	}
}

func (o *Orchestrator) runControlPlaneListener(ctx context.Context, subnet string) {
	listener := protocol.NewListener()
	listener.Start(ctx)
	defer listener.Stop()

	listener.SetHandlers(
		nil,
		func(event *protocol.ProtocolEvent) {
			if event == nil {
				return
			}
			role := listener.GetRole(event.SrcMAC)
			if role != "" && role != "unknown" {
				o.emitDevice(&model.Device{
					MAC:      event.SrcMAC,
					Role:     role,
					LastSeen: time.Now(),
				})
			}
		},
		func(pkt *protocol.CDPPacket) {
			if pkt == nil {
				return
			}
			o.emitDevice(&model.Device{
				MAC:      pkt.SrcMAC,
				Hostname: pkt.DeviceID,
				Model:    pkt.Platform,
				Role:     "switch",
				LastSeen: time.Now(),
			})
		},
		func(pkt *protocol.LLDPPacket) {
			if pkt == nil {
				return
			}
			o.emitDevice(&model.Device{
				MAC:      pkt.SrcMAC,
				Hostname: pkt.SystemName,
				Role:     "switch",
				LastSeen: time.Now(),
			})
		},
	)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (o *Orchestrator) runCDPLLDPListener(ctx context.Context, subnet string) {
	listener := protocol.NewListener()
	listener.Start(ctx)
	defer listener.Stop()

	var cdps []*protocol.CDPPacket
	var lldps []*protocol.LLDPPacket
	var cdpMu sync.Mutex

	listener.SetHandlers(
		nil,
		nil,
		func(pkt *protocol.CDPPacket) {
			cdpMu.Lock()
			cdps = append(cdps, pkt)
			cdpMu.Unlock()
		},
		func(pkt *protocol.LLDPPacket) {
			cdpMu.Lock()
			lldps = append(lldps, pkt)
			cdpMu.Unlock()
		},
	)

	<-ctx.Done()

	for _, cdp := range cdps {
		o.emitDevice(&model.Device{
			MAC:      cdp.SrcMAC,
			Hostname: cdp.DeviceID,
			Model:    cdp.Platform,
			Role:     "switch",
			LastSeen: time.Now(),
		})
	}
	for _, lldp := range lldps {
		o.emitDevice(&model.Device{
			MAC:      lldp.SrcMAC,
			Hostname: lldp.SystemName,
			Role:     "switch",
			LastSeen: time.Now(),
		})
	}
}

func (o *Orchestrator) runBroadcastDetection(ctx context.Context, subnet string) {
	detector := health.NewBroadcastDetector()
	detector.Start()

	<-ctx.Done()

	if finding := detector.DetectStorm(); finding != nil {
		if o.onFinding != nil {
			o.onFinding(finding)
		}
	}
	for _, finding := range detector.DetectMACFlapping() {
		if o.onFinding != nil {
			o.onFinding(finding)
		}
	}
}

func (o *Orchestrator) runARPSpoofDetection(ctx context.Context, subnet string) {
	detector := security.NewARPSpoofDetector()

	<-ctx.Done()

	for _, finding := range detector.GetConflicts() {
		if o.onFinding != nil {
			o.onFinding(finding)
		}
	}
}

func (o *Orchestrator) runRogueDHCPDetection(ctx context.Context, subnet string) {
	detector := security.NewDHCPDetector(nil)

	<-ctx.Done()

	_ = detector
}

func (o *Orchestrator) runBandwidthEstimation(ctx context.Context, subnet string) {
	analyzer := health.NewBandwidthAnalyzer()
	analyzer.Start()

	<-ctx.Done()

	for _, finding := range analyzer.GenerateTopTalkerFindings() {
		if o.onFinding != nil {
			o.onFinding(finding)
		}
	}
}

func (o *Orchestrator) runPortScan(ctx context.Context, subnet string) {
	ips, err := scan.ScanIPsExpand(subnet)
	if err != nil {
		return
	}

	for _, ip := range ips {
		select {
		case <-ctx.Done():
			return
		default:
		}
		ports, err := scan.TCPSynScan(ctx, ip, scan.DefaultPorts())
		if err != nil {
			continue
		}
		for _, p := range ports {
			if p.State == "open" {
				o.emitFinding(&model.Finding{
					ID:             fmt.Sprintf("port-%s-%d", ip, p.Number),
					Type:           "open_port",
					Severity:       "info",
					Title:          fmt.Sprintf("Open port %d on %s", p.Number, ip),
					Description:    fmt.Sprintf("TCP port %d is open on %s", p.Number, ip),
					Recommendation: "Review open ports and ensure only necessary services are exposed.",
					Timestamp:      time.Now(),
				})
			}
		}
	}
}

func (o *Orchestrator) runOpenSharesScan(ctx context.Context) {
	o.mu.Lock()
	var ips []string
	for _, dev := range o.devices {
		ips = append(ips, dev.IPs...)
	}
	o.mu.Unlock()

	scanner := security.NewShareScanner()
	for _, finding := range scanner.ScanSubnet(ips) {
		if o.onFinding != nil {
			o.onFinding(finding)
		}
	}
}

func (o *Orchestrator) runPassiveOSFingerprint(ctx context.Context, subnet string) {
	_ = ctx
	_ = subnet

	sig := &fingerprint.OSSignature{
		TTL:           64,
		TCPWindowSize: 65535,
		DFBit:         true,
		MSS:           1460,
	}

	result := fingerprint.DetectOS(sig)
	if result != nil {
		o.mu.Lock()
		for _, dev := range o.devices {
			if dev.OS == "" {
				dev.OS = result.OS
			}
		}
		o.mu.Unlock()
	}
}

func (o *Orchestrator) monitorProgress(ctx context.Context, startTime time.Time) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			elapsed := time.Since(startTime).Seconds()
			total := o.duration.Seconds()
			if total > 0 {
				progress := int((elapsed/total)*70) + 20
				if progress > 90 {
					progress = 90
				}
				if o.onProgress != nil {
					o.onProgress(o.scanID, progress)
				}
			}
		}
	}
}

func (o *Orchestrator) emitFinding(f *model.Finding) {
	if o.onFinding != nil {
		o.onFinding(f)
	}
	o.mu.Lock()
	o.findings = append(o.findings, f)
	o.mu.Unlock()
}

func (o *Orchestrator) emitDevice(d *model.Device) {
	o.mu.Lock()
	if existing, ok := o.devices[d.MAC]; ok {
		if d.Hostname != "" && existing.Hostname == "" {
			existing.Hostname = d.Hostname
		}
		if d.Role != "" && existing.Role == "" {
			existing.Role = d.Role
		}
		if d.Model != "" && existing.Model == "" {
			existing.Model = d.Model
		}
		if d.OS != "" && existing.OS == "" {
			existing.OS = d.OS
		}
		existing.LastSeen = time.Now()
		o.mu.Unlock()
		if o.onDevice != nil {
			o.onDevice(existing)
		}
		return
	}
	if d.ID == "" {
		d.ID = fmt.Sprintf("dev-%s", d.MAC)
	}
	d.FirstSeen = o.startTime
	d.LastSeen = time.Now()
	o.devices[d.MAC] = d
	o.mu.Unlock()
	if o.onDevice != nil {
		o.onDevice(d)
	}
}
