package impl

import (
	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/data"
)

func NewHealthRepository(d *data.Data) repository.Health {
	return &healthRepositoryImpl{
		d: d,
	}
}

type healthRepositoryImpl struct {
	d *data.Data
}

// Readiness implements repository.Health.
func (h *healthRepositoryImpl) Readiness() error {
	return nil
}
