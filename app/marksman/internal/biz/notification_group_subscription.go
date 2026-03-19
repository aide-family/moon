package biz

import (
	"context"

	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
)

func NewNotificationGroupSubscriptionBiz(
	notificationGroupRepo repository.NotificationGroup,
	subscriptionRepo repository.NotificationGroupSubscription,
	helper *klog.Helper,
) *NotificationGroupSubscriptionBiz {
	return &NotificationGroupSubscriptionBiz{
		notificationGroupRepo: notificationGroupRepo,
		subscriptionRepo:      subscriptionRepo,
		helper:                klog.NewHelper(klog.With(helper.Logger(), "biz", "notification_group_subscription")),
	}
}

type NotificationGroupSubscriptionBiz struct {
	notificationGroupRepo repository.NotificationGroup
	subscriptionRepo      repository.NotificationGroupSubscription
	helper                *klog.Helper
}

func (b *NotificationGroupSubscriptionBiz) GetNotificationGroupSubscription(ctx context.Context, notificationGroupUID snowflake.ID) (*bo.SubscriptionFilterBo, error) {
	if _, err := b.notificationGroupRepo.GetNotificationGroup(ctx, notificationGroupUID); err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.ErrorNotFound("notification group %d not found", notificationGroupUID.Int64())
		}
		b.helper.Errorw("msg", "get notification group failed", "error", err, "uid", notificationGroupUID)
		return nil, merr.ErrorInternalServer("get notification group failed").WithCause(err)
	}
	filter, err := b.subscriptionRepo.GetSubscriptionByNotificationGroupUID(ctx, notificationGroupUID)
	if err != nil {
		if merr.IsNotFound(err) {
			return &bo.SubscriptionFilterBo{}, nil
		}
		b.helper.Errorw("msg", "get subscription failed", "error", err, "notification_group_uid", notificationGroupUID)
		return nil, merr.ErrorInternalServer("get subscription failed").WithCause(err)
	}
	return filter, nil
}

func (b *NotificationGroupSubscriptionBiz) SaveNotificationGroupSubscription(ctx context.Context, notificationGroupUID snowflake.ID, filter *bo.SubscriptionFilterBo) error {
	if _, err := b.notificationGroupRepo.GetNotificationGroup(ctx, notificationGroupUID); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("notification group %d not found", notificationGroupUID.Int64())
		}
		b.helper.Errorw("msg", "get notification group failed", "error", err, "uid", notificationGroupUID)
		return merr.ErrorInternalServer("get notification group failed").WithCause(err)
	}
	_, err := b.subscriptionRepo.GetSubscriptionByNotificationGroupUID(ctx, notificationGroupUID)
	if err != nil {
		if !merr.IsNotFound(err) {
			b.helper.Errorw("msg", "get subscription failed", "error", err, "notification_group_uid", notificationGroupUID)
			return merr.ErrorInternalServer("get subscription failed").WithCause(err)
		}
		if err := b.subscriptionRepo.CreateSubscription(ctx, notificationGroupUID, filter); err != nil {
			b.helper.Errorw("msg", "create subscription failed", "error", err, "notification_group_uid", notificationGroupUID)
			return merr.ErrorInternalServer("create subscription failed").WithCause(err)
		}
		return nil
	}

	if err := b.subscriptionRepo.UpdateSubscription(ctx, notificationGroupUID, filter); err != nil {
		b.helper.Errorw("msg", "update subscription failed", "error", err, "notification_group_uid", notificationGroupUID)
		return merr.ErrorInternalServer("update subscription failed").WithCause(err)
	}
	return nil
}
