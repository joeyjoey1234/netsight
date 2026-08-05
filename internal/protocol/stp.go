package protocol

import (
	"fmt"
	"netsight/internal/model"
)

type BPDU struct {
	ProtocolID   uint16
	Version      uint8
	Type         uint8
	Flags        uint8
	RootID       uint64
	RootPathCost uint32
	BridgeID     uint64
	PortID       uint16
	MessageAge   uint16
	MaxAge       uint16
	HelloTime    uint16
	ForwardDelay uint16
	VLANID       int
	SrcMAC       string
}

type STPTree struct {
	RootBridge string
	RootMAC    string
	Bridges    map[string]*BridgeInfo
}

type BridgeInfo struct {
	BridgeID       string
	MAC            string
	Priority       int
	RootPort       string
	DesignatedPort string
	IsRoot         bool
	PathCostToRoot uint32
}

func ParseBPDU(data []byte, srcMAC string) (*BPDU, error) {
	if len(data) < 35 {
		return nil, fmt.Errorf("BPDU too short: %d bytes", len(data))
	}

	bpdu := &BPDU{
		ProtocolID:   uint16(data[0])<<8 | uint16(data[1]),
		Version:      data[2],
		Type:         data[3],
		Flags:        data[4],
		RootID:       unpackBridgeID(data[5:13]),
		RootPathCost: uint32(data[13])<<24 | uint32(data[14])<<16 | uint32(data[15])<<8 | uint32(data[16]),
		BridgeID:     unpackBridgeID(data[17:25]),
		PortID:       uint16(data[25])<<8 | uint16(data[26]),
		MessageAge:   uint16(data[27])<<8 | uint16(data[28]),
		MaxAge:       uint16(data[29])<<8 | uint16(data[30]),
		HelloTime:    uint16(data[31])<<8 | uint16(data[32]),
		ForwardDelay: uint16(data[33])<<8 | uint16(data[34]),
		SrcMAC:       srcMAC,
	}

	return bpdu, nil
}

func unpackBridgeID(data []byte) uint64 {
	var id uint64
	for i := 0; i < 8 && i < len(data); i++ {
		id = (id << 8) | uint64(data[i])
	}
	return id
}

func BuildSTPTree(bpdus []*BPDU) *STPTree {
	if len(bpdus) == 0 {
		return nil
	}

	tree := &STPTree{
		Bridges: make(map[string]*BridgeInfo),
	}

	var rootBPDU *BPDU
	for _, bpdu := range bpdus {
		bridgeIDStr := fmt.Sprintf("%016X", bpdu.BridgeID)
		if _, exists := tree.Bridges[bridgeIDStr]; !exists {
			tree.Bridges[bridgeIDStr] = &BridgeInfo{
				BridgeID: bridgeIDStr,
				MAC:      bridgeIDToMAC(bpdu.BridgeID),
				Priority: int(bpdu.BridgeID >> 48),
			}
		}
		if bpdu.BridgeID == bpdu.RootID {
			tree.Bridges[bridgeIDStr].IsRoot = true
			rootBPDU = bpdu
		}
	}

	if rootBPDU == nil && len(bpdus) > 0 {
		rootBPDU = bpdus[0]
	}

	tree.RootBridge = fmt.Sprintf("%016X", rootBPDU.RootID)
	tree.RootMAC = bridgeIDToMAC(rootBPDU.RootID)

	for _, bpdu := range bpdus {
		bridgeIDStr := fmt.Sprintf("%016X", bpdu.BridgeID)
		if info, ok := tree.Bridges[bridgeIDStr]; ok {
			if !info.IsRoot {
				info.PathCostToRoot = bpdu.RootPathCost
			}
		}
	}

	return tree
}

func ExportSTPLinks(tree *STPTree) []*model.Link {
	if tree == nil {
		return nil
	}

	var links []*model.Link
	for _, info := range tree.Bridges {
		if info.IsRoot {
			continue
		}
		link := &model.Link{
			SourceID: info.BridgeID,
			TargetID: tree.RootBridge,
			Type:     "STP",
			SrcPort:  info.DesignatedPort,
			DstPort:  info.RootPort,
		}
		links = append(links, link)
	}
	return links
}

func bridgeIDToMAC(id uint64) string {
	bytes := make([]byte, 6)
	for i := 5; i >= 0; i-- {
		bytes[i] = byte(id & 0xFF)
		id >>= 8
	}
	return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X",
		bytes[0], bytes[1], bytes[2], bytes[3], bytes[4], bytes[5])
}
