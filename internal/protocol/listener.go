package protocol

import (
	"context"
	"fmt"
)

type Listener struct {
	ctx          context.Context
	cancel       context.CancelFunc
	roleDetector *DeviceRoleDetection

	onBPDU  func(bpdu *BPDU)
	onEvent func(event *ProtocolEvent)
	onCDP   func(pkt *CDPPacket)
	onLLDP  func(pkt *LLDPPacket)
}

func NewListener() *Listener {
	return &Listener{
		roleDetector: NewDeviceRoleDetection(),
	}
}

func (l *Listener) Start(ctx context.Context) {
	l.ctx, l.cancel = context.WithCancel(ctx)
}

func (l *Listener) Stop() {
	if l.cancel != nil {
		l.cancel()
	}
}

func (l *Listener) SetHandlers(
	bpduFn func(*BPDU),
	eventFn func(*ProtocolEvent),
	cdpFn func(*CDPPacket),
	lldpFn func(*LLDPPacket),
) {
	l.onBPDU = bpduFn
	l.onEvent = eventFn
	l.onCDP = cdpFn
	l.onLLDP = lldpFn
}

func (l *Listener) ProcessPacket(data []byte, srcMAC string) {
	if len(data) < 14 {
		return
	}

	etherType := uint16(data[12])<<8 | uint16(data[13])

	switch etherType {
	case 0x88CC:
		if lldp, err := ParseLLDP(data[14:], srcMAC); err == nil && l.onLLDP != nil {
			l.onLLDP(lldp)
			l.roleDetector.ProcessEvent(&ProtocolEvent{
				Type:    "LLDP",
				SrcMAC:  srcMAC,
				Details: map[string]string{"systemName": lldp.SystemName},
			})
		}
	case 0x8100:
		if len(data) >= 18 {
			innerType := uint16(data[16])<<8 | uint16(data[17])
			if innerType == 0x2000 {
			}
		}
	}

	if len(data) > 23 && etherType == 0x0800 {
		protocol := data[23]
		switch protocol {
		case 89:
			l.roleDetector.ProcessEvent(&ProtocolEvent{
				Type:   "OSPF",
				SrcMAC: srcMAC,
				SrcIP:  fmt.Sprintf("%d.%d.%d.%d", data[26], data[27], data[28], data[29]),
			})
		}
	}
}

func (l *Listener) GetRole(mac string) string {
	return l.roleDetector.GetRole(mac)
}
