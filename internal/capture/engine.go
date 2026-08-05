package capture

import (
	"context"
	"fmt"
	"netsight/internal/model"
	"sync"
	"time"
)

type PacketHandler func(packet *model.PacketSummary)

type StatsHandler func(packetsPerSec, bytesPerSec int64)

type Engine struct {
	ctx         context.Context
	cancel      context.CancelFunc
	handler     PacketHandler
	statsFn     StatsHandler
	isRunning   bool
	mu          sync.Mutex
	iface       string
	filter      string
	packetCount int64
	byteCount   int64
}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) Start(iface string, filter string, handler PacketHandler, statsFn StatsHandler) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.isRunning {
		return fmt.Errorf("capture already running")
	}

	e.ctx, e.cancel = context.WithCancel(context.Background())
	e.handler = handler
	e.statsFn = statsFn
	e.iface = iface
	e.filter = filter
	e.isRunning = true
	e.packetCount = 0
	e.byteCount = 0

	go e.captureLoop()

	go e.statsLoop()

	return nil
}

func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.cancel != nil {
		e.cancel()
	}
	e.isRunning = false
}

func (e *Engine) IsRunning() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.isRunning
}

func (e *Engine) captureLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	count := 0

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			count++
			e.mu.Lock()
			e.packetCount++
			e.byteCount += int64(64 + (count%1400))
			e.mu.Unlock()

			if e.handler != nil {
				e.handler(&model.PacketSummary{
					Number:    count,
					Timestamp: time.Now().Format("15:04:05.000"),
					SrcMAC:    "aa:bb:cc:dd:ee:ff",
					DstMAC:    "ff:ee:dd:cc:bb:aa",
					SrcIP:     "192.168.1.100",
					DstIP:     "192.168.1.1",
					Protocol:  "TCP",
					SrcPort:   54321,
					DstPort:   443,
					Length:    64 + (count % 1400),
					Info:      "Sample capture data — real gopacket coming in Agent E integration",
				})
			}
		}
	}
}

func (e *Engine) statsLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var lastPackets, lastBytes int64

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.mu.Lock()
			pps := e.packetCount - lastPackets
			bps := (e.byteCount - lastBytes)
			lastPackets = e.packetCount
			lastBytes = e.byteCount
			e.mu.Unlock()

			if e.statsFn != nil {
				e.statsFn(pps, bps)
			}
		}
	}
}
