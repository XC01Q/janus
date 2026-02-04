package balancer_test

import (
	"fmt"
	"testing"

	"janus/internal/balancer"
	"janus/internal/domain"
)

func createBenchmarkPool(count int) *domain.ServerPool {
	pool := domain.NewServerPool()
	for i := 0; i < count; i++ {
		server, _ := domain.NewServer(fmt.Sprintf("http://localhost:80%02d", i), i%10+1)
		pool.AddServer(server)
	}
	return pool
}

func BenchmarkRoundRobin_GetNextServer_10(b *testing.B) {
	rr := balancer.NewRoundRobin()
	pool := createBenchmarkPool(10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr.GetNextServer(pool)
	}
}

func BenchmarkRoundRobin_GetNextServer_100(b *testing.B) {
	rr := balancer.NewRoundRobin()
	pool := createBenchmarkPool(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr.GetNextServer(pool)
	}
}

func BenchmarkWeighted_GetNextServer_10(b *testing.B) {
	w := balancer.NewWeighted()
	pool := createBenchmarkPool(10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.GetNextServer(pool)
	}
}

func BenchmarkWeighted_GetNextServer_100(b *testing.B) {
	w := balancer.NewWeighted()
	pool := createBenchmarkPool(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.GetNextServer(pool)
	}
}

func BenchmarkRoundRobin_Parallel(b *testing.B) {
	rr := balancer.NewRoundRobin()
	pool := createBenchmarkPool(50)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rr.GetNextServer(pool)
		}
	})
}

func BenchmarkWeighted_Parallel(b *testing.B) {
	w := balancer.NewWeighted()
	pool := createBenchmarkPool(50)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w.GetNextServer(pool)
		}
	})
}
