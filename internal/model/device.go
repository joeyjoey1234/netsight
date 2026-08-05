package model

import "time"

type Device struct {
	ID        string    `json:"id"`
	MAC       string    `json:"mac"`
	IPs       []string  `json:"ips"`
	Vendor    string    `json:"vendor"`
	Hostname  string    `json:"hostname"`
	OS        string    `json:"os"`
	Role      string    `json:"role"`
	VLANs     []int     `json:"vlans"`
	FirstSeen time.Time `json:"firstSeen"`
	LastSeen  time.Time `json:"lastSeen"`
	Notes     string    `json:"notes"`
	Model     string    `json:"model"`
}

type Port struct {
	DeviceID string `json:"deviceId"`
	Number   int    `json:"number"`
	Protocol string `json:"protocol"`
	Service  string `json:"service"`
	Version  string `json:"version"`
	Banner   string `json:"banner"`
	State    string `json:"state"`
}

type Link struct {
	SourceID string `json:"sourceId"`
	TargetID string `json:"targetId"`
	Type     string `json:"type"`
	SrcPort  string `json:"srcPort"`
	DstPort  string `json:"dstPort"`
	VLAN     int    `json:"vlan"`
}
