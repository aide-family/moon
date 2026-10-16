package biz

import (
	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/merak/internal/biz/notify"
)

var ProviderSet = wire.NewSet(NewHealthUsecase, notify.ProviderSet)

type HealthUsecase struct{}

func NewHealthUsecase() *HealthUsecase {
	return &HealthUsecase{}
}

func (uc *HealthUsecase) Status() string {
	return "ok"
}
