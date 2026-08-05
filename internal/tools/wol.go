package tools

import (
	"fmt"
	"net"
	"strings"
)

func WakeOnLAN(macAddr string) error {
	mac, err := parseMAC(macAddr)
	if err != nil {
		return fmt.Errorf("invalid MAC: %w", err)
	}

	packet := make([]byte, 102)
	for i := 0; i < 6; i++ {
		packet[i] = 0xFF
	}
	for i := 0; i < 16; i++ {
		copy(packet[6+i*6:], mac)
	}

	conn, err := net.Dial("udp", "255.255.255.255:9")
	if err != nil {
		conn, err = net.Dial("udp", "255.255.255.255:7")
		if err != nil {
			return fmt.Errorf("WoL broadcast failed: %w", err)
		}
	}
	defer conn.Close()

	_, err = conn.Write(packet)
	return err
}

func parseMAC(mac string) (net.HardwareAddr, error) {
	mac = strings.ReplaceAll(strings.ReplaceAll(mac, ":", ""), "-", "")
	if len(mac) != 12 {
		return nil, fmt.Errorf("MAC must be 12 hex chars")
	}

	hw := make(net.HardwareAddr, 6)
	_, err := fmt.Sscanf(mac, "%02x%02x%02x%02x%02x%02x", &hw[0], &hw[1], &hw[2], &hw[3], &hw[4], &hw[5])
	return hw, err
}
