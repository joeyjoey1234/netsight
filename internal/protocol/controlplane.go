package protocol

import (
	"netsight/internal/model"
	"sync"
)

type ProtocolEvent struct {
	Type    string
	Subtype string
	SrcMAC  string
	SrcIP   string
	Details map[string]string
}

type DeviceRoleDetection struct {
	Roles map[string]string
	mu    sync.RWMutex
}

func NewDeviceRoleDetection() *DeviceRoleDetection {
	return &DeviceRoleDetection{
		Roles: make(map[string]string),
	}
}

func (d *DeviceRoleDetection) ProcessEvent(event *ProtocolEvent) {
	d.mu.Lock()
	defer d.mu.Unlock()

	key := event.SrcMAC
	if key == "" {
		key = event.SrcIP
	}
	if key == "" {
		return
	}

	current := d.Roles[key]

	switch event.Type {
	case "STP":
		if current != "router" && current != "L3 switch" {
			d.Roles[key] = "switch"
		}
	case "CDP", "LLDP":
		if current == "" || current == "unknown" {
			d.Roles[key] = "switch"
		}
	case "OSPF", "BGP":
		d.Roles[key] = "router"
	case "HSRP", "VRRP":
		if current == "switch" {
			d.Roles[key] = "L3 switch"
		}
	case "DHCP":
	}
}

func (d *DeviceRoleDetection) GetRole(mac string) string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if role, ok := d.Roles[mac]; ok {
		return role
	}
	return "unknown"
}

func (d *DeviceRoleDetection) ApplyToDevices(devices []*model.Device) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	for _, device := range devices {
		if role, ok := d.Roles[device.MAC]; ok && role != "" {
			device.Role = role
		}
	}
}
