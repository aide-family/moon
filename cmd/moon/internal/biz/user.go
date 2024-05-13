package biz

import (
	"github.com/aide-cloud/moon/cmd/moon/internal/biz/repo"
)

type UserBiz struct {
	userRepo repo.UserRepo
	repo.TransactionRepo
}

func NewUserBiz(userRepo repo.UserRepo, transactionRepo repo.TransactionRepo) *UserBiz {
	return &UserBiz{
		userRepo:        userRepo,
		TransactionRepo: transactionRepo,
	}
}
