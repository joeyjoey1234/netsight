package recognition

import (
	"fmt"
	"netsight/internal/model"
	"strings"
)

type Recognizer struct{}

func NewRecognizer() *Recognizer {
	return &Recognizer{}
}

func (r *Recognizer) IdentifyDevice(device *model.Device, cdpInfo map[string]string, lldpInfo map[string]string, httpTitle string) {
	r.detectRole(device)
	r.detectModel(device, cdpInfo, lldpInfo)
	r.detectOS(device, cdpInfo, lldpInfo)
	r.mergeHostname(device, cdpInfo, lldpInfo, httpTitle)
}

func (r *Recognizer) detectRole(device *model.Device) {
	if device.Role != "" && device.Role != "unknown" {
		return
	}

	vendor := strings.ToLower(device.Vendor)
	switch {
	case strings.Contains(vendor, "cisco"),
		strings.Contains(vendor, "juniper"),
		strings.Contains(vendor, "huawei"),
		strings.Contains(vendor, "arista"),
		strings.Contains(vendor, "brocade"),
		strings.Contains(vendor, "aruba"):
		device.Role = "switch"
	case strings.Contains(vendor, "palo alto"),
		strings.Contains(vendor, "fortinet"),
		strings.Contains(vendor, "f5"):
		device.Role = "router"
	case strings.Contains(vendor, "vmware"):
		device.Role = "server"
	case strings.Contains(vendor, "microsoft"),
		strings.Contains(vendor, "linux"),
		strings.Contains(vendor, "red hat"):
		device.Role = "server"
	default:
		device.Role = "workstation"
	}
}

func (r *Recognizer) detectModel(device *model.Device, cdpInfo, lldpInfo map[string]string) {
	if platform, ok := cdpInfo["platform"]; ok && platform != "" {
		device.Model = extractCiscoModel(platform)
		return
	}
	if platform, ok := lldpInfo["systemDesc"]; ok && platform != "" {
		device.Model = extractCiscoModel(platform)
		return
	}

	if device.Vendor != "" {
		device.Model = device.Vendor
	}
}

func (r *Recognizer) detectOS(device *model.Device, cdpInfo, lldpInfo map[string]string) {
	if device.OS != "" {
		return
	}

	if version, ok := cdpInfo["version"]; ok && version != "" {
		device.OS = version
		return
	}

	if desc, ok := lldpInfo["systemDesc"]; ok && desc != "" {
		device.OS = desc
		return
	}
}

func (r *Recognizer) mergeHostname(device *model.Device, cdpInfo, lldpInfo map[string]string, httpTitle string) {
	candidates := []string{}

	if name, ok := cdpInfo["deviceID"]; ok && name != "" {
		candidates = append(candidates, name)
	}
	if name, ok := lldpInfo["systemName"]; ok && name != "" {
		candidates = append(candidates, name)
	}
	if httpTitle != "" {
		candidates = append(candidates, httpTitle)
	}

	for _, c := range candidates {
		if c != "" && device.Hostname == "" {
			device.Hostname = c
			break
		}
	}
}

func extractCiscoModel(platform string) string {
	platform = strings.TrimSpace(platform)
	parts := strings.Fields(platform)
	for _, part := range parts {
		part = strings.ToUpper(part)
		if strings.HasPrefix(part, "WS-C") || strings.HasPrefix(part, "CISCO") {
			model := strings.TrimPrefix(part, "WS-C")
			model = strings.TrimPrefix(model, "CISCO")
			model = strings.ReplaceAll(model, "-", " ")
			return fmt.Sprintf("Cisco %s", model)
		}
	}
	return platform
}

func MapVLANsToDevices(devices []*model.Device) map[int][]*model.Device {
	vlanMap := make(map[int][]*model.Device)
	for _, d := range devices {
		for _, vlan := range d.VLANs {
			vlanMap[vlan] = append(vlanMap[vlan], d)
		}
	}
	return vlanMap
}

func BuildTopologyGraph(devices []*model.Device, links []*model.Link) *TopologyGraph {
	graph := &TopologyGraph{
		Nodes: make([]TopologyNode, 0, len(devices)),
		Edges: make([]TopologyEdge, 0, len(links)),
	}

	for _, d := range devices {
		node := TopologyNode{
			ID:    d.ID,
			Label: d.Hostname,
			Title: formatDeviceTooltip(d),
			Group: roleToGroup(d.Role),
			Color: roleToColor(d.Role),
			Shape: "dot",
			Size:  25,
		}
		if d.Hostname == "" {
			ipLabel := ""
			if len(d.IPs) > 0 {
				ipLabel = d.IPs[0]
			}
			node.Label = ipLabel
		}
		graph.Nodes = append(graph.Nodes, node)
	}

	for _, l := range links {
		edge := TopologyEdge{
			From:   l.SourceID,
			To:     l.TargetID,
			Label:  l.Type,
			Dashes: l.Type == "STP" && l.SrcPort == "Blocked",
			Color:  linkTypeColor(l.Type),
			Width:  2,
		}
		graph.Edges = append(graph.Edges, edge)
	}

	return graph
}

type TopologyGraph struct {
	Nodes []TopologyNode `json:"nodes"`
	Edges []TopologyEdge `json:"edges"`
}

type TopologyNode struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Title string `json:"title"`
	Group string `json:"group"`
	Color string `json:"color"`
	Shape string `json:"shape"`
	Size  int    `json:"size"`
}

type TopologyEdge struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Label  string `json:"label"`
	Dashes bool   `json:"dashes"`
	Color  string `json:"color"`
	Width  int    `json:"width"`
}

func formatDeviceTooltip(d *model.Device) string {
	var parts []string
	if d.Hostname != "" {
		parts = append(parts, fmt.Sprintf("<b>%s</b>", d.Hostname))
	}
	if len(d.IPs) > 0 {
		parts = append(parts, fmt.Sprintf("IP: %s", d.IPs[0]))
	}
	if d.MAC != "" {
		parts = append(parts, fmt.Sprintf("MAC: %s", d.MAC))
	}
	if d.Vendor != "" {
		parts = append(parts, fmt.Sprintf("Vendor: %s", d.Vendor))
	}
	if d.Model != "" {
		parts = append(parts, fmt.Sprintf("Model: %s", d.Model))
	}
	if d.OS != "" {
		parts = append(parts, fmt.Sprintf("OS: %s", d.OS))
	}
	if d.Role != "" && d.Role != "unknown" {
		parts = append(parts, fmt.Sprintf("Role: %s", d.Role))
	}
	if len(d.VLANs) > 0 {
		vlanStrs := make([]string, len(d.VLANs))
		for i, v := range d.VLANs {
			vlanStrs[i] = fmt.Sprintf("%d", v)
		}
		parts = append(parts, fmt.Sprintf("VLANs: %s", strings.Join(vlanStrs, ", ")))
	}
	return strings.Join(parts, "<br>")
}

func roleToGroup(role string) string {
	switch role {
	case "switch", "router", "L3 switch":
		return "network"
	case "server":
		return "server"
	case "workstation":
		return "endpoint"
	default:
		return "unknown"
	}
}

func roleToColor(role string) string {
	switch role {
	case "switch":
		return "#52c41a"
	case "router":
		return "#1677ff"
	case "L3 switch":
		return "#722ed1"
	case "server":
		return "#fa8c16"
	case "workstation":
		return "#13c2c2"
	default:
		return "#8c8c8c"
	}
}

func linkTypeColor(linkType string) string {
	switch linkType {
	case "CDP", "LLDP":
		return "#1677ff"
	case "STP":
		return "#f5222d"
	case "ARP":
		return "#faad14"
	default:
		return "#8c8c8c"
	}
}
