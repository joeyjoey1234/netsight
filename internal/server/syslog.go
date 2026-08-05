package server

import (
	"context"
	"fmt"
	"net"
	"netsight/internal/model"
	"strings"
	"sync"
	"time"
)

type SyslogMessage struct {
	Timestamp time.Time
	SourceIP  string
	Facility  int
	Severity  int
	Message   string
	Raw       string
}

type SyslogBuffer struct {
	mu       sync.RWMutex
	messages []*SyslogMessage
	maxSize  int
}

var SyslogMessages = &SyslogBuffer{
	messages: make([]*SyslogMessage, 0),
	maxSize:  10000,
}

func (b *SyslogBuffer) Add(msg *SyslogMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.messages = append(b.messages, msg)
	if len(b.messages) > b.maxSize {
		b.messages = b.messages[len(b.messages)-b.maxSize:]
	}
}

func (b *SyslogBuffer) GetRecent(n int) []*SyslogMessage {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if n > len(b.messages) {
		n = len(b.messages)
	}
	start := len(b.messages) - n
	result := make([]*SyslogMessage, n)
	copy(result, b.messages[start:])
	return result
}

func (b *SyslogBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.messages = make([]*SyslogMessage, 0)
}

func startSyslog(ctx context.Context, config *model.ServerConfig, onStatus func(*model.ServerState)) error {
	port := config.Port
	if port == 0 {
		port = 514
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
		return fmt.Errorf("syslog listen failed: %w", err)
	}
	defer conn.Close()

	onStatus(&model.ServerState{
		Type:   "syslog",
		Port:   port,
		Status: "running",
	})

	buf := make([]byte, 65535)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			continue
		}

		msg := string(buf[:n])
		msg = strings.TrimRight(msg, "\x00\n\r")

		facility, severity := parseSyslogPriority(msg)

		syslogMsg := &SyslogMessage{
			Timestamp: time.Now(),
			SourceIP:  remoteAddr.IP.String(),
			Facility:  facility,
			Severity:  severity,
			Message:   msg,
			Raw:       msg,
		}

		SyslogMessages.Add(syslogMsg)
	}
}

func parseSyslogPriority(msg string) (facility, severity int) {
	if len(msg) > 1 && msg[0] == '<' {
		end := strings.IndexByte(msg, '>')
		if end > 0 {
			priStr := msg[1:end]
			var pri int
			fmt.Sscanf(priStr, "%d", &pri)
			facility = pri / 8
			severity = pri % 8
		}
	}
	return
}

var SeverityNames = map[int]string{
	0: "EMERGENCY",
	1: "ALERT",
	2: "CRITICAL",
	3: "ERROR",
	4: "WARNING",
	5: "NOTICE",
	6: "INFO",
	7: "DEBUG",
}

var FacilityNames = map[int]string{
	0:  "kernel",
	1:  "user",
	2:  "mail",
	3:  "daemon",
	4:  "auth",
	5:  "syslog",
	6:  "lpr",
	7:  "news",
	8:  "uucp",
	9:  "cron",
	10: "authpriv",
	11: "ftp",
	16: "local0",
	17: "local1",
	18: "local2",
	19: "local3",
	20: "local4",
	21: "local5",
	22: "local6",
	23: "local7",
}
