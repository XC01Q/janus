package domain_test

import (
	"sync"
	"testing"

	"janus/internal/domain"
)

func TestNewServerPool(t *testing.T) {
	pool := domain.NewServerPool()

	if pool.Size() != 0 {
		t.Errorf("new pool size = %d, want 0", pool.Size())
	}

	if pool.HealthyCount() != 0 {
		t.Errorf("new pool healthy count = %d, want 0", pool.HealthyCount())
	}
}

func TestServerPoolAddServer(t *testing.T) {
	pool := domain.NewServerPool()

	server1, _ := domain.NewServer("http://localhost:8081", 1)
	server2, _ := domain.NewServer("http://localhost:8082", 2)

	pool.AddServer(server1)
	if pool.Size() != 1 {
		t.Errorf("pool size after first add = %d, want 1", pool.Size())
	}

	pool.AddServer(server2)
	if pool.Size() != 2 {
		t.Errorf("pool size after second add = %d, want 2", pool.Size())
	}
}

func TestServerPoolGetServers(t *testing.T) {
	pool := domain.NewServerPool()

	server1, _ := domain.NewServer("http://localhost:8081", 1)
	server2, _ := domain.NewServer("http://localhost:8082", 2)

	pool.AddServer(server1)
	pool.AddServer(server2)

	servers := pool.GetServers()
	if len(servers) != 2 {
		t.Errorf("GetServers length = %d, want 2", len(servers))
	}

	servers[0] = nil
	serversAgain := pool.GetServers()
	if serversAgain[0] == nil {
		t.Error("GetServers should return a copy of the slice")
	}
}

func TestServerPoolGetHealthyServers(t *testing.T) {
	pool := domain.NewServerPool()

	server1, _ := domain.NewServer("http://localhost:8081", 1)
	server2, _ := domain.NewServer("http://localhost:8082", 2)
	server3, _ := domain.NewServer("http://localhost:8083", 1)

	pool.AddServer(server1)
	pool.AddServer(server2)
	pool.AddServer(server3)

	healthy := pool.GetHealthyServers()
	if len(healthy) != 3 {
		t.Errorf("healthy count = %d, want 3", len(healthy))
	}

	server2.SetAlive(false)

	healthy = pool.GetHealthyServers()
	if len(healthy) != 2 {
		t.Errorf("healthy count after one down = %d, want 2", len(healthy))
	}

	server1.SetAlive(false)
	server3.SetAlive(false)

	healthy = pool.GetHealthyServers()
	if len(healthy) != 0 {
		t.Errorf("healthy count when all down = %d, want 0", len(healthy))
	}
}

func TestServerPoolMarkServerStatus(t *testing.T) {
	pool := domain.NewServerPool()

	server, _ := domain.NewServer("http://localhost:8081", 1)
	pool.AddServer(server)

	if !server.IsAlive() {
		t.Error("server should be alive initially")
	}

	pool.MarkServerStatus("http://localhost:8081", false)
	if server.IsAlive() {
		t.Error("server should not be alive after MarkServerStatus(false)")
	}

	pool.MarkServerStatus("http://localhost:8081", true)
	if !server.IsAlive() {
		t.Error("server should be alive after MarkServerStatus(true)")
	}

	pool.MarkServerStatus("http://localhost:9999", false)
}

func TestServerPoolHealthyCount(t *testing.T) {
	pool := domain.NewServerPool()

	server1, _ := domain.NewServer("http://localhost:8081", 1)
	server2, _ := domain.NewServer("http://localhost:8082", 2)

	pool.AddServer(server1)
	pool.AddServer(server2)

	if pool.HealthyCount() != 2 {
		t.Errorf("healthy count = %d, want 2", pool.HealthyCount())
	}

	server1.SetAlive(false)
	if pool.HealthyCount() != 1 {
		t.Errorf("healthy count after one down = %d, want 1", pool.HealthyCount())
	}
}

func TestServerPoolConcurrent(t *testing.T) {
	pool := domain.NewServerPool()

	var wg sync.WaitGroup
	iterations := 100

	for i := 0; i < iterations; i++ {
		wg.Add(3)

		go func(idx int) {
			defer wg.Done()
			server, _ := domain.NewServer("http://localhost:8080", 1)
			pool.AddServer(server)
		}(i)

		go func() {
			defer wg.Done()
			_ = pool.GetServers()
		}()

		go func() {
			defer wg.Done()
			_ = pool.GetHealthyServers()
		}()
	}

	wg.Wait()

	if pool.Size() != iterations {
		t.Errorf("pool size = %d, want %d", pool.Size(), iterations)
	}
}
