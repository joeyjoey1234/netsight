package health

import (
	"fmt"
	"net"
	"netsight/internal/model"
	"sync"
	"time"
)

type DuplexCheck struct {
	mu           sync.Mutex
	interfaceMap map[string]*model.InterfaceInfo
}

func NewDuplexCheck() *DuplexCheck {
	return &DuplexCheck{
		interfaceMap: make(map[string]*model.InterfaceInfo),
	}
}

func (d *DuplexCheck) RegisterInterface(info *model.InterfaceInfo) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.interfaceMap[info.Name] = info
}

func (d *DuplexCheck) CheckInterfaces() []*model.Finding {
	d.mu.Lock()
	defer d.mu.Unlock()

	var findings []*model.Finding

	for _, info := range d.interfaceMap {
		if info.Duplex != "" && info.Duplex != "full" {
			findings = append(findings, &model.Finding{
				ID:             generateFindingID("duplex"),
				Type:           "duplex_mismatch",
				Severity:       "high",
				Title:          fmt.Sprintf("Half-duplex detected on %s", info.Name),
				Description:    fmt.Sprintf("Interface %s is running at %d Mbps in %s-duplex mode with %d CRC errors and %d collisions. This may indicate a configuration mismatch with the connected device.", info.Name, info.Speed, info.Duplex, info.CRCErrors, info.Collisions),
				Recommendation: "Configure both sides of the link for auto-negotiation or matching full-duplex settings.",
				Timestamp:      time.Now(),
			})
		}

		if info.CRCErrors > 100 {
			findings = append(findings, &model.Finding{
				ID:             generateFindingID("duplex"),
				Type:           "high_error_rate",
				Severity:       "medium",
				Title:          fmt.Sprintf("High CRC errors on %s", info.Name),
				Description:    fmt.Sprintf("Interface %s has %d CRC errors. This often indicates a duplex mismatch, faulty cable, or electrical interference.", info.Name, info.CRCErrors),
				Recommendation: "Check cable quality, verify duplex settings match on both ends, and test with a known-good cable.",
				Timestamp:      time.Now(),
			})
		}
	}

	return findings
}

func (d *DuplexCheck) CheckLinkDuplex(localInfo *model.InterfaceInfo, speed int64, duplex string) *model.Finding {
	if duplex != "full" {
		return &model.Finding{
			ID:             generateFindingID("duplex"),
			Type:           "duplex_mismatch",
			Severity:       "high",
			Title:          fmt.Sprintf("Half-duplex detected on %s", localInfo.Name),
			Description:    fmt.Sprintf("Interface %s is running at %d Mbps in %s-duplex mode. This may indicate a configuration mismatch with the connected device.", localInfo.Name, speed, duplex),
			Recommendation: "Configure both sides of the link for auto-negotiation or matching full-duplex settings.",
			Timestamp:      time.Now(),
		}
	}
	return nil
}

func (d *DuplexCheck) LocalInterfaces() []*model.InterfaceInfo {
	d.mu.Lock()
	defer d.mu.Unlock()

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	var result []*model.InterfaceInfo
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()
		var ips []string
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ips = append(ips, ipNet.IP.String())
		}

		result = append(result, &model.InterfaceInfo{
			Name:  iface.Name,
			MAC:   iface.HardwareAddr.String(),
			IPs:   ips,
			MTU:   iface.MTU,
			Speed: 0,
		})
	}

	return result
}

func generateFindingID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
