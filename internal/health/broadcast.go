package health

import (
	"fmt"
	"netsight/internal/model"
	"sync"
	"time"
)

type BroadcastDetector struct {
	mu             sync.Mutex
	broadcastCount int64
	multicastCount int64
	arpCount       int64
	totalPackets   int64
	macFlaps       map[string][]string
	startTime      time.Time
	stormThreshold int64
}

func NewBroadcastDetector() *BroadcastDetector {
	return &BroadcastDetector{
		macFlaps:       make(map[string][]string),
		stormThreshold: 1000,
	}
}

func (b *BroadcastDetector) Start() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.startTime = time.Now()
	b.broadcastCount = 0
	b.multicastCount = 0
	b.arpCount = 0
	b.totalPackets = 0
}

func (b *BroadcastDetector) ProcessPacket(dstMAC string, etherType uint16, length int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.totalPackets++

	if dstMAC == "ff:ff:ff:ff:ff:ff" {
		b.broadcastCount++
		if etherType == 0x0806 {
			b.arpCount++
		}
	}

	if len(dstMAC) >= 2 {
		c := dstMAC[1]
		if c == '1' || c == '3' || c == '5' || c == '7' || c == '9' || c == 'b' || c == 'd' || c == 'f' {
			b.multicastCount++
		}
	}
}

func (b *BroadcastDetector) RecordMACFlap(mac string, port string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	ports, exists := b.macFlaps[mac]
	if !exists {
		b.macFlaps[mac] = []string{port}
		return
	}

	if len(ports) > 0 && ports[len(ports)-1] == port {
		return
	}

	b.macFlaps[mac] = append(ports, port)
}

func (b *BroadcastDetector) DetectStorm() *model.Finding {
	b.mu.Lock()
	defer b.mu.Unlock()

	elapsed := time.Since(b.startTime).Seconds()
	if elapsed < 1 {
		return nil
	}

	bps := float64(b.broadcastCount) / elapsed
	if bps > float64(b.stormThreshold) {
		broadcastPct := float64(0)
		if b.totalPackets > 0 {
			broadcastPct = float64(b.broadcastCount) / float64(b.totalPackets) * 100
		}
		return &model.Finding{
			ID:             generateFindingID("storm"),
			Type:           "broadcast_storm",
			Severity:       "critical",
			Title:          "Broadcast storm detected",
			Description:    fmt.Sprintf("Detected %.0f broadcast packets/sec (threshold: %d). %.0f%% of total traffic is broadcast. This may indicate a switching loop.", bps, b.stormThreshold, broadcastPct),
			Recommendation: "Check for spanning tree misconfiguration, unmanaged switches creating loops, or malfunctioning NICs. Review STP topology for blocked ports.",
			Timestamp:      time.Now(),
		}
	}
	return nil
}

func (b *BroadcastDetector) DetectMACFlapping() []*model.Finding {
	var findings []*model.Finding

	b.mu.Lock()
	defer b.mu.Unlock()

	for mac, ports := range b.macFlaps {
		if len(ports) >= 3 {
			findings = append(findings, &model.Finding{
				ID:             generateFindingID("flap"),
				Type:           "mac_flapping",
				Severity:       "high",
				Title:          fmt.Sprintf("MAC flapping detected: %s", mac),
				Description:    fmt.Sprintf("MAC address %s has been seen on %d different ports: %v. This indicates a switching loop.", mac, len(ports), ports),
				Recommendation: "Check the spanning tree configuration. Identify the source of the loop by tracing the flapping MAC through switch MAC tables.",
				Timestamp:      time.Now(),
			})
		}
	}
	return findings
}

func (b *BroadcastDetector) Stats() (broadcastCount, multicastCount, arpCount, totalPackets int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.broadcastCount, b.multicastCount, b.arpCount, b.totalPackets
}
