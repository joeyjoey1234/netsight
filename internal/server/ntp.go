package server

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"netsight/internal/model"
	"time"
)

const ntpEpochOffset = 2208988800

func startNTP(ctx context.Context, config *model.ServerConfig, onStatus func(*model.ServerState)) error {
	port := config.Port
	if port == 0 {
		port = 123
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
		return fmt.Errorf("NTP listen failed: %w", err)
	}
	defer conn.Close()

	onStatus(&model.ServerState{
		Type:   "ntp",
		Port:   port,
		Status: "running",
	})

	buf := make([]byte, 48)

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

		if n < 48 {
			continue
		}

		response := buildNTPResponse(buf)
		conn.WriteToUDP(response, remoteAddr)
	}
}

func buildNTPResponse(request []byte) []byte {
	now := time.Now()
	ntpTime := toNTPTime(now)

	response := make([]byte, 48)

	response[0] = 0x24
	response[1] = 1
	response[2] = 0
	response[3] = 0xEC

	response[12] = 'L'
	response[13] = 'O'
	response[14] = 'C'
	response[15] = 'L'

	binary.BigEndian.PutUint32(response[16:20], ntpTime.Seconds)
	binary.BigEndian.PutUint32(response[20:24], ntpTime.Fraction)

	copy(response[24:32], request[40:48])

	binary.BigEndian.PutUint32(response[32:36], ntpTime.Seconds)
	binary.BigEndian.PutUint32(response[36:40], ntpTime.Fraction)

	binary.BigEndian.PutUint32(response[40:44], ntpTime.Seconds)
	binary.BigEndian.PutUint32(response[44:48], ntpTime.Fraction)

	return response
}

type ntpTimestamp struct {
	Seconds  uint32
	Fraction uint32
}

func toNTPTime(t time.Time) ntpTimestamp {
	secs := t.Unix() + ntpEpochOffset
	frac := uint32(float64(t.Nanosecond()) / 1e9 * float64(1<<32))

	return ntpTimestamp{
		Seconds:  uint32(secs),
		Fraction: frac,
	}
}
