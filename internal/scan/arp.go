package scan

import (
	"context"
	"net"
	"sync"
	"time"
)

func ARPScan(ctx context.Context, subnet string) (map[string]string, error) {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, err
	}

	table := make(map[string]string)
	var mu sync.Mutex

	var ips []net.IP
	for ip := ipNet.IP.Mask(ipNet.Mask); ipNet.Contains(ip); incrementIP(ip) {
		if isNetworkOrBroadcast(ip, ipNet) {
			continue
		}
		ips = append(ips, append(net.IP{}, ip...))
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 50)

	for _, ip := range ips {
		select {
		case <-ctx.Done():
			return table, ctx.Err()
		default:
		}
		wg.Add(1)
		go func(target net.IP) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			mac := arpLookup(target)
			if mac != "" {
				mu.Lock()
				table[target.String()] = mac
				mu.Unlock()
			}
		}(ip)
	}
	wg.Wait()
	return table, nil
}

func arpLookup(ip net.IP) string {
	conn, err := net.DialTimeout("tcp", ip.String()+":7", 500*time.Millisecond)
	if err != nil {
		return ""
	}
	defer conn.Close()

	interfaces, _ := net.Interfaces()
	for _, iface := range interfaces {
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipnet.Contains(ip) {
					return iface.HardwareAddr.String()
				}
			}
		}
	}
	return ""
}
