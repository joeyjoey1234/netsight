package export

import (
	"fmt"
	"netsight/internal/model"
	"os"
	"strings"
	"time"
)

func ExportDrawIO(devices []*model.Device, links []*model.Link) (string, error) {
	outputPath := fmt.Sprintf("netsight-topology-%s.drawio", time.Now().Format("20060102-150405"))

	var sb strings.Builder

	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString(`<mxfile host="NetSight" modified="` + time.Now().Format("2006-01-02T15:04:05") + `" agent="NetSight" version="1.0">`)
	sb.WriteString(`<diagram name="Network Topology" id="topology">`)
	sb.WriteString(`<mxGraphModel dx="1422" dy="794" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="1169" pageHeight="827" math="0" shadow="0">`)
	sb.WriteString(`<root>`)
	sb.WriteString(`<mxCell id="0"/>`)
	sb.WriteString(`<mxCell id="1" parent="0"/>`)

	positions := generateLayout(devices)
	for i, d := range devices {
		x, y := positions[i].X, positions[i].Y
		label := d.Hostname
		if label == "" && len(d.IPs) > 0 {
			label = d.IPs[0]
		}
		if label == "" {
			label = d.MAC
		}

		deviceXML := fmt.Sprintf(
			`<mxCell id="dev-%s" value="%s" style="rounded=1;whiteSpace=wrap;html=1;fillColor=%s;strokeColor=%s;fontColor=#ffffff;" vertex="1" parent="1">`+
				`<mxGeometry x="%d" y="%d" width="120" height="60" as="geometry"/>`+
				`</mxCell>`,
			d.ID,
			escapeXML(label),
			deviceColor(d.Role),
			deviceStrokeColor(d.Role),
			x, y,
		)
		sb.WriteString(deviceXML)
	}

	for i, l := range links {
		edgeXML := fmt.Sprintf(
			`<mxCell id="edge-%d" value="%s" style="edgeStyle=orthogonalEdgeStyle;rounded=0;orthogonalLoop=1;jettySize=auto;html=1;strokeColor=%s;endArrow=classic;" edge="1" parent="1" source="dev-%s" target="dev-%s">`+
				`<mxGeometry relative="1" as="geometry"/>`+
				`</mxCell>`,
			i,
			escapeXML(fmt.Sprintf("%s: %s -> %s", l.Type, l.SrcPort, l.DstPort)),
			linkColor(l.Type),
			l.SourceID,
			l.TargetID,
		)
		sb.WriteString(edgeXML)
	}

	sb.WriteString(`</root>`)
	sb.WriteString(`</mxGraphModel>`)
	sb.WriteString(`</diagram>`)
	sb.WriteString(`</mxfile>`)

	if err := os.WriteFile(outputPath, []byte(sb.String()), 0644); err != nil {
		return "", fmt.Errorf("failed to write Draw.io file: %w", err)
	}

	return outputPath, nil
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

func deviceColor(role string) string {
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

func deviceStrokeColor(role string) string {
	switch role {
	case "switch":
		return "#389e0d"
	case "router":
		return "#0958d9"
	case "L3 switch":
		return "#531dab"
	case "server":
		return "#d46b08"
	case "workstation":
		return "#08979c"
	default:
		return "#595959"
	}
}

func linkColor(linkType string) string {
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

type Position struct {
	X, Y int
}

func generateLayout(devices []*model.Device) []Position {
	positions := make([]Position, len(devices))
	cols := 5
	if len(devices) < 5 {
		cols = len(devices)
	}
	for i := range devices {
		positions[i] = Position{
			X: 50 + (i%cols)*200,
			Y: 50 + (i/cols)*120,
		}
	}
	return positions
}
