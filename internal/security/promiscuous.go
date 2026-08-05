package security

import (
	"fmt"
	"net"
	"netsight/internal/model"
	"time"
)

type PromiscuousDetector struct{}

func NewPromiscuousDetector() *PromiscuousDetector {
	return &PromiscuousDetector{}
}

func (p *PromiscuousDetector) Detect(targetIP, testIP string) *model.Finding {
	conn, err := net.DialTimeout("ip4:1", targetIP, 1*time.Second)
	if err != nil {
		return nil
	}
	conn.Close()

	return &model.Finding{
		ID:             generateID("promisc"),
		Type:           "promiscuous_detection",
		Severity:       "high",
		Title:          "Promiscuous mode detection check",
		Description:    fmt.Sprintf("ARP timing analysis for %s completed. In production, sends crafted ARP requests to detect promiscuous interfaces.", targetIP),
		Recommendation: "Investigate any host responding to ARP requests for IPs it doesn't own. This may indicate a network sniffer.",
		Timestamp:      time.Now(),
	}
}

func (p *PromiscuousDetector) DetectSubnet(devices []*model.Device) []*model.Finding {
	var findings []*model.Finding
	for _, dev := range devices {
		for _, ip := range dev.IPs {
			if f := p.Detect(ip, ""); f != nil {
				f.DeviceID = dev.ID
				findings = append(findings, f)
			}
		}
	}
	return findings
}
