package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/job"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/helper/permission"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/plugin/server/cron_server"
	"github.com/aide-family/moon/pkg/util/safety"
)

func NewTeamBiz(
	cacheRepo repository.Cache,
	userRepo repository.User,
	teamRepo repository.Team,
	teamEmailConfigRepo repository.TeamEmailConfig,
	teamSMSConfigRepo repository.TeamSMSConfig,
	teamRoleRepo repository.TeamRole,
	menuRepo repository.Menu,
	operateLogRepo repository.OperateLog,
	memberRepo repository.Member,
	inviteRepo repository.Invite,
	transaction repository.Transaction,
	eventBus repository.EventBus,
	logger log.Logger,
) *Team {
	teamBiz := &Team{
		helper:              log.NewHelper(log.With(logger, "module", "biz.team")),
		cacheRepo:           cacheRepo,
		userRepo:            userRepo,
		teamRepo:            teamRepo,
		teamEmailConfigRepo: teamEmailConfigRepo,
		teamSMSConfigRepo:   teamSMSConfigRepo,
		teamRoleRepo:        teamRoleRepo,
		menuRepo:            menuRepo,
		operateLogRepo:      operateLogRepo,
		memberRepo:          memberRepo,
		inviteRepo:          inviteRepo,
		transaction:         transaction,
		eventBus:            eventBus,
	}
	do.RegisterGetTeamFunc(func(id uint32) do.Team {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return teamBiz.getTeam(ctx, id)
	})
	do.RegisterGetTeamMemberFunc(func(id uint32) do.TeamMember {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return teamBiz.getTeamMember(ctx, id)
	}, func(ids []uint32) []do.TeamMember {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return teamBiz.getTeamMembers(ctx, ids)
	})
	return teamBiz
}

type Team struct {
	helper              *log.Helper
	cacheRepo           repository.Cache
	userRepo            repository.User
	teamRepo            repository.Team
	teamEmailConfigRepo repository.TeamEmailConfig
	teamSMSConfigRepo   repository.TeamSMSConfig
	teamRoleRepo        repository.TeamRole
	menuRepo            repository.Menu
	operateLogRepo      repository.OperateLog
	memberRepo          repository.Member
	inviteRepo          repository.Invite
	transaction         repository.Transaction
	eventBus            repository.EventBus
}

func (t *Team) publishNoticeEmailConfigDataChangeEvent(ctx context.Context, ids ...uint32) {
	if len(ids) == 0 {
		return
	}
	teamID := permission.GetTeamIDByContextWithZeroValue(safety.CopyValueCtx(ctx))
	safety.Go("publishNoticeEmailConfigDataChangeEvent", func() {
		t.eventBus.PublishDataChangeEvent(vobj.ChangedTypeNoticeEmailConfig, teamID, ids...)
	}, t.helper.Logger())
}

func (t *Team) publishNoticeSMSConfigDataChangeEvent(ctx context.Context, ids ...uint32) {
	if len(ids) == 0 {
		return
	}
	teamID := permission.GetTeamIDByContextWithZeroValue(safety.CopyValueCtx(ctx))
	safety.Go("publishNoticeSMSConfigDataChangeEvent", func() {
		t.eventBus.PublishDataChangeEvent(vobj.ChangedTypeNoticeSMSConfig, teamID, ids...)
	}, t.helper.Logger())
}

func (t *Team) getTeam(ctx context.Context, id uint32) do.Team {
	team, err := t.cacheRepo.GetTeam(ctx, id)
	if err != nil {
		if merr.IsNotFound(err) {
			team, err = t.teamRepo.FindByID(ctx, id)
			if err != nil {
				t.helper.WithContext(ctx).Errorw("msg", "get team fail", "err", err)
			} else {
				if err := t.cacheRepo.CacheTeams(ctx, team); err != nil {
					t.helper.WithContext(ctx).Errorw("msg", "cache team fail", "err", err)
				}
			}
		}
	}
	return team
}

func (t *Team) getTeamMember(ctx context.Context, id uint32) do.TeamMember {
	teamMember, err := t.cacheRepo.GetTeamMember(ctx, id)
	if err != nil {
		if merr.IsNotFound(err) {
			teamMember, err = t.memberRepo.Get(ctx, id)
			if err != nil {
				t.helper.WithContext(ctx).Errorw("msg", "get team member fail", "err", err)
			} else {
				if err := t.cacheRepo.CacheTeamMembers(ctx, teamMember); err != nil {
					t.helper.WithContext(ctx).Errorw("msg", "cache team member fail", "err", err)
				}
			}
		}
	}
	return teamMember
}

func (t *Team) getTeamMembers(ctx context.Context, ids []uint32) []do.TeamMember {
	teamMembers, err := t.cacheRepo.GetTeamMembers(ctx, ids...)
	if err != nil {
		if merr.IsNotFound(err) {
			teamMembers, err = t.memberRepo.Find(ctx, ids)
			if err != nil {
				t.helper.WithContext(ctx).Errorw("msg", "get team members fail", "err", err)
			} else {
				if err := t.cacheRepo.CacheTeamMembers(ctx, teamMembers...); err != nil {
					t.helper.WithContext(ctx).Errorw("msg", "cache team members fail", "err", err)
				}
			}
		}
	}
	return teamMembers
}

func (t *Team) SaveTeam(ctx context.Context, req *bo.SaveOneTeamRequest) error {
	// check team name is unique
	if err := t.teamRepo.CheckNameUnique(ctx, req.GetName(), req.TeamID); err != nil {
		return err
	}

	return t.transaction.MainExec(ctx, func(ctx context.Context) error {
		if req.TeamID <= 0 {
			leaderID, ok := permission.GetUserIDByContext(ctx)
			if !ok {
				return merr.ErrorUnauthorized("user not found in context")
			}
			leaderDo, err := t.userRepo.FindByID(ctx, leaderID)
			if err != nil {
				return err
			}
			createParams := req.WithCreateTeamRequest(leaderDo)
			if err := t.teamRepo.Create(ctx, createParams); err != nil {
				return err
			}
		} else {
			teamInfo, err := t.teamRepo.FindByID(ctx, req.TeamID)
			if err != nil {
				return err
			}
			updateTeamParams := req.WithUpdateTeamRequest(teamInfo)
			if err := t.teamRepo.Update(ctx, updateTeamParams); err != nil {
				return err
			}
		}
		teamDo, err := t.teamRepo.FindByName(ctx, req.GetName())
		if err != nil {
			return err
		}

		if err := t.userRepo.AppendTeam(ctx, teamDo); err != nil {
			t.helper.WithContext(ctx).Errorw("msg", "append team to user fail", "err", err)
			return err
		}
		createMemberParams := &bo.CreateTeamMemberReq{
			Team:     teamDo,
			User:     teamDo.GetLeader(),
			Status:   vobj.MemberStatusNormal,
			Position: vobj.PositionSuperAdmin,
		}
		if err := t.memberRepo.Create(ctx, createMemberParams); err != nil {
			t.helper.WithContext(ctx).Errorw("msg", "create team member fail", "err", err)
			return err
		}
		return nil
	})
}

// SaveEmailConfig saves the email configuration for a team
func (t *Team) SaveEmailConfig(ctx context.Context, req *bo.SaveEmailConfigRequest) (err error) {
	if err := req.Validate(); err != nil {
		return err
	}

	if err := t.teamEmailConfigRepo.CheckNameUnique(ctx, req.GetName(), req.ID); err != nil {
		return err
	}
	emailConfigID := req.ID
	defer t.publishNoticeEmailConfigDataChangeEvent(ctx, emailConfigID)
	return t.transaction.BizExec(ctx, func(ctx context.Context) error {
		if req.ID <= 0 {
			emailConfigID, err = t.teamEmailConfigRepo.Create(ctx, req)
			return err
		}
		emailConfig, err := t.teamEmailConfigRepo.Get(ctx, req.ID)
		if err != nil {
			return err
		}
		return t.teamEmailConfigRepo.Update(ctx, req.WithEmailConfig(emailConfig))
	})
}

// GetEmailConfigs retrieves the email configuration for a team
func (t *Team) GetEmailConfigs(ctx context.Context, req *bo.ListEmailConfigRequest) (*bo.ListEmailConfigListReply, error) {
	configListReply, err := t.teamEmailConfigRepo.List(ctx, req)
	if err != nil {
		return nil, merr.ErrorInternalServer("failed to get email config").WithCause(err)
	}

	return configListReply, nil
}

func (t *Team) GetEmailConfig(ctx context.Context, emailConfigID uint32) (do.TeamEmailConfig, error) {
	return t.teamEmailConfigRepo.Get(ctx, emailConfigID)
}

// SaveSMSConfig saves the SMS configuration for a team
func (t *Team) SaveSMSConfig(ctx context.Context, req *bo.SaveSMSConfigRequest) (err error) {
	if err := req.Validate(); err != nil {
		return err
	}
	if err := t.teamSMSConfigRepo.CheckNameUnique(ctx, req.GetName(), req.ID); err != nil {
		return err
	}
	smsConfigID := req.ID
	defer t.publishNoticeSMSConfigDataChangeEvent(ctx, smsConfigID)
	return t.transaction.BizExec(ctx, func(ctx context.Context) error {
		if req.ID <= 0 {
			smsConfigID, err = t.teamSMSConfigRepo.Create(ctx, req)
			return err
		}
		smsConfig, err := t.teamSMSConfigRepo.Get(ctx, req.ID)
		if err != nil {
			return err
		}
		return t.teamSMSConfigRepo.Update(ctx, req.WithSMSConfig(smsConfig))
	})
}

// GetSMSConfigs retrieves SMS configurations for a team
func (t *Team) GetSMSConfigs(ctx context.Context, req *bo.ListSMSConfigRequest) (*bo.ListSMSConfigListReply, error) {
	return t.teamSMSConfigRepo.List(ctx, req)
}

func (t *Team) GetSMSConfig(ctx context.Context, smsConfigID uint32) (do.TeamSMSConfig, error) {
	return t.teamSMSConfigRepo.Get(ctx, smsConfigID)
}

func (t *Team) SaveTeamRole(ctx context.Context, req *bo.SaveTeamRoleReq) error {
	return t.transaction.BizExec(ctx, func(ctx context.Context) error {
		if req.GetID() <= 0 {
			return t.teamRoleRepo.Create(ctx, req)
		}
		teamRoleDo, err := t.teamRoleRepo.Get(ctx, req.GetID())
		if err != nil {
			return err
		}
		if len(req.GetMenuIds()) > 0 {
			menuDos, err := t.menuRepo.Find(ctx, req.GetMenuIds())
			if err != nil {
				return err
			}
			req.WithMenus(menuDos)
		}

		return t.teamRoleRepo.Update(ctx, req.WithRole(teamRoleDo))
	})
}

func (t *Team) GetTeamRoles(ctx context.Context, req *bo.ListRoleReq) (*bo.ListTeamRoleReply, error) {
	return t.teamRoleRepo.List(ctx, req)
}

func (t *Team) GetTeamRole(ctx context.Context, roleID uint32) (do.TeamRole, error) {
	return t.teamRoleRepo.Get(ctx, roleID)
}

func (t *Team) DeleteTeamRole(ctx context.Context, roleID uint32) error {
	return t.teamRoleRepo.Delete(ctx, roleID)
}

func (t *Team) UpdateTeamRoleStatus(ctx context.Context, req *bo.UpdateRoleStatusReq) error {
	return t.teamRoleRepo.UpdateStatus(ctx, req)
}

func (t *Team) ListTeam(ctx context.Context, req *bo.TeamListRequest) (*bo.TeamListReply, error) {
	return t.teamRepo.List(ctx, req)
}

func (t *Team) OperateLogList(ctx context.Context, req *bo.OperateLogListRequest) (*bo.OperateLogListReply, error) {
	return t.operateLogRepo.TeamList(ctx, req)
}

func (t *Team) GetTeamByID(ctx context.Context, teamID uint32) (do.Team, error) {
	return t.teamRepo.FindByID(ctx, teamID)
}

func (t *Team) GetTeamMembers(ctx context.Context, req *bo.TeamMemberListRequest) (*bo.TeamMemberListReply, error) {
	return t.memberRepo.List(ctx, req)
}

func (t *Team) SelectTeamMembers(ctx context.Context, req *bo.SelectTeamMembersRequest) (*bo.SelectTeamMembersReply, error) {
	return t.memberRepo.Select(ctx, req)
}

func (t *Team) UpdateMemberPosition(ctx context.Context, req *bo.UpdateMemberPositionReq) error {
	userID, ok := permission.GetUserIDByContext(ctx)
	if !ok {
		return merr.ErrorUnauthorized("user not found in context")
	}
	operatorDo, err := t.memberRepo.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}
	req.WithOperator(operatorDo)
	memberDo, err := t.memberRepo.Get(ctx, req.MemberID)
	if err != nil {
		return err
	}
	req.WithMember(memberDo)
	if err := req.Validate(); err != nil {
		return err
	}
	return t.memberRepo.UpdatePosition(ctx, req)
}

func (t *Team) UpdateMemberStatus(ctx context.Context, req *bo.UpdateMemberStatusReq) error {
	userID, ok := permission.GetUserIDByContext(ctx)
	if !ok {
		return merr.ErrorUnauthorized("user not found in context")
	}
	operatorDo, err := t.memberRepo.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}
	req.WithOperator(operatorDo)
	members, err := t.memberRepo.Find(ctx, req.MemberIds)
	if err != nil {
		return err
	}
	req.WithMembers(members)
	if err := req.Validate(); err != nil {
		return err
	}
	return t.memberRepo.UpdateStatus(ctx, req)
}

func (t *Team) UpdateMemberRoles(ctx context.Context, req *bo.UpdateMemberRolesReq) error {
	userID, ok := permission.GetUserIDByContext(ctx)
	if !ok {
		return merr.ErrorUnauthorized("user not found in context")
	}
	operatorDo, err := t.memberRepo.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}
	req.WithOperator(operatorDo)
	memberDo, err := t.memberRepo.Get(ctx, req.MemberID)
	if err != nil {
		return err
	}
	req.WithMember(memberDo)
	roles, err := t.teamRoleRepo.Find(ctx, req.RoleIds)
	if err != nil {
		return err
	}
	req.WithRoles(roles)
	if err := req.Validate(); err != nil {
		return err
	}
	return t.memberRepo.UpdateRoles(ctx, req)
}

func (t *Team) RemoveMember(ctx context.Context, req *bo.RemoveMemberReq) error {
	userID, ok := permission.GetUserIDByContext(ctx)
	if !ok {
		return merr.ErrorUnauthorized("user not found in context")
	}
	operatorDo, err := t.memberRepo.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}
	req.WithOperator(operatorDo)
	memberDo, err := t.memberRepo.Get(ctx, req.MemberID)
	if err != nil {
		return err
	}
	req.WithMember(memberDo)
	if err := req.Validate(); err != nil {
		return err
	}
	return t.memberRepo.UpdateStatus(ctx, req)
}

func (t *Team) InviteMember(ctx context.Context, req *bo.InviteMemberReq) error {
	userID, ok := permission.GetUserIDByContext(ctx)
	if !ok {
		return merr.ErrorUnauthorized("user not found in context")
	}
	operatorDo, err := t.memberRepo.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}
	req.WithOperator(operatorDo)
	inviterDo, err := t.userRepo.FindByEmail(ctx, req.UserEmail)
	if err != nil {
		return err
	}
	req.WithInviteUser(inviterDo)
	teamID, ok := permission.GetTeamIDByContext(ctx)
	if !ok {
		return merr.ErrorUnauthorized("team not found in context")
	}
	teamDo, err := t.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return err
	}
	req.WithTeam(teamDo)
	if err := req.Validate(); err != nil {
		return err
	}
	return t.transaction.MainExec(ctx, func(ctx context.Context) error {
		return t.inviteRepo.TeamInviteUser(ctx, req)
	})
}

func (t *Team) Jobs() []cron_server.CronJob {
	return []cron_server.CronJob{
		job.NewTeamJob(t.teamRepo, t.cacheRepo, t.helper.Logger()),
		job.NewTeamMemberJob(t.memberRepo, t.cacheRepo, t.helper.Logger()),
	}
}
