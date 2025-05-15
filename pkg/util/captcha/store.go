package captcha

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
)

var _ base64Captcha.Store = (*store)(nil)

const (
	defaultPrefix  = "moon:palace:captcha"
	defaultExpire  = 3 * time.Minute
	defaultTimeout = 10 * time.Second
)

func NewStore(cli *redis.Client, opts ...StoreOption) base64Captcha.Store {
	s := &store{
		cli:     cli,
		prefix:  defaultPrefix,
		timeout: defaultTimeout,
		expire:  defaultExpire,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func WithPrefix(prefix string) StoreOption {
	return func(s *store) {
		if prefix != "" {
			s.prefix = prefix
		}
	}
}

func WithTimeout(timeout time.Duration) StoreOption {
	return func(s *store) {
		if timeout > 0 {
			s.timeout = timeout
		}
	}
}

func WithExpire(expire time.Duration) StoreOption {
	return func(s *store) {
		if expire > 0 {
			s.expire = expire
		}
	}
}

type store struct {
	cli     *redis.Client
	prefix  string
	timeout time.Duration
	expire  time.Duration
}

type StoreOption func(s *store)

func (s *store) getKey(id string) string {
	return fmt.Sprintf("%s:%s", s.prefix, id)
}

func (s *store) Set(id string, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	return s.cli.Set(ctx, s.getKey(id), value, s.expire).Err()
}

func (s *store) Get(id string, clear bool) string {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	if clear {
		defer s.cli.Del(ctx, s.getKey(id))
	}
	return s.cli.Get(ctx, s.getKey(id)).Val()
}

func (s *store) Verify(id, answer string, clear bool) bool {
	if id == "" || answer == "" {
		return false
	}
	v := s.Get(id, clear)
	return strings.EqualFold(v, answer)
}
