package service

import (
	"github.com/aide-family/marksman/internal/biz"
)

func NewUserService(userBiz *biz.User) *UserService {
	return &UserService{
		User: userBiz,
	}
}

type UserService struct {
	*biz.User
}
