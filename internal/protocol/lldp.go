package protocol

import (
	"fmt"
	"strings"
)

type LLDPPacket struct {
	ChassisID     string
	PortID        string
	TTL           int
	SystemName    string
	SystemDesc    string
	PortDesc      string
	Capabilities  []string
	MgmtAddresses []string
	VLANID        int
	VLANName      string
	SrcMAC        string
}

func ParseLLDP(data []byte, srcMAC string) (*LLDPPacket, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("LLDP packet too short: %d bytes", len(data))
	}

	packet := &LLDPPacket{SrcMAC: srcMAC}

	offset := 0
	for offset < len(data)-2 {
		tlvType := data[offset] >> 1
		tlvLen := int(uint16(data[offset]&0x01)<<8 | uint16(data[offset+1]))
		if tlvLen < 2 || offset+tlvLen > len(data)+2 {
			break
		}
		value := data[offset+2 : offset+tlvLen]

		switch tlvType {
		case 0:
			return packet, nil
		case 1:
			packet.ChassisID = parseLLDPChassisID(value)
		case 2:
			packet.PortID = parseLLDPPortID(value)
		case 3:
			if len(value) >= 2 {
				packet.TTL = int(uint16(value[0])<<8 | uint16(value[1]))
			}
		case 4:
			packet.PortDesc = strings.TrimRight(string(value), "\x00")
		case 5:
			packet.SystemName = strings.TrimRight(string(value), "\x00")
		case 6:
			packet.SystemDesc = strings.TrimRight(string(value), "\x00")
		case 7:
			packet.Capabilities = parseLLDPCapabilities(value)
		case 8:
			packet.MgmtAddresses = append(packet.MgmtAddresses, parseLLDPMgmtAddress(value))
		case 127:
			if len(value) >= 7 {
				oui := fmt.Sprintf("%02X%02X%02X", value[0], value[1], value[2])
				subtype := value[3]
				if oui == "0080C2" && subtype == 1 && len(value) >= 7 {
					packet.VLANID = int(uint16(value[4])<<8 | uint16(value[5]))
				}
			}
		}

		offset += tlvLen
	}

	return packet, nil
}

func parseLLDPChassisID(data []byte) string {
	if len(data) < 3 {
		return string(data)
	}
	if data[0] == 0x04 && len(data) >= 7 {
		return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", data[1], data[2], data[3], data[4], data[5], data[6])
	}
	if data[0] == 0x05 && len(data) >= 6 {
		return fmt.Sprintf("%d.%d.%d.%d", data[2], data[3], data[4], data[5])
	}
	return strings.TrimRight(string(data[1:]), "\x00")
}

func parseLLDPPortID(data []byte) string {
	if len(data) < 2 {
		return string(data)
	}
	return strings.TrimRight(string(data[1:]), "\x00")
}

func parseLLDPCapabilities(data []byte) []string {
	caps := []string{}
	capNames := map[uint16]string{
		0x0001: "Other",
		0x0002: "Repeater",
		0x0004: "Bridge",
		0x0008: "WLAN AP",
		0x0010: "Router",
		0x0020: "Telephone",
		0x0040: "DOCSIS",
		0x0080: "Station Only",
	}

	if len(data) >= 4 {
		capBits := uint16(data[0])<<8 | uint16(data[1])
		for bit, name := range capNames {
			if capBits&bit != 0 {
				caps = append(caps, name)
			}
		}
	}
	return caps
}

func parseLLDPMgmtAddress(data []byte) string {
	if len(data) < 8 || data[0] != 0x01 {
		return ""
	}
	return fmt.Sprintf("%d.%d.%d.%d", data[7], data[8], data[9], data[10])
}
