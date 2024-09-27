package cache

import (
	"context"
	"time"
)

type (
	ICacher interface {
		Close() error
		Delete(ctx context.Context, key string) error
		Exist(ctx context.Context, key string) (bool, error)
		Get(ctx context.Context, key string) (string, error)
		Set(ctx context.Context, key string, value string, expiration time.Duration) error
		SetNX(ctx context.Context, key string, value string, expiration time.Duration) (bool, error)
		Inc(ctx context.Context, key string) (int64, error)
		Dec(ctx context.Context, key string) (int64, error)
	}

	defaultCache struct {
	}
)

func (d *defaultCache) Close() error {
	return nil
}

func (d *defaultCache) Delete(ctx context.Context, key string) error {
	return nil
}

func (d *defaultCache) Exist(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (d *defaultCache) Get(ctx context.Context, key string) (string, error) {
	return "", nil
}

func (d *defaultCache) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return nil
}

func (d *defaultCache) SetNX(ctx context.Context, key string, value string, expiration time.Duration) (bool, error) {
	return false, nil
}

func (d *defaultCache) Inc(ctx context.Context, key string) (int64, error) {
	return 0, nil
}

func (d *defaultCache) Dec(ctx context.Context, key string) (int64, error) {
	return 0, nil
}

func NewDefaultCache() ICacher {
	return &defaultCache{}
}
