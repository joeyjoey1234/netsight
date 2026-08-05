package tools

import (
	"fmt"
	"net"
	"netsight/internal/model"
	"os"

	"golang.org/x/sys/windows"
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
