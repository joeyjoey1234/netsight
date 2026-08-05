package server

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"netsight/internal/model"
	"strings"
	"time"
)

func startDNS(ctx context.Context, config *model.ServerConfig, onStatus func(*model.ServerState)) error {
	port := config.Port
	if port == 0 {
		port = 53
	}

	forwarder := config.DNS
	if forwarder == "" {
		forwarder = "8.8.8.8:53"
	}

	addr := &net.UDPAddr{
		IP:   net.ParseIP(config.Interface),
		Port: port,
	}
	if addr.IP == nil {
		addr.IP = net.IPv4zero
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("DNS listen failed: %w", err)
	}
	defer conn.Close()

	onStatus(&model.ServerState{
		Type:   "dns",
		Port:   port,
		Status: "running",
	})

	buf := make([]byte, 512)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		go func() {
			response, err := forwardDNSQuery(buf[:n], forwarder)
			if err != nil {
				return
			}
			conn.WriteToUDP(response, remoteAddr)
		}()
	}
}

func forwardDNSQuery(query []byte, forwarder string) ([]byte, error) {
	conn, err := net.DialTimeout("udp", forwarder, 3*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	if _, err := conn.Write(query); err != nil {
		return nil, err
	}

	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	response := make([]byte, 512)
	n, err := conn.Read(response)
	if err != nil {
		return nil, err
	}

	return response[:n], nil
}

func extractDNSName(data []byte, offset int) (string, int) {
	var parts []string
	pos := offset
	hasPointer := false

	for {
		if pos >= len(data) {
			break
		}
		length := int(data[pos])
		if length == 0 {
			pos++
			break
		}
		if length&0xC0 == 0xC0 {
			if !hasPointer && pos+1 < len(data) {
				ptr := int(binary.BigEndian.Uint16(data[pos:pos+2]) & 0x3FFF)
				name, _ := extractDNSName(data, ptr)
				parts = append(parts, name)
				pos += 2
				hasPointer = true
				break
			}
			pos += 2
			break
		}
		pos++
		if pos+length > len(data) {
			break
		}
		parts = append(parts, string(data[pos:pos+length]))
		pos += length
	}

	return strings.Join(parts, "."), pos
}
