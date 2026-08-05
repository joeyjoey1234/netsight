package protocol

import (
	"fmt"
	"strings"
)

type CDPPacket struct {
	Version      string
	TTL          int
	DeviceID     string
	Platform     string
	Capabilities []string
	PortID       string
	NativeVLAN   int
	Addresses    []string
	SrcMAC       string
}

func ParseCDP(data []byte, srcMAC string) (*CDPPacket, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("CDP packet too short: %d bytes", len(data))
	}

	packet := &CDPPacket{
		Version: fmt.Sprintf("%d", data[0]),
		TTL:     int(data[1]),
		SrcMAC:  srcMAC,
	}

	offset := 4
	for offset < len(data)-4 {
		tlvType := uint16(data[offset])<<8 | uint16(data[offset+1])
		tlvLen := int(uint16(data[offset+2])<<8 | uint16(data[offset+3]))
		if tlvLen < 4 || offset+tlvLen > len(data) {
			break
		}
		value := string(data[offset+4 : offset+tlvLen])

		switch tlvType {
		case 0x0001:
			packet.DeviceID = strings.TrimRight(value, "\x00")
		case 0x0002:
			packet.Addresses = append(packet.Addresses, parseCDPAddress(data[offset+4:offset+tlvLen]))
		case 0x0003:
			packet.PortID = strings.TrimRight(value, "\x00")
		case 0x0004:
			packet.Capabilities = parseCDPCapabilities(data[offset+4 : offset+tlvLen])
		case 0x0005:
			packet.Version = strings.TrimRight(value, "\x00")
		case 0x0006:
			packet.Platform = strings.TrimRight(value, "\x00")
		case 0x000A:
			if len(data[offset+4:]) >= 2 {
				packet.NativeVLAN = int(uint16(data[offset+4])<<8 | uint16(data[offset+5]))
			}
		}

		offset += tlvLen
	}

	return packet, nil
}

func parseCDPAddress(data []byte) string {
	if len(data) < 8 {
		return ""
	}
	if len(data) >= 13 && data[6] == 0x01 && data[7] == 0x01 {
		return fmt.Sprintf("%d.%d.%d.%d", data[10], data[11], data[12], data[13])
	}
	return ""
}

func parseCDPCapabilities(data []byte) []string {
	caps := []string{}
	capNames := map[uint32]string{
		0x0001: "Router",
		0x0002: "Transparent Bridge",
		0x0004: "Source Route Bridge",
		0x0008: "Switch",
		0x0010: "Host",
		0x0020: "IGMP",
		0x0040: "Repeater",
		0x0080: "VoIP Phone",
		0x0100: "Remotely Managed",
	}

	if len(data) >= 4 {
		capBits := uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])
		for bit, name := range capNames {
			if capBits&bit != 0 {
				caps = append(caps, name)
			}
		}
	}
	return caps
}
