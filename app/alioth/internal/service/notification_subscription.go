package service

import (
	"context"
	"time"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/notification"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

type NotificationSubscriptionService struct {
	v1.UnimplementedNotificationSubscriptionServiceServer
	center *notification.Center
}

func NewNotificationSubscriptionService(center *notification.Center) *NotificationSubscriptionService {
	return &NotificationSubscriptionService{center: center}
}

func (s *NotificationSubscriptionService) CreateNotificationSubscription(ctx context.Context, req *v1.CreateNotificationSubscriptionRequest) (*v1.CreateNotificationSubscriptionResponse, error) {
	item, err := s.center.CreateSubscription(ctx, subscriptionInput("", req.GetName(), req.GetLabels(), req.GetLevelUids(), req.GetDatasourceUids(), req.GetGroupUids(), req.GetRemark()))
	if err != nil {
		return nil, err
	}
	return &v1.CreateNotificationSubscriptionResponse{Subscription: subscriptionItem(*item)}, nil
}

func (s *NotificationSubscriptionService) UpdateNotificationSubscription(ctx context.Context, req *v1.UpdateNotificationSubscriptionRequest) (*v1.UpdateNotificationSubscriptionResponse, error) {
	item, err := s.center.UpdateSubscription(ctx, subscriptionInput(req.GetUid(), req.GetName(), req.GetLabels(), req.GetLevelUids(), req.GetDatasourceUids(), req.GetGroupUids(), req.GetRemark()))
	if err != nil {
		return nil, err
	}
	return &v1.UpdateNotificationSubscriptionResponse{Subscription: subscriptionItem(*item)}, nil
}

func (s *NotificationSubscriptionService) UpdateNotificationSubscriptionStatus(ctx context.Context, req *v1.UpdateNotificationSubscriptionStatusRequest) (*v1.UpdateNotificationSubscriptionStatusResponse, error) {
	item, err := s.center.UpdateSubscriptionStatus(ctx, req.GetUid(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateNotificationSubscriptionStatusResponse{Subscription: subscriptionItem(*item)}, nil
}

func (s *NotificationSubscriptionService) DeleteNotificationSubscription(ctx context.Context, req *v1.DeleteNotificationSubscriptionRequest) (*v1.DeleteNotificationSubscriptionResponse, error) {
	if err := s.center.DeleteSubscription(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteNotificationSubscriptionResponse{}, nil
}

func (s *NotificationSubscriptionService) GetNotificationSubscription(ctx context.Context, req *v1.GetNotificationSubscriptionRequest) (*v1.GetNotificationSubscriptionResponse, error) {
	item, err := s.center.GetSubscription(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetNotificationSubscriptionResponse{Subscription: subscriptionItem(*item)}, nil
}

func (s *NotificationSubscriptionService) ListNotificationSubscriptions(ctx context.Context, req *v1.ListNotificationSubscriptionsRequest) (*v1.ListNotificationSubscriptionsResponse, error) {
	items, total, page, pageSize, err := s.center.ListSubscriptions(ctx, req.GetPage(), req.GetPageSize(), req.GetName())
	if err != nil {
		return nil, err
	}
	out := make([]*v1.NotificationSubscriptionItem, 0, len(items))
	for i := range items {
		out = append(out, subscriptionItem(items[i]))
	}
	return &v1.ListNotificationSubscriptionsResponse{Items: out, Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func subscriptionInput(uid, name, labels string, levelUIDs, datasourceUIDs, groupUIDs []string, remark string) bo.NotificationSubscription {
	return bo.NotificationSubscription{
		UID: uid, Name: name, Labels: labels, LevelUIDs: levelUIDs,
		DatasourceUIDs: datasourceUIDs, GroupUIDs: groupUIDs, Remark: remark,
	}
}

func subscriptionItem(item bo.NotificationSubscription) *v1.NotificationSubscriptionItem {
	return &v1.NotificationSubscriptionItem{
		Uid: item.UID, Name: item.Name, Status: item.Status, Labels: item.Labels,
		LevelUids: item.LevelUIDs, DatasourceUids: item.DatasourceUIDs, GroupUids: item.GroupUIDs,
		Remark: item.Remark, CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339),
	}
}
