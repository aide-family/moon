package repoimpl

import (
	"context"

	"github.com/aide-cloud/moon/cmd/moon/internal/biz/do/model"
	"github.com/aide-cloud/moon/cmd/moon/internal/biz/do/query"
	"github.com/aide-cloud/moon/cmd/moon/internal/biz/repo"
	"github.com/aide-cloud/moon/cmd/moon/internal/data"
)

func NewTeamRepo(data *data.Data) repo.TeamRepo {
	return &teamRepoImpl{
		data: data,
	}
}

type teamRepoImpl struct {
	data *data.Data
}

func (l *teamRepoImpl) GetUserTeamByID(ctx context.Context, userID, teamID uint32) (*model.SysTeamMember, error) {
	return query.Use(l.data.GetBizDB(ctx)).WithContext(ctx).SysTeamMember.Where(
		query.SysTeamMember.TeamID.Eq(teamID),
		query.SysTeamMember.UserID.Eq(userID),
	).First()
}

func (l *teamRepoImpl) GetTeamRoleByUserID(ctx context.Context, userID, teamID uint32) ([]*model.SysTeamMemberRole, error) {
	return query.Use(l.data.GetBizDB(ctx)).WithContext(ctx).SysTeamMemberRole.Where(
		query.SysTeamMemberRole.TeamID.Eq(teamID),
		query.SysTeamMemberRole.UserID.Eq(userID),
	).Find()
}

func (l *teamRepoImpl) GetUserTeamRole(ctx context.Context, userID, teamID, roleID uint32) (*model.SysTeamMemberRole, error) {
	return query.Use(l.data.GetBizDB(ctx)).WithContext(ctx).SysTeamMemberRole.Where(
		query.SysTeamMemberRole.TeamID.Eq(teamID),
		query.SysTeamMemberRole.UserID.Eq(userID),
		query.SysTeamMemberRole.RoleID.Eq(roleID),
	).First()
}
