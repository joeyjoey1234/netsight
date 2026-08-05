package scan

import (
	"context"
	"netsight/internal/model"
	"time"
)

type Scanner interface {
	Scan(ctx context.Context, target string) (*ScanResult, error)
}

type ScanResult struct {
	Devices []*model.Device
	Ports   map[string][]*model.Port
}

type ScanOrchestrator struct {
	OnProgress    func(scanID string, progress int)
	OnDeviceFound func(device *model.Device)
	OnComplete    func(scanID string, status string)

	ctx    context.Context
	cancel context.CancelFunc
}

func NewScanOrchestrator() *ScanOrchestrator {
	return &ScanOrchestrator{}
}

func (o *ScanOrchestrator) StartScan(subnet, preset string) (string, error) {
	scanID := generateID()
	o.ctx, o.cancel = context.WithCancel(context.Background())

	go o.runScan(o.ctx, scanID, subnet, preset)
	return scanID, nil
}

func (o *ScanOrchestrator) StopScan(scanID string) error {
	if o.cancel != nil {
		o.cancel()
	}
	return nil
}

func (o *ScanOrchestrator) runScan(ctx context.Context, scanID, subnet, preset string) {
	_ = preset
	start := time.Now()

	if o.OnProgress != nil {
		o.OnProgress(scanID, 10)
	}
	liveHosts, err := PingSweep(ctx, subnet)
	if err != nil {
		if o.OnComplete != nil {
			o.OnComplete(scanID, "failed")
		}
		return
	}

	if o.OnProgress != nil {
		o.OnProgress(scanID, 30)
	}
	arpTable, err := ARPScan(ctx, subnet)
	_ = arpTable
	_ = err

	if o.OnProgress != nil {
		o.OnProgress(scanID, 50)
	}
	totalHosts := len(liveHosts)
	for i, ip := range liveHosts {
		select {
		case <-ctx.Done():
			return
		default:
		}

		device := &model.Device{
			ID:        generateDeviceID(ip),
			IPs:       []string{ip},
			Role:      "unknown",
			FirstSeen: start,
			LastSeen:  time.Now(),
		}

		if mac, ok := arpTable[ip]; ok {
			device.MAC = mac
			device.Vendor = LookupOUI(mac)
		}

		ports, _ := TCPSynScan(ctx, ip, DefaultPorts())
		for _, p := range ports {
			if p.State == "open" && isBannerable(p.Number) {
				banner := GrabBanner(ctx, ip, p.Number)
				if banner != "" {
					p.Banner = banner
					p.Service, p.Version = ParseBanner(p.Number, banner)
				}
			}
		}

		title := GrabHTTPTitle(ctx, ip)
		if title != "" {
			device.Hostname = title
		}

		if o.OnDeviceFound != nil {
			o.OnDeviceFound(device)
		}

		progress := 50 + int(float64(i+1)/float64(totalHosts)*40)
		if o.OnProgress != nil {
			o.OnProgress(scanID, progress)
		}
	}

	if o.OnProgress != nil {
		o.OnProgress(scanID, 100)
	}
	if o.OnComplete != nil {
		o.OnComplete(scanID, "completed")
	}
}
