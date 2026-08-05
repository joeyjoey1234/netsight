package model

import "time"

type Project struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Created  time.Time       `json:"created"`
	Settings ProjectSettings `json:"settings"`
}

type ProjectSettings struct {
	DefaultSubnet string   `json:"defaultSubnet"`
	ScanPorts     []int    `json:"scanPorts"`
	ExcludeIPs    []string `json:"excludeIps"`
}
