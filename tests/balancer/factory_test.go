package balancer_test

import (
	"testing"

	"janus/internal/balancer"
)

func TestNewStrategyRoundRobin(t *testing.T) {
	strategy, err := balancer.NewStrategy("round_robin")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strategy == nil {
		t.Fatal("expected strategy, got nil")
	}

	if strategy.Name() != "round_robin" {
		t.Errorf("name = %s, want round_robin", strategy.Name())
	}
}

func TestNewStrategyWeighted(t *testing.T) {
	strategy, err := balancer.NewStrategy("weighted")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strategy == nil {
		t.Fatal("expected strategy, got nil")
	}

	if strategy.Name() != "weighted" {
		t.Errorf("name = %s, want weighted", strategy.Name())
	}
}

func TestNewStrategyLeastConnections(t *testing.T) {
	strategy, err := balancer.NewStrategy("least_connections")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strategy == nil {
		t.Fatal("expected strategy, got nil")
	}

	if strategy.Name() != "least_connections" {
		t.Errorf("name = %s, want least_connections", strategy.Name())
	}
}

func TestNewStrategyUnknown(t *testing.T) {
	strategy, err := balancer.NewStrategy("unknown_strategy")

	if err == nil {
		t.Error("expected error for unknown strategy, got nil")
	}

	if strategy != nil {
		t.Error("expected nil strategy for error case")
	}
}

func TestNewStrategyEmpty(t *testing.T) {
	strategy, err := balancer.NewStrategy("")

	if err == nil {
		t.Error("expected error for empty strategy, got nil")
	}

	if strategy != nil {
		t.Error("expected nil strategy for error case")
	}
}

func TestNewStrategyAllValid(t *testing.T) {
	validStrategies := []string{
		"round_robin",
		"weighted",
		"least_connections",
	}

	for _, name := range validStrategies {
		t.Run(name, func(t *testing.T) {
			strategy, err := balancer.NewStrategy(name)

			if err != nil {
				t.Errorf("strategy %s should be valid: %v", name, err)
			}

			if strategy == nil {
				t.Errorf("expected non-nil strategy for %s", name)
			}

			if strategy.Name() != name {
				t.Errorf("strategy.Name() = %s, want %s", strategy.Name(), name)
			}
		})
	}
}
