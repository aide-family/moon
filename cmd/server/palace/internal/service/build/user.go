package build

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	"github.com/aide-family/moon/cmd/server/palace/internal/data/runtimecache"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
)

type UserBuilder struct {
	*model.SysUser
}

func NewUserBuilder(user *model.SysUser) *UserBuilder {
	return &UserBuilder{
		SysUser: user,
	}
}

// ToApi 转换成api
func (b *UserBuilder) ToApi() *admin.User {
	if types.IsNil(b) || types.IsNil(b.SysUser) {
		return nil
	}
	return &admin.User{
		Id:        b.ID,
		Name:      b.Username,
		Nickname:  b.Nickname,
		Email:     b.Email,
		Phone:     b.Phone,
		Status:    api.Status(b.Status),
		Gender:    api.Gender(b.Gender),
		Role:      api.Role(b.Role),
		Avatar:    b.Avatar,
		Remark:    b.Remark,
		CreatedAt: b.CreatedAt.String(),
		UpdatedAt: b.UpdatedAt.String(),
	}
}

type TeamMemberBuilder struct {
	*bizmodel.SysTeamMember
}

func NewTeamMemberBuilder(member *bizmodel.SysTeamMember) *TeamMemberBuilder {
	return &TeamMemberBuilder{
		SysTeamMember: member,
	}
}

func (b *TeamMemberBuilder) ToApi(ctx context.Context) *admin.TeamMember {
	if types.IsNil(b) || types.IsNil(b.SysTeamMember) {
		return nil
	}
	cache := runtimecache.GetRuntimeCache()
	return &admin.TeamMember{
		UserId:    b.UserID,
		Id:        b.ID,
		Role:      api.Role(b.Role),
		Status:    api.Status(b.Status),
		CreatedAt: b.CreatedAt.String(),
		UpdatedAt: b.UpdatedAt.String(),
		User:      NewUserBuilder(cache.GetUser(ctx, b.UserID)).ToApi(),
	}
}
