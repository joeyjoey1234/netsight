package server

import (
	"context"
	"fmt"
	"net"
	"netsight/internal/model"
	"sync"
	"time"
)

type NetcatMessage struct {
	Timestamp  time.Time
	RemoteAddr string
	Data       []byte
	HexDump    string
	ASCII      string
	Direction  string
}

var NetcatBuffer = &NetcatMessageBuffer{
	messages: make([]*NetcatMessage, 0),
	maxSize:  5000,
}

type NetcatMessageBuffer struct {
	mu       sync.RWMutex
	messages []*NetcatMessage
	maxSize  int
}

func (b *NetcatMessageBuffer) Add(msg *NetcatMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.messages = append(b.messages, msg)
	if len(b.messages) > b.maxSize {
		b.messages = b.messages[len(b.messages)-b.maxSize:]
	}
}

func (b *NetcatMessageBuffer) GetMessages() []*NetcatMessage {
	b.mu.RLock()
	defer b.mu.RUnlock()
	result := make([]*NetcatMessage, len(b.messages))
	copy(result, b.messages)
	return result
}

func (b *NetcatMessageBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.messages = make([]*NetcatMessage, 0)
}

func startNetcat(ctx context.Context, config *model.ServerConfig, onStatus func(*model.ServerState)) error {
	port := config.Port
	if port == 0 {
		port = 4444
	}

	addr := fmt.Sprintf("%s:%d", config.Interface, port)

	go func() {
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			onStatus(&model.ServerState{
				Type:   "netcat",
				Port:   port,
				Status: "error",
				Error:  fmt.Sprintf("TCP listen failed: %v", err),
			})
			return
		}
		defer listener.Close()

		onStatus(&model.ServerState{
			Type:   "netcat",
			Port:   port,
			Status: "running",
		})

		go func() {
			<-ctx.Done()
			listener.Close()
		}()

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			listener.(*net.TCPListener).SetDeadline(time.Now().Add(1 * time.Second))
			conn, err := listener.Accept()
			if err != nil {
				continue
			}

			go handleNetcatTCP(ctx, conn)
		}
	}()

	<-ctx.Done()
	return nil
}

func handleNetcatTCP(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	buf := make([]byte, 4096)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, err := conn.Read(buf)
		if err != nil {
			return
		}

		msg := &NetcatMessage{
			Timestamp:  time.Now(),
			RemoteAddr: remoteAddr,
			Data:       make([]byte, n),
			Direction:  "in",
		}
		copy(msg.Data, buf[:n])
		msg.HexDump = hexDump(buf[:n])
		msg.ASCII = toASCII(buf[:n])

		NetcatBuffer.Add(msg)

		conn.Write(buf[:n])
	}
}

func hexDump(data []byte) string {
	var result string
	for i, b := range data {
		if i > 0 && i%16 == 0 {
			result += "\n"
		} else if i > 0 {
			result += " "
		}
		result += fmt.Sprintf("%02X", b)
	}
	return result
}

func toASCII(data []byte) string {
	var result string
	for _, b := range data {
		if b >= 32 && b < 127 {
			result += string(b)
		} else {
			result += "."
		}
	}
	return result
}
