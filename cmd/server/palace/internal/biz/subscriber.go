package biz

import (
	"context"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
)

// NewSubscriptionStrategyBiz 创建订阅相关业务
func NewSubscriptionStrategyBiz(strategyRepo repository.SubscriberStrategy) *SubscriberBiz {
	return &SubscriberBiz{
		subscriberRepo: strategyRepo,
	}
}

type (
	// SubscriberBiz 订阅相关业务
	SubscriberBiz struct {
		subscriberRepo repository.SubscriberStrategy
	}
)

// UserSubscriptionStrategy 当前用户订阅策略
func (s *SubscriberBiz) UserSubscriptionStrategy(ctx context.Context, params *bo.SubscriberStrategyParams) error {
	err := s.subscriberRepo.UserSubscriberStrategy(ctx, params)
	if !types.IsNil(err) {
		return merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return nil
}

// UnSubscriptionStrategy 取消订阅策略
func (s *SubscriberBiz) UnSubscriptionStrategy(ctx context.Context, params *bo.UnSubscriberStrategyParams) error {
	err := s.subscriberRepo.UserUnSubscriberStrategy(ctx, params)
	if !types.IsNil(err) {
		return merr.ErrorI18nUserNotSubscribedErr(ctx)
	}
	return nil
}

// UserSubscriptionStrategyList 当前用户订阅策略列表
func (s *SubscriberBiz) UserSubscriptionStrategyList(ctx context.Context, params *bo.QueryUserSubscriberParams) ([]*bizmodel.StrategySubscribers, error) {
	strategyList, err := s.subscriberRepo.UserSubscriberStrategyList(ctx, params)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return strategyList, nil
}

// StrategySubscribersList 策略订阅用户列表
func (s *SubscriberBiz) StrategySubscribersList(ctx context.Context, params *bo.QueryStrategySubscriberParams) ([]*bizmodel.StrategySubscribers, error) {
	subscriberList, err := s.subscriberRepo.StrategySubscriberList(ctx, params)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return subscriberList, nil
}
