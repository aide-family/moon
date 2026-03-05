package service

import (
	"github.com/aide-family/rabbit/internal/biz"
)

func NewUserService(userBiz *biz.User) (*UserService, error) {
	return &UserService{User: userBiz}, nil
}

type UserService struct {
	*biz.User
}
