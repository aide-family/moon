package biz

import (
	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/alkaid/internal/biz/agent"
)

var ProviderSet = wire.NewSet(NewHealthUsecase, agent.ProviderSet)

type HealthUsecase struct{}

func NewHealthUsecase() *HealthUsecase {
	return &HealthUsecase{}
}

func (uc *HealthUsecase) Status() string {
	return "ok"
}
