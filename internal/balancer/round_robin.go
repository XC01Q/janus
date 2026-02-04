package balancer

import (
	"janus/internal/domain"
	"sync/atomic"
)

type RoundRobin struct {
	current atomic.Uint64
}

func NewRoundRobin() *RoundRobin {
	return &RoundRobin{}
}

func (r *RoundRobin) GetNextServer(pool *domain.ServerPool) *domain.Server {
	servers := pool.GetHealthyServers()
	if len(servers) == 0 {
		return nil
	}

	next := r.current.Add(1)
	idx := (next - 1) % uint64(len(servers))

	return servers[idx]
}

func (r *RoundRobin) Name() string {
	return "round_robin"
}
