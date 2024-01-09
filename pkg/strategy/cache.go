package strategy

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

var _ CacheAlerter = (*RedisCacheAlerter)(nil)

type CacheAlerter interface {
	Set(ctx context.Context, rule *Rule, result []*Result) error
	Del(ctx context.Context, rule *Rule) error
	Get(ctx context.Context, rule *Rule) (*Result, error)
}

type RedisCacheAlerter struct {
	client *redis.Client

	log *log.Helper
}

func (r *RedisCacheAlerter) Set(ctx context.Context, rule *Rule, result []*Result) error {
	return nil
}

func (r *RedisCacheAlerter) Del(ctx context.Context, rule *Rule) error {
	return nil
}

func (r *RedisCacheAlerter) Get(ctx context.Context, rule *Rule) (*Result, error) {
	return nil, nil
}

func NewRedisCacheAlerter(client *redis.Client, logger log.Logger) *RedisCacheAlerter {
	return &RedisCacheAlerter{
		client: client,
		log:    log.NewHelper(logger),
	}
}
