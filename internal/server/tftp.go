package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"netsight/internal/model"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TFTP opcodes
const (
	tftpRRQ   = 1 // Read Request
	tftpWRQ   = 2 // Write Request
	tftpDATA  = 3 // Data
	tftpACK   = 4 // Acknowledgment
	tftpERROR = 5 // Error
)

func startTFTP(ctx context.Context, config *model.ServerConfig, onStatus func(*model.ServerState)) error {
	port := config.Port
	if port == 0 {
		port = 69
	}

	addr := fmt.Sprintf("%s:%d", config.Interface, port)
	_ = addr
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP(config.Interface),
		Port: port,
	})
	if err != nil {
		return fmt.Errorf("TFTP listen failed: %w", err)
	}
	defer conn.Close()

	onStatus(&model.ServerState{
		Type:   "tftp",
		Port:   port,
		Status: "running",
	})

	rootDir := config.RootDir
	if rootDir == "" {
		rootDir = "."
	}

	// Process TFTP requests
	buf := make([]byte, 516) // max TFTP data block: 512 + 4 header bytes

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

		if n < 2 {
			sendTFTPError(conn, remoteAddr, 4, "Malformed request")
			continue
		}
		opcode := uint16(buf[0])<<8 | uint16(buf[1])

		switch opcode {
		case tftpRRQ:
			go handleTFTPRead(ctx, conn, remoteAddr, buf[:n], rootDir)
		case tftpWRQ:
			go handleTFTPWrite(ctx, conn, remoteAddr, buf[:n], rootDir)
		}
	}
}

func handleTFTPRead(ctx context.Context, conn *net.UDPConn, remoteAddr *net.UDPAddr, packet []byte, rootDir string) {
	if len(packet) < 3 {
		sendTFTPError(conn, remoteAddr, 4, "Malformed request")
		return
	}
	// Parse filename from RRQ packet
	filename := ""
	for i := 2; i < len(packet); i++ {
		if packet[i] == 0 {
			filename = string(packet[2:i])
			break
		}
	}

	if filename == "" {
		sendTFTPError(conn, remoteAddr, 0, "No filename specified")
		return
	}

	// Prevent directory traversal
	cleanPath, ok := safeTFTPPath(rootDir, filename)
	if !ok {
		sendTFTPError(conn, remoteAddr, 2, "Access denied")
		return
	}

	file, err := os.Open(cleanPath)
	if err != nil {
		sendTFTPError(conn, remoteAddr, 1, "File not found")
		return
	}
	defer file.Close()

	block := uint16(1)
	buf := make([]byte, 512)

	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			// Send last block
			dataPacket := make([]byte, 4+n)
			dataPacket[0] = 0x00
			dataPacket[1] = tftpDATA
			dataPacket[2] = byte(block >> 8)
			dataPacket[3] = byte(block)
			copy(dataPacket[4:], buf[:n])
			conn.WriteToUDP(dataPacket, remoteAddr)
			// Wait for final ACK
			return
		}
		if err != nil {
			sendTFTPError(conn, remoteAddr, 0, "Read error")
			return
		}

		dataPacket := make([]byte, 4+n)
		dataPacket[0] = 0x00
		dataPacket[1] = tftpDATA
		dataPacket[2] = byte(block >> 8)
		dataPacket[3] = byte(block)
		copy(dataPacket[4:], buf[:n])

		// Send DATA packet
		conn.WriteToUDP(dataPacket, remoteAddr)

		// Wait for ACK
		ackBuf := make([]byte, 4)
		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		n, _, err = conn.ReadFromUDP(ackBuf)
		if err != nil || n < 4 || ackBuf[1] != tftpACK {
			continue // Retry or timeout
		}

		ackBlock := uint16(ackBuf[2])<<8 | uint16(ackBuf[3])
		if ackBlock == block {
			block++
		}
	}
}

func handleTFTPWrite(ctx context.Context, conn *net.UDPConn, remoteAddr *net.UDPAddr, packet []byte, rootDir string) {
	// TFTP writes are disabled until an explicit authenticated write policy exists.
	sendTFTPError(conn, remoteAddr, 2, "Writes are disabled")
	return
	/*
		// Parse filename from WRQ packet
		filename := ""
		for i := 2; i < len(packet); i++ {
			if packet[i] == 0 {
				filename = string(packet[2:i])
				break
			}
		}

		if filename == "" {
			sendTFTPError(conn, remoteAddr, 0, "No filename specified")
			return
		}

		// Prevent directory traversal
		cleanPath := filepath.Clean(filepath.Join(rootDir, filename))
		if !filepath.HasPrefix(cleanPath, filepath.Clean(rootDir)) {
			sendTFTPError(conn, remoteAddr, 2, "Access denied")
			return
		}

		// Create directory if needed
		os.MkdirAll(filepath.Dir(cleanPath), 0755)

		file, err := os.Create(cleanPath)
		if err != nil {
			sendTFTPError(conn, remoteAddr, 0, "Cannot create file")
			return
		}
		defer file.Close()

		// Send ACK 0 to start transfer
		ackPacket := []byte{0x00, tftpACK, 0x00, 0x00}
		conn.WriteToUDP(ackPacket, remoteAddr)

		dataBuf := make([]byte, 516)
		for {
			conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			n, _, err := conn.ReadFromUDP(dataBuf)
			if err != nil {
				return // Timeout — transfer complete or failed
			}

			if n < 4 || dataBuf[1] != tftpDATA {
				continue
			}

			block := uint16(dataBuf[2])<<8 | uint16(dataBuf[3])
			file.Write(dataBuf[4:n])

			// Send ACK
			ackPacket := []byte{0x00, tftpACK, byte(block >> 8), byte(block)}
			conn.WriteToUDP(ackPacket, remoteAddr)

			// If data block < 512, transfer complete
			if n < 516 {
				return
			}
		}
	*/
}

func safeTFTPPath(rootDir, name string) (string, bool) {
	if name == "" || filepath.IsAbs(name) {
		return "", false
	}
	root, err := filepath.Abs(rootDir)
	if err != nil {
		return "", false
	}
	path, err := filepath.Abs(filepath.Join(root, filepath.FromSlash(name)))
	if err != nil {
		return "", false
	}
	rel, err := filepath.Rel(root, path)
	return path, err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

func sendTFTPError(conn *net.UDPConn, addr *net.UDPAddr, code uint16, msg string) {
	packet := []byte{0x00, tftpERROR, byte(code >> 8), byte(code)}
	packet = append(packet, []byte(msg)...)
	packet = append(packet, 0)
	conn.WriteToUDP(packet, addr)
}
