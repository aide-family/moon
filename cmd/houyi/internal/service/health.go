package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/houyi/internal/biz"
	"github.com/aide-family/moon/pkg/api/common"
	"github.com/aide-family/moon/pkg/hello"
	"github.com/aide-family/moon/pkg/util/timex"
)

type HealthService struct {
	common.UnimplementedHealthServer

	healthBiz   *biz.HealthBiz
	registerBiz *biz.RegisterBiz
	helper      *log.Helper
}

func NewHealthService(healthBiz *biz.HealthBiz, registerBiz *biz.RegisterBiz, logger log.Logger) *HealthService {
	return &HealthService{
		healthBiz:   healthBiz,
		registerBiz: registerBiz,
		helper:      log.NewHelper(log.With(logger, "module", "service.health")),
	}
}

func (s *HealthService) Check(ctx context.Context, req *common.CheckRequest) (*common.CheckReply, error) {
	if err := s.healthBiz.Check(ctx); err != nil {
		return nil, err
	}
	return &common.CheckReply{
		Healthy: true,
		Version: hello.GetEnv().Version(),
		Time:    timex.Format(timex.Now()),
	}, nil
}

func (s *HealthService) Register(ctx context.Context, isOnline bool) error {
	if isOnline {
		return s.registerBiz.Online(ctx)
	}
	return s.registerBiz.Offline(ctx)
}
