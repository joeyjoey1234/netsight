package health

import (
	"fmt"
	"netsight/internal/model"
	"sort"
	"sync"
	"time"
)

type TopTalker struct {
	IP         string
	TotalBytes int64
	Packets    int64
	LastSeen   time.Time
}

type BandwidthAnalyzer struct {
	mu        sync.Mutex
	talkers   map[string]*TopTalker
	startTime time.Time
}

func NewBandwidthAnalyzer() *BandwidthAnalyzer {
	return &BandwidthAnalyzer{
		talkers: make(map[string]*TopTalker),
	}
}

func (ba *BandwidthAnalyzer) Start() {
	ba.mu.Lock()
	defer ba.mu.Unlock()
	ba.startTime = time.Now()
	ba.talkers = make(map[string]*TopTalker)
}

func (ba *BandwidthAnalyzer) RecordTraffic(srcIP, dstIP string, length int) {
	ba.mu.Lock()
	defer ba.mu.Unlock()

	ba.recordIP(srcIP, length)
	ba.recordIP(dstIP, length)
}

func (ba *BandwidthAnalyzer) recordIP(ip string, length int) {
	if ip == "" || ip == "0.0.0.0" {
		return
	}

	t, exists := ba.talkers[ip]
	if !exists {
		t = &TopTalker{IP: ip}
		ba.talkers[ip] = t
	}
	t.TotalBytes += int64(length)
	t.Packets++
	t.LastSeen = time.Now()
}

func (ba *BandwidthAnalyzer) GetTopTalkers(n int) []*TopTalker {
	ba.mu.Lock()
	defer ba.mu.Unlock()

	talkers := make([]*TopTalker, 0, len(ba.talkers))
	for _, t := range ba.talkers {
		talkers = append(talkers, t)
	}

	sort.Slice(talkers, func(i, j int) bool {
		return talkers[i].TotalBytes > talkers[j].TotalBytes
	})

	if n > len(talkers) {
		n = len(talkers)
	}
	return talkers[:n]
}

func (ba *BandwidthAnalyzer) GenerateTopTalkerFindings() []*model.Finding {
	var findings []*model.Finding
	top := ba.GetTopTalkers(5)

	ba.mu.Lock()
	elapsed := time.Since(ba.startTime).Seconds()
	ba.mu.Unlock()

	for _, t := range top {
		if elapsed < 1 {
			continue
		}
		bps := float64(t.TotalBytes*8) / elapsed
		if bps > 10_000_000 {
			findings = append(findings, &model.Finding{
				ID:             generateFindingID("bandwidth"),
				Type:           "bandwidth_heavy_user",
				Severity:       "medium",
				Title:          fmt.Sprintf("High bandwidth usage: %s", t.IP),
				Description:    fmt.Sprintf("IP %s has transferred %.2f MB at %.2f Mbps over %.0f seconds (%d packets).", t.IP, float64(t.TotalBytes)/1e6, bps/1e6, elapsed, t.Packets),
				Recommendation: "Investigate the traffic source. Consider QoS policies if this is sustained.",
				Timestamp:      time.Now(),
			})
		}
	}
	return findings
}

func (ba *BandwidthAnalyzer) Elapsed() time.Duration {
	ba.mu.Lock()
	defer ba.mu.Unlock()
	return time.Since(ba.startTime)
}
