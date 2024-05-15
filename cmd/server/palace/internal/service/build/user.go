package build

import (
	"github.com/aide-cloud/moon/api"
	"github.com/aide-cloud/moon/api/admin"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/do/model"
	"github.com/aide-cloud/moon/pkg/types"
)

type UserBuild struct {
	*model.SysUser
}

func NewUserBuild(user *model.SysUser) *UserBuild {
	return &UserBuild{
		SysUser: user,
	}
}

// ToApi 转换成api
func (b *UserBuild) ToApi() *admin.User {
	if types.IsNil(b) {
		return nil
	}
	return &admin.User{
		Id:        b.ID,
		Name:      b.Username,
		Nickname:  b.Nickname,
		Email:     string(b.Email),
		Phone:     string(b.Phone),
		Status:    api.Status(b.Status),
		Gender:    api.Gender(b.Gender),
		Role:      api.Role(b.Role),
		Avatar:    b.Avatar,
		Remark:    b.Remark,
		CreatedAt: b.CreatedAt.String(),
		UpdatedAt: b.UpdatedAt.String(),
	}
}
