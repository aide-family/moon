package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
)

var _ UserRepo = (*UnimplementedUserRepo)(nil)

type (
	UserRepo interface {
		mustEmbedUnimplemented()
		Get(ctx context.Context, scopes ...query.ScopeMethod) (*dobo.UserDO, error)
		List(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.UserDO, error)
		Create(ctx context.Context, user *dobo.UserDO) (*dobo.UserDO, error)
		Update(ctx context.Context, user *dobo.UserDO, scopes ...query.ScopeMethod) (*dobo.UserDO, error)
		Delete(ctx context.Context, scopes ...query.ScopeMethod) error
		RelateRoles(ctx context.Context, userDo *dobo.UserDO, roleList []*dobo.RoleDO) error
	}

	UnimplementedUserRepo struct{}
)

func (UnimplementedUserRepo) RelateRoles(_ context.Context, _ *dobo.UserDO, _ []*dobo.RoleDO) error {
	return status.Errorf(codes.Unimplemented, "method RelateRoles not implemented")
}

func (UnimplementedUserRepo) mustEmbedUnimplemented() {}

func (UnimplementedUserRepo) Get(_ context.Context, _ ...query.ScopeMethod) (*dobo.UserDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}

func (UnimplementedUserRepo) List(_ context.Context, _ query.Pagination, _ ...query.ScopeMethod) ([]*dobo.UserDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}

func (UnimplementedUserRepo) Create(_ context.Context, _ *dobo.UserDO) (*dobo.UserDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}

func (UnimplementedUserRepo) Update(_ context.Context, _ *dobo.UserDO, _ ...query.ScopeMethod) (*dobo.UserDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Update not implemented")
}

func (UnimplementedUserRepo) Delete(_ context.Context, _ ...query.ScopeMethod) error {
	return status.Errorf(codes.Unimplemented, "method Delete not implemented")
}
