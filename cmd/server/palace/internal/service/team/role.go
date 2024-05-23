package team

import (
	"context"

	"github.com/aide-cloud/moon/api/admin"
	pb "github.com/aide-cloud/moon/api/admin/team"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/do/model"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/service/build"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/aide-cloud/moon/pkg/vobj"
)

type RoleService struct {
	pb.UnimplementedRoleServer

	teamRoleBiz *biz.TeamRoleBiz
}

func NewRoleService(teamRoleBiz *biz.TeamRoleBiz) *RoleService {
	return &RoleService{
		teamRoleBiz: teamRoleBiz,
	}
}

func (s *RoleService) CreateRole(ctx context.Context, req *pb.CreateRoleRequest) (*pb.CreateRoleReply, error) {
	params := &bo.CreateTeamRoleParams{
		TeamID:      req.GetTeamId(),
		Name:        req.GetName(),
		Remark:      req.GetRemark(),
		Status:      vobj.StatusEnable,
		Permissions: req.GetPermissions(),
	}
	_, err := s.teamRoleBiz.CreateTeamRole(ctx, params)
	if err != nil {
		return nil, err
	}
	return &pb.CreateRoleReply{}, nil
}

func (s *RoleService) UpdateRole(ctx context.Context, req *pb.UpdateRoleRequest) (*pb.UpdateRoleReply, error) {
	data := req.GetData()
	params := &bo.UpdateTeamRoleParams{
		ID:          req.GetId(),
		Name:        data.GetName(),
		Remark:      data.GetRemark(),
		Permissions: data.GetPermissions(),
	}
	if err := s.teamRoleBiz.UpdateTeamRole(ctx, params); err != nil {
		return nil, err
	}
	return &pb.UpdateRoleReply{}, nil
}

func (s *RoleService) DeleteRole(ctx context.Context, req *pb.DeleteRoleRequest) (*pb.DeleteRoleReply, error) {
	if err := s.teamRoleBiz.DeleteTeamRole(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &pb.DeleteRoleReply{}, nil
}

func (s *RoleService) GetRole(ctx context.Context, req *pb.GetRoleRequest) (*pb.GetRoleReply, error) {
	roleDetail, err := s.teamRoleBiz.GetTeamRole(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &pb.GetRoleReply{
		Role: build.NewTeamRoleBuild(roleDetail).ToApi(),
	}, nil
}

func (s *RoleService) ListRole(ctx context.Context, req *pb.ListRoleRequest) (*pb.ListRoleReply, error) {
	params := &bo.ListTeamRoleParams{
		TeamID:  req.GetTeamId(),
		Keyword: req.GetKeyword(),
	}
	teamRoles, err := s.teamRoleBiz.ListTeamRole(ctx, params)
	if err != nil {
		return nil, err
	}
	return &pb.ListRoleReply{
		List: types.SliceTo(teamRoles, func(item *model.SysTeamRole) *admin.TeamRole {
			return build.NewTeamRoleBuild(item).ToApi()
		}),
	}, nil
}

func (s *RoleService) UpdateRoleStatus(ctx context.Context, req *pb.UpdateRoleStatusRequest) (*pb.UpdateRoleStatusReply, error) {
	if err := s.teamRoleBiz.UpdateTeamRoleStatus(ctx, vobj.Status(req.GetStatus()), req.GetId()); err != nil {
		return nil, err
	}
	return &pb.UpdateRoleStatusReply{}, nil
}

func (s *RoleService) GetRoleSelectList(ctx context.Context, req *pb.GetRoleSelectListRequest) (*pb.GetRoleSelectListReply, error) {
	params := &bo.ListTeamRoleParams{
		TeamID:  req.GetTeamId(),
		Keyword: req.GetKeyword(),
	}
	teamRoles, err := s.teamRoleBiz.ListTeamRole(ctx, params)
	if err != nil {
		return nil, err
	}
	return &pb.GetRoleSelectListReply{
		List: types.SliceTo(teamRoles, func(item *model.SysTeamRole) *admin.Select {
			return build.NewTeamRoleBuild(item).ToSelect()
		}),
	}, nil
}
