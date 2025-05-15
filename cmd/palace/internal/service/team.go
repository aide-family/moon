package service

import (
	"context"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/cmd/palace/internal/helper/permission"
	"github.com/moon-monitor/moon/cmd/palace/internal/service/build"
	"github.com/moon-monitor/moon/pkg/api/palace"
	"github.com/moon-monitor/moon/pkg/api/palace/common"
	"github.com/moon-monitor/moon/pkg/merr"
)

type TeamService struct {
	palace.UnimplementedTeamServer

	teamBiz    *biz.Team
	messageBiz *biz.Message
}

func NewTeamService(
	teamBiz *biz.Team,
	messageBiz *biz.Message,
) *TeamService {
	return &TeamService{
		teamBiz:    teamBiz,
		messageBiz: messageBiz,
	}
}

func (s *TeamService) SaveTeam(ctx context.Context, req *palace.SaveTeamRequest) (*common.EmptyReply, error) {
	params := build.ToSaveOneTeamRequest(req)
	if err := s.teamBiz.SaveTeam(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{Message: "保存团队信息成功"}, nil
}

func (s *TeamService) GetTeam(ctx context.Context, _ *common.EmptyRequest) (*common.TeamItem, error) {
	teamId, ok := permission.GetTeamIDByContext(ctx)
	if !ok {
		return nil, merr.ErrorPermissionDenied("please select team")
	}
	teamDo, err := s.teamBiz.GetTeamByID(ctx, teamId)
	if err != nil {
		return nil, err
	}
	return build.ToTeamItem(teamDo), nil
}

func (s *TeamService) GetTeamResources(ctx context.Context, req *common.EmptyRequest) (*palace.GetTeamResourcesReply, error) {
	return &palace.GetTeamResourcesReply{}, nil
}

func (s *TeamService) TransferTeam(ctx context.Context, req *palace.TransferTeamRequest) (*common.EmptyReply, error) {
	return &common.EmptyReply{}, nil
}

func (s *TeamService) InviteMember(ctx context.Context, req *palace.InviteMemberRequest) (*common.EmptyReply, error) {
	params := &bo.InviteMemberReq{
		UserEmail:    req.GetUserEmail(),
		Position:     vobj.Role(req.GetPosition()),
		RoleIds:      req.GetRoleIds(),
		SendEmailFun: s.messageBiz.SendEmail,
	}
	if err := s.teamBiz.InviteMember(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{Message: "邀请团队成员成功"}, nil
}

func (s *TeamService) RemoveMember(ctx context.Context, req *palace.RemoveMemberRequest) (*common.EmptyReply, error) {
	params := &bo.RemoveMemberReq{
		MemberID: req.GetMemberId(),
	}
	if err := s.teamBiz.RemoveMember(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{Message: "移除团队成员成功"}, nil
}

func (s *TeamService) GetTeamMembers(ctx context.Context, req *palace.GetTeamMembersRequest) (*palace.GetTeamMembersReply, error) {
	teamId, ok := permission.GetTeamIDByContext(ctx)
	if !ok {
		return nil, merr.ErrorPermissionDenied("please select team")
	}
	params := build.ToTeamMemberListRequest(req, teamId)
	membersReply, err := s.teamBiz.GetTeamMembers(ctx, params)
	if err != nil {
		return nil, err
	}
	return &palace.GetTeamMembersReply{
		Items:      build.ToTeamMemberItems(membersReply.Items),
		Pagination: build.ToPaginationReply(membersReply.PaginationReply),
	}, nil
}

func (s *TeamService) UpdateMemberPosition(ctx context.Context, req *palace.UpdateMemberPositionRequest) (*common.EmptyReply, error) {
	params := &bo.UpdateMemberPositionReq{
		MemberID: req.GetMemberId(),
		Position: vobj.Role(req.GetPosition()),
	}
	if err := s.teamBiz.UpdateMemberPosition(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{Message: "更新团队成员职位成功"}, nil
}

func (s *TeamService) UpdateMemberStatus(ctx context.Context, req *palace.UpdateMemberStatusRequest) (*common.EmptyReply, error) {
	params := &bo.UpdateMemberStatusReq{
		MemberIds: req.GetMemberIds(),
		Status:    vobj.MemberStatus(req.GetStatus()),
	}
	if err := s.teamBiz.UpdateMemberStatus(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{Message: "更新团队成员状态成功"}, nil
}

func (s *TeamService) UpdateMemberRoles(ctx context.Context, req *palace.UpdateMemberRolesRequest) (*common.EmptyReply, error) {
	params := &bo.UpdateMemberRolesReq{
		MemberId: req.GetMemberId(),
		RoleIds:  req.GetRoleIds(),
	}
	if err := s.teamBiz.UpdateMemberRoles(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{Message: "更新团队成员角色成功"}, nil
}

func (s *TeamService) GetTeamRoles(ctx context.Context, req *palace.GetTeamRolesRequest) (*palace.GetTeamRolesReply, error) {
	params := build.ToListRoleRequest(req)
	roleReply, err := s.teamBiz.GetTeamRoles(ctx, params)
	if err != nil {
		return nil, err
	}
	return &palace.GetTeamRolesReply{
		Items:      build.ToTeamRoleItems(roleReply.Items),
		Pagination: build.ToPaginationReply(roleReply.PaginationReply),
	}, nil
}

func (s *TeamService) SaveTeamRole(ctx context.Context, req *palace.SaveTeamRoleRequest) (*common.EmptyReply, error) {
	params := build.ToSaveTeamRoleRequest(req)
	if err := s.teamBiz.SaveTeamRole(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{Message: "保存团队角色成功"}, nil
}

func (s *TeamService) DeleteTeamRole(ctx context.Context, req *palace.DeleteTeamRoleRequest) (*common.EmptyReply, error) {
	if err := s.teamBiz.DeleteTeamRole(ctx, req.GetRoleId()); err != nil {
		return nil, err
	}
	return &common.EmptyReply{Message: "删除团队角色成功"}, nil
}

func (s *TeamService) UpdateTeamRoleStatus(ctx context.Context, req *palace.UpdateTeamRoleStatusRequest) (*common.EmptyReply, error) {
	params := &bo.UpdateRoleStatusReq{
		RoleID: req.GetRoleId(),
		Status: vobj.GlobalStatus(req.GetStatus()),
	}
	if err := s.teamBiz.UpdateTeamRoleStatus(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{Message: "更新团队角色状态成功"}, nil
}

func (s *TeamService) SaveEmailConfig(ctx context.Context, req *palace.SaveEmailConfigRequest) (*common.EmptyReply, error) {
	if err := s.teamBiz.SaveEmailConfig(ctx, build.ToSaveEmailConfigRequest(req)); err != nil {
		return nil, err
	}
	return &common.EmptyReply{Message: "保存邮件配置成功"}, nil
}

func (s *TeamService) GetEmailConfigs(ctx context.Context, req *palace.GetEmailConfigsRequest) (*palace.GetEmailConfigsReply, error) {
	params := build.ToListEmailConfigRequest(req)
	config, err := s.teamBiz.GetEmailConfigs(ctx, params)
	if err != nil {
		return nil, err
	}
	return build.ToEmailConfigReply(config), nil
}

func (s *TeamService) SaveSMSConfig(ctx context.Context, req *palace.SaveSMSConfigRequest) (*common.EmptyReply, error) {
	if err := s.teamBiz.SaveSMSConfig(ctx, build.ToSaveSMSConfigRequest(req)); err != nil {
		return nil, err
	}
	return &common.EmptyReply{Message: "保存短信配置成功"}, nil
}

func (s *TeamService) GetSMSConfigs(ctx context.Context, req *palace.GetSMSConfigsRequest) (*palace.GetSMSConfigsReply, error) {
	params := build.ToListSMSConfigRequest(req)
	config, err := s.teamBiz.GetSMSConfigs(ctx, params)
	if err != nil {
		return nil, err
	}
	return build.ToSMSConfigReply(config), nil
}

func (s *TeamService) OperateLogList(ctx context.Context, req *palace.TeamOperateLogListRequest) (*palace.TeamOperateLogListReply, error) {
	params := build.ToOperateLogListRequest(req)
	operateLogReply, err := s.teamBiz.OperateLogList(ctx, params)
	if err != nil {
		return nil, err
	}
	return &palace.TeamOperateLogListReply{
		Items:      build.ToOperateLogItems(operateLogReply.Items),
		Pagination: build.ToPaginationReply(operateLogReply.PaginationReply),
	}, nil
}
