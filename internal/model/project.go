package model

import "time"

type Project struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Created  time.Time       `json:"created"`
	Settings ProjectSettings `json:"settings"`
	Devices  []*Device       `json:"devices,omitempty"`
	Scans    []*Scan         `json:"scans,omitempty"`
	Findings []Finding       `json:"findings,omitempty"`
}

type ProjectSettings struct {
	DefaultSubnet string   `json:"defaultSubnet"`
	ScanPorts     []int    `json:"scanPorts"`
	ExcludeIPs    []string `json:"excludeIps"`
}
