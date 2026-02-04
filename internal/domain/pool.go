package domain

import (
	"sync"
)

type ServerPool struct {
	servers []*Server
	mu      sync.RWMutex
}

func NewServerPool() *ServerPool {
	return &ServerPool{
		servers: make([]*Server, 0),
	}
}

func (p *ServerPool) AddServer(server *Server) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.servers = append(p.servers, server)
}

func (p *ServerPool) GetServers() []*Server {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]*Server, len(p.servers))
	copy(result, p.servers)
	return result
}

func (p *ServerPool) GetHealthyServers() []*Server {
	p.mu.RLock()
	defer p.mu.RUnlock()

	healthy := make([]*Server, 0, len(p.servers))
	for _, s := range p.servers {
		if s.IsAlive() {
			healthy = append(healthy, s)
		}
	}
	return healthy
}

func (p *ServerPool) MarkServerStatus(serverURL string, alive bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, s := range p.servers {
		if s.URL.String() == serverURL {
			s.SetAlive(alive)
			return
		}
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
