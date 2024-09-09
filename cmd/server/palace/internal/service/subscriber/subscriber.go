package subscriber

import (
	"context"

	sbscriberapi "github.com/aide-family/moon/api/admin/subscriber"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
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
	param := builder.NewParamsBuild().WithContext(ctx).SubscriberModuleBuilder().WithSubscriberStrategyRequest(req).ToBo()
	if err := s.subscriberBiz.UserSubscriptionStrategy(ctx, param); !types.IsNil(err) {
		return nil, err
	}
	return &sbscriberapi.SubscriberStrategyReply{}, nil
}

// UnSubscriber unsubscribe
func (s *Service) UnSubscriber(ctx context.Context, req *sbscriberapi.UnSubscriberRequest) (*sbscriberapi.UnSubscriberReply, error) {
	param := builder.NewParamsBuild().SubscriberModuleBuilder().WithUnSubscriberRequest(req).ToBo()
	if err := s.subscriberBiz.UnSubscriptionStrategy(ctx, param); !types.IsNil(err) {
		return nil, err
	}
	return &sbscriberapi.UnSubscriberReply{}, nil
}

// UserSubscriberList user subscriber list
func (s *Service) UserSubscriberList(ctx context.Context, req *sbscriberapi.UserSubscriberListRequest) (*sbscriberapi.UserSubscriberListReply, error) {
	param := builder.NewParamsBuild().SubscriberModuleBuilder().WithUserSubscriberListRequest(req).ToBo()
	strategyList, err := s.subscriberBiz.UserSubscriptionStrategyList(ctx, param)
	if !types.IsNil(err) {
		return nil, err
	}
	return &sbscriberapi.UserSubscriberListReply{
		Pagination: builder.NewParamsBuild().PaginationModuleBuilder().ToAPI(param.Page),
		Strategies: builder.NewParamsBuild().SubscriberModuleBuilder().DoSubscriberBuilder().ToStrategies(strategyList),
	}, nil
}

// GetStrategySubscriber get strategy subscriber
func (s *Service) GetStrategySubscriber(ctx context.Context, req *sbscriberapi.StrategySubscriberRequest) (*sbscriberapi.StrategySubscriberReply, error) {
	param := builder.NewParamsBuild().SubscriberModuleBuilder().WithStrategySubscriberRequest(req).ToBo()
	subscribersList, err := s.subscriberBiz.StrategySubscribersList(ctx, param)
	if err != nil {
		return nil, err
	}

	return &sbscriberapi.StrategySubscriberReply{
		Pagination:  builder.NewParamsBuild().PaginationModuleBuilder().ToAPI(param.Page),
		Subscribers: builder.NewParamsBuild().SubscriberModuleBuilder().DoSubscriberBuilder().ToAPIs(subscribersList),
	}, nil
}
