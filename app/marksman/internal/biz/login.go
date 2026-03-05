package biz

import (
	"github.com/aide-family/marksman/internal/biz/repository"
)

func NewLoginBiz(authRepo repository.LoginRepository) *LoginBiz {
	return &LoginBiz{LoginRepository: authRepo}
}

type LoginBiz struct {
	repository.LoginRepository
}
