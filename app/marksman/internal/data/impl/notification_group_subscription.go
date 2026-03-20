package impl

import (
	"context"

	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/data"
	"github.com/aide-family/marksman/internal/data/impl/convert"
	"github.com/aide-family/marksman/internal/data/impl/query"
)

func NewNotificationGroupSubscriptionRepository(d *data.Data) (repository.NotificationGroupSubscription, error) {
	query.SetDefault(d.DB())
	return &notificationGroupSubscriptionRepository{db: d.DB()}, nil
}

type notificationGroupSubscriptionRepository struct {
	db *gorm.DB
}

func (r *notificationGroupSubscriptionRepository) GetSubscriptionByNotificationGroupUID(ctx context.Context, notificationGroupUID snowflake.ID) (*bo.SubscriptionFilterBo, error) {
	n := query.NotificationGroupSubscription
	m, err := n.WithContext(ctx).Where(n.NotificationGroupUID.Eq(notificationGroupUID.Int64())).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("subscription not found")
		}
		return nil, err
	}
	return convert.ToSubscriptionFilterBo(m), nil
}

func (r *notificationGroupSubscriptionRepository) CreateSubscription(ctx context.Context, notificationGroupUID snowflake.ID, filter *bo.SubscriptionFilterBo) error {
	do := convert.ToNotificationGroupSubscriptionDO(ctx, notificationGroupUID, filter)
	return query.NotificationGroupSubscription.WithContext(ctx).Create(do)
}

func (r *notificationGroupSubscriptionRepository) UpdateSubscription(ctx context.Context, notificationGroupUID snowflake.ID, filter *bo.SubscriptionFilterBo) error {
	n := query.NotificationGroupSubscription
	do := convert.ToNotificationGroupSubscriptionDO(ctx, notificationGroupUID, filter)
	info, err := n.WithContext(ctx).Where(n.NotificationGroupUID.Eq(notificationGroupUID.Int64())).UpdateColumnSimple(
		n.StrategyGroupUIDs.Value(do.StrategyGroupUIDs),
		n.StrategyUIDs.Value(do.StrategyUIDs),
		n.StrategyLevels.Value(do.StrategyLevels),
		n.Labels.Value(do.Labels),
		n.ExcludeLabels.Value(do.ExcludeLabels),
		n.DatasourceUIDs.Value(do.DatasourceUIDs),
	)
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return merr.ErrorNotFound("subscription not found")
	}
	return nil
}
