package model

type InterfaceInfo struct {
	Name       string   `json:"name"`
	IPs        []string `json:"ips"`
	MAC        string   `json:"mac"`
	Gateway    string   `json:"gateway"`
	DNS        []string `json:"dns"`
	PublicIP   string   `json:"publicIp"`
	MTU        int      `json:"mtu"`
	Speed      int64    `json:"speed"`
	Duplex     string   `json:"duplex"`
	CRCErrors  uint64   `json:"crcErrors"`
	Collisions uint64   `json:"collisions"`
}

type PingResult struct {
	Target    string  `json:"target"`
	Sequence  int     `json:"sequence"`
	TTL       int     `json:"ttl"`
	LatencyMs float64 `json:"latencyMs"`
	Bytes     int     `json:"bytes"`
	TimedOut  bool    `json:"timedOut"`
}

type Hop struct {
	Number    int      `json:"number"`
	IP        string   `json:"ip"`
	Hostname  string   `json:"hostname"`
	LatencyMs float64  `json:"latencyMs"`
	AllIPs    []string `json:"allIps"`
}

type PacketSummary struct {
	Number   int    `json:"number"`
	Timestamp string `json:"timestamp"`
	SrcMAC   string `json:"srcMac"`
	DstMAC   string `json:"dstMac"`
	SrcIP    string `json:"srcIp"`
	DstIP    string `json:"dstIp"`
	Protocol string `json:"protocol"`
	SrcPort  int    `json:"srcPort,omitempty"`
	DstPort  int    `json:"dstPort,omitempty"`
	Length   int    `json:"length"`
	Info     string `json:"info"`
}

type IPerfResult struct {
	Interval      float64 `json:"interval"`
	TransferBytes int64   `json:"transferBytes"`
	BandwidthBps  int64   `json:"bandwidthBps"`
	JitterMs      float64 `json:"jitterMs,omitempty"`
	LostPackets   int     `json:"lostPackets,omitempty"`
	TotalPackets  int     `json:"totalPackets,omitempty"`
}
