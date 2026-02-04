package balancer_test

import (
	"reflect"
	"testing"

	"janus/internal/balancer"
	"janus/internal/domain"
)

func TestWeightedName(t *testing.T) {
	w := balancer.NewWeighted()

	if w.Name() != "weighted" {
		t.Errorf("name = %s, want weighted", w.Name())
	}
}

func TestWeightedEmptyPool(t *testing.T) {
	w := balancer.NewWeighted()
	pool := domain.NewServerPool()

	server := w.GetNextServer(pool)
	if server != nil {
		t.Error("expected nil for empty pool")
	}
}

func TestWeightedSingleServer(t *testing.T) {
	w := balancer.NewWeighted()
	pool := domain.NewServerPool()

	server, _ := domain.NewServer("http://localhost:8081", 5)
	pool.AddServer(server)

	for i := 0; i < 5; i++ {
		selected := w.GetNextServer(pool)
		if selected == nil {
			t.Fatal("expected server, got nil")
		}
		if selected.URL.String() != "http://localhost:8081" {
			t.Errorf("unexpected server URL: %s", selected.URL)
		}
	}
}

func TestWeightedDistribution(t *testing.T) {
	w := balancer.NewWeighted()
	pool := domain.NewServerPool()

	server1, _ := domain.NewServer("http://localhost:8081", 1)
	server2, _ := domain.NewServer("http://localhost:8082", 2)
	server3, _ := domain.NewServer("http://localhost:8083", 1)

	pool.AddServer(server1)
	pool.AddServer(server2)
	pool.AddServer(server3)

	counts := make(map[string]int)
	totalRequests := 400

	for i := 0; i < totalRequests; i++ {
		selected := w.GetNextServer(pool)
		if selected == nil {
			t.Fatal("expected server, got nil")
		}
		counts[selected.URL.String()]++
	}

	count1 := counts["http://localhost:8081"]
	count2 := counts["http://localhost:8082"]
	count3 := counts["http://localhost:8083"]

	if count2 <= count1 || count2 <= count3 {
		t.Errorf("weighted distribution incorrect: %d:%d:%d (expected ratio ~1:2:1)",
			count1, count2, count3)
	}

	expected1 := totalRequests / 4
	expected2 := totalRequests / 2

	tolerance := totalRequests / 10

	if abs(count1-expected1) > tolerance {
		t.Errorf("server 1 count = %d, expected ~%d", count1, expected1)
	}

	if abs(count2-expected2) > tolerance {
		t.Errorf("server 2 count = %d, expected ~%d", count2, expected2)
	}
}

func TestWeightedSkipsUnhealthy(t *testing.T) {
	w := balancer.NewWeighted()
	pool := domain.NewServerPool()

	server1, _ := domain.NewServer("http://localhost:8081", 1)
	server2, _ := domain.NewServer("http://localhost:8082", 10)
	server3, _ := domain.NewServer("http://localhost:8083", 1)

	pool.AddServer(server1)
	pool.AddServer(server2)
	pool.AddServer(server3)

	pool.SetServerStatus(server2, false)

	for i := 0; i < 20; i++ {
		selected := w.GetNextServer(pool)
		if selected == nil {
			t.Fatal("expected server, got nil")
		}
		if selected.URL.String() == "http://localhost:8082" {
			t.Error("unhealthy server should not receive requests")
		}
	}
}

func TestWeightedAllUnhealthy(t *testing.T) {
	w := balancer.NewWeighted()
	pool := domain.NewServerPool()

	server, _ := domain.NewServer("http://localhost:8081", 1)
	pool.SetServerStatus(server, false)
	pool.AddServer(server)

	result := w.GetNextServer(pool)
	if result != nil {
		t.Error("expected nil when all servers are unhealthy")
	}
}

func TestWeightedReset(t *testing.T) {
	w := balancer.NewWeighted()
	pool := domain.NewServerPool()

	server, _ := domain.NewServer("http://localhost:8081", 1)
	pool.AddServer(server)

	for i := 0; i < 5; i++ {
		w.GetNextServer(pool)
	}

	w.Reset()

	selected := w.GetNextServer(pool)
	if selected == nil {
		t.Error("expected server after reset, got nil")
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func TestWeightedCleanup(t *testing.T) {
	w := balancer.NewWeighted()
	pool := domain.NewServerPool()

	s1, _ := domain.NewServer("http://localhost:8081", 1)
	s2, _ := domain.NewServer("http://localhost:8082", 1)
	pool.AddServer(s1)
	pool.AddServer(s2)

	w.GetNextServer(pool)

	if getMapSize(w) != 2 {
		t.Fatalf("expected map size 2, got %d", getMapSize(w))
	}

	pool.SetServerStatus(s2, false)

	w.GetNextServer(pool)

	if getMapSize(w) != 1 {
		t.Errorf("expected map size 1 after cleanup, got %d", getMapSize(w))
	}
}

func getMapSize(w *balancer.Weighted) int {
	val := reflect.Indirect(reflect.ValueOf(w))
	field := val.FieldByName("currentWeights")
	return field.Len()
}
