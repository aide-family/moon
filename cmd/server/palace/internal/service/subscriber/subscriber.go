package subscriber

import (
	"context"

	sbscriberapi "github.com/aide-family/moon/api/admin/subscriber"
	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/util/types"
)

// Service subscriber service
type Service struct {
	sbscriberapi.UnimplementedSubscriberServer
	subscriberBiz *biz.SubscriberBiz
}

// NewSubscriberService new subscriber service
func NewSubscriberService(subscriberBiz *biz.SubscriberBiz) *Service {
	return &Service{subscriberBiz: subscriberBiz}
}

// UserSubscriberStrategy user subscriber strategy
func (s *Service) UserSubscriberStrategy(ctx context.Context, req *sbscriberapi.SubscriberStrategyRequest) (*sbscriberapi.SubscriberStrategyReply, error) {
	subscriberID, err := getSubscriberID(ctx)
	if err != nil {
		return nil, err
	}
	param := build.NewBuilder().WithContext(ctx).
		SubscriberStrategyModuleBuilder().
		WithAPISubscriberStrategyRequest(req).
		ToSubscriberBo()
	param.UserID = subscriberID
	if err := s.subscriberBiz.UserSubscriptionStrategy(ctx, param); !types.IsNil(err) {
		return nil, err
	}
	return &sbscriberapi.SubscriberStrategyReply{}, nil
}

// UnSubscriber unsubscribe
func (s *Service) UnSubscriber(ctx context.Context, req *sbscriberapi.UnSubscriberRequest) (*sbscriberapi.UnSubscriberReply, error) {
	subscriberID, err := getSubscriberID(ctx)
	if err != nil {
		return nil, err
	}
	param := build.NewBuilder().
		WithContext(ctx).
		SubscriberStrategyModuleBuilder().
		WithAPIUnSubscriberStrategyRequest(req).
		ToUnSubscriberBo()
	param.UserID = subscriberID
	if err = s.subscriberBiz.UnSubscriptionStrategy(ctx, param); !types.IsNil(err) {
		return nil, err
	}
	return &sbscriberapi.UnSubscriberReply{}, nil
}

// UserSubscriberList user subscriber list
func (s *Service) UserSubscriberList(ctx context.Context, req *sbscriberapi.UserSubscriberListRequest) (*sbscriberapi.UserSubscriberListReply, error) {
	subscriberID, err := getSubscriberID(ctx)
	if err != nil {
		return nil, err
	}
	param := build.NewBuilder().
		WithContext(ctx).SubscriberStrategyModuleBuilder().
		WithAPIUserSubscriberStrategyListRequest(req).
		ToUserSubscriberListBo()
	param.UserID = subscriberID
	strategyList, err := s.subscriberBiz.UserSubscriptionStrategyList(ctx, param)
	if !types.IsNil(err) {
		return nil, err
	}
	return &sbscriberapi.UserSubscriberListReply{
		Pagination: build.NewPageBuilder(param.Page).ToAPI(),
		Strategies: build.NewBuilder().
			WithContext(ctx).
			SubscriberStrategyModuleBuilder().
			WithDosUserSubscriberStrategy(strategyList).ToAPIs(),
	}, nil
}

// GetStrategySubscriber get strategy subscriber
func (s *Service) GetStrategySubscriber(ctx context.Context, req *sbscriberapi.StrategySubscriberRequest) (*sbscriberapi.StrategySubscriberReply, error) {
	param := build.NewBuilder().SubscriberStrategyModuleBuilder().
		WithAPISubscriberStrategyListRequest(req).
		ToStrategySubscriberListBo()
	subscribersList, err := s.subscriberBiz.StrategySubscribersList(ctx, param)
	if err != nil {
		return nil, err
	}

	return &sbscriberapi.StrategySubscriberReply{
		Pagination: build.NewPageBuilder(param.Page).ToAPI(),
		Subscribers: build.NewBuilder().
			WithContext(ctx).SubscriberStrategyModuleBuilder().
			WithDosSubscriberStrategy(subscribersList).
			ToAPIs(),
	}, nil
}

func getSubscriberID(ctx context.Context) (uint32, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return 0, merr.ErrorI18nUnLoginErr(ctx)
	}
	return claims.UserID, nil
}
