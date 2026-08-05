package server

import (
	"context"
	"fmt"
	"netsight/internal/model"
	"sync"
)

// StatusCallback is called when a server's status changes
type StatusCallback func(state *model.ServerState)

// Manager coordinates lifecycle for all built-in servers
type Manager struct {
	mu       sync.RWMutex
	servers  map[string]*serverInstance
	onStatus StatusCallback
}

type serverInstance struct {
	config  *model.ServerConfig
	state   *model.ServerState
	cancel  context.CancelFunc
	running bool
}

// NewManager creates a server manager
func NewManager(onStatus StatusCallback) *Manager {
	return &Manager{
		servers:  make(map[string]*serverInstance),
		onStatus: onStatus,
	}
}

// StartServer launches a server by type
func (m *Manager) StartServer(serverType string, config *model.ServerConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if inst, exists := m.servers[serverType]; exists && inst.running {
		return fmt.Errorf("%s server already running", serverType)
	}

	ctx, cancel := context.WithCancel(context.Background())
	state := &model.ServerState{
		Type:   serverType,
		Port:   config.Port,
		Status: "starting",
	}

	inst := &serverInstance{
		config: config,
		state:  state,
		cancel: cancel,
	}
	m.servers[serverType] = inst

	m.emitStatus(state)

	go func() {
		var err error
		switch serverType {
		case "tftp":
			err = startTFTP(ctx, config, func(s *model.ServerState) { m.emitStatus(s) })
		case "http":
			err = startHTTP(ctx, config, func(s *model.ServerState) { m.emitStatus(s) })
		case "ftp":
			err = startFTP(ctx, config, func(s *model.ServerState) { m.emitStatus(s) })
		case "syslog":
			err = startSyslog(ctx, config, func(s *model.ServerState) { m.emitStatus(s) })
		case "netcat":
			err = startNetcat(ctx, config, func(s *model.ServerState) { m.emitStatus(s) })
		case "dhcp":
			err = startDHCP(ctx, config, func(s *model.ServerState) { m.emitStatus(s) })
		case "ntp":
			err = startNTP(ctx, config, func(s *model.ServerState) { m.emitStatus(s) })
		case "dns":
			err = startDNS(ctx, config, func(s *model.ServerState) { m.emitStatus(s) })
		default:
			err = fmt.Errorf("unknown server type: %s", serverType)
		}

		m.mu.Lock()
		inst.running = false
		if err != nil {
			inst.state.Status = "error"
			inst.state.Error = err.Error()
		} else {
			inst.state.Status = "stopped"
		}
		m.mu.Unlock()
		m.emitStatus(inst.state)
	}()

	m.mu.Lock()
	inst.running = true
	inst.state.Status = "running"
	m.mu.Unlock()

	return nil
}

// StopServer stops a running server
func (m *Manager) StopServer(serverType string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	inst, exists := m.servers[serverType]
	if !exists || !inst.running {
		return fmt.Errorf("%s server not running", serverType)
	}

	inst.cancel()
	inst.running = false
	inst.state.Status = "stopped"
	m.emitStatus(inst.state)
	return nil
}

// GetStatus returns the current state of all servers
func (m *Manager) GetStatus() []*model.ServerState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	states := make([]*model.ServerState, 0, len(m.servers))
	for _, inst := range m.servers {
		states = append(states, inst.state)
	}
	return states
}

// IsRunning returns whether a specific server is running
func (m *Manager) IsRunning(serverType string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	inst, exists := m.servers[serverType]
	return exists && inst.running
}

// Shutdown stops all running servers
func (m *Manager) Shutdown() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, inst := range m.servers {
		if inst.running {
			inst.cancel()
			inst.running = false
			inst.state.Status = "stopped"
			m.emitStatus(inst.state)
		}
	}
}

func (m *Manager) emitStatus(state *model.ServerState) {
	if m.onStatus != nil {
		m.onStatus(state)
	}
}
