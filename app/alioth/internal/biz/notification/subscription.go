package notification

import (
	"context"
	"strings"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/page"
	"github.com/aide-family/Alioth/app/alioth/pkg/perm"
	labelsutil "github.com/aide-family/Alioth/pkg/labels"
	"github.com/aide-family/Alioth/pkg/merr"
)

const maxSubscriptionFilterValues = 200

func (center *Center) CreateSubscription(ctx context.Context, in bo.NotificationSubscription) (*bo.NotificationSubscription, error) {
	ns, err := center.access.Allow(ctx, perm.NotificationWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	if err := center.normalizeSubscription(ctx, &in); err != nil {
		return nil, err
	}
	item, err := center.repo.CreateSubscription(ctx, in)
	return item, center.access.Wrap(err)
}

func (center *Center) UpdateSubscription(ctx context.Context, in bo.NotificationSubscription) (*bo.NotificationSubscription, error) {
	ns, err := center.access.Allow(ctx, perm.NotificationWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	if err := center.normalizeSubscription(ctx, &in); err != nil {
		return nil, err
	}
	item, err := center.repo.UpdateSubscription(ctx, in)
	return item, center.access.Wrap(err)
}

func (center *Center) UpdateSubscriptionStatus(ctx context.Context, uid, status string) (*bo.NotificationSubscription, error) {
	ns, err := center.access.Allow(ctx, perm.NotificationWrite)
	if err != nil {
		return nil, err
	}
	item, err := center.repo.UpdateSubscriptionStatus(ctx, ns, uid, status)
	return item, center.access.Wrap(err)
}

func (center *Center) DeleteSubscription(ctx context.Context, uid string) error {
	ns, err := center.access.Allow(ctx, perm.NotificationDelete)
	if err != nil {
		return err
	}
	return center.access.Wrap(center.repo.DeleteSubscription(ctx, ns, uid))
}

func (center *Center) GetSubscription(ctx context.Context, uid string) (*bo.NotificationSubscription, error) {
	ns, err := center.access.Allow(ctx, perm.NotificationRead)
	if err != nil {
		return nil, err
	}
	item, err := center.repo.GetSubscription(ctx, ns, uid)
	return item, center.access.Wrap(err)
}

func (center *Center) ListSubscriptions(ctx context.Context, pageNo, pageSize int32, name string) ([]bo.NotificationSubscription, int, int, int, error) {
	ns, err := center.access.Allow(ctx, perm.NotificationRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := center.repo.ListSubscriptions(ctx, ns, p, s, strings.TrimSpace(name))
	return items, total, p, s, center.access.Wrap(err)
}

func (center *Center) normalizeSubscription(ctx context.Context, in *bo.NotificationSubscription) error {
	if strings.TrimSpace(in.Name) == "" {
		return merr.Invalid("name is required")
	}
	in.Name = strings.TrimSpace(in.Name)
	in.LevelUIDs = uniqueStrings(in.LevelUIDs)
	in.DatasourceUIDs = uniqueStrings(in.DatasourceUIDs)
	in.GroupUIDs = uniqueStrings(in.GroupUIDs)
	if len(in.LevelUIDs) > maxSubscriptionFilterValues || len(in.DatasourceUIDs) > maxSubscriptionFilterValues {
		return merr.Invalid("too many subscription filters")
	}
	if len(in.GroupUIDs) == 0 {
		return merr.Invalid("notification group is required")
	}
	for _, groupUID := range in.GroupUIDs {
		if _, err := center.repo.GetGroup(ctx, in.NamespaceUID, groupUID); err != nil {
			return center.access.Wrap(err)
		}
	}
	normalized, err := labelsutil.NormalizeMatchers(in.Labels)
	if err != nil {
		return err
	}
	if len(normalized) > 4096 {
		return merr.Invalid("labels is too long")
	}
	in.Labels = normalized
	return nil
}
