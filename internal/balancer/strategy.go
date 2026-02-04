package balancer

import (
	"janus/internal/domain"
)

type Strategy interface {
	GetNextServer(pool *domain.ServerPool) *domain.Server
	Name() string
}
