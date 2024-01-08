package repository

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"

	"prometheus-manager/app/prom_server/internal/biz/bo"
)

var _ RoleRepo = (*UnimplementedRoleRepo)(nil)

type (
	RoleRepo interface {
		mustEmbedUnimplemented()
		Create(ctx context.Context, role *bo.RoleBO) (*bo.RoleBO, error)
		Update(ctx context.Context, role *bo.RoleBO, scopes ...basescopes.ScopeMethod) (*bo.RoleBO, error)
		UpdateAll(ctx context.Context, role *bo.RoleBO, scopes ...basescopes.ScopeMethod) error
		Delete(ctx context.Context, scopes ...basescopes.ScopeMethod) error
		Get(ctx context.Context, scopes ...basescopes.ScopeMethod) (*bo.RoleBO, error)
		Find(ctx context.Context, scopes ...basescopes.ScopeMethod) ([]*bo.RoleBO, error)
		List(ctx context.Context, pgInfo basescopes.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.RoleBO, error)
		RelateApi(ctx context.Context, roleId uint32, apiList []*bo.ApiBO) error
	}

	UnimplementedRoleRepo struct{}
)

func (UnimplementedRoleRepo) UpdateAll(_ context.Context, _ *bo.RoleBO, _ ...basescopes.ScopeMethod) error {
	return status.Error(codes.Unimplemented, "method UpdateAll not implemented")
}

func (UnimplementedRoleRepo) mustEmbedUnimplemented() {}

func (UnimplementedRoleRepo) Create(_ context.Context, _ *bo.RoleBO) (*bo.RoleBO, error) {
	return nil, status.Error(codes.Unimplemented, "method Create not implemented")
}

func (UnimplementedRoleRepo) Update(_ context.Context, _ *bo.RoleBO, _ ...basescopes.ScopeMethod) (*bo.RoleBO, error) {
	return nil, status.Error(codes.Unimplemented, "method Update not implemented")
}

func (UnimplementedRoleRepo) Delete(_ context.Context, _ ...basescopes.ScopeMethod) error {
	return status.Error(codes.Unimplemented, "method Delete not implemented")
}

func (UnimplementedRoleRepo) Get(_ context.Context, _ ...basescopes.ScopeMethod) (*bo.RoleBO, error) {
	return nil, status.Error(codes.Unimplemented, "method Get not implemented")
}

func (UnimplementedRoleRepo) Find(_ context.Context, _ ...basescopes.ScopeMethod) ([]*bo.RoleBO, error) {
	return nil, status.Error(codes.Unimplemented, "method Find not implemented")
}

func (UnimplementedRoleRepo) List(_ context.Context, _ basescopes.Pagination, _ ...basescopes.ScopeMethod) ([]*bo.RoleBO, error) {
	return nil, status.Error(codes.Unimplemented, "method List not implemented")
}

func (UnimplementedRoleRepo) RelateApi(_ context.Context, _ uint32, _ []*bo.ApiBO) error {
	return status.Error(codes.Unimplemented, "method RelateApi not implemented")
}
