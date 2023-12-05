package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"prometheus-manager/app/prom_server/internal/biz/bo"
)

var _ UserRepo = (*UnimplementedUserRepo)(nil)

type (
	UserRepo interface {
		mustEmbedUnimplemented()
		Get(ctx context.Context, scopes ...query.ScopeMethod) (*bo.UserBO, error)
		Find(ctx context.Context, scopes ...query.ScopeMethod) ([]*bo.UserBO, error)
		Count(ctx context.Context, scopes ...query.ScopeMethod) (int64, error)
		List(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*bo.UserBO, error)
		Create(ctx context.Context, user *bo.UserBO) (*bo.UserBO, error)
		Update(ctx context.Context, user *bo.UserBO, scopes ...query.ScopeMethod) (*bo.UserBO, error)
		Delete(ctx context.Context, scopes ...query.ScopeMethod) error
		RelateRoles(ctx context.Context, userBO *bo.UserBO, roleList []*bo.RoleBO) error
	}

	UnimplementedUserRepo struct{}
)

func (UnimplementedUserRepo) RelateRoles(_ context.Context, _ *bo.UserBO, _ []*bo.RoleBO) error {
	return status.Errorf(codes.Unimplemented, "method RelateRoles not implemented")
}

func (UnimplementedUserRepo) mustEmbedUnimplemented() {}

func (UnimplementedUserRepo) Get(_ context.Context, _ ...query.ScopeMethod) (*bo.UserBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}

func (UnimplementedUserRepo) Find(_ context.Context, _ ...query.ScopeMethod) ([]*bo.UserBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}

func (UnimplementedUserRepo) Count(_ context.Context, _ ...query.ScopeMethod) (int64, error) {
	return 0, status.Errorf(codes.Unimplemented, "method Count not implemented")
}

func (UnimplementedUserRepo) List(_ context.Context, _ query.Pagination, _ ...query.ScopeMethod) ([]*bo.UserBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}

func (UnimplementedUserRepo) Create(_ context.Context, _ *bo.UserBO) (*bo.UserBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}

func (UnimplementedUserRepo) Update(_ context.Context, _ *bo.UserBO, _ ...query.ScopeMethod) (*bo.UserBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Update not implemented")
}

func (UnimplementedUserRepo) Delete(_ context.Context, _ ...query.ScopeMethod) error {
	return status.Errorf(codes.Unimplemented, "method Delete not implemented")
}
