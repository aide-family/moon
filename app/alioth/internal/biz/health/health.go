package health

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewUsecase)

type Usecase struct{}

func NewUsecase() *Usecase {
	return &Usecase{}
}

func (uc *Usecase) Status() string {
	return "ok"
}
