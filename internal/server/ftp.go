package server

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"netsight/internal/model"
	"os"
	"strings"
	"time"
)

func startFTP(ctx context.Context, config *model.ServerConfig, onStatus func(*model.ServerState)) error {
	port := config.Port
	if port == 0 {
		port = 21
	}

	addr := fmt.Sprintf("%s:%d", config.Interface, port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("FTP listen failed: %w", err)
	}
	defer listener.Close()

	onStatus(&model.ServerState{
		Type:   "ftp",
		Port:   port,
		Status: "running",
	})

	rootDir := config.RootDir
	if rootDir == "" {
		rootDir = "."
	}

	readOnly := config.ReadOnly

	go func() {
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
			go handleFTPConnection(ctx, conn, rootDir, readOnly)
		}
	}()

	<-ctx.Done()
	return nil
}

func handleFTPConnection(ctx context.Context, conn net.Conn, rootDir string, readOnly bool) {
	defer conn.Close()

	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	// Send welcome message
	sendFTP(writer, "220 NetSight FTP Server ready\r\n")
	writer.Flush()

	var currentDir = "/"
	var authenticated bool

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		conn.SetReadDeadline(time.Now().Add(5 * time.Minute))
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		line = strings.TrimSpace(line)
		parts := strings.SplitN(line, " ", 2)
		cmd := strings.ToUpper(parts[0])

		switch cmd {
		case "USER":
			// Accept any user for now (anonymous FTP-like)
			sendFTP(writer, "331 User name okay, need password\r\n")
			authenticated = true

		case "PASS":
			if authenticated {
				sendFTP(writer, "230 User logged in\r\n")
			} else {
				sendFTP(writer, "530 Not logged in\r\n")
			}

		case "PWD":
			sendFTP(writer, fmt.Sprintf("257 \"%s\" is the current directory\r\n", currentDir))

		case "CWD":
			if len(parts) > 1 {
				newDir := strings.Trim(parts[1], "\"")
				sendFTP(writer, fmt.Sprintf("250 Changed to %s\r\n", newDir))
				currentDir = newDir
			} else {
				sendFTP(writer, "501 Syntax error\r\n")
			}

		case "TYPE":
			sendFTP(writer, "200 Type set to I\r\n")

		case "PASV":
			// Passive mode — use the same connection for data
			sendFTP(writer, "227 Entering Passive Mode (127,0,0,1,0,0)\r\n")

		case "LIST", "NLST":
			sendFTP(writer, "150 Opening ASCII mode data connection\r\n")
			// Send directory listing inline
			entries, _ := os.ReadDir(rootDir)
			for _, entry := range entries {
				info, _ := entry.Info()
				if entry.IsDir() {
					fmt.Fprintf(conn, "drwxr-xr-x 1 owner group %12d %s %s\r\n",
						info.Size(), info.ModTime().Format("Jan 02 15:04"), entry.Name())
				} else {
					fmt.Fprintf(conn, "-rw-r--r-- 1 owner group %12d %s %s\r\n",
						info.Size(), info.ModTime().Format("Jan 02 15:04"), entry.Name())
				}
			}
			sendFTP(writer, "226 Transfer complete\r\n")

		case "RETR":
			if readOnly {
				sendFTP(writer, "550 Permission denied (read-only mode)\r\n")
			} else if len(parts) > 1 {
				filename := strings.TrimSpace(parts[1])
				sendFTP(writer, fmt.Sprintf("150 Opening data connection for %s\r\n", filename))
				_ = filename // TODO: send file contents
				sendFTP(writer, "226 Transfer complete\r\n")
			} else {
				sendFTP(writer, "501 Syntax error\r\n")
			}

		case "STOR":
			sendFTP(writer, "550 Permission denied (read-only server)\r\n")

		case "QUIT":
			sendFTP(writer, "221 Goodbye\r\n")
			writer.Flush()
			return

		default:
			sendFTP(writer, fmt.Sprintf("502 Command not implemented: %s\r\n", cmd))
		}
		writer.Flush()
	}
}

func sendFTP(w *bufio.Writer, msg string) {
	w.WriteString(msg)
}
