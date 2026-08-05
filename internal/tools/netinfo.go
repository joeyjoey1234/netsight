package tools

import (
	"fmt"
	"net"
	"netsight/internal/model"
	"os"
	"strings"
)

func GetNetworkInfo() (*model.InterfaceInfo, error) {
	info := &model.InterfaceInfo{
		Name: "unknown",
		MTU:  1500,
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to list interfaces: %w", err)
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipnet.IP.To4() != nil {
					info.Name = iface.Name
					info.IPs = append(info.IPs, ipnet.IP.String())
					info.MAC = iface.HardwareAddr.String()
					info.MTU = iface.MTU
				}
			}
		}

		if len(info.IPs) > 0 {
			break
		}
	}

	info.Gateway = detectGateway()
	info.DNS = detectDNS()
	info.PublicIP = detectPublicIP()

	return info, nil
}

// GetAllInterfaces returns all network interfaces including loopback and down ones
func GetAllInterfaces() []*model.InterfaceInfo {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	var result []*model.InterfaceInfo
	for _, iface := range interfaces {
		info := &model.InterfaceInfo{
			Name: iface.Name,
			MAC:  iface.HardwareAddr.String(),
			MTU:  iface.MTU,
			IPs:  []string{},
		}

		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				info.IPs = append(info.IPs, ipnet.IP.String())
			}
		}

		result = append(result, info)
	}

	return result
}

// GetAvailableSubnets returns CIDR strings for all active interfaces
func GetAvailableSubnets() ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var subnets []string
	seen := make(map[string]bool)

	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipnet.IP.To4() != nil {
					ones, _ := ipnet.Mask.Size()
					cidr := fmt.Sprintf("%s/%d", ipnet.IP.Mask(ipnet.Mask).String(), ones)
					if !seen[cidr] {
						seen[cidr] = true
						subnets = append(subnets, cidr)
					}
				}
			}
		}
	}

	if len(subnets) == 0 {
		subnets = append(subnets, "192.168.1.0/24")
	}

	return subnets, nil
}

func detectGateway() string {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	if localAddr != nil {
		ip := localAddr.IP.To4()
		if ip != nil {
			ip[3] = 1
			return ip.String()
		}
	}
	return ""
}

func detectDNS() []string {
	host, _ := os.Hostname()
	_ = host
	return []string{"8.8.8.8", "1.1.1.1"}
}

func detectPublicIP() string {
	return "detecting..."
}

// GetInterfaceDisplayName returns a human-readable name for an interface
func GetInterfaceDisplayName(info *model.InterfaceInfo) string {
	ipStr := ""
	if len(info.IPs) > 0 {
		for _, ip := range info.IPs {
			if strings.Contains(ip, ":") {
				continue
			}
			ipStr = ip
			break
		}
	}
	macShort := ""
	if len(info.MAC) >= 8 {
		macShort = info.MAC[:8]
	}
	_ = macShort
	if ipStr != "" {
		return fmt.Sprintf("%s (%s)", info.Name, ipStr)
	}
	return info.Name
}
