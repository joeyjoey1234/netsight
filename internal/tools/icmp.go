package tools

import (
	"context"
	"fmt"
	"net"
	"netsight/internal/model"
	"time"
)

func Ping(ctx context.Context, target string, count int, size int, ttl int, onResult func(*model.PingResult)) error {
	if count <= 0 {
		count = 4
	}
	if size < 16 {
		size = 56
	}

	ipAddr, err := net.ResolveIPAddr("ip4", target)
	if err != nil {
		addrs, lookupErr := net.LookupHost(target)
		if lookupErr != nil || len(addrs) == 0 {
			return fmt.Errorf("cannot resolve %s: %w", target, err)
		}
		ipAddr = &net.IPAddr{IP: net.ParseIP(addrs[0])}
	}

	conn, err := net.DialIP("ip4:icmp", nil, ipAddr)
	if err != nil {
		return fmt.Errorf("ICMP dial failed: %w", err)
	}
	defer conn.Close()

	payload := make([]byte, size)
	for i := range payload {
		payload[i] = byte(i % 256)
	}

	for seq := 0; seq < count || count == 0; seq++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		msg := buildICMPEcho(uint16(seq), payload, ttl)
		start := time.Now()

		conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
		if _, err := conn.Write(msg); err != nil {
			onResult(&model.PingResult{
				Target:   target,
				Sequence: seq,
				TimedOut: true,
			})
			continue
		}

		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		reply := make([]byte, 1500)
		n, err := conn.Read(reply)
		elapsed := time.Since(start)

		result := &model.PingResult{
			Target:   target,
			Sequence: seq,
		}

		if err != nil {
			result.TimedOut = true
		} else if n >= 28 {
			result.TTL = int(reply[8])
			result.LatencyMs = float64(elapsed.Microseconds()) / 1000.0
			result.Bytes = n
			result.TimedOut = false
		} else {
			result.TimedOut = true
		}

		onResult(result)

		if count > 0 && seq < count-1 && !result.TimedOut {
			time.Sleep(1 * time.Second)
		}
	}

	return nil
}

func buildICMPEcho(seq uint16, payload []byte, ttl int) []byte {
	totalLen := 8 + len(payload)
	msg := make([]byte, totalLen)
	msg[0] = 8
	msg[1] = 0
	msg[4] = 0xAB
	msg[5] = 0xCD
	msg[6] = byte(seq >> 8)
	msg[7] = byte(seq)
	copy(msg[8:], payload)

	cs := checksum(msg)
	msg[2] = byte(cs >> 8)
	msg[3] = byte(cs)

	return msg
}

func checksum(data []byte) uint16 {
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
