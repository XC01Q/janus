package server

import (
	"context"
	"log"
	"net"
	"time"

	"janus/internal/domain"
)

type HealthChecker struct {
	pool     *domain.ServerPool
	interval time.Duration
	timeout  time.Duration
}

func NewHealthChecker(pool *domain.ServerPool, interval time.Duration) *HealthChecker {
	return &HealthChecker{
		pool:     pool,
		interval: interval,
		timeout:  2 * time.Second,
	}
}

func (h *HealthChecker) Start(ctx context.Context) {
	h.checkAll()

	go func() {
		ticker := time.NewTicker(h.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("[INFO] Health checker stopped")
				return
			case <-ticker.C:
				h.checkAll()
			}
		}
	}()

	log.Printf("[INFO] Health checker started (interval: %v)", h.interval)
}

func (h *HealthChecker) checkAll() {
	servers := h.pool.GetServers()

	for _, server := range servers {
		go h.checkServer(server)
	}
}

func (h *HealthChecker) checkServer(server *domain.Server) {
	address := server.URL.Host

	if server.URL.Port() == "" {
		if server.URL.Scheme == "https" {
			address = address + ":443"
		} else {
			address = address + ":80"
		}
	}

	conn, err := net.DialTimeout("tcp", address, h.timeout)

	wasAlive := server.IsAlive()

	if err != nil {
		h.pool.SetServerStatus(server, false)

		if wasAlive {
			log.Printf("[WARN] Server %s is DOWN: %v", server.URL, err)
		}
		return
	}

	conn.Close()
	h.pool.SetServerStatus(server, true)

	if !wasAlive {
		log.Printf("[INFO] Server %s is UP", server.URL)
	}
}

func (h *HealthChecker) CheckOnce() {
	servers := h.pool.GetServers()

	for _, server := range servers {
		h.checkServer(server)
	}
}
