package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
)

var _ ChatGroupRepo = (*UnimplementedChatGroupRepo)(nil)

type (
	ChatGroupRepo interface {
		mustEmbedUnimplemented()
		Create(ctx context.Context, chatGroup *dobo.ChatGroupDO) (*dobo.ChatGroupDO, error)
		Get(ctx context.Context, scopes ...query.ScopeMethod) (*dobo.ChatGroupDO, error)
		Find(ctx context.Context, scopes ...query.ScopeMethod) ([]*dobo.ChatGroupDO, error)
		Update(ctx context.Context, chatGroup *dobo.ChatGroupDO, scopes ...query.ScopeMethod) error
		Delete(ctx context.Context, scopes ...query.ScopeMethod) error
		List(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.ChatGroupDO, error)
	}

	UnimplementedChatGroupRepo struct{}
)

func (UnimplementedChatGroupRepo) mustEmbedUnimplemented() {}

func (UnimplementedChatGroupRepo) Create(_ context.Context, _ *dobo.ChatGroupDO) (*dobo.ChatGroupDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}

func (UnimplementedChatGroupRepo) Get(_ context.Context, _ ...query.ScopeMethod) (*dobo.ChatGroupDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}

func (UnimplementedChatGroupRepo) Find(_ context.Context, _ ...query.ScopeMethod) ([]*dobo.ChatGroupDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Find not implemented")
}

func (UnimplementedChatGroupRepo) Update(_ context.Context, _ *dobo.ChatGroupDO, _ ...query.ScopeMethod) error {
	return status.Errorf(codes.Unimplemented, "method Update not implemented")
}

func (UnimplementedChatGroupRepo) Delete(_ context.Context, _ ...query.ScopeMethod) error {
	return status.Errorf(codes.Unimplemented, "method Delete not implemented")
}

func (UnimplementedChatGroupRepo) List(_ context.Context, _ query.Pagination, _ ...query.ScopeMethod) ([]*dobo.ChatGroupDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
