package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
)

var _ RoleRepo = (*UnimplementedRoleRepo)(nil)

type (
	RoleRepo interface {
		mustEmbedUnimplemented()
		Create(ctx context.Context, role *dobo.RoleDO) (*dobo.RoleDO, error)
		Update(ctx context.Context, role *dobo.RoleDO, scopes ...query.ScopeMethod) (*dobo.RoleDO, error)
		Delete(ctx context.Context, scopes ...query.ScopeMethod) error
		Get(ctx context.Context, scopes ...query.ScopeMethod) (*dobo.RoleDO, error)
		List(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.RoleDO, error)
	}

	UnimplementedRoleRepo struct{}
)

func (UnimplementedRoleRepo) mustEmbedUnimplemented() {}

func (UnimplementedRoleRepo) Create(_ context.Context, _ *dobo.RoleDO) (*dobo.RoleDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}

func (UnimplementedRoleRepo) Update(_ context.Context, _ *dobo.RoleDO, _ ...query.ScopeMethod) (*dobo.RoleDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Update not implemented")
}

func (UnimplementedRoleRepo) Delete(_ context.Context, _ ...query.ScopeMethod) error {
	return status.Errorf(codes.Unimplemented, "method Delete not implemented")
}

func (UnimplementedRoleRepo) Get(_ context.Context, _ ...query.ScopeMethod) (*dobo.RoleDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}

func (UnimplementedRoleRepo) List(_ context.Context, _ query.Pagination, _ ...query.ScopeMethod) ([]*dobo.RoleDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
