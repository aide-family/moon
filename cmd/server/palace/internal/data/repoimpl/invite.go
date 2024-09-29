package repoimpl

import (
	"context"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/query"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"gorm.io/gen"
)

// NewInviteRepository 创建邀请仓库
func NewInviteRepository(data *data.Data) repository.TeamInvite {
	return &InviteRepositoryImpl{
		data: data,
	}
}

type (
	InviteRepositoryImpl struct {
		data *data.Data
	}
)

func (i *InviteRepositoryImpl) DeleteInvite(ctx context.Context, inviteId uint32) error {
	mainQuery := query.Use(i.data.GetMainDB(ctx))
	_, err := mainQuery.SysTeamInvite.WithContext(ctx).Where(mainQuery.SysTeamInvite.ID.Eq(inviteId)).Delete()
	return err
}

func (i *InviteRepositoryImpl) GetInviteDetail(ctx context.Context, inviteId uint32) (*model.SysTeamInvite, error) {
	mainQuery := query.Use(i.data.GetMainDB(ctx))
	return mainQuery.SysTeamInvite.WithContext(ctx).Where(mainQuery.SysTeamInvite.ID.Eq(inviteId)).First()
}

func (i *InviteRepositoryImpl) GetInviteUserByUserIdAndType(ctx context.Context, params *bo.InviteUserParams) (*model.SysTeamInvite, error) {
	mainQuery := query.Use(i.data.GetMainDB(ctx))
	var wheres []gen.Condition
	wheres = append(wheres, mainQuery.SysTeamInvite.UserID.Eq(params.UserID))
	wheres = append(wheres, mainQuery.SysTeamInvite.TeamID.Eq(params.TeamID))
	return mainQuery.SysTeamInvite.WithContext(ctx).Where(wheres...).First()
}

func (i *InviteRepositoryImpl) InviteUser(ctx context.Context, params *bo.InviteUserParams) error {
	mainQuery := query.Use(i.data.GetMainDB(ctx))
	teamInvite, _ := mainQuery.SysTeamInvite.WithContext(ctx).
		Where(mainQuery.SysTeamInvite.UserID.Eq(params.UserID),
			mainQuery.SysTeamInvite.TeamID.Eq(params.TeamID)).First()

	if !types.IsNil(teamInvite) {
		teamInvite.RolesIds = params.TeamRoleIds
		teamInvite.InviteType = vobj.InviteTypeUnderReview
		if _, err := mainQuery.WithContext(ctx).SysTeamInvite.Updates(teamInvite); !types.IsNil(err) {
			return err
		}
	} else {
		teamInvite = &model.SysTeamInvite{
			TeamID:     params.TeamID,
			UserID:     params.UserID,
			InviteType: vobj.InviteTypeUnderReview,
			RolesIds:   params.TeamRoleIds,
		}
		if err := mainQuery.SysTeamInvite.WithContext(ctx).Create(teamInvite); !types.IsNil(err) {
			return err
		}
	}
	return nil
}

func (i *InviteRepositoryImpl) UpdateInviteStatus(ctx context.Context, params *bo.UpdateInviteStatusParams) error {
	mainQuery := query.Use(i.data.GetMainDB(ctx))
	if _, err := mainQuery.SysTeamInvite.WithContext(ctx).Where(mainQuery.SysTeamInvite.ID.Eq(params.InviteID)).Update(mainQuery.SysTeamInvite.InviteType, params.InviteType.GetValue()); err != nil {
		return err
	}

	// 如果邀请类型是加入团队，则创建团队成员信息
	if params.InviteType.IsJoined() {
		teamInvite, err := i.GetInviteDetail(ctx, params.InviteID)

		if !types.IsNil(err) {
			return err
		}

		if err = i.createTeamMemberInfo(ctx, teamInvite); !types.IsNil(err) {
			return err
		}
	}
	return nil
}

func (i *InviteRepositoryImpl) UserInviteList(ctx context.Context, params *bo.QueryInviteListParams) ([]*model.SysTeamInvite, error) {
	mainQuery := query.Use(i.data.GetMainDB(ctx))
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnauthorized(ctx)
	}
	var wheres []gen.Condition

	wheres = append(wheres, mainQuery.SysTeamInvite.UserID.Eq(claims.UserID))

	if !params.InviteType.IsUnknown() {
		wheres = append(wheres, mainQuery.SysTeamInvite.InviteType.Eq(params.InviteType.GetValue()))
	}

	queryWrapper := mainQuery.SysTeamInvite.WithContext(ctx).Where(wheres...)
	return queryWrapper.Order(mainQuery.SysTeamInvite.ID.Desc()).Find()
}

func (i *InviteRepositoryImpl) createTeamMemberInfo(ctx context.Context, invite *model.SysTeamInvite) error {
	bizQuery, err := getTeamIdBizQuery(i.data, invite.TeamID)
	if !types.IsNil(err) {
		return err
	}
	teamMember := &bizmodel.SysTeamMember{
		UserID: invite.UserID,
		TeamID: invite.TeamID,
		Status: vobj.StatusEnable,
		TeamRoles: types.SliceTo(invite.RolesIds.ToSlice(), func(roleID uint32) *bizmodel.SysTeamRole {
			return &bizmodel.SysTeamRole{
				AllFieldModel: model.AllFieldModel{ID: roleID},
			}
		}),
	}
	return bizQuery.SysTeamMember.WithContext(ctx).Create(teamMember)
}
