package service

import (
	"context"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/role"
)

type RoleService struct {
	v1.UnimplementedRoleServiceServer
	authority *role.Authority
}

func NewRoleService(authority *role.Authority) *RoleService {
	return &RoleService{authority: authority}
}

func (s *RoleService) CreateRole(ctx context.Context, req *v1.CreateRoleRequest) (*v1.CreateRoleResponse, error) {
	item, err := s.authority.CreateRole(ctx, req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.CreateRoleResponse{Role: roleItem(*item)}, nil
}

func (s *RoleService) UpdateRole(ctx context.Context, req *v1.UpdateRoleRequest) (*v1.UpdateRoleResponse, error) {
	item, err := s.authority.UpdateRole(ctx, req.GetUid(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateRoleResponse{Role: roleItem(*item)}, nil
}

func (s *RoleService) UpdateRoleStatus(ctx context.Context, req *v1.UpdateRoleStatusRequest) (*v1.UpdateRoleStatusResponse, error) {
	item, err := s.authority.UpdateRoleStatus(ctx, req.GetUid(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateRoleStatusResponse{Role: roleItem(*item)}, nil
}

func (s *RoleService) DeleteRole(ctx context.Context, req *v1.DeleteRoleRequest) (*v1.DeleteRoleResponse, error) {
	if err := s.authority.DeleteRole(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteRoleResponse{}, nil
}

func (s *RoleService) GetRole(ctx context.Context, req *v1.GetRoleRequest) (*v1.GetRoleResponse, error) {
	item, err := s.authority.GetRole(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetRoleResponse{Role: roleItem(*item)}, nil
}

func (s *RoleService) ListRoles(ctx context.Context, req *v1.ListRolesRequest) (*v1.ListRolesResponse, error) {
	items, total, page, pageSize, err := s.authority.ListRoles(ctx, req.GetPage(), req.GetPageSize(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.ListRolesResponse{Items: roleItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *RoleService) SelectRoles(ctx context.Context, req *v1.SelectRolesRequest) (*v1.SelectRolesResponse, error) {
	items, err := s.authority.SelectRoles(ctx, req.GetLimit(), req.GetLastUid())
	if err != nil {
		return nil, err
	}
	return &v1.SelectRolesResponse{Items: selectItems(items), LastUid: lastUID(items)}, nil
}

func (s *RoleService) UpdateRolePermissions(ctx context.Context, req *v1.UpdateRolePermissionsRequest) (*v1.UpdateRolePermissionsResponse, error) {
	item, err := s.authority.UpdateRolePermissions(ctx, req.GetUid(), req.GetPermissions())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateRolePermissionsResponse{Role: roleItem(*item)}, nil
}

func (s *RoleService) ListPermissions(ctx context.Context, _ *v1.ListPermissionsRequest) (*v1.ListPermissionsResponse, error) {
	items, err := s.authority.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*v1.PermissionItem, 0, len(items))
	for _, item := range items {
		out = append(out, &v1.PermissionItem{Uid: item.UID, Name: item.Name, Resource: item.Resource, Action: item.Action})
	}
	return &v1.ListPermissionsResponse{Items: out}, nil
}
