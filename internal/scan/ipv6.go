package scan

import (
	"context"
	"fmt"
	"net"
	"time"

	"netsight/internal/model"
)

type IPv6Neighbor struct {
	IP        string    `json:"ip"`
	MAC       string    `json:"mac"`
	Interface string    `json:"interface"`
	State     string    `json:"state"`
	LastSeen  time.Time `json:"lastSeen"`
}

func IPv6NDScan(ctx context.Context, iface string) ([]*IPv6Neighbor, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to list interfaces: %w", err)
	}

	var targetIface *net.Interface
	if iface == "" {
		for _, i := range interfaces {
			if i.Flags&net.FlagLoopback != 0 || i.Flags&net.FlagUp == 0 {
				continue
			}
			addrs, _ := i.Addrs()
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() == nil {
					targetIface = &i
					break
				}
			}
			if targetIface != nil {
				break
			}
		}
	} else {
		for _, i := range interfaces {
			if i.Name == iface {
				targetIface = &i
				break
			}
		}
	}

	if targetIface == nil {
		return nil, fmt.Errorf("no suitable IPv6 interface found")
	}

	addrs, _ := targetIface.Addrs()
	var linkLocal string
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if ipnet.IP.To4() == nil && ipnet.IP.IsLinkLocalUnicast() {
				linkLocal = ipnet.IP.String()
				break
			}
		}
	}

	neighbors := make(map[string]*IPv6Neighbor)

	neighs, _ := getOSNeighborCache()
	for _, n := range neighs {
		neighbors[n.IP] = n
	}

	_ = linkLocal

	var result []*IPv6Neighbor
	for _, n := range neighbors {
		result = append(result, n)
	}

	return result, nil
}

func getOSNeighborCache() ([]*IPv6Neighbor, error) {
	return nil, nil
}

func IPv6Ping(ctx context.Context, target string, count int) ([]*model.PingResult, error) {
	return nil, fmt.Errorf("IPv6 ping: use dual-stack ping tool for IPv6 targets")
}

func IPv6Traceroute(ctx context.Context, target string, maxHops int) ([]*model.Hop, error) {
	return nil, fmt.Errorf("IPv6 traceroute: in production uses raw sockets with hop limit")
}
