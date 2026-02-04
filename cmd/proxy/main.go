package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"janus/internal/balancer"
	"janus/internal/config"
	"janus/internal/domain"
	"janus/internal/server"
)

func main() {
	configPath := flag.String("config", "config.json", "path to configuration file")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("[INFO] Starting Janus Reverse Proxy...")

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("[FATAL] Failed to load config: %v", err)
	}

	log.Printf("[INFO] Configuration loaded: port=%d, strategy=%s, servers=%d, health_check=%ds",
		cfg.Port, cfg.Strategy, len(cfg.Servers), cfg.HealthCheckTime)

	pool := createServerPool(cfg)

	strategy, err := balancer.NewStrategy(cfg.Strategy)
	if err != nil {
		log.Fatalf("[FATAL] Failed to create strategy: %v", err)
	}

	log.Printf("[INFO] Using balancing strategy: %s", strategy.Name())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	healthChecker := server.NewHealthChecker(pool, time.Duration(cfg.HealthCheckTime)*time.Second)
	healthChecker.Start(ctx)

	proxyHandler := server.NewProxyHandler(pool, strategy)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      proxyHandler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("[INFO] Proxy server listening on :%d", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[FATAL] Server error: %v", err)
		}
	}()

	gracefulShutdown(httpServer, cancel)
}

func createServerPool(cfg *config.Config) *domain.ServerPool {
	pool := domain.NewServerPool()

	for _, serverCfg := range cfg.Servers {
		srv, err := domain.NewServer(serverCfg.URL, serverCfg.Weight)
		if err != nil {
			log.Printf("[WARN] Invalid server URL %s: %v", serverCfg.URL, err)
			continue
		}

		pool.AddServer(srv)
		log.Printf("[INFO] Added server: %s (weight: %d)", serverCfg.URL, serverCfg.Weight)
	}

	if pool.Size() == 0 {
		log.Fatal("[FATAL] No valid servers configured")
	}

	return pool
}

func gracefulShutdown(srv *http.Server, cancel context.CancelFunc) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Printf("[INFO] Received signal %v, shutting down...", sig)

	cancel()

	ctx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[ERROR] Server shutdown error: %v", err)
	}

	log.Println("[INFO] Server stopped gracefully")
}
