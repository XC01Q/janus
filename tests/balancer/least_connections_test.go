package balancer_test

import (
	"testing"

	"janus/internal/balancer"
	"janus/internal/domain"
)

func TestLeastConnectionsName(t *testing.T) {
	lc := balancer.NewLeastConnections()

	if lc.Name() != "least_connections" {
		t.Errorf("name = %s, want least_connections", lc.Name())
	}
}

func TestLeastConnectionsEmptyPool(t *testing.T) {
	lc := balancer.NewLeastConnections()
	pool := domain.NewServerPool()

	server := lc.GetNextServer(pool)
	if server != nil {
		t.Error("expected nil for empty pool")
	}
}

func TestLeastConnectionsSingleServer(t *testing.T) {
	lc := balancer.NewLeastConnections()
	pool := domain.NewServerPool()

	server, _ := domain.NewServer("http://localhost:8081", 1)
	pool.AddServer(server)

	for i := 0; i < 5; i++ {
		selected := lc.GetNextServer(pool)
		if selected == nil {
			t.Fatal("expected server, got nil")
		}
		if selected.URL.String() != "http://localhost:8081" {
			t.Errorf("unexpected server URL: %s", selected.URL)
		}
	}
}

func TestLeastConnectionsSelection(t *testing.T) {
	lc := balancer.NewLeastConnections()
	pool := domain.NewServerPool()

	server1, _ := domain.NewServer("http://localhost:8081", 1)
	server2, _ := domain.NewServer("http://localhost:8082", 1)
	server3, _ := domain.NewServer("http://localhost:8083", 1)

	pool.AddServer(server1)
	pool.AddServer(server2)
	pool.AddServer(server3)

	server1.IncrementConnections()
	server1.IncrementConnections()

	server2.IncrementConnections()

	selected := lc.GetNextServer(pool)
	if selected == nil {
		t.Fatal("expected server, got nil")
	}

	if selected.URL.String() != "http://localhost:8083" {
		t.Errorf("expected server with 0 connections, got %s with %d connections",
			selected.URL, selected.GetConnections())
	}
}

func TestLeastConnectionsMultipleSelections(t *testing.T) {
	lc := balancer.NewLeastConnections()
	pool := domain.NewServerPool()

	server1, _ := domain.NewServer("http://localhost:8081", 1)
	server2, _ := domain.NewServer("http://localhost:8082", 1)

	pool.AddServer(server1)
	pool.AddServer(server2)

	selected1 := lc.GetNextServer(pool)
	selected1.IncrementConnections()

	selected2 := lc.GetNextServer(pool)

	if selected2.URL.String() == selected1.URL.String() {
		t.Error("second selection should choose different server with fewer connections")
	}
}

func TestLeastConnectionsSkipsUnhealthy(t *testing.T) {
	lc := balancer.NewLeastConnections()
	pool := domain.NewServerPool()

	server1, _ := domain.NewServer("http://localhost:8081", 1)
	server2, _ := domain.NewServer("http://localhost:8082", 1)

	pool.AddServer(server1)
	pool.AddServer(server2)

	server1.SetAlive(false)

	for i := 0; i < 10; i++ {
		server2.IncrementConnections()
	}

	selected := lc.GetNextServer(pool)
	if selected == nil {
		t.Fatal("expected server, got nil")
	}

	if selected.URL.String() != "http://localhost:8082" {
		t.Errorf("expected healthy server, got %s", selected.URL)
	}
}

func TestLeastConnectionsAllUnhealthy(t *testing.T) {
	lc := balancer.NewLeastConnections()
	pool := domain.NewServerPool()

	server, _ := domain.NewServer("http://localhost:8081", 1)
	server.SetAlive(false)
	pool.AddServer(server)

	result := lc.GetNextServer(pool)
	if result != nil {
		t.Error("expected nil when all servers are unhealthy")
	}
}

func TestLeastConnectionsEqualConnections(t *testing.T) {
	lc := balancer.NewLeastConnections()
	pool := domain.NewServerPool()

	server1, _ := domain.NewServer("http://localhost:8081", 1)
	server2, _ := domain.NewServer("http://localhost:8082", 1)
	server3, _ := domain.NewServer("http://localhost:8083", 1)

	pool.AddServer(server1)
	pool.AddServer(server2)
	pool.AddServer(server3)

	for _, s := range pool.GetServers() {
		for i := 0; i < 5; i++ {
			s.IncrementConnections()
		}
	}

	selected := lc.GetNextServer(pool)
	if selected == nil {
		t.Fatal("expected server, got nil")
	}

	if selected.URL.String() != "http://localhost:8081" {
		t.Logf("Note: selected %s (any is valid when equal)", selected.URL)
	}
}
