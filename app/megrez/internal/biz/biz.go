package biz

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewHealthUsecase)

type HealthUsecase struct{}

func NewHealthUsecase() *HealthUsecase {
	return &HealthUsecase{}
}

func (uc *HealthUsecase) Status() string {
	return "ok"
}
