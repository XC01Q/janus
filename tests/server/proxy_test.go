package server_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"janus/internal/balancer"
	"janus/internal/domain"
	"janus/internal/server"
)

func TestProxyHandlerNoServers(t *testing.T) {
	pool := domain.NewServerPool()
	strategy := balancer.NewRoundRobin()
	handler := server.NewProxyHandler(pool, strategy)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestProxyHandlerAllUnhealthy(t *testing.T) {
	pool := domain.NewServerPool()
	srv, _ := domain.NewServer("http://localhost:9999", 1)
	srv.SetAlive(false)
	pool.AddServer(srv)

	strategy := balancer.NewRoundRobin()
	handler := server.NewProxyHandler(pool, strategy)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestProxyHandlerForwardsRequest(t *testing.T) {
	backendHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Server", "test-server")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from server"))
	})

	backendServer := httptest.NewServer(backendHandler)
	defer backendServer.Close()

	pool := domain.NewServerPool()
	srv, _ := domain.NewServer(backendServer.URL, 1)
	pool.AddServer(srv)

	strategy := balancer.NewRoundRobin()
	handler := server.NewProxyHandler(pool, strategy)

	req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if rec.Header().Get("X-Server") != "test-server" {
		t.Error("response should include server header")
	}

	body, _ := io.ReadAll(rec.Body)
	if string(body) != "Hello from server" {
		t.Errorf("body = %s, want 'Hello from server'", string(body))
	}
}

func TestProxyHandlerForwardsHeaders(t *testing.T) {
	var receivedHeaders http.Header

	backendHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	})

	backendServer := httptest.NewServer(backendHandler)
	defer backendServer.Close()

	pool := domain.NewServerPool()
	srv, _ := domain.NewServer(backendServer.URL, 1)
	pool.AddServer(srv)

	strategy := balancer.NewRoundRobin()
	handler := server.NewProxyHandler(pool, strategy)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Custom-Header", "custom-value")
	req.Header.Set("Authorization", "Bearer token123")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if receivedHeaders.Get("X-Custom-Header") != "custom-value" {
		t.Error("custom header should be forwarded")
	}

	if receivedHeaders.Get("Authorization") != "Bearer token123" {
		t.Error("authorization header should be forwarded")
	}
}

func TestProxyHandlerAddsForwardedHeaders(t *testing.T) {
	var receivedHeaders http.Header

	backendHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	})

	backendServer := httptest.NewServer(backendHandler)
	defer backendServer.Close()

	pool := domain.NewServerPool()
	srv, _ := domain.NewServer(backendServer.URL, 1)
	pool.AddServer(srv)

	strategy := balancer.NewRoundRobin()
	handler := server.NewProxyHandler(pool, strategy)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Host = "original-host.com"

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if receivedHeaders.Get("X-Forwarded-Host") != "original-host.com" {
		t.Errorf("X-Forwarded-Host = %s, want original-host.com",
			receivedHeaders.Get("X-Forwarded-Host"))
	}

	if receivedHeaders.Get("X-Forwarded-Proto") == "" {
		t.Error("X-Forwarded-Proto should be set")
	}
}

func TestProxyHandlerConnectionCount(t *testing.T) {
	backendHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	backendServer := httptest.NewServer(backendHandler)
	defer backendServer.Close()

	pool := domain.NewServerPool()
	srv, _ := domain.NewServer(backendServer.URL, 1)
	pool.AddServer(srv)

	if srv.GetConnections() != 0 {
		t.Errorf("initial connections = %d, want 0", srv.GetConnections())
	}

	strategy := balancer.NewRoundRobin()
	handler := server.NewProxyHandler(pool, strategy)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if srv.GetConnections() != 0 {
		t.Errorf("connections after request = %d, want 0", srv.GetConnections())
	}
}

func TestProxyHandlerRoundRobinDistribution(t *testing.T) {
	requestCounts := make(map[string]int)

	handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("server1"))
	})
	handler2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("server2"))
	})

	server1 := httptest.NewServer(handler1)
	defer server1.Close()
	server2 := httptest.NewServer(handler2)
	defer server2.Close()

	pool := domain.NewServerPool()
	srv1, _ := domain.NewServer(server1.URL, 1)
	srv2, _ := domain.NewServer(server2.URL, 1)
	pool.AddServer(srv1)
	pool.AddServer(srv2)

	strategy := balancer.NewRoundRobin()
	handler := server.NewProxyHandler(pool, strategy)

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		body, _ := io.ReadAll(rec.Body)
		requestCounts[string(body)]++
	}

	if requestCounts["server1"] != 5 {
		t.Errorf("server1 got %d requests, want 5", requestCounts["server1"])
	}
	if requestCounts["server2"] != 5 {
		t.Errorf("server2 got %d requests, want 5", requestCounts["server2"])
	}
}
