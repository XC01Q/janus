package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"janus/internal/balancer"
	"janus/internal/domain"
	"janus/internal/server"
)

func BenchmarkProxyHandler_RoundRobin(b *testing.B) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer backend.Close()

	pool := domain.NewServerPool()
	srv, _ := domain.NewServer(backend.URL, 1)
	pool.AddServer(srv)

	strategy := balancer.NewRoundRobin()
	handler := server.NewProxyHandler(pool, strategy)

	req := httptest.NewRequest("GET", "http://localhost:8080/", nil)
	w := &mockResponseWriter{}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(w, req)
	}
}

type mockResponseWriter struct{}

func (m *mockResponseWriter) Header() http.Header         { return http.Header{} }
func (m *mockResponseWriter) Write(p []byte) (int, error) { return len(p), nil }
func (m *mockResponseWriter) WriteHeader(statusCode int)  {}
