package scan

import (
	"context"
	"fmt"
	"net"
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
	OnDeviceFound func(scanID string, device *model.Device)
	OnPortsFound  func(scanID, deviceID string, ports []*model.Port)
	OnStarted     func(*model.Scan)
	OnComplete    func(scanID string, status string)
	OnError       func(scanID string, err error)

	ctx    context.Context
	cancel context.CancelFunc
}

func NewScanOrchestrator() *ScanOrchestrator {
	return &ScanOrchestrator{}
}

func (o *ScanOrchestrator) StartScan(subnet, preset string) (string, error) {
	if _, _, err := net.ParseCIDR(subnet); err != nil {
		return "", fmt.Errorf("invalid subnet: %w", err)
	}
	if _, err := presetConfig(preset); err != nil {
		return "", err
	}
	scanID := generateID()
	o.ctx, o.cancel = context.WithCancel(context.Background())

	if o.OnStarted != nil {
		o.OnStarted(&model.Scan{ID: scanID, Timestamp: time.Now(), Subnet: subnet, Preset: preset, Status: "running"})
	}
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
	start := time.Now()
	ports := DefaultPorts()

	if o.OnProgress != nil {
		o.OnProgress(scanID, 10)
	}
	liveHosts, err := PingSweep(ctx, subnet)
	if err != nil {
		if o.OnError != nil {
			o.OnError(scanID, err)
		}
		if o.OnComplete != nil {
			o.OnComplete(scanID, "failed")
		}
		return
	}

	if o.OnProgress != nil {
		o.OnProgress(scanID, 30)
	}
	arpTable, err := ARPScan(ctx, subnet)
	if err != nil {
		if o.OnError != nil {
			o.OnError(scanID, err)
		}
		arpTable = make(map[string]string)
	}

	if o.OnProgress != nil {
		o.OnProgress(scanID, 50)
	}
	totalHosts := len(liveHosts)
	for i, ip := range liveHosts {
		select {
		case <-ctx.Done():
			if o.OnComplete != nil {
				o.OnComplete(scanID, "cancelled")
			}
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

		scanPorts, err := TCPSynScan(ctx, ip, ports)
		if err != nil {
			if o.OnError != nil {
				o.OnError(scanID, err)
			}
			if o.OnComplete != nil {
				o.OnComplete(scanID, "failed")
			}
			return
		}
		for _, p := range scanPorts {
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
			o.OnDeviceFound(scanID, device)
		}
		if o.OnPortsFound != nil {
			o.OnPortsFound(scanID, device.ID, scanPorts)
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

func presetConfig(name string) (string, error) {
	switch name {
	case "quick", "short", "long", "manual":
		return name, nil
	default:
		return "", fmt.Errorf("unknown preset: %s", name)
	}
}
