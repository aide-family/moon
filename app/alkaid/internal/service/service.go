package service

import (
	"context"

	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/alkaid/internal/biz"
)

var ProviderSet = wire.NewSet(NewHealthService)

type HealthService struct {
	uc *biz.HealthUsecase
}

func NewHealthService(uc *biz.HealthUsecase) *HealthService {
	return &HealthService{uc: uc}
}

func (s *HealthService) Status(context.Context) string {
	return s.uc.Status()
}
