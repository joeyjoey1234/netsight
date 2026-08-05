package scan

import (
	"fmt"
	"net"
	"strings"
)

type SubnetInfo struct {
	CIDR        string `json:"cidr"`
	Netmask     string `json:"netmask"`
	Wildcard    string `json:"wildcard"`
	Network     string `json:"network"`
	Broadcast   string `json:"broadcast"`
	FirstHost   string `json:"firstHost"`
	LastHost    string `json:"lastHost"`
	TotalHosts  int    `json:"totalHosts"`
	UsableHosts int    `json:"usableHosts"`
	Binary      string `json:"binary"`
}

func CalculateSubnet(cidr string) (*SubnetInfo, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR: %w", err)
	}

	ones, bits := ipNet.Mask.Size()
	mask := ipNet.Mask

	info := &SubnetInfo{
		CIDR:     cidr,
		Netmask:  mask.String(),
		Wildcard: wildcardMask(mask).String(),
		Network:  ipNet.IP.Mask(mask).String(),
	}

	broadcast := make(net.IP, len(ipNet.IP))
	for i := range broadcast {
		broadcast[i] = ipNet.IP[i] | ^mask[i]
	}
	info.Broadcast = broadcast.String()

	info.TotalHosts = 1 << (bits - ones)
	if info.TotalHosts > 2 {
		info.UsableHosts = info.TotalHosts - 2
		first := make(net.IP, len(ipNet.IP))
		copy(first, ipNet.IP.Mask(mask))
		first[len(first)-1]++
		info.FirstHost = first.String()
		last := make(net.IP, len(broadcast))
		copy(last, broadcast)
		last[len(last)-1]--
		info.LastHost = last.String()
	} else if info.TotalHosts == 2 {
		info.UsableHosts = info.TotalHosts
		info.FirstHost = ipNet.IP.Mask(mask).String()
		info.LastHost = broadcast.String()
	} else {
		info.UsableHosts = info.TotalHosts
		info.FirstHost = ipNet.IP.String()
		info.LastHost = ipNet.IP.String()
	}

	var binaryParts []string
	for _, octet := range mask {
		binaryParts = append(binaryParts, fmt.Sprintf("%08b", octet))
	}
	info.Binary = strings.Join(binaryParts, ".")

	return info, nil
}

func wildcardMask(mask net.IPMask) net.IPMask {
	wildcard := make(net.IPMask, len(mask))
	for i, b := range mask {
		wildcard[i] = ^b
	}
	return wildcard
}
