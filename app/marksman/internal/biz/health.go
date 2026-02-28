package biz

import "github.com/aide-family/marksman/internal/biz/repository"

func NewHealth(healthRepo repository.Health) *Health {
	return &Health{
		healthRepo: healthRepo,
	}
}

type Health struct {
	healthRepo repository.Health
}

func (h *Health) Readiness() error {
	return h.healthRepo.Readiness()
}
