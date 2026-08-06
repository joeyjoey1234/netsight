package scan

import (
	"context"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func ARPScan(ctx context.Context, subnet string) (map[string]string, error) {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, err
	}

	cacheEntries := readARPCache()
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

			mac := ""
			if cached, ok := cacheEntries[target.String()]; ok {
				mac = cached
			}
			if mac != "" {
				mu.Lock()
				table[target.String()] = mac
				mu.Unlock()
			}
		}(ip)
	}
	wg.Wait()

	if len(table) == 0 {
		cacheEntries := readARPCache()
		for _, ipStr := range ips {
			select {
			case <-ctx.Done():
				return table, ctx.Err()
			default:
			}
			if mac, ok := cacheEntries[ipStr.String()]; ok {
				mu.Lock()
				table[ipStr.String()] = mac
				mu.Unlock()
			}
		}
	}

	return table, nil
}

func arpLookup(ip net.IP) string {
	dstAddr := &net.UDPAddr{
		IP:   ip,
		Port: 7,
	}

	conn, err := net.DialTimeout("udp", dstAddr.String(), 300*time.Millisecond)
	if err != nil {
		return ""
	}
	conn.Close()

	return ""
}

func readARPCache() map[string]string {
	result := make(map[string]string)

	out, err := exec.Command("arp", "-a").Output()
	if err != nil {
		return result
	}

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		for i, field := range fields {
			ipStr := strings.Trim(field, "()")
			if net.ParseIP(ipStr) != nil && i+1 < len(fields) {
				mac := strings.TrimSpace(fields[i+1])
				mac = strings.ReplaceAll(mac, "-", ":")
				if isValidMAC(mac) {
					result[ipStr] = mac
				}
			}
		}
	}

	return result
}

func isValidMAC(mac string) bool {
	_, err := net.ParseMAC(mac)
	return err == nil
}

func arpLookupByPing(ip net.IP) {
	_, _ = exec.Command("ping", "-c", "1", "-W", "1", ip.String()).Output()
}
