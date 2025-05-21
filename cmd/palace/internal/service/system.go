package service

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz"
	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/api/palace"
	"github.com/aide-family/moon/pkg/api/palace/common"
)

func NewSystemService(
	userBiz *biz.UserBiz,
	messageBiz *biz.Message,
	teamBiz *biz.Team,
	systemBiz *biz.System,
	logsBiz *biz.Logs,
) *SystemService {
	return &SystemService{
		userBiz:    userBiz,
		messageBiz: messageBiz,
		teamBiz:    teamBiz,
		systemBiz:  systemBiz,
		logsBiz:    logsBiz,
	}
}

type SystemService struct {
	palace.UnimplementedSystemServer
	userBiz    *biz.UserBiz
	messageBiz *biz.Message
	teamBiz    *biz.Team
	systemBiz  *biz.System
	logsBiz    *biz.Logs
}

func (s *SystemService) UpdateUser(ctx context.Context, req *palace.UpdateUserRequest) (*common.EmptyReply, error) {
	params := build.ToUserUpdateInfo(req)
	if err := s.userBiz.UpdateUserBaseInfo(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (s *SystemService) UpdateUserStatus(ctx context.Context, req *palace.UpdateUserStatusRequest) (*common.EmptyReply, error) {
	params := &bo.UpdateUserStatusRequest{
		UserIds: req.GetUserIds(),
		Status:  vobj.UserStatus(req.GetStatus()),
	}
	if err := s.userBiz.UpdateUserStatus(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (s *SystemService) ResetUserPassword(ctx context.Context, req *palace.ResetUserPasswordRequest) (*common.EmptyReply, error) {
	params := &bo.ResetUserPasswordRequest{
		UserId:       req.GetUserId(),
		SendEmailFun: s.messageBiz.SendEmail,
	}
	if err := s.userBiz.ResetUserPassword(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (s *SystemService) UpdateUserPosition(ctx context.Context, req *palace.UpdateUserPositionRequest) (*common.EmptyReply, error) {
	params := &bo.UpdateUserPositionRequest{
		UserId:   req.GetUserId(),
		Position: vobj.Position(req.GetPosition()),
	}
	if err := s.userBiz.UpdateUserPosition(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (s *SystemService) GetUser(ctx context.Context, req *palace.GetUserRequest) (*common.UserItem, error) {
	userDo, err := s.userBiz.GetUser(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}

	return build.ToUserItem(userDo), nil
}

func (s *SystemService) GetUserList(ctx context.Context, req *palace.GetUserListRequest) (*palace.GetUserListReply, error) {
	params := build.ToUserListRequest(req)
	userReply, err := s.userBiz.ListUser(ctx, params)
	if err != nil {
		return nil, err
	}
	return &palace.GetUserListReply{
		Items:      build.ToUserItems(userReply.Items),
		Pagination: build.ToPaginationReply(userReply.PaginationReply),
	}, nil
}

func (s *SystemService) GetTeamList(ctx context.Context, req *palace.GetTeamListRequest) (*palace.GetTeamListReply, error) {
	params := build.ToTeamListRequest(req)
	teamReply, err := s.teamBiz.ListTeam(ctx, params)
	if err != nil {
		return nil, err
	}

	return &palace.GetTeamListReply{
		Items:      build.ToTeamItems(teamReply.Items),
		Pagination: build.ToPaginationReply(teamReply.PaginationReply),
	}, nil
}

func (s *SystemService) GetTeam(ctx context.Context, req *palace.GetTeamRequest) (*common.TeamItem, error) {
	teamDo, err := s.teamBiz.GetTeamByID(ctx, req.GetTeamId())
	if err != nil {
		return nil, err
	}
	return build.ToTeamItem(teamDo), nil
}

func (s *SystemService) GetSystemRole(ctx context.Context, req *palace.GetSystemRoleRequest) (*common.SystemRoleItem, error) {
	roleDo, err := s.systemBiz.GetRole(ctx, req.GetRoleId())
	if err != nil {
		return nil, err
	}
	return build.ToSystemRoleItem(roleDo), nil
}

func (s *SystemService) GetSystemRoles(ctx context.Context, req *palace.GetSystemRolesRequest) (*palace.GetSystemRolesReply, error) {
	params := build.ToListRoleRequest(req)
	roleReply, err := s.systemBiz.GetRoles(ctx, params)
	if err != nil {
		return nil, err
	}
	return &palace.GetSystemRolesReply{
		Items:      build.ToSystemRoleItems(roleReply.Items),
		Pagination: build.ToPaginationReply(roleReply.PaginationReply),
	}, nil
}

func (s *SystemService) SaveRole(ctx context.Context, req *palace.SaveRoleRequest) (*common.EmptyReply, error) {
	params := build.ToSaveRoleRequest(req)
	if err := s.systemBiz.SaveRole(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (s *SystemService) UpdateRoleStatus(ctx context.Context, req *palace.UpdateRoleStatusRequest) (*common.EmptyReply, error) {
	params := &bo.UpdateRoleStatusReq{
		RoleID: req.GetRoleId(),
		Status: vobj.GlobalStatus(req.GetStatus()),
	}
	if err := s.systemBiz.UpdateRoleStatus(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (s *SystemService) UpdateUserRoles(ctx context.Context, req *palace.UpdateUserRolesRequest) (*common.EmptyReply, error) {
	params := &bo.UpdateUserRolesReq{
		RoleIDs: req.GetRoleIds(),
		UserID:  req.GetUserId(),
	}
	if err := s.systemBiz.UpdateUserRoles(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (s *SystemService) UpdateRoleUsers(ctx context.Context, req *palace.UpdateRoleUsersRequest) (*common.EmptyReply, error) {
	params := &bo.UpdateRoleUsersReq{
		RoleID:  req.GetRoleId(),
		UserIDs: req.GetUserIds(),
	}
	if err := s.systemBiz.UpdateRoleUsers(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (s *SystemService) GetTeamAuditList(ctx context.Context, req *palace.GetTeamAuditListRequest) (*palace.GetTeamAuditListReply, error) {
	params := build.ToTeamAuditListRequest(req)
	teamAuditReply, err := s.systemBiz.GetTeamAuditList(ctx, params)
	if err != nil {
		return nil, err
	}
	return &palace.GetTeamAuditListReply{
		Items:      build.ToTeamAuditItems(teamAuditReply.Items),
		Pagination: build.ToPaginationReply(teamAuditReply.PaginationReply),
	}, nil
}

func (s *SystemService) UpdateTeamAuditStatus(ctx context.Context, req *palace.UpdateTeamAuditStatusRequest) (*common.EmptyReply, error) {
	params := &bo.UpdateTeamAuditStatusReq{
		AuditID: req.GetAuditId(),
		Status:  vobj.StatusAudit(req.GetStatus()),
		Reason:  req.GetReason(),
	}
	if err := s.systemBiz.UpdateTeamAuditStatus(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (s *SystemService) OperateLogList(ctx context.Context, req *common.OperateLogListRequest) (*common.OperateLogListReply, error) {
	params := build.ToOperateLogListRequest(req)
	operateLogReply, err := s.systemBiz.OperateLogList(ctx, params)
	if err != nil {
		return nil, err
	}
	return &common.OperateLogListReply{
		Items:      build.ToOperateLogItems(operateLogReply.Items),
		Pagination: build.ToPaginationReply(operateLogReply.PaginationReply),
	}, nil
}

func (s *SystemService) GetSendMessageLogs(ctx context.Context, req *palace.GetSendMessageLogsRequest) (*palace.GetSendMessageLogsReply, error) {
	params := build.ToListSendMessageLogParams(req)
	logsReply, err := s.logsBiz.GetSendMessageLogs(ctx, params)
	if err != nil {
		return nil, err
	}
	return &palace.GetSendMessageLogsReply{
		Items:      build.ToSendMessageLogs(logsReply.Items),
		Pagination: build.ToPaginationReply(logsReply.PaginationReply),
	}, nil
}

func (s *SystemService) GetSendMessageLog(ctx context.Context, req *palace.OperateOneSendMessageRequest) (*common.SendMessageLogItem, error) {
	params := build.ToGetSendMessageLogParams(req.GetRequestId())
	logDo, err := s.logsBiz.GetSendMessageLog(ctx, params)
	if err != nil {
		return nil, err
	}
	return build.ToSendMessageLog(logDo), nil
}

func (s *SystemService) RetrySendMessage(ctx context.Context, req *palace.OperateOneSendMessageRequest) (*common.EmptyReply, error) {
	params := build.ToRetrySendMessageParams(req.GetRequestId())
	if err := s.logsBiz.RetrySendMessage(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}
