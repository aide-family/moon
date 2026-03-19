package service

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/marksman/internal/biz"
	"github.com/aide-family/marksman/internal/biz/bo"
	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
)

func NewNotificationGroupSubscriptionService(subscriptionBiz *biz.NotificationGroupSubscriptionBiz) *NotificationGroupSubscriptionService {
	return &NotificationGroupSubscriptionService{
		subscriptionBiz: subscriptionBiz,
	}
}

type NotificationGroupSubscriptionService struct {
	apiv1.UnimplementedNotificationGroupSubscriptionServer

	subscriptionBiz *biz.NotificationGroupSubscriptionBiz
}

func (s *NotificationGroupSubscriptionService) GetNotificationGroupSubscription(ctx context.Context, req *apiv1.GetNotificationGroupSubscriptionRequest) (*apiv1.GetNotificationGroupSubscriptionReply, error) {
	uid := snowflake.ParseInt64(req.GetNotificationGroupUid())
	filter, err := s.subscriptionBiz.GetNotificationGroupSubscription(ctx, uid)
	if err != nil {
		return nil, err
	}
	return &apiv1.GetNotificationGroupSubscriptionReply{
		Filter: bo.ToAPIV1SubscriptionFilter(filter),
	}, nil
}

func (s *NotificationGroupSubscriptionService) SaveNotificationGroupSubscription(ctx context.Context, req *apiv1.SaveNotificationGroupSubscriptionRequest) (*apiv1.SaveNotificationGroupSubscriptionReply, error) {
	uid := snowflake.ParseInt64(req.GetNotificationGroupUid())
	filterBo := bo.NewSubscriptionFilterBo(req.GetFilter())
	if err := s.subscriptionBiz.SaveNotificationGroupSubscription(ctx, uid, filterBo); err != nil {
		return nil, err
	}
	return &apiv1.SaveNotificationGroupSubscriptionReply{}, nil
}
