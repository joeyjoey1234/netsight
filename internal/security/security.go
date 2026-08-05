package security

import (
	"fmt"
	"netsight/internal/model"
	"time"
)

func RunAllSecurityChecks(devices []*model.Device, expectedDHCPServers []string) []*model.Finding {
	var findings []*model.Finding

	credChecker := NewDefaultCredentialChecker()
	for _, device := range devices {
		findings = append(findings, credChecker.CheckDefaultCredentials(device)...)
	}

	shareScanner := NewShareScanner()
	var ips []string
	for _, d := range devices {
		ips = append(ips, d.IPs...)
	}
	findings = append(findings, shareScanner.ScanSubnet(ips)...)

	arpDetector := NewARPSpoofDetector()
	if arpDetector.GetARPConflictCount() > 0 {
		findings = append(findings, &model.Finding{
			ID:             generateID("arp-summary"),
			Type:           "arp_spoofing",
			Severity:       "high",
			Title:          fmt.Sprintf("ARP spoofing summary: %d conflicts detected", arpDetector.GetARPConflictCount()),
			Description:    "Multiple ARP conflicts detected during the scan. Check individual findings for details.",
			Recommendation: "Enable Dynamic ARP Inspection (DAI) on managed switches.",
			Timestamp:      time.Now(),
		})
	}

	_ = NewDHCPDetector(expectedDHCPServers)

	return findings
}

type DefaultCredentialChecker struct{}

func NewDefaultCredentialChecker() *DefaultCredentialChecker {
	return &DefaultCredentialChecker{}
}

func (d *DefaultCredentialChecker) CheckDefaultCredentials(device *model.Device) []*model.Finding {
	return CheckDefaultCredentials(device)
}
