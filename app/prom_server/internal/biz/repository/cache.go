package repository

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"prometheus-manager/pkg/util/cache"
)

var _ CacheRepo = (*UnimplementedCacheRepo)(nil)

type (
	CacheRepo interface {
		mustEmbedUnimplemented()
		Client() (cache.GlobalCache, error)
		Set(ctx context.Context, key string, value []byte, expiration time.Duration) error
	}

	UnimplementedCacheRepo struct{}
)

func (UnimplementedCacheRepo) Client() (cache.GlobalCache, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Client not implemented")
}

func (UnimplementedCacheRepo) Set(_ context.Context, _ string, _ []byte, _ time.Duration) error {
	return status.Errorf(codes.Unimplemented, "method Set not implemented")
}

func (UnimplementedCacheRepo) mustEmbedUnimplemented() {}
