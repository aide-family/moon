package repository

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"prometheus-manager/app/prom_server/internal/biz/bo"
)

var _ PingRepo = (*UnimplementedPingRepo)(nil)

type (
	// PingRepo is a Greater repo.
	PingRepo interface {
		mustEmbedUnimplemented()
		Ping(ctx context.Context, g *bo.Ping) (*bo.Ping, error)
	}

	UnimplementedPingRepo struct{}
)

func (UnimplementedPingRepo) mustEmbedUnimplemented() {}

func (UnimplementedPingRepo) Ping(_ context.Context, _ *bo.Ping) (*bo.Ping, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
