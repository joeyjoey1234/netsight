package fingerprint

import (
	"fmt"
	"strings"
)

type OSSignature struct {
	TTL           int
	TCPWindowSize int
	DFBit         bool
	MSS           int
	TCPOptions    string
}

type OSResult struct {
	OS         string
	Confidence float64
	Details    map[string]string
}

var osDB = []struct {
	OS    string
	Match func(sig *OSSignature) bool
}{
	{
		OS: "Linux (recent kernel)",
		Match: func(s *OSSignature) bool {
			return s.TTL == 64 && s.DFBit && s.TCPWindowSize > 60000 && s.MSS == 1460
		},
	},
	{
		OS: "Linux (older kernel)",
		Match: func(s *OSSignature) bool {
			return s.TTL == 64 && !s.DFBit && s.TCPWindowSize == 5840 && s.MSS == 1460
		},
	},
	{
		OS: "Windows 10/11",
		Match: func(s *OSSignature) bool {
			return s.TTL == 128 && s.DFBit && s.TCPWindowSize == 65535 && (s.MSS == 1460 || s.MSS == 0)
		},
	},
	{
		OS: "Windows Server 2016/2019/2022",
		Match: func(s *OSSignature) bool {
			return s.TTL == 128 && s.DFBit && s.TCPWindowSize == 65535 && s.MSS == 1460
		},
	},
	{
		OS: "Cisco IOS",
		Match: func(s *OSSignature) bool {
			return s.TTL == 255 && s.TCPWindowSize == 4128
		},
	},
	{
		OS: "Cisco IOS XR",
		Match: func(s *OSSignature) bool {
			return s.TTL == 255 && s.TCPWindowSize == 16384
		},
	},
	{
		OS: "Juniper JUNOS",
		Match: func(s *OSSignature) bool {
			return s.TTL == 64 && s.TCPWindowSize == 65535 && s.MSS == 1460
		},
	},
	{
		OS: "FreeBSD",
		Match: func(s *OSSignature) bool {
			return s.TTL == 64 && s.DFBit && s.TCPWindowSize == 65535 && s.MSS == 1460
		},
	},
	{
		OS: "macOS",
		Match: func(s *OSSignature) bool {
			return s.TTL == 64 && s.DFBit && s.TCPWindowSize == 65535 && s.MSS == 1460
		},
	},
	{
		OS: "VMware ESXi",
		Match: func(s *OSSignature) bool {
			return s.TTL == 64 && s.TCPWindowSize == 65535
		},
	},
	{
		OS: "Aruba Switch",
		Match: func(s *OSSignature) bool {
			return s.TTL == 64 && s.TCPWindowSize == 4096
		},
	},
}

func DetectOS(sig *OSSignature) *OSResult {
	for _, entry := range osDB {
		if entry.Match(sig) {
			return &OSResult{
				OS:         entry.OS,
				Confidence: 0.7,
				Details: map[string]string{
					"ttl":    fmt.Sprintf("%d", sig.TTL),
					"window": fmt.Sprintf("%d", sig.TCPWindowSize),
					"df":     fmt.Sprintf("%v", sig.DFBit),
					"mss":    fmt.Sprintf("%d", sig.MSS),
				},
			}
		}
	}

	os := "unknown"
	switch {
	case sig.TTL <= 64:
		os = "likely Linux/Unix"
	case sig.TTL <= 128:
		os = "likely Windows"
	case sig.TTL <= 255:
		os = "likely network device (Cisco/Juniper)"
	}

	return &OSResult{
		OS:         os,
		Confidence: 0.3,
	}
}

func ExtractSignatureFromPacket(ttl int, winSize int, df bool, mss int, tcpOpts string) *OSSignature {
	return &OSSignature{
		TTL:           ttl,
		TCPWindowSize: winSize,
		DFBit:         df,
		MSS:           mss,
		TCPOptions:    strings.TrimSpace(tcpOpts),
	}
}
