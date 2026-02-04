package balancer_test

import (
	"strconv"
	"testing"

	"janus/internal/balancer"
	"janus/internal/domain"
)

func createTestPoolRR(count int) *domain.ServerPool {
	pool := domain.NewServerPool()

	for i := 0; i < count; i++ {
		server, _ := domain.NewServer("http://localhost:"+strconv.Itoa(8081+i), 1)
		pool.AddServer(server)
	}

	return pool
}

func TestRoundRobinName(t *testing.T) {
	rr := balancer.NewRoundRobin()

	if rr.Name() != "round_robin" {
		t.Errorf("name = %s, want round_robin", rr.Name())
	}
}

func TestRoundRobinEmptyPool(t *testing.T) {
	rr := balancer.NewRoundRobin()
	pool := domain.NewServerPool()

	server := rr.GetNextServer(pool)
	if server != nil {
		t.Error("expected nil for empty pool")
	}
}

func TestRoundRobinSingleServer(t *testing.T) {
	rr := balancer.NewRoundRobin()
	pool := createTestPoolRR(1)

	for i := 0; i < 5; i++ {
		server := rr.GetNextServer(pool)
		if server == nil {
			t.Fatal("expected server, got nil")
		}
		if server.URL.String() != "http://localhost:8081" {
			t.Errorf("unexpected server URL: %s", server.URL)
		}
	}
}

func TestRoundRobinDistribution(t *testing.T) {
	rr := balancer.NewRoundRobin()
	pool := createTestPoolRR(3)

	counts := make(map[string]int)

	for i := 0; i < 9; i++ {
		server := rr.GetNextServer(pool)
		if server == nil {
			t.Fatal("expected server, got nil")
		}
		counts[server.URL.String()]++
	}

	expected := 3
	for url, count := range counts {
		if count != expected {
			t.Errorf("server %s got %d requests, want %d", url, count, expected)
		}
	}
}

func TestRoundRobinSkipsUnhealthy(t *testing.T) {
	rr := balancer.NewRoundRobin()
	pool := domain.NewServerPool()

	server1, _ := domain.NewServer("http://localhost:8081", 1)
	server2, _ := domain.NewServer("http://localhost:8082", 1)
	server3, _ := domain.NewServer("http://localhost:8083", 1)

	pool.AddServer(server1)
	pool.AddServer(server2)
	pool.AddServer(server3)

	pool.SetServerStatus(server2, false)

	seen := make(map[string]bool)
	for i := 0; i < 10; i++ {
		server := rr.GetNextServer(pool)
		seen[server.URL.String()] = true
	}

	if seen["http://localhost:8082"] {
		t.Error("unhealthy server should not receive requests")
	}

	if !seen["http://localhost:8081"] || !seen["http://localhost:8083"] {
		t.Error("healthy servers should receive requests")
	}
}

func TestRoundRobinAllUnhealthy(t *testing.T) {
	rr := balancer.NewRoundRobin()
	pool := domain.NewServerPool()

	server, _ := domain.NewServer("http://localhost:8081", 1)
	pool.SetServerStatus(server, false)
	pool.AddServer(server)

	result := rr.GetNextServer(pool)
	if result != nil {
		t.Error("expected nil when all servers are unhealthy")
	}
}
