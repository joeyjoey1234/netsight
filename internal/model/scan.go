package model

import "time"

type Scan struct {
	ID           string        `json:"id"`
	Timestamp    time.Time     `json:"timestamp"`
	Subnet       string        `json:"subnet"`
	Duration     time.Duration `json:"duration"`
	Preset       string        `json:"preset"`
	Status       string        `json:"status"`
	DevicesFound int           `json:"devicesFound"`
	Findings     []Finding     `json:"findings"`
}

type Finding struct {
	ID             string    `json:"id"`
	Type           string    `json:"type"`
	Severity       string    `json:"severity"`
	DeviceID       string    `json:"deviceId"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Recommendation string    `json:"recommendation"`
	Timestamp      time.Time `json:"timestamp"`
}
