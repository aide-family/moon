package build

import (
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	"github.com/aide-family/moon/pkg/helper/model"
	"github.com/aide-family/moon/pkg/helper/model/bizmodel"
	"github.com/aide-family/moon/pkg/types"
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

type TeamMemberBuild struct {
	*bizmodel.SysTeamMember
}

func NewTeamMemberBuild(member *bizmodel.SysTeamMember) *TeamMemberBuild {
	return &TeamMemberBuild{
		SysTeamMember: member,
	}
}

func (b *TeamMemberBuild) ToApi() *admin.TeamMember {
	if types.IsNil(b) || types.IsNil(b.SysTeamMember) {
		return nil
	}
	return &admin.TeamMember{
		UserId:    b.UserID,
		Id:        b.ID,
		Roles:     nil,
		Status:    api.Status(b.Status),
		CreatedAt: b.CreatedAt.String(),
		UpdatedAt: b.UpdatedAt.String(),
		// TODO 从全局变量获取
		//User:      NewUserBuild(b.Member).ToApi(),
	}
}
