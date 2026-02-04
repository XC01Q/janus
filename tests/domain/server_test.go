package domain_test

import (
	"sync"
	"testing"

	"janus/internal/domain"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		weight     int
		wantWeight int
		wantErr    bool
	}{
		{
			name:       "valid URL with weight",
			url:        "http://localhost:8080",
			weight:     2,
			wantWeight: 2,
			wantErr:    false,
		},
		{
			name:       "valid URL with zero weight defaults to 1",
			url:        "http://localhost:8081",
			weight:     0,
			wantWeight: 1,
			wantErr:    false,
		},
		{
			name:       "valid URL with negative weight defaults to 1",
			url:        "http://localhost:8082",
			weight:     -5,
			wantWeight: 1,
			wantErr:    false,
		},
		{
			name:    "invalid URL",
			url:     "://invalid",
			weight:  1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := domain.NewServer(tt.url, tt.weight)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if server.Weight != tt.wantWeight {
				t.Errorf("weight = %d, want %d", server.Weight, tt.wantWeight)
			}

			if !server.IsAlive() {
				t.Error("new server should be alive by default")
			}

			if server.GetConnections() != 0 {
				t.Errorf("connections = %d, want 0", server.GetConnections())
			}
		})
	}
}

func TestServerAliveStatus(t *testing.T) {
	server, err := domain.NewServer("http://localhost:8080", 1)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	if !server.IsAlive() {
		t.Error("server should be alive initially")
	}

	server.SetAlive(false)
	if server.IsAlive() {
		t.Error("server should not be alive after SetAlive(false)")
	}

	server.SetAlive(true)
	if !server.IsAlive() {
		t.Error("server should be alive after SetAlive(true)")
	}
}

func TestServerAliveStatusConcurrent(t *testing.T) {
	server, _ := domain.NewServer("http://localhost:8080", 1)

	var wg sync.WaitGroup
	iterations := 1000

	for i := 0; i < iterations; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()
			server.SetAlive(true)
			_ = server.IsAlive()
		}()

		go func() {
			defer wg.Done()
			server.SetAlive(false)
			_ = server.IsAlive()
		}()
	}

	wg.Wait()
}

func TestServerConnections(t *testing.T) {
	server, _ := domain.NewServer("http://localhost:8080", 1)

	if c := server.GetConnections(); c != 0 {
		t.Errorf("initial connections = %d, want 0", c)
	}

	server.IncrementConnections()
	if c := server.GetConnections(); c != 1 {
		t.Errorf("after increment connections = %d, want 1", c)
	}

	server.IncrementConnections()
	if c := server.GetConnections(); c != 2 {
		t.Errorf("after second increment connections = %d, want 2", c)
	}

	server.DecrementConnections()
	if c := server.GetConnections(); c != 1 {
		t.Errorf("after decrement connections = %d, want 1", c)
	}
}

func TestServerConnectionsConcurrent(t *testing.T) {
	server, _ := domain.NewServer("http://localhost:8080", 1)

	var wg sync.WaitGroup
	iterations := 10000

	for i := 0; i < iterations; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()
			server.IncrementConnections()
		}()

		go func() {
			defer wg.Done()
			server.DecrementConnections()
		}()
	}

	wg.Wait()

	if c := server.GetConnections(); c != 0 {
		t.Errorf("final connections = %d, want 0", c)
	}
}
