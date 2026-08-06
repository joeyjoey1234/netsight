package server

import (
	"context"
	"fmt"
	"netsight/internal/model"
	"sync"
	"time"
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
	if config == nil {
		return fmt.Errorf("server config is required")
	}
	switch serverType {
	case "tftp", "http", "ftp", "syslog", "netcat", "dhcp", "ntp", "dns":
	default:
		return fmt.Errorf("unknown server type: %s", serverType)
	}
	if config.Port < 0 || config.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Port)
	}
	m.mu.Lock()
	if inst, exists := m.servers[serverType]; exists && (inst.running || inst.state.Status == "starting") {
		m.mu.Unlock()
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

	m.mu.Unlock()
	m.emitStatus(cloneState(state))

	go func() {
		var err error
		switch serverType {
		case "tftp":
			err = startTFTP(ctx, config, func(s *model.ServerState) { m.markRunning(serverType, inst, s) })
		case "http":
			err = startHTTP(ctx, config, func(s *model.ServerState) { m.markRunning(serverType, inst, s) })
		case "ftp":
			err = startFTP(ctx, config, func(s *model.ServerState) { m.markRunning(serverType, inst, s) })
		case "syslog":
			err = startSyslog(ctx, config, func(s *model.ServerState) { m.markRunning(serverType, inst, s) })
		case "netcat":
			err = startNetcat(ctx, config, func(s *model.ServerState) { m.markRunning(serverType, inst, s) })
		case "dhcp":
			err = startDHCP(ctx, config, func(s *model.ServerState) { m.markRunning(serverType, inst, s) })
		case "ntp":
			err = startNTP(ctx, config, func(s *model.ServerState) { m.markRunning(serverType, inst, s) })
		case "dns":
			err = startDNS(ctx, config, func(s *model.ServerState) { m.markRunning(serverType, inst, s) })
		default:
			err = fmt.Errorf("unknown server type: %s", serverType)
		}

		m.mu.Lock()
		if current, ok := m.servers[serverType]; !ok || current != inst {
			m.mu.Unlock()
			return
		}
		inst.running = false
		if ctx.Err() != nil {
			inst.state.Status = "stopped"
		} else if err != nil {
			inst.state.Status = "error"
			inst.state.Error = err.Error()
		} else {
			inst.state.Status = "stopped"
		}
		finalState := cloneState(inst.state)
		m.mu.Unlock()
		m.emitStatus(finalState)
	}()

	return nil
}

// StopServer stops a running server
func (m *Manager) StopServer(serverType string) error {
	m.mu.Lock()

	inst, exists := m.servers[serverType]
	if !exists || (!inst.running && inst.state.Status != "starting") {
		m.mu.Unlock()
		return fmt.Errorf("%s server not running", serverType)
	}

	inst.cancel()
	inst.running = false
	inst.state.Status = "stopped"
	state := cloneState(inst.state)
	m.mu.Unlock()
	m.emitStatus(state)
	return nil
}

// GetStatus returns the current state of all servers
func (m *Manager) GetStatus() []*model.ServerState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	states := make([]*model.ServerState, 0, len(m.servers))
	for _, inst := range m.servers {
		states = append(states, cloneState(inst.state))
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
	var states []*model.ServerState
	for _, inst := range m.servers {
		if inst.running || inst.state.Status == "starting" {
			inst.cancel()
			inst.running = false
			inst.state.Status = "stopped"
			states = append(states, cloneState(inst.state))
		}
	}
	m.mu.Unlock()
	for _, state := range states {
		m.emitStatus(state)
	}
}

func (m *Manager) markRunning(serverType string, inst *serverInstance, state *model.ServerState) {
	m.mu.Lock()
	if current, ok := m.servers[serverType]; !ok || current != inst {
		m.mu.Unlock()
		return
	}
	if inst.state.Status == "stopped" {
		m.mu.Unlock()
		return
	}
	inst.running = true
	*inst.state = *state
	inst.state.StartedAt = time.Now()
	copy := cloneState(inst.state)
	m.mu.Unlock()
	m.emitStatus(copy)
}

func cloneState(state *model.ServerState) *model.ServerState {
	copy := *state
	return &copy
}

func (m *Manager) emitStatus(state *model.ServerState) {
	if m.onStatus != nil {
		m.onStatus(state)
	}
}
