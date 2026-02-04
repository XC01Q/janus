package server_test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"janus/internal/domain"
	"janus/internal/server"
)

func TestHealthCheckerCreation(t *testing.T) {
	pool := domain.NewServerPool()
	checker := server.NewHealthChecker(pool, 5*time.Second)

	if checker == nil {
		t.Fatal("expected health checker, got nil")
	}
}

func TestHealthCheckerCheckOnce(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	pool := domain.NewServerPool()
	srv, _ := domain.NewServer(testServer.URL, 1)
	pool.AddServer(srv)

	checker := server.NewHealthChecker(pool, 1*time.Second)
	checker.CheckOnce()

	if !srv.IsAlive() {
		t.Error("server should be alive after successful check")
	}
}

func TestHealthCheckerDetectsDownServer(t *testing.T) {
	pool := domain.NewServerPool()

	srv, _ := domain.NewServer("http://localhost:59999", 1)
	pool.AddServer(srv)

	checker := server.NewHealthChecker(pool, 1*time.Second)
	checker.CheckOnce()

	time.Sleep(100 * time.Millisecond)

	if srv.IsAlive() {
		t.Error("server should not be alive after failed check")
	}
}

func TestHealthCheckerDetectsRecovery(t *testing.T) {
	pool := domain.NewServerPool()

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	srv, _ := domain.NewServer("http://localhost:"+strconv.Itoa(port), 1)
	pool.AddServer(srv)

	checker := server.NewHealthChecker(pool, 1*time.Second)

	checker.CheckOnce()
	time.Sleep(100 * time.Millisecond)

	if srv.IsAlive() {
		t.Error("server should be down initially")
	}

	testServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	newListener, err := net.Listen("tcp", "localhost:"+strconv.Itoa(port))
	if err != nil {
		t.Skip("cannot recreate listener on same port")
	}

	testServer.Listener = newListener
	testServer.Start()
	defer testServer.Close()

	pool2 := domain.NewServerPool()
	srv2, _ := domain.NewServer(testServer.URL, 1)
	srv2.SetAlive(false)
	pool2.AddServer(srv2)

	checker2 := server.NewHealthChecker(pool2, 1*time.Second)

	checker2.CheckOnce()
	time.Sleep(100 * time.Millisecond)

	if !srv2.IsAlive() {
		t.Error("server should recover after server starts")
	}
}

func TestHealthCheckerStart(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	pool := domain.NewServerPool()
	srv, _ := domain.NewServer(testServer.URL, 1)
	srv.SetAlive(false)
	pool.AddServer(srv)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	checker := server.NewHealthChecker(pool, 100*time.Millisecond)
	checker.Start(ctx)

	time.Sleep(200 * time.Millisecond)

	if !srv.IsAlive() {
		t.Error("server should be alive after health check")
	}
}

func TestHealthCheckerStops(t *testing.T) {
	pool := domain.NewServerPool()
	srv, _ := domain.NewServer("http://localhost:59999", 1)
	pool.AddServer(srv)

	ctx, cancel := context.WithCancel(context.Background())

	checker := server.NewHealthChecker(pool, 50*time.Millisecond)
	checker.Start(ctx)

	time.Sleep(100 * time.Millisecond)

	cancel()

	time.Sleep(100 * time.Millisecond)
}

func TestHealthCheckerMultipleServers(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server1.Close()

	pool := domain.NewServerPool()

	srv1, _ := domain.NewServer(server1.URL, 1)
	srv2, _ := domain.NewServer("http://localhost:59999", 1)

	pool.AddServer(srv1)
	pool.AddServer(srv2)

	checker := server.NewHealthChecker(pool, 1*time.Second)
	checker.CheckOnce()

	time.Sleep(500 * time.Millisecond)

	if !srv1.IsAlive() {
		t.Error("srv1 should be alive")
	}

	if srv2.IsAlive() {
		t.Error("srv2 should not be alive")
	}
}
