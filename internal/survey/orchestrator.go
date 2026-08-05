package survey

import (
	"context"
	"fmt"
	"sync"
	"time"

	"netsight/internal/model"
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
	select {
	case <-ctx.Done():
		return
	case <-time.After(500 * time.Millisecond):
	}
	if o.onProgress != nil {
		o.onProgress(o.scanID, 10)
	}
}

func (o *Orchestrator) runControlPlaneListener(ctx context.Context, subnet string) {
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
	<-ctx.Done()
}

func (o *Orchestrator) runBroadcastDetection(ctx context.Context, subnet string) {
	<-ctx.Done()
}

func (o *Orchestrator) runARPSpoofDetection(ctx context.Context, subnet string) {
	<-ctx.Done()
}

func (o *Orchestrator) runRogueDHCPDetection(ctx context.Context, subnet string) {
	<-ctx.Done()
}

func (o *Orchestrator) runBandwidthEstimation(ctx context.Context, subnet string) {
	<-ctx.Done()
}

func (o *Orchestrator) runPortScan(ctx context.Context, subnet string) {
	<-ctx.Done()
}

func (o *Orchestrator) runOpenSharesScan(ctx context.Context) {
	<-ctx.Done()
}

func (o *Orchestrator) runPassiveOSFingerprint(ctx context.Context, subnet string) {
	<-ctx.Done()
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
