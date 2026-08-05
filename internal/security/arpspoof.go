package security

import (
	"fmt"
	"netsight/internal/model"
	"sync"
	"time"
)

type ARPSpoofDetector struct {
	mu       sync.RWMutex
	arpCache map[string]*arpEntry
}

type arpEntry struct {
	MAC       string
	FirstSeen time.Time
	LastSeen  time.Time
	Count     int
	MACs      map[string]int
}

func NewARPSpoofDetector() *ARPSpoofDetector {
	return &ARPSpoofDetector{
		arpCache: make(map[string]*arpEntry),
	}
}

func (a *ARPSpoofDetector) ProcessARPPacket(ip, mac string) *model.Finding {
	a.mu.Lock()
	defer a.mu.Unlock()

	entry, exists := a.arpCache[ip]
	if !exists {
		a.arpCache[ip] = &arpEntry{
			MAC:       mac,
			FirstSeen: time.Now(),
			LastSeen:  time.Now(),
			Count:     1,
			MACs:      map[string]int{mac: 1},
		}
		return nil
	}

	entry.Count++
	entry.LastSeen = time.Now()
	entry.MACs[mac]++

	if len(entry.MACs) > 1 {
		var macList string
		for m := range entry.MACs {
			macList += m + " "
		}

		return &model.Finding{
			ID:             generateID("arp-spoof"),
			Type:           "arp_spoofing",
			Severity:       "critical",
			Title:          fmt.Sprintf("ARP spoofing detected for IP %s", ip),
			Description:    fmt.Sprintf("Multiple MAC addresses claiming IP %s: %s. Original MAC was %s. This indicates an active ARP spoofing attack or ARP cache poisoning.", ip, macList, entry.MAC),
			Recommendation: "Immediately investigate the conflicting MAC addresses. Enable Dynamic ARP Inspection (DAI) on managed switches. Track down the attacker's physical port.",
			Timestamp:      time.Now(),
		}
	}

	return nil
}

func (a *ARPSpoofDetector) GetARPConflictCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	count := 0
	for _, entry := range a.arpCache {
		if len(entry.MACs) > 1 {
			count++
		}
	}
	return count
}

func (a *ARPSpoofDetector) GetConflicts() []*model.Finding {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var findings []*model.Finding
	for ip, entry := range a.arpCache {
		if len(entry.MACs) > 1 {
			var macList string
			for m := range entry.MACs {
				macList += m + " "
			}

			findings = append(findings, &model.Finding{
				ID:             generateID("arp-spoof"),
				Type:           "arp_spoofing",
				Severity:       "critical",
				Title:          fmt.Sprintf("ARP spoofing detected for IP %s", ip),
				Description:    fmt.Sprintf("Multiple MAC addresses claiming IP %s: %s. Original MAC was %s.", ip, macList, entry.MAC),
				Recommendation: "Enable Dynamic ARP Inspection (DAI) on managed switches.",
				Timestamp:      entry.LastSeen,
			})
		}
	}
	return findings
}

func (a *ARPSpoofDetector) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.arpCache = make(map[string]*arpEntry)
}
