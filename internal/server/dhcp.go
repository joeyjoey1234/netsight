package server

import (
	"context"
	"fmt"
	"net"
	"netsight/internal/model"
	"sync"
	"time"
)

type DHCPLease struct {
	IP        string
	MAC       string
	Hostname  string
	LeasedAt  time.Time
	ExpiresAt time.Time
	Active    bool
}

type DHCPServer struct {
	mu      sync.Mutex
	leases  map[string]*DHCPLease
	pool    []string
	nextIP  int
	gateway string
	dns     string
	subnet  *net.IPNet
}

func startDHCP(ctx context.Context, config *model.ServerConfig, onStatus func(*model.ServerState)) error {
	port := config.Port
	if port == 0 {
		port = 67
	}

	poolStart := net.ParseIP(config.PoolStart)
	poolEnd := net.ParseIP(config.PoolEnd)
	if poolStart == nil {
		poolStart = net.ParseIP("192.168.100.100")
	}
	if poolEnd == nil {
		poolEnd = net.ParseIP("192.168.100.200")
	}

	var pool []string
	for ip := copyIP(poolStart); !ip.Equal(nextIP(poolEnd)); incrementIP(ip) {
		pool = append(pool, ip.String())
	}

	gateway := config.Gateway
	if gateway == "" {
		gateway = "192.168.100.1"
	}

	dns := config.DNS
	if dns == "" {
		dns = "8.8.8.8"
	}

	server := &DHCPServer{
		leases:  make(map[string]*DHCPLease),
		pool:    pool,
		nextIP:  0,
		gateway: gateway,
		dns:     dns,
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
		return fmt.Errorf("DHCP listen failed: %w", err)
	}
	defer conn.Close()

	onStatus(&model.ServerState{
		Type:   "dhcp",
		Port:   port,
		Status: "running",
	})

	buf := make([]byte, 1500)

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

		response := server.handleDHCPPacket(buf[:n], remoteAddr)
		if response != nil {
			conn.WriteToUDP(response, remoteAddr)
		}
	}
}

func (s *DHCPServer) handleDHCPPacket(packet []byte, remoteAddr *net.UDPAddr) []byte {
	if len(packet) < 240 {
		return nil
	}

	msgType := byte(0)
	for i := 240; i < len(packet); {
		opt := packet[i]
		if opt == 255 {
			break
		}
		if opt == 0 {
			i++
			continue
		}
		if i+1 >= len(packet) {
			break
		}
		optLen := int(packet[i+1])
		if opt == 53 && optLen >= 1 && i+2 < len(packet) {
			msgType = packet[i+2]
			break
		}
		i += 2 + optLen
	}

	var response []byte

	switch msgType {
	case 1:
		response = s.buildDHCPOffer(packet)
	case 3:
		response = s.buildDHCPAck(packet)
	}

	return response
}

func (s *DHCPServer) buildDHCPOffer(packet []byte) []byte {
	mac := extractClientMAC(packet)
	ip := s.allocateIP(string(mac))
	_ = ip

	offer := make([]byte, 300)
	copy(offer[0:4], packet[0:4])

	offer[0] = 2
	offer[1] = 1
	offer[2] = 6
	offer[10] = 0
	offer[11] = 0
	offer[12] = 128
	offer[13] = 63

	copy(offer[16:20], net.ParseIP(s.gateway).To4())

	for i := 0; i < 6; i++ {
		offer[28+i] = mac[i]
	}

	copy(offer[236:240], []byte{99, 130, 83, 99})

	offer[240] = 53
	offer[241] = 1
	offer[242] = 2
	offer[243] = 54
	offer[244] = 4
	copy(offer[245:249], net.ParseIP(s.gateway).To4())
	offer[249] = 51
	offer[250] = 4
	offer[251] = 0
	offer[252] = 0
	offer[253] = 14
	offer[254] = 16
	offer[255] = 1
	offer[256] = 4
	copy(offer[257:261], []byte{255, 255, 255, 0})
	offer[261] = 3
	offer[262] = 4
	copy(offer[263:267], net.ParseIP(s.gateway).To4())
	offer[267] = 6
	offer[268] = 4
	copy(offer[269:273], net.ParseIP(s.dns).To4())
	offer[273] = 255

	return offer[:274]
}

func (s *DHCPServer) buildDHCPAck(packet []byte) []byte {
	ack := make([]byte, 300)
	copy(ack[0:4], packet[0:4])

	ack[0] = 2
	ack[1] = 1
	ack[2] = 6

	copy(ack[236:240], []byte{99, 130, 83, 99})
	ack[240] = 53
	ack[241] = 1
	ack[242] = 5
	ack[243] = 255

	return ack[:244]
}

func (s *DHCPServer) allocateIP(mac string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if lease, exists := s.leases[mac]; exists {
		lease.Active = true
		lease.LeasedAt = time.Now()
		lease.ExpiresAt = time.Now().Add(1 * time.Hour)
		return lease.IP
	}

	if s.nextIP >= len(s.pool) {
		s.nextIP = 0
	}

	ip := s.pool[s.nextIP]
	s.nextIP++

	s.leases[mac] = &DHCPLease{
		IP:        ip,
		MAC:       mac,
		LeasedAt:  time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Active:    true,
	}

	return ip
}

func extractClientMAC(packet []byte) []byte {
	if len(packet) < 34 {
		return []byte{0, 0, 0, 0, 0, 0}
	}
	return packet[28:34]
}

func copyIP(ip net.IP) net.IP {
	dup := make(net.IP, len(ip))
	copy(dup, ip)
	return dup
}

func nextIP(ip net.IP) net.IP {
	next := copyIP(ip)
	for i := len(next) - 1; i >= 0; i-- {
		next[i]++
		if next[i] != 0 {
			break
		}
	}
	return next
}

func incrementIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			break
		}
	}
}
