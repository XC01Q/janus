package balancer

import (
	"fmt"
)

func NewStrategy(name string) (Strategy, error) {
	switch name {
	case "round_robin":
		return NewRoundRobin(), nil
	case "weighted":
		return NewWeighted(), nil
	case "least_connections":
		return NewLeastConnections(), nil
	default:
		return nil, fmt.Errorf("unknown balancing strategy: %s", name)
	}
}
