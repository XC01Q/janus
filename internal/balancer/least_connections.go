package balancer

import (
	"janus/internal/domain"
)

type LeastConnections struct{}

func NewLeastConnections() *LeastConnections {
	return &LeastConnections{}
}

func (l *LeastConnections) GetNextServer(pool *domain.ServerPool) *domain.Server {
	servers := pool.GetHealthyServers()
	if len(servers) == 0 {
		return nil
	}

	var selected *domain.Server
	minConnections := int64(-1)

	for _, s := range servers {
		connections := s.GetConnections()

		if minConnections == -1 || connections < minConnections {
			minConnections = connections
			selected = s
		}
	}

	return selected
}

func (l *LeastConnections) Name() string {
	return "least_connections"
}
