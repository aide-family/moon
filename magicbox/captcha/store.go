package captcha

import (
	"context"
	"strings"
	"time"

	"github.com/mojocn/base64Captcha"

	"github.com/aide-family/magicbox/plugin/cache"
)

var _ base64Captcha.Store = (*store)(nil)

const (
	defaultPrefix  cache.K = "moon:captcha"
	defaultExpire          = 3 * time.Minute
	defaultTimeout         = 10 * time.Second
)

func NewStore(cli cache.Interface, opts ...StoreOption) base64Captcha.Store {
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

type store struct {
	cli     cache.Interface
	prefix  cache.K
	timeout time.Duration
	expire  time.Duration
}

type StoreOption func(s *store)

func (s *store) Set(id string, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	return s.cli.Set(ctx, s.prefix.Joins(id), value, s.expire)
}

func (s *store) Get(id string, clear bool) string {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	if clear {
		defer s.cli.Del(ctx, s.prefix.Joins(id))
	}
	v, err := s.cli.Get(ctx, s.prefix.Joins(id))
	if err != nil {
		return ""
	}
	return v
}

func (s *store) Verify(id, answer string, clear bool) bool {
	if id == "" || answer == "" {
		return false
	}
	v := s.Get(id, clear)
	return strings.EqualFold(v, answer)
}

func WithPrefix(prefix cache.K) StoreOption {
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
