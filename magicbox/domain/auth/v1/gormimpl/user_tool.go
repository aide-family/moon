package gormimpl

import (
	"strings"
	"time"

	apiv1 "github.com/aide-family/magicbox/api/v1"
	"github.com/aide-family/magicbox/domain/auth/v1/gormimpl/model"
	"github.com/aide-family/magicbox/enum"
)

func convertUserItem(u *model.User) *apiv1.UserItem {
	if u == nil {
		return nil
	}
	return &apiv1.UserItem{
		Uid:       u.ID.Int64(),
		Email:     u.Email,
		Phone:     "", // 模型无 phone 字段，可由上游补齐
		Status:    u.Status,
		CreatedAt: u.CreatedAt.Format(time.DateTime),
		UpdatedAt: u.UpdatedAt.Format(time.DateTime),
		Name:      u.Name,
		Nickname:  u.Nickname,
		Avatar:    u.Avatar,
		Remark:    u.Remark,
	}
}

func convertUserSelectItem(u *model.User) *apiv1.SelectUserItem {
	if u == nil {
		return nil
	}
	return &apiv1.SelectUserItem{
		Value:    u.ID.Int64(),
		Label:    strings.Join([]string{u.Name, "(", u.Nickname, ")"}, " "),
		Disabled: u.Status != enum.UserStatus_ACTIVE || u.DeletedAt.Valid,
		Tooltip:  u.Remark,
	}
}
