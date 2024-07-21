package team

import (
	"context"

	"github.com/aide-family/moon/api/admin"
	teamapi "github.com/aide-family/moon/api/admin/team"
	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type RoleService struct {
	teamapi.UnimplementedRoleServer

	teamRoleBiz *biz.TeamRoleBiz
}

func NewRoleService(teamRoleBiz *biz.TeamRoleBiz) *RoleService {
	return &RoleService{
		teamRoleBiz: teamRoleBiz,
	}
}

func (s *RoleService) CreateRole(ctx context.Context, req *teamapi.CreateRoleRequest) (*teamapi.CreateRoleReply, error) {
	params := &bo.CreateTeamRoleParams{
		TeamID:      req.GetTeamId(),
		Name:        req.GetName(),
		Remark:      req.GetRemark(),
		Status:      vobj.StatusEnable,
		Permissions: req.GetPermissions(),
	}
	_, err := s.teamRoleBiz.CreateTeamRole(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &teamapi.CreateRoleReply{}, nil
}

func (s *RoleService) UpdateRole(ctx context.Context, req *teamapi.UpdateRoleRequest) (*teamapi.UpdateRoleReply, error) {
	data := req.GetData()
	params := &bo.UpdateTeamRoleParams{
		ID:          req.GetId(),
		Name:        data.GetName(),
		Remark:      data.GetRemark(),
		Permissions: data.GetPermissions(),
	}
	if err := s.teamRoleBiz.UpdateTeamRole(ctx, params); !types.IsNil(err) {
		return nil, err
	}
	return &teamapi.UpdateRoleReply{}, nil
}

func (s *RoleService) DeleteRole(ctx context.Context, req *teamapi.DeleteRoleRequest) (*teamapi.DeleteRoleReply, error) {
	if err := s.teamRoleBiz.DeleteTeamRole(ctx, req.GetId()); !types.IsNil(err) {
		return nil, err
	}
	return &teamapi.DeleteRoleReply{}, nil
}

func (s *RoleService) GetRole(ctx context.Context, req *teamapi.GetRoleRequest) (*teamapi.GetRoleReply, error) {
	roleDetail, err := s.teamRoleBiz.GetTeamRole(ctx, req.GetId())
	if !types.IsNil(err) {
		return nil, err
	}
	return &teamapi.GetRoleReply{
		Role: build.NewBuilder().WithApiTeamRole(roleDetail).ToApi(),
	}, nil
}

func (s *RoleService) ListRole(ctx context.Context, req *teamapi.ListRoleRequest) (*teamapi.ListRoleReply, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	params := &bo.ListTeamRoleParams{
		TeamID:  claims.GetTeam(),
		Keyword: req.GetKeyword(),
	}
	teamRoles, err := s.teamRoleBiz.ListTeamRole(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &teamapi.ListRoleReply{
		List: types.SliceTo(teamRoles, func(item *bizmodel.SysTeamRole) *admin.TeamRole {
			return build.NewBuilder().WithApiTeamRole(item).ToApi()
		}),
	}, nil
}

func (s *RoleService) UpdateRoleStatus(ctx context.Context, req *teamapi.UpdateRoleStatusRequest) (*teamapi.UpdateRoleStatusReply, error) {
	if err := s.teamRoleBiz.UpdateTeamRoleStatus(ctx, vobj.Status(req.GetStatus()), req.GetId()); !types.IsNil(err) {
		return nil, err
	}
	return &teamapi.UpdateRoleStatusReply{}, nil
}

func (s *RoleService) GetRoleSelectList(ctx context.Context, req *teamapi.GetRoleSelectListRequest) (*teamapi.GetRoleSelectListReply, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	params := &bo.ListTeamRoleParams{
		TeamID:  claims.GetTeam(),
		Keyword: req.GetKeyword(),
	}
	teamRoles, err := s.teamRoleBiz.ListTeamRole(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &teamapi.GetRoleSelectListReply{
		List: types.SliceTo(teamRoles, func(item *bizmodel.SysTeamRole) *admin.Select {
			return build.NewBuilder().WithSelectTeamRole(item).ToSelect()
		}),
	}, nil
}
