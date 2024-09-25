package repoimpl

import (
	"context"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"gorm.io/gen"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
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

func (i *InviteRepositoryImpl) GetInviteUserByUserIdAndType(ctx context.Context, params *bo.InviteUserParams) (*bizmodel.SysTeamInvite, error) {
	bizQuery, err := getBizQuery(ctx, i.data)
	if !types.IsNil(err) {
		return nil, err
	}
	var wheres []gen.Condition
	wheres = append(wheres, bizQuery.SysTeamInvite.UserID.Eq(params.UserID))
	// 已加入或邀请中状态
	wheres = append(wheres, bizQuery.SysTeamInvite.InviteType.Eq(vobj.GetInviteTypeJoined().GetValue()))
	wheres = append(wheres, bizQuery.SysTeamInvite.InviteType.Eq(vobj.GetInviteTypeUnderReview().GetValue()))
	wheres = append(wheres, bizQuery.SysTeamInvite.TeamID.Eq(params.TeamID))
	return bizQuery.SysTeamInvite.
		WithContext(ctx).
		Where(wheres...).First()
}

func (i *InviteRepositoryImpl) InviteUser(ctx context.Context, params *bo.InviteUserParams) error {
	bizQuery, err := getBizQuery(ctx, i.data)
	if !types.IsNil(err) {
		return err
	}
	err = bizQuery.SysTeamInvite.WithContext(ctx).Create(createInviteToModel(ctx, params))
	if !types.IsNil(err) {
		return err
	}
	return nil
}

func (i *InviteRepositoryImpl) UpdateInviteStatus(ctx context.Context, params *bo.UpdateInviteStatusParams) error {
	bizQuery, err := getBizQuery(ctx, i.data)
	if !types.IsNil(err) {
		return err
	}
	if _, err = bizQuery.SysTeamInvite.WithContext(ctx).Where(bizQuery.SysTeamInvite.ID.Eq(params.InviteID)).Update(bizQuery.SysTeamInvite.InviteType, params.InviteType.GetValue()); err != nil {
		return err
	}
	return nil
}

func (i *InviteRepositoryImpl) InviteList(ctx context.Context, params *bo.QueryInviteListParams) {
	//TODO implement me
	panic("implement me")
}

func createInviteToModel(ctx context.Context, params *bo.InviteUserParams) *bizmodel.SysTeamInvite {
	invite := &bizmodel.SysTeamInvite{
		TeamID:        params.TeamID,
		UserID:        params.UserID,
		SysTeamRoleID: params.TeamRoleID,
		InviteType:    vobj.GetInviteTypeJoined(), // 邀请中
	}
	invite.WithContext(ctx)
	return invite
}
