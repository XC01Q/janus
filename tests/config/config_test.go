package config_test

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"janus/internal/config"
)

func createTempConfig(t *testing.T, content string) string {
	t.Helper()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp config: %v", err)
	}

	return configPath
}

func TestLoadConfigValid(t *testing.T) {
	content := `{
		"port": 8080,
		"health_check_time": 5,
		"strategy": "round_robin",
		"backends": [
			{"url": "http://localhost:8081", "weight": 1},
			{"url": "http://localhost:8082", "weight": 2}
		]
	}`

	configPath := createTempConfig(t, content)
	cfg, err := config.LoadConfig(configPath)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("port = %d, want 8080", cfg.Port)
	}

	if cfg.HealthCheckTime != 5 {
		t.Errorf("health_check_time = %d, want 5", cfg.HealthCheckTime)
	}

	if cfg.Strategy != "round_robin" {
		t.Errorf("strategy = %s, want round_robin", cfg.Strategy)
	}

	if len(cfg.Servers) != 2 {
		t.Errorf("servers count = %d, want 2", len(cfg.Servers))
	}

	if cfg.Servers[1].Weight != 2 {
		t.Errorf("second server weight = %d, want 2", cfg.Servers[1].Weight)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	content := `{
		"backends": [
			{"url": "http://localhost:8081"}
		]
	}`

	configPath := createTempConfig(t, content)
	cfg, err := config.LoadConfig(configPath)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != config.DefaultPort {
		t.Errorf("default port = %d, want %d", cfg.Port, config.DefaultPort)
	}

	if cfg.HealthCheckTime != config.DefaultHealthCheckTime {
		t.Errorf("default health_check_time = %d, want %d",
			cfg.HealthCheckTime, config.DefaultHealthCheckTime)
	}

	if cfg.Strategy != config.DefaultStrategy {
		t.Errorf("default strategy = %s, want %s", cfg.Strategy, config.DefaultStrategy)
	}

	if cfg.Servers[0].Weight != 1 {
		t.Errorf("default weight = %d, want 1", cfg.Servers[0].Weight)
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := config.LoadConfig("/nonexistent/path/config.json")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	content := `{invalid json}`
	configPath := createTempConfig(t, content)

	_, err := config.LoadConfig(configPath)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestLoadConfigInvalidPort(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{"negative port", -1, true},
		{"zero port is defaulted", 0, false},
		{"port too high", 70000, true},
		{"valid port", 3000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := `{
				"port": ` + strconv.Itoa(tt.port) + `,
				"backends": [{"url": "http://localhost:8081"}]
			}`

			configPath := createTempConfig(t, content)
			_, err := config.LoadConfig(configPath)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestLoadConfigInvalidStrategy(t *testing.T) {
	content := `{
		"strategy": "unknown_strategy",
		"backends": [{"url": "http://localhost:8081"}]
	}`

	configPath := createTempConfig(t, content)
	_, err := config.LoadConfig(configPath)

	if err == nil {
		t.Error("expected error for unknown strategy, got nil")
	}
}

func TestLoadConfigValidStrategies(t *testing.T) {
	strategies := []string{"round_robin", "weighted", "least_connections"}

	for _, strategy := range strategies {
		t.Run(strategy, func(t *testing.T) {
			content := `{
				"strategy": "` + strategy + `",
				"backends": [{"url": "http://localhost:8081"}]
			}`

			configPath := createTempConfig(t, content)
			cfg, err := config.LoadConfig(configPath)

			if err != nil {
				t.Errorf("strategy %s should be valid: %v", strategy, err)
			}

			if cfg.Strategy != strategy {
				t.Errorf("strategy = %s, want %s", cfg.Strategy, strategy)
			}
		})
	}
}

func TestLoadConfigNoServers(t *testing.T) {
	content := `{
		"backends": []
	}`

	configPath := createTempConfig(t, content)
	_, err := config.LoadConfig(configPath)

	if err == nil {
		t.Error("expected error for empty servers, got nil")
	}
}

func TestLoadConfigEmptyServerURL(t *testing.T) {
	content := `{
		"backends": [{"url": "", "weight": 1}]
	}`

	configPath := createTempConfig(t, content)
	_, err := config.LoadConfig(configPath)

	if err == nil {
		t.Error("expected error for empty server URL, got nil")
	}
}
