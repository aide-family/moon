package repository

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

var _ EndpointRepo = (*UnimplementedEndpointRepo)(nil)

type (
	EndpointRepo interface {
		mustEmbedUnimplemented()
		Append(ctx context.Context, endpoint *bo.EndpointBO) (*bo.EndpointBO, error)
		Update(ctx context.Context, endpoint *bo.EndpointBO) (*bo.EndpointBO, error)
		UpdateStatus(ctx context.Context, ids []uint32, status vo.Status) error
		Delete(ctx context.Context, ids []uint32) error
		List(ctx context.Context, pagination bo.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.EndpointBO, error)
		Get(ctx context.Context, scopes ...basescopes.ScopeMethod) (*bo.EndpointBO, error)
		GetByParams(ctx context.Context, scopes ...basescopes.ScopeMethod) ([]*bo.EndpointBO, error)
	}

	UnimplementedEndpointRepo struct{}
)

func (UnimplementedEndpointRepo) GetByParams(_ context.Context, _ ...basescopes.ScopeMethod) ([]*bo.EndpointBO, error) {
	return nil, status.Error(codes.Unimplemented, "method GetByParams not implemented")
}

func (r UnimplementedEndpointRepo) Get(ctx context.Context, scopes ...basescopes.ScopeMethod) (*bo.EndpointBO, error) {
	return nil, status.Error(codes.Unimplemented, "method Get not implemented")
}

func (UnimplementedEndpointRepo) UpdateStatus(_ context.Context, _ []uint32, _ vo.Status) error {
	return status.Error(codes.Unimplemented, "method UpdateStatus not implemented")
}

func (UnimplementedEndpointRepo) Update(_ context.Context, _ *bo.EndpointBO) (*bo.EndpointBO, error) {
	return nil, status.Error(codes.Unimplemented, "method Update not implemented")
}

func (UnimplementedEndpointRepo) Append(_ context.Context, _ *bo.EndpointBO) (*bo.EndpointBO, error) {
	return nil, status.Error(codes.Unimplemented, "method Append not implemented")
}

func (UnimplementedEndpointRepo) Delete(_ context.Context, _ []uint32) error {
	return status.Error(codes.Unimplemented, "method Delete not implemented")
}

func (UnimplementedEndpointRepo) List(_ context.Context, _ bo.Pagination, _ ...basescopes.ScopeMethod) ([]*bo.EndpointBO, error) {
	return nil, status.Error(codes.Unimplemented, "method List not implemented")
}

func (UnimplementedEndpointRepo) mustEmbedUnimplemented() {}
