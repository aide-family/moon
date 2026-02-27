// Package convert is the convert package for the auth service.
package convert

import (
	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/data/impl/do"
)

func UserToBo(user *do.User) *bo.UserItemBo {
	if user == nil {
		return nil
	}
	return &bo.UserItemBo{
		UID:       user.UID,
		Email:     user.Email,
		Name:      user.Name,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Status:    user.Status,
		Remark:    user.Remark,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
