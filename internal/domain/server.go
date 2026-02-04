package domain

import (
	"net/url"
	"sync"
	"sync/atomic"
)

type Server struct {
	URL         *url.URL
	Weight      int
	alive       bool
	mu          sync.RWMutex
	connections atomic.Int64
}

func NewServer(rawURL string, weight int) (*Server, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	if weight < 1 {
		weight = 1
	}

	return &Server{
		URL:    parsedURL,
		Weight: weight,
		alive:  true,
	}, nil
}

func (s *Server) SetAlive(alive bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.alive = alive
}

func (s *Server) IsAlive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.alive
}

func (s *Server) IncrementConnections() {
	s.connections.Add(1)
}

func (s *Server) DecrementConnections() {
	s.connections.Add(-1)
}

func (s *Server) GetConnections() int64 {
	return s.connections.Load()
}
