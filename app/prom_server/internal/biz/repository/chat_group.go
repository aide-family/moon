package repository

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"

	"prometheus-manager/app/prom_server/internal/biz/bo"
)

var _ ChatGroupRepo = (*UnimplementedChatGroupRepo)(nil)

type (
	ChatGroupRepo interface {
		mustEmbedUnimplemented()
		Create(ctx context.Context, chatGroup *bo.ChatGroupBO) (*bo.ChatGroupBO, error)
		Get(ctx context.Context, scopes ...basescopes.ScopeMethod) (*bo.ChatGroupBO, error)
		Find(ctx context.Context, scopes ...basescopes.ScopeMethod) ([]*bo.ChatGroupBO, error)
		Update(ctx context.Context, chatGroup *bo.ChatGroupBO, scopes ...basescopes.ScopeMethod) error
		Delete(ctx context.Context, scopes ...basescopes.ScopeMethod) error
		List(ctx context.Context, pgInfo bo.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.ChatGroupBO, error)
	}

	UnimplementedChatGroupRepo struct{}
)

func (UnimplementedChatGroupRepo) mustEmbedUnimplemented() {}

func (UnimplementedChatGroupRepo) Create(_ context.Context, _ *bo.ChatGroupBO) (*bo.ChatGroupBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}

func (UnimplementedChatGroupRepo) Get(_ context.Context, _ ...basescopes.ScopeMethod) (*bo.ChatGroupBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}

func (UnimplementedChatGroupRepo) Find(_ context.Context, _ ...basescopes.ScopeMethod) ([]*bo.ChatGroupBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Find not implemented")
}

func (UnimplementedChatGroupRepo) Update(_ context.Context, _ *bo.ChatGroupBO, _ ...basescopes.ScopeMethod) error {
	return status.Errorf(codes.Unimplemented, "method Update not implemented")
}

func (UnimplementedChatGroupRepo) Delete(_ context.Context, _ ...basescopes.ScopeMethod) error {
	return status.Errorf(codes.Unimplemented, "method Delete not implemented")
}

func (UnimplementedChatGroupRepo) List(_ context.Context, _ bo.Pagination, _ ...basescopes.ScopeMethod) ([]*bo.ChatGroupBO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
