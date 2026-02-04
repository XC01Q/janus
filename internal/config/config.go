package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

const (
	DefaultPort            = 8080
	DefaultHealthCheckTime = 5
	DefaultStrategy        = "round_robin"
)

var ValidStrategies = map[string]bool{
	"round_robin":       true,
	"weighted":          true,
	"least_connections": true,
}

type Config struct {
	Port            int            `json:"port"`
	HealthCheckTime int            `json:"health_check_time"`
	Strategy        string         `json:"strategy"`
	Servers         []ServerConfig `json:"backends"`
}

type ServerConfig struct {
	URL    string `json:"url"`
	Weight int    `json:"weight"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	cfg.applyDefaults()

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Port == 0 {
		c.Port = DefaultPort
	}
	if c.HealthCheckTime == 0 {
		c.HealthCheckTime = DefaultHealthCheckTime
	}
	if c.Strategy == "" {
		c.Strategy = DefaultStrategy
	}

	for i := range c.Servers {
		if c.Servers[i].Weight == 0 {
			c.Servers[i].Weight = 1
		}
	}
}

func (c *Config) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}

	if c.HealthCheckTime < 1 {
		return errors.New("health_check_time must be at least 1 second")
	}

	if !ValidStrategies[c.Strategy] {
		return fmt.Errorf("unknown strategy: %s (valid: round_robin, weighted, least_connections)", c.Strategy)
	}

	if len(c.Servers) == 0 {
		return errors.New("at least one server is required")
	}

	for i, server := range c.Servers {
		if server.URL == "" {
			return fmt.Errorf("server %d: URL is required", i)
		}
		if server.Weight < 1 {
			return fmt.Errorf("server %d: weight must be at least 1", i)
		}
	}

	return nil
}
