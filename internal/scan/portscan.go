package scan

import (
	"context"
	"fmt"
	"net"
	"netsight/internal/model"
	"strings"
	"sync"
	"time"
)

func DefaultPorts() []int {
	return []int{21, 22, 23, 25, 53, 80, 139, 443, 445, 3389, 8080, 8443}
}

func TCPSynScan(ctx context.Context, target string, ports []int) ([]*model.Port, error) {
	var results []*model.Port
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 100)

	for _, port := range ports {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			portState := tcpConnect(target, p)
			mu.Lock()
			results = append(results, &model.Port{
				DeviceID: "",
				Number:   p,
				Protocol: "tcp",
				State:    portState,
			})
			mu.Unlock()
		}(port)
	}
	wg.Wait()
	return results, nil
}

func tcpConnect(target string, port int) string {
	addr := fmt.Sprintf("%s:%d", target, port)
	conn, err := net.DialTimeout("tcp", addr, 800*time.Millisecond)
	if err != nil {
		return "closed"
	}
	conn.Close()
	return "open"
}

func GrabBanner(ctx context.Context, target string, port int) string {
	_ = ctx
	addr := fmt.Sprintf("%s:%d", target, port)
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return ""
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(buf[:n]))
}

func isBannerable(port int) bool {
	bannerPorts := map[int]bool{21: true, 22: true, 23: true, 25: true, 80: true, 110: true, 143: true, 443: true, 993: true, 995: true, 3306: true, 5432: true, 6379: true, 8080: true, 27017: true}
	return bannerPorts[port]
}

func ParseBanner(port int, banner string) (service, version string) {
	lower := strings.ToLower(banner)

	switch port {
	case 22:
		if strings.Contains(lower, "ssh") {
			service = "ssh"
			version = strings.TrimPrefix(lower, "ssh-")
			if idx := strings.Index(version, " "); idx > 0 {
				version = version[:idx]
			}
		}
	case 80, 8080:
		if strings.Contains(lower, "apache") {
			service = "http"
			version = "Apache"
		} else if strings.Contains(lower, "nginx") {
			service = "http"
			version = "nginx"
		} else if strings.Contains(lower, "iis") || strings.Contains(lower, "microsoft") {
			service = "http"
			version = "IIS"
		}
	case 21:
		service = "ftp"
		version = strings.TrimSpace(banner)
	case 3306:
		service = "mysql"
		version = strings.TrimSpace(banner)
	case 5432:
		service = "postgresql"
	}
	return
}
