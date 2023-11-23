package biz

import (
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/repository"
)

type (
	UserBiz struct {
		log *log.Helper

		userRepo repository.UserRepo
	}
)

func NewUserBiz(userRepo repository.UserRepo, logger log.Logger) *UserBiz {
	return &UserBiz{
		log: log.NewHelper(logger),

		userRepo: userRepo,
	}
}
