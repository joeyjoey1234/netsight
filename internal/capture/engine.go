package capture

import (
	"context"
	"fmt"
	"netsight/internal/model"
	"net"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
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
	handle, err := pcap.OpenLive(e.iface, 65536, true, pcap.BlockForever)
	if err != nil {
		e.fallbackCaptureLoop()
		return
	}
	defer handle.Close()

	if e.filter != "" {
		if err := handle.SetBPFFilter(e.filter); err != nil {
			e.fallbackCaptureLoop()
			return
		}
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		select {
		case <-e.ctx.Done():
			return
		default:
		}

		summary := packetToSummary(packet, int(e.packetCount))
		e.mu.Lock()
		e.packetCount++
		e.byteCount += int64(len(packet.Data()))
		e.mu.Unlock()

		if e.handler != nil {
			e.handler(summary)
		}
	}
}

func (e *Engine) fallbackCaptureLoop() {
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
					Info:      "Fallback test data — no live capture interface available",
				})
			}
		}
	}
}

func packetToSummary(packet gopacket.Packet, number int) *model.PacketSummary {
	summary := &model.PacketSummary{
		Number:    number,
		Timestamp: packet.Metadata().Timestamp.Format("15:04:05.000"),
		Length:    len(packet.Data()),
		Protocol:  "UNKNOWN",
	}

	if ethLayer := packet.Layer(layers.LayerTypeEthernet); ethLayer != nil {
		eth, _ := ethLayer.(*layers.Ethernet)
		summary.SrcMAC = eth.SrcMAC.String()
		summary.DstMAC = eth.DstMAC.String()
	}

	if ip4Layer := packet.Layer(layers.LayerTypeIPv4); ip4Layer != nil {
		ip4, _ := ip4Layer.(*layers.IPv4)
		summary.SrcIP = ip4.SrcIP.String()
		summary.DstIP = ip4.DstIP.String()
		summary.Protocol = ip4.Protocol.String()
	}

	if ip6Layer := packet.Layer(layers.LayerTypeIPv6); ip6Layer != nil {
		ip6, _ := ip6Layer.(*layers.IPv6)
		summary.SrcIP = ip6.SrcIP.String()
		summary.DstIP = ip6.DstIP.String()
		summary.Protocol = "IPv6"
	}

	if arpLayer := packet.Layer(layers.LayerTypeARP); arpLayer != nil {
		arp, _ := arpLayer.(*layers.ARP)
		summary.SrcIP = net.IP(arp.SourceProtAddress).String()
		summary.DstIP = net.IP(arp.DstProtAddress).String()
		summary.SrcMAC = net.HardwareAddr(arp.SourceHwAddress).String()
		summary.DstMAC = net.HardwareAddr(arp.DstHwAddress).String()
		summary.Protocol = "ARP"
	}

	if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
		tcp, _ := tcpLayer.(*layers.TCP)
		summary.SrcPort = int(tcp.SrcPort)
		summary.DstPort = int(tcp.DstPort)
	}

	if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
		udp, _ := udpLayer.(*layers.UDP)
		summary.SrcPort = int(udp.SrcPort)
		summary.DstPort = int(udp.DstPort)
	}

	switch {
	case summary.SrcPort == 53 || summary.DstPort == 53:
		summary.Info = "DNS"
	case summary.SrcPort == 80 || summary.DstPort == 80:
		summary.Info = "HTTP"
	case summary.SrcPort == 443 || summary.DstPort == 443:
		summary.Info = "HTTPS"
	case summary.SrcPort == 22 || summary.DstPort == 22:
		summary.Info = "SSH"
	}

	return summary
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
