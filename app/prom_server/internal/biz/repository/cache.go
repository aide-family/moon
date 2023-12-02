package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ CacheRepo = (*UnimplementedCacheRepo)(nil)

type (
	CacheRepo interface {
		mustEmbedUnimplemented()
		Client() (*redis.Client, error)
		Set(ctx context.Context, key string, value any, expiration time.Duration) error
	}

	UnimplementedCacheRepo struct{}
)

func (UnimplementedCacheRepo) Client() (*redis.Client, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Client not implemented")
}

func (UnimplementedCacheRepo) Set(_ context.Context, _ string, _ any, _ time.Duration) error {
	return status.Errorf(codes.Unimplemented, "method Set not implemented")
}

func (UnimplementedCacheRepo) mustEmbedUnimplemented() {}
