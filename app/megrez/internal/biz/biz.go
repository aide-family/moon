package biz

import (
	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/megrez/internal/biz/eval"
)

var ProviderSet = wire.NewSet(NewHealthUsecase, eval.ProviderSet)

type HealthUsecase struct{}

func NewHealthUsecase() *HealthUsecase {
	return &HealthUsecase{}
}

func (uc *HealthUsecase) Status() string {
	return "ok"
}
