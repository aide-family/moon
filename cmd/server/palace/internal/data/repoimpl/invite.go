package repoimpl

import (
	"context"
	// 导入邀请邮件模板
	_ "embed"
	"strings"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-family/moon/pkg/helper"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/query"
	"github.com/aide-family/moon/pkg/util/format"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/go-kratos/kratos/v2/log"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gen"
	"gorm.io/gorm"
)

// NewInviteRepository 创建团队邀请实现
func NewInviteRepository(bc *palaceconf.Bootstrap, data *data.Data, cacheRepo repository.Cache) repository.TeamInvite {
	return &InviteRepositoryImpl{
		bc:        bc,
		data:      data,
		cacheRepo: cacheRepo,
	}
}

type (
	// InviteRepositoryImpl 团队邀请实现
	InviteRepositoryImpl struct {
		data      *data.Data
		bc        *palaceconf.Bootstrap
		cacheRepo repository.Cache
	}
)

// DeleteInvite 删除邀请
func (i *InviteRepositoryImpl) DeleteInvite(ctx context.Context, inviteID uint32) error {
	mainQuery := query.Use(i.data.GetMainDB(ctx))
	_, err := mainQuery.SysTeamInvite.WithContext(ctx).Where(mainQuery.SysTeamInvite.ID.Eq(inviteID)).Delete()
	return err
}

// GetInviteDetail 获取邀请详情
func (i *InviteRepositoryImpl) GetInviteDetail(ctx context.Context, inviteID uint32) (*model.SysTeamInvite, error) {
	mainQuery := query.Use(i.data.GetMainDB(ctx))
	return mainQuery.SysTeamInvite.WithContext(ctx).Where(mainQuery.SysTeamInvite.ID.Eq(inviteID)).First()
}

// GetInviteUserByUserIDAndType 获取邀请用户
func (i *InviteRepositoryImpl) GetInviteUserByUserIDAndType(ctx context.Context, params *bo.InviteUserParams) (*model.SysTeamInvite, error) {
	mainQuery := query.Use(i.data.GetMainDB(ctx))
	var wheres []gen.Condition
	wheres = append(wheres, mainQuery.SysTeamInvite.UserID.Eq(params.UserID))
	wheres = append(wheres, mainQuery.SysTeamInvite.TeamID.Eq(params.TeamID))
	return mainQuery.SysTeamInvite.WithContext(ctx).Where(wheres...).First()
}

// InviteUser 邀请用户
func (i *InviteRepositoryImpl) InviteUser(ctx context.Context, params *bo.InviteUserParams) (teamInvite *model.SysTeamInvite, err error) {
	mainQuery := query.Use(i.data.GetMainDB(ctx))
	teamInvite, err = mainQuery.SysTeamInvite.WithContext(ctx).
		Where(mainQuery.SysTeamInvite.UserID.Eq(params.UserID),
			mainQuery.SysTeamInvite.TeamID.Eq(params.TeamID)).First()
	if !types.IsNil(err) && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if !types.IsNil(teamInvite) {
		teamInvite.RolesIds = params.TeamRoleIds
		teamInvite.InviteType = vobj.InviteTypeUnderReview
		if _, err = mainQuery.WithContext(ctx).SysTeamInvite.Updates(teamInvite); !types.IsNil(err) {
			return nil, err
		}
		return teamInvite, nil
	}
	teamInvite = &model.SysTeamInvite{
		TeamID:     params.TeamID,
		UserID:     params.UserID,
		InviteType: vobj.InviteTypeUnderReview,
		RolesIds:   params.TeamRoleIds,
		Role:       params.Role,
	}
	teamInvite.WithContext(ctx)
	if err = mainQuery.SysTeamInvite.WithContext(ctx).Create(teamInvite); !types.IsNil(err) {
		return nil, err
	}
	i.cacheRepo.SyncUserTeamList(ctx, teamInvite.UserID)
	return
}

// UpdateInviteStatus 更新邀请状态
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

// UserInviteList 用户邀请列表
func (i *InviteRepositoryImpl) UserInviteList(ctx context.Context, params *bo.QueryInviteListParams) ([]*model.SysTeamInvite, error) {
	mainQuery := query.Use(i.data.GetMainDB(ctx))
	var wheres []gen.Condition

	wheres = append(wheres, mainQuery.SysTeamInvite.UserID.Eq(middleware.GetUserID(ctx)))

	if !params.InviteType.IsUnknown() {
		wheres = append(wheres, mainQuery.SysTeamInvite.InviteType.Eq(params.InviteType.GetValue()))
	}

	queryWrapper := mainQuery.SysTeamInvite.WithContext(ctx).Where(wheres...)
	return queryWrapper.Order(mainQuery.SysTeamInvite.ID.Desc()).Find()
}

// createTeamMemberInfo 创建团队成员信息
func (i *InviteRepositoryImpl) createTeamMemberInfo(ctx context.Context, invite *model.SysTeamInvite) error {
	bizQuery, err := getTeamIDBizQuery(i.data, invite.TeamID)
	if !types.IsNil(err) {
		return err
	}
	teamMember := &bizmodel.SysTeamMember{
		UserID: invite.UserID,
		Role:   invite.Role,
		Status: vobj.StatusEnable,
		TeamRoles: types.SliceTo(invite.RolesIds.ToSlice(), func(roleID uint32) *bizmodel.SysTeamRole {
			return &bizmodel.SysTeamRole{
				AllFieldModel: bizmodel.AllFieldModel{AllFieldModel: model.AllFieldModel{ID: roleID}},
			}
		}),
	}
	defer i.cacheRepo.SyncUserTeamList(ctx, invite.UserID)
	return bizQuery.SysTeamMember.WithContext(ctx).Create(teamMember)
}

//go:embed invite_email.html
var inviteEmailHTML string

// SendInviteEmail 发送邀请邮件
func (i *InviteRepositoryImpl) SendInviteEmail(ctx context.Context, params *bo.InviteUserParams, opUser, user *model.SysUser) error {
	email := user.Email
	if err := helper.CheckEmail(email); err != nil {
		return err
	}
	bizQuery, err := getTeamIDBizQuery(i.data, params.TeamID)
	if !types.IsNil(err) {
		return err
	}
	// 获取团队详情
	teamInfo := i.cacheRepo.GetTeam(ctx, params.TeamID)
	// 获取团队角色
	teamRoles, err := bizQuery.SysTeamRole.WithContext(ctx).Where(bizQuery.SysTeamRole.ID.In(params.TeamRoleIds.ToSlice()...)).Find()
	if !types.IsNil(err) {
		log.Errorw("method", "SendInviteEmail", "err", err)
	}
	// 发送验证码到用户邮箱
	emailBody := format.Formatter(inviteEmailHTML, map[string]string{
		"Inviter":  opUser.Username,
		"TeamName": teamInfo.Name,
		"TeamRole": params.Role.String(),
		"BusinessRoles": strings.Join(types.SliceTo(teamRoles, func(role *bizmodel.SysTeamRole) string {
			return role.Name
		}), "、"),
		"RedirectURI": i.bc.GetOauth2().GetRedirectUri(),
		"APP":         i.bc.GetServer().GetName(),
		"Remark":      i.bc.GetServer().GetMetadata()["description"],
	})
	return i.data.GetEmail().SetSubject("欢迎使用 Moon 监控系统").SetTo(email).SetBody(emailBody, "text/html").Send()
}
