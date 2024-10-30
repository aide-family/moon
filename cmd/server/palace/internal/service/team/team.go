package team

import (
	"context"

	teamapi "github.com/aide-family/moon/api/admin/team"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

// Service 团队管理服务
type Service struct {
	teamapi.UnimplementedTeamServer

	teamBiz *biz.TeamBiz
}

// NewTeamService 创建团队服务
func NewTeamService(teamBiz *biz.TeamBiz) *Service {
	return &Service{
		teamBiz: teamBiz,
	}
}

// CreateTeam 创建团队
func (s *Service) CreateTeam(ctx context.Context, req *teamapi.CreateTeamRequest) (*teamapi.CreateTeamReply, error) {
	params := builder.NewParamsBuild().WithContext(ctx).WithContext(ctx).TeamModuleBuilder().WithCreateTeamRequest(req).ToBo()
	teamDo, err := s.teamBiz.CreateTeam(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &teamapi.CreateTeamReply{
		Detail: builder.NewParamsBuild().WithContext(ctx).TeamModuleBuilder().DoTeamBuilder().ToAPI(teamDo),
	}, nil
}

// UpdateTeam 更新团队
func (s *Service) UpdateTeam(ctx context.Context, req *teamapi.UpdateTeamRequest) (*teamapi.UpdateTeamReply, error) {
	params := builder.NewParamsBuild().WithContext(ctx).TeamModuleBuilder().WithUpdateTeamRequest(req).ToBo()
	if err := s.teamBiz.UpdateTeam(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &teamapi.UpdateTeamReply{}, nil
}

// GetTeam 获取团队
func (s *Service) GetTeam(ctx context.Context, req *teamapi.GetTeamRequest) (*teamapi.GetTeamReply, error) {
	teamInfo, err := s.teamBiz.GetTeam(ctx, req.GetId())
	if !types.IsNil(err) {
		return nil, err
	}
	return &teamapi.GetTeamReply{
		Detail: builder.NewParamsBuild().WithContext(ctx).TeamModuleBuilder().DoTeamBuilder().ToAPI(teamInfo),
	}, nil
}

// ListTeam 获取团队列表
func (s *Service) ListTeam(ctx context.Context, req *teamapi.ListTeamRequest) (*teamapi.ListTeamReply, error) {
	param := builder.NewParamsBuild().WithContext(ctx).TeamModuleBuilder().WithListTeamRequest(req).ToBo()
	teamList, err := s.teamBiz.ListTeam(ctx, param)
	if !types.IsNil(err) {
		return nil, err
	}
	return &teamapi.ListTeamReply{
		Pagination: builder.NewParamsBuild().WithContext(ctx).PaginationModuleBuilder().ToAPI(param.Page),
		List:       builder.NewParamsBuild().WithContext(ctx).TeamModuleBuilder().DoTeamBuilder().ToAPIs(teamList),
	}, nil
}

// UpdateTeamStatus 更新团队状态
func (s *Service) UpdateTeamStatus(ctx context.Context, req *teamapi.UpdateTeamStatusRequest) (*teamapi.UpdateTeamStatusReply, error) {
	if err := s.teamBiz.UpdateTeamStatus(ctx, vobj.Status(req.GetStatus()), req.GetId()); !types.IsNil(err) {
		return nil, err
	}
	return &teamapi.UpdateTeamStatusReply{}, nil
}

// MyTeam 获取当前用户团队列表
func (s *Service) MyTeam(ctx context.Context, _ *teamapi.MyTeamRequest) (*teamapi.MyTeamReply, error) {
	teamList, err := s.teamBiz.GetUserTeamList(ctx, middleware.GetUserID(ctx))
	if !types.IsNil(err) {
		return nil, err
	}
	return &teamapi.MyTeamReply{
		List: builder.NewParamsBuild().WithContext(ctx).TeamModuleBuilder().DoTeamBuilder().ToAPIs(teamList),
	}, nil
}

// RemoveTeamMember 移除团队成员
func (s *Service) RemoveTeamMember(ctx context.Context, req *teamapi.RemoveTeamMemberRequest) (*teamapi.RemoveTeamMemberReply, error) {
	params := builder.NewParamsBuild().WithContext(ctx).TeamModuleBuilder().WithRemoveTeamMemberRequest(req).ToBo()
	if err := s.teamBiz.RemoveTeamMember(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &teamapi.RemoveTeamMemberReply{}, nil
}

// SetTeamAdmin 设置团队管理员
func (s *Service) SetTeamAdmin(ctx context.Context, req *teamapi.SetTeamAdminRequest) (*teamapi.SetTeamAdminReply, error) {
	params := builder.NewParamsBuild().WithContext(ctx).TeamModuleBuilder().WithSetTeamAdminRequest(req).ToBo()
	if err := s.teamBiz.SetTeamAdmin(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &teamapi.SetTeamAdminReply{}, nil
}

// RemoveTeamAdmin 移除团队管理员
func (s *Service) RemoveTeamAdmin(ctx context.Context, req *teamapi.RemoveTeamAdminRequest) (*teamapi.RemoveTeamAdminReply, error) {
	params := builder.NewParamsBuild().WithContext(ctx).TeamModuleBuilder().WithRemoveTeamAdminRequest(req).ToBo()
	if err := s.teamBiz.SetTeamAdmin(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &teamapi.RemoveTeamAdminReply{}, nil
}

// SetMemberRole 设置团队成员角色
func (s *Service) SetMemberRole(ctx context.Context, req *teamapi.SetMemberRoleRequest) (*teamapi.SetMemberRoleReply, error) {
	params := builder.NewParamsBuild().WithContext(ctx).TeamModuleBuilder().WithSetMemberRoleRequest(req).ToBo()
	if err := s.teamBiz.SetMemberRole(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &teamapi.SetMemberRoleReply{}, nil
}

// ListTeamMember 获取团队成员列表
func (s *Service) ListTeamMember(ctx context.Context, req *teamapi.ListTeamMemberRequest) (*teamapi.ListTeamMemberReply, error) {
	params := builder.NewParamsBuild().WithContext(ctx).TeamModuleBuilder().WithListTeamMemberRequest(req).ToBo()
	memberList, err := s.teamBiz.ListTeamMember(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &teamapi.ListTeamMemberReply{
		Pagination: builder.NewParamsBuild().WithContext(ctx).PaginationModuleBuilder().ToAPI(params.Page),
		List:       builder.NewParamsBuild().WithContext(ctx).TeamMemberModuleBuilder().DoTeamMemberBuilder().ToAPIs(memberList),
	}, nil
}

// TransferTeamLeader 转移团队负责人
func (s *Service) TransferTeamLeader(ctx context.Context, req *teamapi.TransferTeamLeaderRequest) (*teamapi.TransferTeamLeaderReply, error) {
	params := builder.NewParamsBuild().WithContext(ctx).TeamModuleBuilder().WithTransferTeamLeaderRequest(req).ToBo()
	if err := s.teamBiz.TransferTeamLeader(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &teamapi.TransferTeamLeaderReply{}, nil
}

// SetTeamMailConfig 设置团队邮件配置
func (s *Service) SetTeamMailConfig(ctx context.Context, req *teamapi.SetTeamMailConfigRequest) (*teamapi.SetTeamMailConfigReply, error) {
	params := builder.NewParamsBuild().WithContext(ctx).TeamModuleBuilder().WithSetTeamMailConfigRequest(req).ToBo()
	if err := s.teamBiz.SetTeamMailConfig(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &teamapi.SetTeamMailConfigReply{}, nil
}

// GetTeamMailConfig 获取团队邮件配置
func (s *Service) GetTeamMailConfig(ctx context.Context, _ *teamapi.GetTeamMailConfigRequest) (*teamapi.GetTeamMailConfigReply, error) {
	config, err := s.teamBiz.GetTeamMailConfig(ctx)
	if !types.IsNil(err) {
		return nil, err
	}
	return &teamapi.GetTeamMailConfigReply{
		Config: &conf.EmailConfig{
			User: config.User,
			Pass: config.Pass,
			Host: config.Host,
			Port: config.Port,
		},
		Remark: config.Remark,
	}, nil
}

// UpdateTeamMemberStatus 更新团队成员状态
func (s *Service) UpdateTeamMemberStatus(ctx context.Context, req *teamapi.UpdateTeamMemberStatusRequest) (*teamapi.UpdateTeamMemberStatusReply, error) {
	if err := s.teamBiz.UpdateTeamMemberStatus(ctx, vobj.Status(req.GetStatus()), req.GetMemberIds()...); !types.IsNil(err) {
		return nil, err
	}
	return &teamapi.UpdateTeamMemberStatusReply{}, nil
}

// GetTeamMemberDetail 获取团队成员详情
func (s *Service) GetTeamMemberDetail(ctx context.Context, req *teamapi.GetTeamMemberDetailRequest) (*teamapi.GetTeamMemberDetailReply, error) {
	member, err := s.teamBiz.GetTeamMemberDetail(ctx, req.GetId())
	if !types.IsNil(err) {
		return nil, err
	}
	return &teamapi.GetTeamMemberDetailReply{
		Detail: builder.NewParamsBuild().WithContext(ctx).TeamMemberModuleBuilder().DoTeamMemberBuilder().ToAPI(member),
	}, nil
}
