package scan

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

func PingSweep(ctx context.Context, subnet string) ([]string, error) {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, fmt.Errorf("invalid subnet: %w", err)
	}

	var ips []string
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 64)
	for ip := ipNet.IP.Mask(ipNet.Mask); ipNet.Contains(ip); incrementIP(ip) {
		if isNetworkOrBroadcast(ip, ipNet) {
			continue
		}
		select {
		case <-ctx.Done():
			return ips, ctx.Err()
		default:
			ipStr := ip.String()
			wg.Add(1)
			go func(target string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				if pingHost(target) || tcpProbeHost(target) {
					mu.Lock()
					ips = append(ips, target)
					mu.Unlock()
				}
			}(ipStr)
		}
	}
	wg.Wait()
	return ips, nil
}

func pingHost(ip string) bool {
	conn, err := net.DialTimeout("ip4:icmp", ip, 500*time.Millisecond)
	if err != nil {
		return false
	}
	defer conn.Close()

	msg := buildICMPEchoRequest()
	conn.SetDeadline(time.Now().Add(500 * time.Millisecond))
	_, err = conn.Write(msg)
	if err != nil {
		return false
	}

	reply := make([]byte, 1500)
	_, err = conn.Read(reply)
	return err == nil
}

func tcpProbeHost(ip string) bool {
	for _, port := range []int{80, 443, 22, 445, 3389, 8080} {
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, fmt.Sprint(port)), 250*time.Millisecond)
		if err == nil {
			conn.Close()
			return true
		}
	}
	return false
}

func buildICMPEchoRequest() []byte {
	msg := make([]byte, 8)
	msg[0] = 8
	msg[1] = 0
	msg[4] = 0x00
	msg[5] = 0x01
	msg[6] = 0x00
	msg[7] = 0x01

	checksum := icmpChecksum(msg)
	msg[2] = byte(checksum >> 8)
	msg[3] = byte(checksum)

	return msg
}

func icmpChecksum(data []byte) uint16 {
	var sum uint32
	for i := 0; i < len(data)-1; i += 2 {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
	}
	if len(data)%2 == 1 {
		sum += uint32(data[len(data)-1]) << 8
	}
	sum = (sum >> 16) + (sum & 0xffff)
	sum += sum >> 16
	return uint16(^sum)
}

func isNetworkOrBroadcast(ip net.IP, network *net.IPNet) bool {
	if ip.Equal(network.IP) {
		return true
	}
	broadcast := make(net.IP, len(network.IP))
	for i := range broadcast {
		broadcast[i] = network.IP[i] | ^network.Mask[i]
	}
	return ip.Equal(broadcast)
}

func incrementIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] > 0 {
			break
		}
	}
}
