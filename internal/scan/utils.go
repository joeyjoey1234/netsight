package scan

import (
	"crypto/rand"
	"fmt"
	"log"
	"net"
)

func generateID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		log.Printf("scan: failed to generate random ID: %v", err)
		return fmt.Sprintf("fallback-%d", len(b))
	}
	return fmt.Sprintf("%x", b)
}

func generateDeviceID(ip string) string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("dev-%s", ip)
	}
	return fmt.Sprintf("dev-%s-%x", ip, b[:4])
}

func GeneratePassword(length int, includeSpecial bool) string {
	const alpha = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	const digits = "0123456789"
	const special = "!@#$%^&*()-_=+[]{}|;:,.<>?"

	charset := alpha + digits
	if includeSpecial {
		charset += special
	}

	b := make([]byte, length)
	for i := range b {
		val := make([]byte, 1)
		if _, err := rand.Read(val); err != nil {
			b[i] = charset[i%len(charset)]
			continue
		}
		b[i] = charset[int(val[0])%len(charset)]
	}
	return string(b)
}

func ScanIPsExpand(cidr string) ([]string, error) {
	info, err := CalculateSubnet(cidr)
	if err != nil {
		return nil, err
	}
	if info.TotalHosts > 65536 {
		return nil, fmt.Errorf("subnet too large: %d hosts", info.TotalHosts)
	}

	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ipNet.IP.Mask(ipNet.Mask); ipNet.Contains(ip); incrementIP(ip) {
		if isNetworkOrBroadcast(ip, ipNet) {
			continue
		}
		ips = append(ips, ip.String())
	}
	return ips, nil
}
