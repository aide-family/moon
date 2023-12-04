package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/bo"
)

var _ RoleRepo = (*UnimplementedRoleRepo)(nil)

type (
	RoleRepo interface {
		mustEmbedUnimplemented()
		Create(ctx context.Context, role *bo.RoleBO) (*bo.RoleBO, error)
		Update(ctx context.Context, role *bo.RoleBO, scopes ...query.ScopeMethod) (*bo.RoleBO, error)
		UpdateAll(ctx context.Context, role *bo.RoleBO, scopes ...query.ScopeMethod) error
		Delete(ctx context.Context, scopes ...query.ScopeMethod) error
		Get(ctx context.Context, scopes ...query.ScopeMethod) (*bo.RoleBO, error)
		Find(ctx context.Context, scopes ...query.ScopeMethod) ([]*bo.RoleBO, error)
		List(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*bo.RoleBO, error)
		RelateApi(ctx context.Context, roleId uint, apiList []*bo.ApiBO) error
	}

	UnimplementedRoleRepo struct{}
)

func (UnimplementedRoleRepo) UpdateAll(_ context.Context, _ *bo.RoleBO, _ ...query.ScopeMethod) error {
	return status.Error(codes.Unimplemented, "method UpdateAll not implemented")
}

func (UnimplementedRoleRepo) mustEmbedUnimplemented() {}

func (UnimplementedRoleRepo) Create(_ context.Context, _ *bo.RoleBO) (*bo.RoleBO, error) {
	return nil, status.Error(codes.Unimplemented, "method Create not implemented")
}

func (UnimplementedRoleRepo) Update(_ context.Context, _ *bo.RoleBO, _ ...query.ScopeMethod) (*bo.RoleBO, error) {
	return nil, status.Error(codes.Unimplemented, "method Update not implemented")
}

func (UnimplementedRoleRepo) Delete(_ context.Context, _ ...query.ScopeMethod) error {
	return status.Error(codes.Unimplemented, "method Delete not implemented")
}

func (UnimplementedRoleRepo) Get(_ context.Context, _ ...query.ScopeMethod) (*bo.RoleBO, error) {
	return nil, status.Error(codes.Unimplemented, "method Get not implemented")
}

func (UnimplementedRoleRepo) Find(_ context.Context, _ ...query.ScopeMethod) ([]*bo.RoleBO, error) {
	return nil, status.Error(codes.Unimplemented, "method Find not implemented")
}

func (UnimplementedRoleRepo) List(_ context.Context, _ query.Pagination, _ ...query.ScopeMethod) ([]*bo.RoleBO, error) {
	return nil, status.Error(codes.Unimplemented, "method List not implemented")
}

func (UnimplementedRoleRepo) RelateApi(_ context.Context, _ uint, _ []*bo.ApiBO) error {
	return status.Error(codes.Unimplemented, "method RelateApi not implemented")
}
