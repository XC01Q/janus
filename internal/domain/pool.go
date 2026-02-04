package domain

import (
	"sync"
	"sync/atomic"
)

type ServerPool struct {
	servers        []*Server
	mu             sync.RWMutex
	healthyServers atomic.Value
}

func NewServerPool() *ServerPool {
	p := &ServerPool{
		servers: make([]*Server, 0),
	}
	p.healthyServers.Store(make([]*Server, 0))
	return p
}

func (p *ServerPool) AddServer(server *Server) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.servers = append(p.servers, server)
	p.refreshHealthyCache()
}

func (p *ServerPool) GetServers() []*Server {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]*Server, len(p.servers))
	copy(result, p.servers)
	return result
}

func (p *ServerPool) GetHealthyServers() []*Server {
	return p.healthyServers.Load().([]*Server)
}

func (p *ServerPool) refreshHealthyCache() {
	healthy := make([]*Server, 0, len(p.servers))
	for _, s := range p.servers {
		if s.IsAlive() {
			healthy = append(healthy, s)
		}
	}
	p.healthyServers.Store(healthy)
}

func (p *ServerPool) SetServerStatus(server *Server, alive bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if server.IsAlive() == alive {
		return
	}

	server.SetAlive(alive)
	p.refreshHealthyCache()
}

func (p *ServerPool) MarkServerStatus(serverURL string, alive bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	changed := false
	for _, s := range p.servers {
		if s.URL.String() == serverURL {
			if s.IsAlive() != alive {
				s.SetAlive(alive)
				changed = true
			}
			break
		}
	}

	if changed {
		p.refreshHealthyCache()
	}
}

func (p *ServerPool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.servers)
}

func (p *ServerPool) HealthyCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	count := 0
	for _, s := range p.servers {
		if s.IsAlive() {
			count++
		}
	}
	return count
}
