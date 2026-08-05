package survey

import (
	"fmt"
	"time"
)

type Preset struct {
	Name     string
	Duration time.Duration
	Tools    []string
}

var (
	Quick = Preset{
		Name:     "quick",
		Duration: 3 * time.Minute,
		Tools: []string{
			"control_plane",
			"arp_scan",
			"cdp_lldp",
			"broadcast_detection",
			"arp_spoof_detection",
			"rogue_dhcp_detection",
			"bandwidth_estimate",
		},
	}

	Short = Preset{
		Name:     "short",
		Duration: 10 * time.Minute,
		Tools: []string{
			"control_plane",
			"arp_scan",
			"cdp_lldp",
			"broadcast_detection",
			"arp_spoof_detection",
			"rogue_dhcp_detection",
			"bandwidth_estimate",
		},
	}

	Long = Preset{
		Name:     "long",
		Duration: 0,
		Tools: []string{
			"control_plane",
			"arp_scan",
			"cdp_lldp",
			"broadcast_detection",
			"arp_spoof_detection",
			"rogue_dhcp_detection",
			"bandwidth_estimate",
			"port_scan",
			"banner_grab",
			"http_title",
			"passive_os_fingerprint",
			"open_shares",
		},
	}
)

func GetPreset(name string) (*Preset, error) {
	switch name {
	case "quick":
		return &Quick, nil
	case "short":
		return &Short, nil
	case "long":
		return &Long, nil
	default:
		return nil, fmt.Errorf("unknown preset: %s", name)
	}
}
