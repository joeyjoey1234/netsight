package tools

import (
	"context"
	"fmt"
	"net"
	"netsight/internal/model"
	"time"
)

func Traceroute(ctx context.Context, target string, mode string, maxHops int, onHop func(*model.Hop)) error {
	if maxHops <= 0 {
		maxHops = 30
	}

	addrs, err := net.LookupHost(target)
	if err != nil {
		return fmt.Errorf("cannot resolve %s: %w", target, err)
	}
	targetIP := net.ParseIP(addrs[0])

	for ttl := 1; ttl <= maxHops; ttl++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		hop := &model.Hop{
			Number: ttl,
		}

		var hopIP string
		var elapsed time.Duration

		switch mode {
		case "tcp":
			hopIP, elapsed, err = tcpProbe(targetIP, ttl)
		case "udp":
			hopIP, elapsed, err = udpProbe(targetIP, ttl)
		default:
			hopIP, elapsed, err = icmpProbe(targetIP, ttl)
		}

		hop.LatencyMs = float64(elapsed.Microseconds()) / 1000.0

		if err != nil {
			hop.IP = "*"
			hop.Hostname = "*"
			onHop(hop)
			continue
		}

		hop.IP = hopIP
		hop.AllIPs = []string{hopIP}

		names, _ := net.LookupAddr(hopIP)
		if len(names) > 0 {
			hop.Hostname = names[0]
		}

		onHop(hop)

		if hopIP == targetIP.String() {
			return nil
		}

		time.Sleep(200 * time.Millisecond)
	}

	return nil
}

func icmpProbe(target net.IP, ttl int) (string, time.Duration, error) {
	conn, err := net.DialIP("ip4:icmp", nil, &net.IPAddr{IP: target})
	if err != nil {
		return "", 0, err
	}
	defer conn.Close()

	start := time.Now()
	msg := buildICMPEcho(uint16(ttl), []byte{0x00}, ttl)
	conn.Write(msg)

	reply := make([]byte, 1500)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := conn.Read(reply)
	elapsed := time.Since(start)

	if err != nil {
		return "", elapsed, err
	}

	if n >= 20 {
		srcIP := net.IP(reply[12:16])
		return srcIP.String(), elapsed, nil
	}

	return "", elapsed, fmt.Errorf("short reply")
}

func tcpProbe(target net.IP, ttl int) (string, time.Duration, error) {
	addr := fmt.Sprintf("%s:80", target.String())
	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	elapsed := time.Since(start)

	if err != nil {
		return "", elapsed, err
	}
	conn.Close()

	return target.String(), elapsed, nil
}

func udpProbe(target net.IP, ttl int) (string, time.Duration, error) {
	addr := fmt.Sprintf("%s:33434", target.String())
	conn, err := net.DialTimeout("udp", addr, 2*time.Second)
	if err != nil {
		return "", 0, err
	}
	defer conn.Close()

	conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
	start := time.Now()
	conn.Write([]byte{0x00})
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 512)
	_, err = conn.Read(buf)
	elapsed := time.Since(start)

	return target.String(), elapsed, nil
}
