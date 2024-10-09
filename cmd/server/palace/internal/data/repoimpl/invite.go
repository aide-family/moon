package repoimpl

import (
	"context"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"

	"gorm.io/gen"
	"gorm.io/gen/field"
)

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
	bizQuery, err := getBizQuery(ctx, i.data)
	if !types.IsNil(err) {
		return err
	}
	_, err = bizQuery.SysTeamInvite.WithContext(ctx).Where(bizQuery.SysTeamInvite.ID.Eq(inviteId)).Delete()
	return err
}

func (i *InviteRepositoryImpl) GetInviteDetail(ctx context.Context, inviteId uint32) (*bizmodel.SysTeamInvite, error) {
	bizQuery, err := getBizQuery(ctx, i.data)
	if !types.IsNil(err) {
		return nil, err
	}
	return bizQuery.SysTeamInvite.WithContext(ctx).Preload(field.Associations).Where(bizQuery.SysTeamInvite.ID.Eq(inviteId)).First()
}

func (i *InviteRepositoryImpl) GetInviteUserByUserIdAndType(ctx context.Context, params *bo.InviteUserParams) (*bizmodel.SysTeamInvite, error) {
	bizQuery, err := getBizQuery(ctx, i.data)
	if !types.IsNil(err) {
		return nil, err
	}
	var wheres []gen.Condition
	wheres = append(wheres, bizQuery.SysTeamInvite.UserID.Eq(params.UserID))
	wheres = append(wheres, bizQuery.SysTeamInvite.TeamID.Eq(params.TeamID))
	return bizQuery.SysTeamInvite.WithContext(ctx).Where(wheres...).First()
}

func (i *InviteRepositoryImpl) InviteUser(ctx context.Context, params *bo.InviteUserParams) error {
	bizQuery, err := getBizQuery(ctx, i.data)
	if !types.IsNil(err) {
		return err
	}
	teamInvite, _ := bizQuery.SysTeamInvite.WithContext(ctx).
		Where(bizQuery.SysTeamInvite.UserID.Eq(params.UserID),
			bizQuery.SysTeamInvite.TeamID.Eq(params.TeamID)).First()

	if !types.IsNil(teamInvite) {
		teamInvite.TeamRoles = types.SliceTo(params.TeamRoleIds, func(roleID uint32) *bizmodel.SysTeamRole {
			return &bizmodel.SysTeamRole{
				AllFieldModel: model.AllFieldModel{ID: roleID},
				TeamID:        params.TeamID,
			}
		})
		teamInvite.InviteType = vobj.InviteTypeUnderReview
		_, err = bizQuery.WithContext(ctx).SysTeamInvite.Updates(teamInvite)
	} else {
		teamInvite = &bizmodel.SysTeamInvite{
			TeamID:     params.TeamID,
			UserID:     params.UserID,
			InviteType: vobj.InviteTypeUnderReview,
			TeamRoles: types.SliceTo(params.TeamRoleIds, func(roleID uint32) *bizmodel.SysTeamRole {
				return &bizmodel.SysTeamRole{
					AllFieldModel: model.AllFieldModel{ID: roleID},
					TeamID:        params.TeamID,
				}
			}),
		}
		err = bizQuery.SysTeamInvite.WithContext(ctx).Create(teamInvite)
	}
	return err
}

func (i *InviteRepositoryImpl) UpdateInviteStatus(ctx context.Context, params *bo.UpdateInviteStatusParams) error {
	bizQuery, err := getBizQuery(ctx, i.data)
	if !types.IsNil(err) {
		return err
	}
	if _, err = bizQuery.SysTeamInvite.WithContext(ctx).Where(bizQuery.SysTeamInvite.ID.Eq(params.InviteID)).Update(bizQuery.SysTeamInvite.InviteType, params.InviteType.GetValue()); err != nil {
		return err
	}

	// 如果邀请类型是加入团队，则创建团队成员信息
	if params.InviteType.IsJoined() {
		teamInvite, err := i.GetInviteDetail(ctx, params.InviteID)
		if !types.IsNil(err) {
			return err
		}
		err = i.createTeamMemberInfo(ctx, teamInvite)
		if !types.IsNil(err) {
			return err
		}
	}
	return nil
}

func (i *InviteRepositoryImpl) InviteList(ctx context.Context) ([]*bizmodel.SysTeamInvite, error) {
	bizQuery, err := getBizQuery(ctx, i.data)
	if !types.IsNil(err) {
		return nil, err
	}
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnauthorized(ctx)
	}

	return bizQuery.SysTeamInvite.WithContext(ctx).Where(bizQuery.SysTeamInvite.UserID.Eq(claims.UserID)).Find()
}

func (i *InviteRepositoryImpl) createTeamMemberInfo(ctx context.Context, invite *bizmodel.SysTeamInvite) error {
	bizQuery, err := getBizQuery(ctx, i.data)
	if !types.IsNil(err) {
		return err
	}
	teamMember := &bizmodel.SysTeamMember{
		UserID:    invite.UserID,
		TeamID:    invite.TeamID,
		Status:    vobj.StatusEnable,
		TeamRoles: invite.TeamRoles,
	}
	return bizQuery.SysTeamMember.WithContext(ctx).Create(teamMember)
}
