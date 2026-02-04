package server

import (
	"log"
	"net/http"
	"net/http/httputil"

	"janus/internal/balancer"
	"janus/internal/domain"
)

type ProxyHandler struct {
	pool     *domain.ServerPool
	strategy balancer.Strategy
}

func NewProxyHandler(pool *domain.ServerPool, strategy balancer.Strategy) *ProxyHandler {
	return &ProxyHandler{
		pool:     pool,
		strategy: strategy,
	}
}

func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server := h.strategy.GetNextServer(h.pool)
	if server == nil {
		log.Printf("[ERROR] No available servers")
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}

	server.IncrementConnections()
	defer server.DecrementConnections()

	log.Printf("[INFO] Forwarding request to %s (connections: %d, strategy: %s)",
		server.URL, server.GetConnections(), h.strategy.Name())

	proxy := h.createReverseProxy(server)
	proxy.ServeHTTP(w, r)
}

func (h *ProxyHandler) createReverseProxy(server *domain.Server) *httputil.ReverseProxy {
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = server.URL.Scheme
			req.URL.Host = server.URL.Host

			if req.Header.Get("X-Forwarded-Host") == "" {
				req.Header.Set("X-Forwarded-Host", req.Host)
			}

			if clientIP := req.RemoteAddr; clientIP != "" {
				if prior := req.Header.Get("X-Forwarded-For"); prior != "" {
					req.Header.Set("X-Forwarded-For", prior+", "+clientIP)
				} else {
					req.Header.Set("X-Forwarded-For", clientIP)
				}
			}

			if req.TLS != nil {
				req.Header.Set("X-Forwarded-Proto", "https")
			} else {
				req.Header.Set("X-Forwarded-Proto", "http")
			}

			req.Host = server.URL.Host

			log.Printf("[DEBUG] Proxying: %s %s -> %s%s",
				req.Method, req.URL.Path, server.URL, req.URL.Path)
		},

		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("[ERROR] Proxy error for %s: %v", server.URL, err)
			server.SetAlive(false)
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
		},
	}

	return proxy
}
