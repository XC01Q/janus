package balancer

import (
	"janus/internal/domain"
	"sync"
)

type Weighted struct {
	mu             sync.Mutex
	currentWeights map[*domain.Server]int
	initialized    bool
}

func NewWeighted() *Weighted {
	return &Weighted{
		currentWeights: make(map[*domain.Server]int),
	}
}

func (w *Weighted) GetNextServer(pool *domain.ServerPool) *domain.Server {
	w.mu.Lock()
	defer w.mu.Unlock()

	servers := pool.GetHealthyServers()
	if len(servers) == 0 {
		return nil
	}

	if len(w.currentWeights) > len(servers) {
		healthyMap := make(map[*domain.Server]bool, len(servers))
		for _, s := range servers {
			healthyMap[s] = true
		}

		for s := range w.currentWeights {
			if !healthyMap[s] {
				delete(w.currentWeights, s)
			}
		}
	}

	totalWeight := 0
	for _, s := range servers {
		totalWeight += s.Weight
	}

	for _, s := range servers {
		if _, exists := w.currentWeights[s]; !exists {
			w.currentWeights[s] = 0
		}
		w.currentWeights[s] += s.Weight
	}

	var selected *domain.Server
	maxWeight := 0

	for _, s := range servers {
		if w.currentWeights[s] > maxWeight {
			maxWeight = w.currentWeights[s]
			selected = s
		}
	}

	if selected == nil {
		return servers[0]
	}

	w.currentWeights[selected] -= totalWeight

	return selected
}

func (w *Weighted) Name() string {
	return "weighted"
}

func (w *Weighted) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.currentWeights = make(map[*domain.Server]int)
	w.initialized = false
}
