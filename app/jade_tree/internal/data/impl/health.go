package impl

import (
	"github.com/aide-family/jade_tree/internal/biz/repository"
	"github.com/aide-family/jade_tree/internal/data"
)

func NewHealthRepository(d *data.Data) repository.Health {
	return &healthRepositoryImpl{d: d}
}

type healthRepositoryImpl struct {
	d *data.Data
}

func (h *healthRepositoryImpl) Readiness() error { return nil }
