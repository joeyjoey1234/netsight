package recognition

import (
	"fmt"
	"netsight/internal/model"
)

type VLANMapEntry struct {
	ID          int             `json:"id"`
	Name        string          `json:"name"`
	Devices     []*model.Device `json:"devices"`
	TrunkPorts  []string        `json:"trunkPorts"`
	AccessPorts []string        `json:"accessPorts"`
	NativeVLAN  int             `json:"nativeVlan,omitempty"`
}

func BuildVLANMap(devices []*model.Device, links []*model.Link) map[int]*VLANMapEntry {
	vlanMap := make(map[int]*VLANMapEntry)

	vlanSet := make(map[int]bool)
	for _, d := range devices {
		for _, vlan := range d.VLANs {
			vlanSet[vlan] = true
		}
	}
	for _, l := range links {
		if l.VLAN > 0 {
			vlanSet[l.VLAN] = true
		}
	}

	for vlan := range vlanSet {
		vlanMap[vlan] = &VLANMapEntry{
			ID:          vlan,
			Name:        defaultVLANName(vlan),
			Devices:     make([]*model.Device, 0),
			TrunkPorts:  make([]string, 0),
			AccessPorts: make([]string, 0),
		}
	}

	for _, d := range devices {
		for _, vlan := range d.VLANs {
			if entry, ok := vlanMap[vlan]; ok {
				entry.Devices = append(entry.Devices, d)
			}
		}
	}

	for _, l := range links {
		if entry, ok := vlanMap[l.VLAN]; ok {
			entry.AccessPorts = append(entry.AccessPorts, fmt.Sprintf("%s (%s)", l.SrcPort, l.Type))
		}
	}

	return vlanMap
}

func defaultVLANName(vlanID int) string {
	if vlanID == 1 {
		return "Default"
	}
	return fmt.Sprintf("VLAN %d", vlanID)
}
