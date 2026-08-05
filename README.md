# NetSight

Portable network analysis tool for Windows — network scanning, topology mapping, packet capture, built-in servers, and security auditing. Single `.exe`, zero install.

## Tech Stack

- **Backend:** Go 1.22+ / Wails v2
- **Frontend:** React 18 + TypeScript + Ant Design 5 + vis-network
- **Packet capture:** google/gopacket + Npcap
- **Storage:** SQLite

## Download

Pre-built binaries: [GitHub Releases](https://github.com/joeyjoey1234/netsight/releases)

Requires [Npcap](https://npcap.com) (free, one-time install).

## Build (Windows only)

```bash
# Prerequisites: Go 1.22+, Node.js 18+, Wails CLI, MinGW (for CGO)
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails build -platform windows/amd64 -o netsight.exe
```

## Run

**GUI:** Double-click `netsight.exe`

**CLI:**
```bash
netsight scan --subnet 192.168.1.0/24 --preset quick
netsight export --input scan.json --format pdf
netsight version
```

## Features

- Network scan with visual topology map
- ARP scan + MAC vendor identification
- CDP/LLDP/STP/OSPF/BGP protocol parsers
- Passive OS fingerprinting
- Packet capture with BPF filters
- 8 built-in servers (TFTP, HTTP, FTP, Syslog, Netcat, DHCP, NTP, DNS)
- Troubleshooting tools (ping, traceroute, nslookup, WoL, iPerf)
- One-click survey presets (Quick 3m / Short 10m / Long)
- Security auditing (rogue DHCP, ARP spoofing, open shares, default creds)
- PDF report + Draw.io export
- IPv6 support

## License

MIT
