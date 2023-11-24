package repository

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
)

var _ EndpointRepo = (*UnimplementedEndpointRepo)(nil)

type (
	EndpointRepo interface {
		mustEmbedUnimplemented()
		Append(ctx context.Context, endpoint []*dobo.EndpointDO) error
		Delete(ctx context.Context, endpoint []*dobo.EndpointDO) error
		List(ctx context.Context) ([]*dobo.EndpointDO, error)
	}

	UnimplementedEndpointRepo struct{}
)

func (UnimplementedEndpointRepo) Append(_ context.Context, _ []*dobo.EndpointDO) error {
	return status.Errorf(codes.Unimplemented, "method Append not implemented")
}

func (UnimplementedEndpointRepo) Delete(_ context.Context, _ []*dobo.EndpointDO) error {
	return status.Errorf(codes.Unimplemented, "method Delete not implemented")
}

func (UnimplementedEndpointRepo) List(_ context.Context) ([]*dobo.EndpointDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}

func (UnimplementedEndpointRepo) mustEmbedUnimplemented() {}
