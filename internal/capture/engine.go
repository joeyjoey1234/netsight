package capture

import (
	"context"
	"fmt"
	"net"
	"netsight/internal/model"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type PacketHandler func(packet *model.PacketSummary)

type StatsHandler func(packetsPerSec, bytesPerSec int64)

// ResolveInterfaceName accepts either an Npcap device name or the friendly
// adapter name shown by net.Interfaces and returns the pcap device identifier.
func ResolveInterfaceName(name string) (string, error) {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		return "", err
	}
	for _, device := range devices {
		if device.Name == name || device.Description == name {
			return device.Name, nil
		}
	}
	return "", fmt.Errorf("capture interface not found: %s", name)
}

type Engine struct {
	ctx         context.Context
	cancel      context.CancelFunc
	handler     PacketHandler
	statsFn     StatsHandler
	isRunning   bool
	mu          sync.Mutex
	iface       string
	filter      string
	handle      *pcap.Handle
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
	handle, err := pcap.OpenLive(iface, 65536, true, pcap.BlockForever)
	if err != nil {
		e.isRunning = false
		return fmt.Errorf("packet capture unavailable: %w", err)
	}
	e.handle = handle
	if filter != "" {
		if err := handle.SetBPFFilter(filter); err != nil {
			handle.Close()
			e.isRunning = false
			return fmt.Errorf("invalid capture filter: %w", err)
		}
	}

	go e.captureLoop(handle)

	go e.statsLoop()

	return nil
}

func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.cancel != nil {
		e.cancel()
	}
	if e.handle != nil {
		e.handle.Close()
		e.handle = nil
	}
	e.isRunning = false
}

func (e *Engine) IsRunning() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.isRunning
}

func (e *Engine) captureLoop(handle *pcap.Handle) {
	defer handle.Close()

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

func (e *Engine) finishWithError(err error) {
	e.mu.Lock()
	e.isRunning = false
	e.mu.Unlock()
	if e.handler != nil {
		e.handler(&model.PacketSummary{Protocol: "ERROR", Info: err.Error()})
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
