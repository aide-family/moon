package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/moon-monitor/moon/cmd/palace/internal/helper/permission"
	"github.com/moon-monitor/moon/pkg/merr"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/repository"
)

func NewSystem(
	roleRepo repository.Role,
	userRepo repository.User,
	auditRepo repository.Audit,
	operateLogRepo repository.OperateLog,
	transactionRepo repository.Transaction,
	logger log.Logger,
) *System {
	return &System{
		roleRepo:        roleRepo,
		userRepo:        userRepo,
		auditRepo:       auditRepo,
		operateLogRepo:  operateLogRepo,
		transactionRepo: transactionRepo,
		helper:          log.NewHelper(log.With(logger, "module", "biz.system")),
	}
}

type System struct {
	roleRepo        repository.Role
	userRepo        repository.User
	auditRepo       repository.Audit
	operateLogRepo  repository.OperateLog
	transactionRepo repository.Transaction
	helper          *log.Helper
}

func (s *System) GetRole(ctx context.Context, roleId uint32) (do.Role, error) {
	return s.roleRepo.Get(ctx, roleId)
}

func (s *System) GetRoles(ctx context.Context, req *bo.ListRoleReq) (*bo.ListRoleReply, error) {
	return s.roleRepo.List(ctx, req)
}

func (s *System) SaveRole(ctx context.Context, req *bo.SaveRoleReq) error {
	return s.transactionRepo.BizExec(ctx, func(ctx context.Context) error {
		if req.GetID() <= 0 {
			return s.roleRepo.Create(ctx, req)
		}
		roleDo, err := s.roleRepo.Get(ctx, req.GetID())
		if err != nil {
			return err
		}
		req.WithRole(roleDo)
		return s.roleRepo.Update(ctx, req)
	})
}

func (s *System) UpdateRoleStatus(ctx context.Context, req *bo.UpdateRoleStatusReq) error {
	return s.roleRepo.UpdateStatus(ctx, req)
}

func (s *System) UpdateRoleUsers(ctx context.Context, req *bo.UpdateRoleUsersReq) error {
	operatorId, ok := permission.GetUserIDByContext(ctx)
	if !ok {
		return merr.ErrorUnauthorized("user not found in context")
	}
	operatorDo, err := s.userRepo.Get(ctx, operatorId)
	if err != nil {
		return err
	}
	req.WithOperator(operatorDo)
	roleDo, err := s.roleRepo.Get(ctx, req.RoleID)
	if err != nil {
		return err
	}
	req.WithRole(roleDo)
	userDos, err := s.userRepo.Find(ctx, req.UserIDs)
	if err != nil {
		return err
	}
	req.WithUsers(userDos)
	if err := req.Validate(); err != nil {
		return err
	}
	return s.roleRepo.UpdateUsers(ctx, req)
}

func (s *System) UpdateUserRoles(ctx context.Context, req *bo.UpdateUserRolesReq) error {
	userId, ok := permission.GetUserIDByContext(ctx)
	if !ok {
		return merr.ErrorUnauthorized("user not found in context")
	}
	operatorDo, err := s.userRepo.Get(ctx, userId)
	if err != nil {
		return err
	}
	req.WithOperator(operatorDo)
	userDo, err := s.userRepo.Get(ctx, req.UserID)
	if err != nil {
		return err
	}
	req.WithUser(userDo)
	roleDos, err := s.roleRepo.Find(ctx, req.RoleIDs)
	if err != nil {
		return err
	}
	req.WithRoles(roleDos)
	if err := req.Validate(); err != nil {
		return err
	}
	return s.userRepo.UpdateUserRoles(ctx, req)
}

func (s *System) GetTeamAuditList(ctx context.Context, req *bo.TeamAuditListRequest) (*bo.TeamAuditListReply, error) {
	return s.auditRepo.TeamAuditList(ctx, req)
}

func (s *System) UpdateTeamAuditStatus(ctx context.Context, req *bo.UpdateTeamAuditStatusReq) error {
	teamAuditDo, err := s.auditRepo.Get(ctx, req.GetAuditID())
	if err != nil {
		return err
	}
	req.WithTeamAudit(teamAuditDo)
	if err := req.Validate(); err != nil {
		return err
	}
	return s.auditRepo.UpdateTeamAuditStatus(ctx, req)
}

func (s *System) OperateLogList(ctx context.Context, req *bo.OperateLogListRequest) (*bo.OperateLogListReply, error) {
	return s.operateLogRepo.List(ctx, req)
}
