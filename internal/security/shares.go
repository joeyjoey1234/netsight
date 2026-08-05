package security

import (
	"fmt"
	"net"
	"netsight/internal/model"
	"time"
)

type ShareScanner struct{}

func NewShareScanner() *ShareScanner {
	return &ShareScanner{}
}

func (s *ShareScanner) ScanHost(ip string) []*model.Finding {
	var findings []*model.Finding

	for _, port := range []int{139, 445} {
		addr := fmt.Sprintf("%s:%d", ip, port)
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err != nil {
			continue
		}
		conn.Close()

		findings = append(findings, &model.Finding{
			ID:             generateID("share"),
			Type:           "open_share",
			Severity:       "medium",
			Title:          fmt.Sprintf("SMB service accessible on %s:%d", ip, port),
			Description:    fmt.Sprintf("SMB (Server Message Block) is accessible on %s port %d. Open shares may expose sensitive data. In production, this enumerates shares, permissions, and accessible files using sharefinder.", ip, port),
			Recommendation: "Audit SMB shares on this host. Remove anonymous/guest access. Ensure share permissions follow least-privilege principle. Disable SMBv1.",
			Timestamp:      time.Now(),
		})
	}

	for _, port := range []int{2049, 548} {
		addr := fmt.Sprintf("%s:%d", ip, port)
		conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
		if err != nil {
			continue
		}
		conn.Close()

		serviceName := "NFS"
		if port == 548 {
			serviceName = "AFP"
		}

		findings = append(findings, &model.Finding{
			ID:             generateID("share"),
			Type:           "open_share",
			Severity:       "info",
			Title:          fmt.Sprintf("%s service accessible on %s:%d", serviceName, ip, port),
			Description:    fmt.Sprintf("%s is accessible on %s port %d. Verify share permissions.", serviceName, ip, port),
			Recommendation: fmt.Sprintf("Audit %s exports/shares for appropriate access controls.", serviceName),
			Timestamp:      time.Now(),
		})
	}

	return findings
}

func (s *ShareScanner) ScanSubnet(targets []string) []*model.Finding {
	var findings []*model.Finding
	for _, ip := range targets {
		findings = append(findings, s.ScanHost(ip)...)
	}
	return findings
}
