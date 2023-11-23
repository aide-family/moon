package biz

import (
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/repository"
)

type (
	RoleBiz struct {
		log *log.Helper

		roleRepo repository.RoleRepo
	}
)

func NewRoleBiz(roleRepo repository.RoleRepo, logger log.Logger) *RoleBiz {
	return &RoleBiz{
		log:      log.NewHelper(logger),
		roleRepo: roleRepo,
	}
}
