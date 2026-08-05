package security

import (
	"fmt"
	"netsight/internal/model"
	"time"
)

type DHCPDetector struct {
	expectedServers map[string]bool
}

func NewDHCPDetector(expectedServers []string) *DHCPDetector {
	d := &DHCPDetector{
		expectedServers: make(map[string]bool),
	}
	for _, ip := range expectedServers {
		d.expectedServers[ip] = true
	}
	return d
}

func (d *DHCPDetector) ProcessDHCPOffer(serverIP, offeredIP, offeredGateway string, offeredDNS []string) *model.Finding {
	if d.expectedServers[serverIP] {
		return nil
	}

	return &model.Finding{
		ID:             generateID("rogue-dhcp"),
		Type:           "rogue_dhcp",
		Severity:       "critical",
		Title:          fmt.Sprintf("Rogue DHCP server detected: %s", serverIP),
		Description:    fmt.Sprintf("DHCP offer received from unauthorized server %s offering IP %s, gateway %s, DNS %v. This may be a rogue device or misconfigured router.", serverIP, offeredIP, offeredGateway, offeredDNS),
		Recommendation: "Immediately locate and disconnect the rogue DHCP server. Check for unauthorized routers, travel routers, or misconfigured devices. Enable DHCP snooping on managed switches.",
		Timestamp:      time.Now(),
	}
}

func (d *DHCPDetector) ProcessDHCPAck(serverIP string, clientIP string) *model.Finding {
	if d.expectedServers[serverIP] {
		return nil
	}
	return &model.Finding{
		ID:             generateID("rogue-dhcp-ack"),
		Type:           "rogue_dhcp",
		Severity:       "critical",
		Title:          fmt.Sprintf("Rogue DHCP server leased address: %s -> %s", serverIP, clientIP),
		Description:    fmt.Sprintf("Device %s received a DHCP lease from unauthorized server %s.", clientIP, serverIP),
		Recommendation: "Locate and remove the rogue DHCP server. Check DHCP snooping database on managed switches.",
		Timestamp:      time.Now(),
	}
}

func (d *DHCPDetector) IsAuthorized(serverIP string) bool {
	return d.expectedServers[serverIP]
}

func (d *DHCPDetector) AddAuthorized(serverIP string) {
	d.expectedServers[serverIP] = true
}
