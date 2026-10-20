package notification

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/page"
	"github.com/aide-family/Alioth/app/alioth/pkg/perm"
	"github.com/aide-family/Alioth/pkg/merr"
)

const maxSubscriptionFilterValues = 200

func (uc *Usecase) CreateSubscription(ctx context.Context, in bo.NotificationSubscription) (*bo.NotificationSubscription, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	if err := uc.normalizeSubscription(ctx, &in); err != nil {
		return nil, err
	}
	item, err := uc.repo.CreateSubscription(ctx, in)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) UpdateSubscription(ctx context.Context, in bo.NotificationSubscription) (*bo.NotificationSubscription, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	if err := uc.normalizeSubscription(ctx, &in); err != nil {
		return nil, err
	}
	item, err := uc.repo.UpdateSubscription(ctx, in)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) UpdateSubscriptionStatus(ctx context.Context, uid, status string) (*bo.NotificationSubscription, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationWrite)
	if err != nil {
		return nil, err
	}
	item, err := uc.repo.UpdateSubscriptionStatus(ctx, ns, uid, status)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) DeleteSubscription(ctx context.Context, uid string) error {
	ns, err := uc.access.Allow(ctx, perm.NotificationWrite)
	if err != nil {
		return err
	}
	return uc.access.Wrap(uc.repo.DeleteSubscription(ctx, ns, uid))
}

func (uc *Usecase) GetSubscription(ctx context.Context, uid string) (*bo.NotificationSubscription, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationRead)
	if err != nil {
		return nil, err
	}
	item, err := uc.repo.GetSubscription(ctx, ns, uid)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) ListSubscriptions(ctx context.Context, pageNo, pageSize int32, name string) ([]bo.NotificationSubscription, int, int, int, error) {
	ns, err := uc.access.Allow(ctx, perm.NotificationRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := uc.repo.ListSubscriptions(ctx, ns, p, s, strings.TrimSpace(name))
	return items, total, p, s, uc.access.Wrap(err)
}

func (uc *Usecase) normalizeSubscription(ctx context.Context, in *bo.NotificationSubscription) error {
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
		if _, err := uc.repo.GetGroup(ctx, in.NamespaceUID, groupUID); err != nil {
			return uc.access.Wrap(err)
		}
	}
	labels := map[string]string{}
	raw := strings.TrimSpace(in.Labels)
	if raw != "" && raw != "{}" {
		if err := json.Unmarshal([]byte(raw), &labels); err != nil {
			return merr.Invalid("invalid labels")
		}
	}
	if labels == nil {
		labels = map[string]string{}
	}
	normalized, err := json.Marshal(labels)
	if err != nil {
		return merr.Invalid("invalid labels")
	}
	if len(normalized) > 4096 {
		return merr.Invalid("labels is too long")
	}
	in.Labels = string(normalized)
	return nil
}
