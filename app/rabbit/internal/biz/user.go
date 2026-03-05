package biz

import (
	"github.com/aide-family/rabbit/internal/biz/repository"
)

func NewUser(userRepo repository.User) *User {
	return &User{
		User: userRepo,
	}
}

type User struct {
	repository.User
}
