# NetSight — Feature List v5 (Final)

## Core Discovery Engine

| # | Feature | Notes |
|---|---|---|
| 1 | **Network scan → visual topology map** | Ping sweep → live hosts → port scan → auto-layout map. Drag to organize. |
| 2 | **ARP scan** | MAC ↔ IP table per subnet. Populates device map. |
| 3 | **Reverse ARP scan** | RARP listener. Still used in PXE boot envs and by attackers mapping networks silently. Find misconfigured boot servers AND detect RARP-based recon. |
| 4 | **MAC vendor identification** | OUI lookup inline on every device. |
| 5 | **CDP/LLDP listener + parser** | Discovers switch-to-switch and switch-to-device links. Auto-draws topology edges. |
| 6 | **Control plane listener** | Passive capture. Identifies STP BPDUs, OSPF hellos, BGP, HSRP/VRRP, DHCP offers, ARP storms. Tags device roles: switch, router, L3 switch, DHCP server, HSRP primary. |
| 7 | **Service version + banner grab** | Not just "port 22 open" — "OpenSSH 7.4, RHEL 7, EOL 2024." Grabs banners on common ports during scan. |
| 8 | **Passive OS fingerprinting** | TTL, TCP window size, DF bit, MSS from captured traffic. Tags devices: "likely Linux," "likely Windows Server 2019," "likely Cisco IOS." No packets sent. |
| 9 | **HTTP/HTTPS title grabber** | Hit port 80/443/8080/8443 on discovered devices. Grab page title, status code, redirect chain. Instant context. |
| 10 | **Device recognition (smart)** | Combine MAC OUI + banner + HTTP title + OS fingerprint + CDP/LLDP platform info → "Cisco 2960-X, IOS 15.2(2)E7" instead of just "Cisco switch." |

## Layer 1–2

| # | Feature | Notes |
|---|---|---|
| 11 | **Duplex/speed mismatch detection** | Query NIC stats via WMI: CRC errors, late collisions, runts, FullDuplex flag. "This link is probably half-duplex." |
| 12 | **STP topology tree** | From BPDU capture: build spanning tree. Show root bridge, root ports, blocked ports. Instant visual of L2 topology. |
| 13 | **Broadcast storm / loop detection** | During capture: flag excessive broadcasts, multicast floods, MAC flapping. "Something's wrong on VLAN 200." |
| 14 | **Bandwidth / top-talkers** | Lightweight NetFlow/sFlow listener or per-IP bandwidth estimation during capture. "Who's saturating the WAN link?" |

## Manual Traffic Inspection

| # | Feature | Notes |
|---|---|---|
| 15 | **Packet capture (tcpdump/wireshark-lite)** | Select interface → capture with BPF filter → live packet list → click to inspect packet headers. Raw pcap export for Wireshark. |

## Troubleshooting Tools

| # | Feature | Notes |
|---|---|---|
| 16 | **Ping** | Configurable count, size, TTL, DF bit. Live output. IPv4 and IPv6. |
| 17 | **Traceroute** | ICMP, TCP, and UDP modes. Per-hop latency. IPv4 and IPv6. |
| 18 | **NSlookup / DNS query** | A, AAAA, MX, NS, TXT, PTR, SOA. Quick widget. |
| 19 | **Quick network info** | Interfaces, gateway, DNS, public IP, link speed, MTU. IPv4 and IPv6. |
| 20 | **Wake-on-LAN** | Select device from map → send magic packet. |
| 21 | **iPerf throughput test** | Set one device as server, another as client → measure TCP/UDP throughput, jitter, packet loss. Built-in. |
| 22 | **Latency/jitter monitor** | Continuous ping to target → live graph. Poor man's IP SLA. Shows loss %, jitter, latency over time. Exportable. |

## Survey Presets (One-Click Data Collection)

| # | Feature | Notes |
|---|---|---|
| 23 | **Quick survey (3 min)** | One button. Runs all passive monitoring tools for 3 minutes: control plane listener, ARP scan, CDP/LLDP, broadcast detection, ARP spoofing detection, rogue DHCP detection, bandwidth estimate. Fills out the topology map and findings panel. Fast first look. |
| 24 | **Short survey (10 min)** | Same as Quick but runs for 10 minutes. Deeper data. Catches slower BPDU/OSPF/BGP hellos. Better bandwidth estimate. Default preset for most onsite surveys. |
| 25 | **Long survey (custom time)** | User-configurable duration. Runs all monitoring tools for as long as you set. Overnight captures, intermittent issues, deep passive profiling. |

## Organization & Persistence

| # | Feature | Notes |
|---|---|---|
| 26 | **Free-use mode** | Quick scan, no persistence. Close = gone. |
| 27 | **Project mode** | Saved surveys. Scan history. Device notes. Per-project settings. |
| 28 | **Scan diff** | Project-only. New devices, missing devices, changed IPs/MACs, new open ports. |
| 29 | **VLAN-to-device mapping** | CDP/LLDP + control plane → group by VLAN. Trunk ports, access ports, native VLAN mismatches. |

## Hacker / Auditor Perspective

| # | Feature | Notes |
|---|---|---|
| 30 | **Promiscuous detection** | ARP sweep + timing analysis. Flags interfaces in promiscuous mode (likely sniffers). |
| 31 | **Default credential flagging** | Port scan hits known default cred ports (telnet 23, SSH common banners, SNMP public/private, HTTP on switches). Highlights easy entry points. |
| 32 | **Rogue DHCP detection** | Control plane listener flags DHCP offers from unexpected IPs. Someone plugged in a travel router. |
| 33 | **ARP spoofing detection** | Multiple MACs claiming same IP. Flags in real time during control plane listen. |
| 34 | **Open network shares scan** | SMB/NFS/AFP discovery. Shows writable shares, anonymous access. Audit AND exposure awareness. |

## Built-in Servers (One-Click Toggle)

| # | Feature | Notes |
|---|---|---|
| 35 | **TFTP server** | One-click toggle. Config backups, firmware transfers, phone configs. Bind to interface, set root directory. |
| 36 | **HTTP server** | One-click toggle. Serve files from a directory over HTTP. Quick config sharing, firmware hosting. |
| 37 | **FTP server** | One-click toggle. Read-only by default. Legacy devices that only speak FTP. Quick file push to old kit. |
| 38 | **Syslog server** | One-click toggle. Listen on UDP 514. Capture and display syslog from switches, routers, firewalls in real time during troubleshooting. |
| 39 | **Netcat listener** | One-click toggle. TCP or UDP listener on any port. Buffer incoming data, display hex/ASCII. Test ACLs, firewall rules, app connectivity. |
| 40 | **DHCP server** | One-click toggle. Quick temporary DHCP for isolated lab segments or recovery scenarios. Configurable pool range, gateway, DNS. |
| 41 | **NTP server** | One-click toggle. Quick time source for lab gear or devices that drift. Stratum 10, just enough for isolated networks. |
| 42 | **DNS server** | One-click toggle. Simple authoritative or forwarding DNS for lab/testing. Point test devices at it and control what they resolve. |

## Output & Automation

| # | Feature | Notes |
|---|---|---|
| 43 | **One-click PDF report** | Full survey summary: topology map, device list, open ports, findings, VLAN map, STP tree, top-talkers. Clients and managers love reports. |
| 44 | **CLI / headless mode** | `netsight scan --subnet 10.0.0.0/24 --output survey.json` — run from SSH, get structured output. Enables scripting and automation. |

## Utilities

| # | Feature | Notes |
|---|---|---|
| 45 | **Subnet calculator** | CIDR to range, wildcard mask, usable hosts, binary. Quick widget. |
| 46 | **Export to diagram** | Draw.io XML. Importable into draw.io, Lucidchart, Visio. Preserves layout, labels, VLAN colors, link types, device roles. |
| 47 | **Password/certificate generator** | Generate random passwords, PSKs, pre-shared keys. Quick widget. |
| 48 | **IPv6 support (global)** | IPv6 neighbor discovery scan, IPv6 ping, IPv6 traceroute, AAAA lookups. IPv6-aware in every tool. |

---

## Feature Map by Persona

| Engineer | Hacker / Auditor | Both |
|---|---|---|
| CDP/LLDP | Reverse ARP scan | Network scan + topology |
| VLAN mapping | Promiscuous detection | ARP scan |
| STP topology tree | Rogue DHCP detection | MAC vendor ID |
| Duplex mismatch detection | ARP spoofing detection | Control plane listener |
| Wake-on-LAN | Default credential flagging | Service version + banner |
| Broadcast storm detection | Open share scan | Passive OS fingerprinting |
| Bandwidth / top-talkers | | HTTP title grabber |
| iPerf throughput test | | Device recognition |
| Latency/jitter monitor | | Packet capture |
| Survey presets (all 3) | | Troubleshooting tools |
| Built-in servers (all 8) | | Subnet calculator |
| CLI mode | | Draw.io export |
| PDF report | | IPv6 support |
| Scan diff | | Password generator |

---

## Survey Presets — Quick Reference

| Preset | Duration | Runs | Best For |
|---|---|---|---|
| Quick | 3 min | Control plane, ARP, CDP/LLDP, broadcast, ARP spoofing, rogue DHCP, bandwidth estimate | Fast first look, pre-meeting scan |
| Short | 10 min | Same as Quick, longer capture | Default onsite survey, catches slow hellos |
| Long | Custom | Same as Short, user-defined time | Overnight, intermittent issues, deep profiling |

---

## Built-in Servers — Quick Reference

| Server | Default Port | Use Case |
|---|---|---|
| TFTP | UDP 69 | Config backup/restore, firmware push, VoIP phone configs |
| HTTP | TCP 8080 | Quick file sharing, serving configs, firmware to devices with web fetch |
| FTP | TCP 21 | Legacy devices, old IP cameras, printers that only do FTP |
| Syslog | UDP 514 | Capture real-time logs from switches/routers/firewalls during troubleshooting |
| Netcat | Any | Test ACLs, firewall rules, app connectivity. Hex/ASCII viewer |
| DHCP | UDP 67 | Isolated lab segments, recovery mode, replacing a dead DHCP server temporarily |
| NTP | UDP 123 | Lab gear time sync, devices that drift, isolated networks |
| DNS | UDP 53 | Lab/testing DNS. Authoritative or forwarding. Control what devices resolve |

---

## Tech Stack

- **Language:** Go 1.22+
- **GUI:** Wails v3 (WebView2-based, built into Windows 10+)
- **Topology visualization:** vis-network or Cytoscape.js (rendered in Wails WebView)
- **Packet capture:** gopacket (Google) + Npcap/WinPcap on Windows
- **Packet crafting/dissection:** goscapy (pure-Go Scapy port) — BGP, OSPF, STP, VRRP, HSRP, CDP, LLDP, DHCP layers built-in
- **Scanning:** Custom ARP/ICMP/TCP scanning via gopacket + goscapy
- **Diagram export:** Custom draw.io XML builder
- **Persistence:** SQLite via modernc.org/sqlite (pure Go, no CGo) or mattn/go-sqlite3
- **Servers:** Custom lightweight implementations using Go stdlib + existing libraries (see research below)
- **IPv6:** Native via Go stdlib `net` + goscapy IPv6 support
- **PDF reports:** jung-kurt/gofpdf or signintech/gopdf
- **Windows system queries:** StackExchange/wmi (WMI queries for NIC stats, duplex, adapters)

### Distribution

- Single `.exe` binary (~20–30 MB) via `GOOS=windows GOARCH=amd64 go build`
- Zero installation required. Download, double-click, run.
- Only external dependency: **Npcap** (free, one-time install). Bundled installer can be shipped alongside.

---

## Library Research & Feasibility

All 48 features were audited against Go's package ecosystem. Results:

### Solid (mature Go libraries exist) — 46 features

| Feature | Go Library | Notes |
|---|---|---|
| Packet capture (pcap/bpf) | `google/gopacket` | Gold standard, wraps Npcap on Windows |
| ARP scanning | `mdlayher/arp` or gopacket layers | Pure Go ARP handling |
| CDP parsing | gopacket CDP layer | Cisco Discovery Protocol, built-in |
| LLDP parsing | gopacket LLDP layer | Link Layer Discovery Protocol, built-in |
| STP/BPDU parsing | gopacket BPDU layer | Spanning Tree, built-in |
| OSPF parsing | goscapy / gopacket OSPFv2/v3 layers | Built-in |
| BGP parsing | `osrg/gobgp` (CNCF) or goscapy BGP layer | Full BGP daemon library available |
| VRRP/HSRP parsing | goscapy VRRPv2/v3 / HSRP layers | Built-in |
| DHCP client + server | `insomniacslk/dhcp` | Full DHCPv4/v6, widely used |
| DNS server | `miekg/dns` | De facto standard, full server + client |
| NTP client | `beevik/ntp` | 713 importers, mature |
| NTP server | `soypat/lneto/examples/ntp-server` + `facebook/time` server code | Minimal server to Facebook-grade |
| Passive OS fingerprinting | `smallnest/goscapy/pkg/p0f` | p0f signature format compatible |
| SMB share discovery | `vflame6/sharefinder` | Active shares enumeration, guest/null/auth |
| iPerf throughput | `BGrewell/go-iperf` | Wraps + embeds iperf3 binary, Windows support |
| TFTP server | `pin/tftp` | RFC 1350 compliant, pure Go |
| FTP server | `fclairamb/ftpserverlib` | Mature server framework |
| Syslog server | `influxdata/go-syslog` | RFC 5424/3164 parser |
| PDF generation | `jung-kurt/gofpdf` or `signintech/gopdf` | Mature PDF libs |
| MAC OUI lookup | Static OUI table (IEEE-published) | Trivial, many Go helpers exist |
| WMI queries (NIC stats, duplex) | `StackExchange/wmi` | WQL interface for Windows WMI |
| SQLite | `modernc.org/sqlite` | Pure Go, no CGo required |

### Doable (custom work required, clear path) — 1 feature

| Feature | Approach | Effort |
|---|---|---|
| Duplex/speed mismatch (#11) | Query `Win32_NetworkAdapter` (Speed) + `MSFT_NetAdapter` (FullDuplex) + `Win32_PerfRawData` (CRC errors, collisions, runts) via `StackExchange/wmi` | ~2 days |

### Windows Platform Limitation — 1 feature removed

| Feature | Reason |
|---|---|
| ~~WiFi survey~~ | **Removed.** WiFi monitor mode is impossible on Windows — the kernel simply does not expose raw 802.11 frame capture. Basic SSID/BSSID scanning via `wlanapi.dll` syscall is possible but provides no value without monitor mode for the intended auditor/engineer use cases (channel analysis, hidden SSID detection, security type audit). Specialized AirPcap hardware is required. This feature is deferred to a future Linux build. |

---

## Build Order

1. Network scan + visual topology map + draw.io export (the core)
2. ARP scan + MAC vendor ID
3. CDP/LLDP integration into topology map
4. Control plane listener + device role tagging
5. Project mode + scan diff
6. VLAN-to-device mapping
7. Troubleshooting tools (ping, traceroute, nslookup, quick info)
8. Survey presets (Quick / Short / Long)
9. Packet capture (tcpdump-lite)
10. Built-in servers (TFTP, HTTP, FTP — most-used first)
11. STP topology tree + broadcast storm detection
12. Service version + banner grab + HTTP title grabber + passive OS fingerprinting
13. Device recognition (smart)
14. Hacker perspective features
15. Wake-on-LAN
16. iPerf throughput test + latency/jitter monitor
17. Bandwidth / top-talkers + duplex mismatch detection
18. Additional servers (Syslog, Netcat, DHCP, NTP, DNS)
19. IPv6 support (global — add to all existing tools)
20. CLI mode
21. PDF report
22. Subnet calculator + password generator
