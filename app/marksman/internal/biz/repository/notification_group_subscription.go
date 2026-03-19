package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/marksman/internal/biz/bo"
)

// NotificationGroupSubscription defines data access for notification group subscription (filter) per group.
type NotificationGroupSubscription interface {
	GetSubscriptionByNotificationGroupUID(ctx context.Context, notificationGroupUID snowflake.ID) (*bo.SubscriptionFilterBo, error)
	CreateSubscription(ctx context.Context, notificationGroupUID snowflake.ID, filter *bo.SubscriptionFilterBo) error
	UpdateSubscription(ctx context.Context, notificationGroupUID snowflake.ID, filter *bo.SubscriptionFilterBo) error
}
