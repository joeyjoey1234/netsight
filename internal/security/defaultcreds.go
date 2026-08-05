package security

import (
	"fmt"
	"netsight/internal/model"
	"time"
)

type DefaultCredentialEntry struct {
	Port        int
	Service     string
	Description string
	Risk        string
}

var knownDefaultCreds = []DefaultCredentialEntry{
	{23, "Telnet", "Telnet sends credentials in cleartext", "critical"},
	{21, "FTP", "FTP sends credentials in cleartext", "critical"},
	{80, "HTTP-Management", "HTTP management interface may use default credentials", "high"},
	{443, "HTTPS-Management", "HTTPS management interface may use default credentials", "high"},
	{161, "SNMP", "SNMP may use 'public'/'private' community strings", "high"},
	{22, "SSH", "SSH with default credentials or weak keys", "medium"},
	{3389, "RDP", "RDP enabled without Network Level Authentication", "medium"},
	{8080, "HTTP-Alt", "Alternate HTTP management port", "high"},
	{8443, "HTTPS-Alt", "Alternate HTTPS management port", "high"},
}

func CheckDefaultCredentials(device *model.Device) []*model.Finding {
	var findings []*model.Finding

	deviceName := device.Hostname
	if deviceName == "" {
		if len(device.IPs) > 0 {
			deviceName = device.IPs[0]
		} else {
			deviceName = device.MAC
		}
	}

	for _, dc := range knownDefaultCreds {
		findings = append(findings, &model.Finding{
			ID:             generateID("defcred"),
			Type:           "default_credential",
			Severity:       dc.Risk,
			DeviceID:       device.ID,
			Title:          fmt.Sprintf("%s on port %d (potential default credentials)", dc.Service, dc.Port),
			Description:    fmt.Sprintf("Device %s has %s open on port %d. %s.", deviceName, dc.Service, dc.Port, dc.Description),
			Recommendation: fmt.Sprintf("Disable %s if not needed, or ensure strong non-default credentials are in use.", dc.Service),
			Timestamp:      time.Now(),
		})
	}

	return findings
}

func KnownDefaultServices() []DefaultCredentialEntry {
	return knownDefaultCreds
}
