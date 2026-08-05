package tools

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

type LatencyMonitor struct {
	mu      sync.Mutex
	running bool
	cancel  context.CancelFunc
	results []LatencyPoint
}

type LatencyPoint struct {
	Timestamp time.Time `json:"timestamp"`
	LatencyMs float64   `json:"latencyMs"`
	TimedOut  bool      `json:"timedOut"`
	SeqNum    int       `json:"seqNum"`
}

func NewLatencyMonitor() *LatencyMonitor {
	return &LatencyMonitor{}
}

func (m *LatencyMonitor) Start(ctx context.Context, target string, interval time.Duration, onPoint func(point *LatencyPoint)) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return fmt.Errorf("monitor already running")
	}
	m.running = true
	ctx, m.cancel = context.WithCancel(ctx)
	m.mu.Unlock()

	go func() {
		seq := 0
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				start := time.Now()

				conn, err := net.DialTimeout("ip4:icmp", target, time.Duration(interval))
				if err != nil {
					onPoint(&LatencyPoint{
						Timestamp: start,
						LatencyMs: 0,
						TimedOut:  true,
						SeqNum:    seq,
					})
					seq++
					continue
				}

				msg := buildICMPEcho(uint16(seq), []byte{0x00}, 64)
				conn.Write(msg)

				reply := make([]byte, 1500)
				conn.SetReadDeadline(time.Now().Add(time.Duration(interval)))
				_, err = conn.Read(reply)
				elapsed := time.Since(start)
				conn.Close()

				point := &LatencyPoint{
					Timestamp: start,
					SeqNum:    seq,
				}

				if err != nil {
					point.TimedOut = true
				} else {
					point.LatencyMs = float64(elapsed.Microseconds()) / 1000.0
				}

				onPoint(point)
				seq++
			}
		}
	}()

	return nil
}

func (m *LatencyMonitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cancel != nil {
		m.cancel()
	}
	m.running = false
}

func (m *LatencyMonitor) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}
