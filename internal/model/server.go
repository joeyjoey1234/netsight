package model

import "time"

type ServerState struct {
	Type      string    `json:"type"`
	Port      int       `json:"port"`
	Status    string    `json:"status"`
	Interface string    `json:"interface"`
	Error     string    `json:"error,omitempty"`
	StartedAt time.Time `json:"startedAt,omitempty"`
}

type ServerConfig struct {
	Port      int    `json:"port"`
	Interface string `json:"interface"`
	RootDir   string `json:"rootDir,omitempty"`
	ReadOnly  bool   `json:"readOnly,omitempty"`
	PoolStart string `json:"poolStart,omitempty"`
	PoolEnd   string `json:"poolEnd,omitempty"`
	Gateway   string `json:"gateway,omitempty"`
	DNS       string `json:"dns,omitempty"`
}
