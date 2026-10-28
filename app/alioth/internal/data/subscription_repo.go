package data

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationsubscription"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationsubscriptiongroup"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/pkg/merr"
)

func (r *notificationRepo) CreateSubscription(ctx context.Context, in bo.NotificationSubscription) (*bo.NotificationSubscription, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	row, err := tx.NotificationSubscription.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetName(in.Name).
		SetStatus(notificationsubscription.StatusACTIVE).
		SetLabels(in.Labels).
		SetLevelUids(in.LevelUIDs).
		SetDatasourceUids(in.DatasourceUIDs).
		SetDatasourceLevelUids(in.DatasourceLevelUIDs).
		SetRepeatIntervalSeconds(in.RepeatIntervalSeconds).
		SetRemark(in.Remark).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if err := replaceSubscriptionGroups(ctx, tx, in.NamespaceUID, row.UID, in.GroupUIDs); err != nil {
		return nil, r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetSubscription(ctx, in.NamespaceUID, row.UID)
}

func (r *notificationRepo) UpdateSubscription(ctx context.Context, in bo.NotificationSubscription) (*bo.NotificationSubscription, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	n, err := tx.NotificationSubscription.Update().Where(
		notificationsubscription.NamespaceUIDEQ(in.NamespaceUID),
		notificationsubscription.UIDEQ(in.UID),
		notificationsubscription.DeletedAtIsNil(),
	).
		SetName(in.Name).
		SetLabels(in.Labels).
		SetLevelUids(in.LevelUIDs).
		SetDatasourceUids(in.DatasourceUIDs).
		SetDatasourceLevelUids(in.DatasourceLevelUIDs).
		SetRepeatIntervalSeconds(in.RepeatIntervalSeconds).
		SetRemark(in.Remark).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	if err := replaceSubscriptionGroups(ctx, tx, in.NamespaceUID, in.UID, in.GroupUIDs); err != nil {
		return nil, r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetSubscription(ctx, in.NamespaceUID, in.UID)
}

func (r *notificationRepo) UpdateSubscriptionStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.NotificationSubscription, error) {
	value := notificationsubscription.StatusACTIVE
	if status == "DISABLED" {
		value = notificationsubscription.StatusDISABLED
	} else if status != "ACTIVE" {
		return nil, merr.Invalid("invalid request")
	}
	n, err := r.client.NotificationSubscription.Update().Where(
		notificationsubscription.NamespaceUIDEQ(namespaceUID),
		notificationsubscription.UIDEQ(uid),
		notificationsubscription.DeletedAtIsNil(),
	).SetStatus(value).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	return r.GetSubscription(ctx, namespaceUID, uid)
}

func (r *notificationRepo) DeleteSubscription(ctx context.Context, namespaceUID, uid string) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	now := time.Now()
	if _, err := tx.NotificationSubscriptionGroup.Update().Where(
		notificationsubscriptiongroup.NamespaceUIDEQ(namespaceUID),
		notificationsubscriptiongroup.SubscriptionUIDEQ(uid),
		notificationsubscriptiongroup.DeletedAtIsNil(),
	).SetDeletedAt(now).Save(ctx); err != nil {
		return r.fail(err)
	}
	n, err := tx.NotificationSubscription.Update().Where(
		notificationsubscription.NamespaceUIDEQ(namespaceUID),
		notificationsubscription.UIDEQ(uid),
		notificationsubscription.DeletedAtIsNil(),
	).SetDeletedAt(now).Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("not found")
	}
	if err := tx.Commit(); err != nil {
		return r.fail(err)
	}
	return nil
}

func (r *notificationRepo) GetSubscription(ctx context.Context, namespaceUID, uid string) (*bo.NotificationSubscription, error) {
	row, err := r.client.NotificationSubscription.Query().Where(
		notificationsubscription.NamespaceUIDEQ(namespaceUID),
		notificationsubscription.UIDEQ(uid),
		notificationsubscription.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	links, err := r.client.NotificationSubscriptionGroup.Query().Where(
		notificationsubscriptiongroup.NamespaceUIDEQ(namespaceUID),
		notificationsubscriptiongroup.SubscriptionUIDEQ(uid),
		notificationsubscriptiongroup.DeletedAtIsNil(),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	groupUIDs := make([]string, 0, len(links))
	for _, link := range links {
		groupUIDs = append(groupUIDs, link.GroupUID)
	}
	item := toNotificationSubscription(row, groupUIDs)
	return &item, nil
}

func (r *notificationRepo) ListSubscriptions(ctx context.Context, namespaceUID string, page, pageSize int, name string) ([]bo.NotificationSubscription, int, error) {
	query := r.client.NotificationSubscription.Query().Where(
		notificationsubscription.NamespaceUIDEQ(namespaceUID),
		notificationsubscription.DeletedAtIsNil(),
	)
	if name != "" {
		query = query.Where(notificationsubscription.NameContainsFold(name))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Asc(notificationsubscription.FieldUID)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	if len(rows) == 0 {
		return []bo.NotificationSubscription{}, total, nil
	}
	uids := make([]string, len(rows))
	for i, row := range rows {
		uids[i] = row.UID
	}
	links, err := r.client.NotificationSubscriptionGroup.Query().Where(
		notificationsubscriptiongroup.NamespaceUIDEQ(namespaceUID),
		notificationsubscriptiongroup.SubscriptionUIDIn(uids...),
		notificationsubscriptiongroup.DeletedAtIsNil(),
	).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	groupsBySubscription := make(map[string][]string, len(links))
	for _, link := range links {
		groupsBySubscription[link.SubscriptionUID] = append(groupsBySubscription[link.SubscriptionUID], link.GroupUID)
	}
	items := make([]bo.NotificationSubscription, 0, len(rows))
	for _, row := range rows {
		items = append(items, toNotificationSubscription(row, groupsBySubscription[row.UID]))
	}
	return items, total, nil
}

func replaceSubscriptionGroups(ctx context.Context, tx *ent.Tx, namespaceUID, subscriptionUID string, groupUIDs []string) error {
	if _, err := tx.NotificationSubscriptionGroup.Update().Where(
		notificationsubscriptiongroup.NamespaceUIDEQ(namespaceUID),
		notificationsubscriptiongroup.SubscriptionUIDEQ(subscriptionUID),
		notificationsubscriptiongroup.DeletedAtIsNil(),
	).SetDeletedAt(time.Now()).Save(ctx); err != nil {
		return err
	}
	seen := map[string]struct{}{}
	for _, groupUID := range groupUIDs {
		if groupUID == "" {
			continue
		}
		if _, ok := seen[groupUID]; ok {
			continue
		}
		seen[groupUID] = struct{}{}
		if err := tx.NotificationSubscriptionGroup.Create().
			SetUID(id.New()).
			SetNamespaceUID(namespaceUID).
			SetSubscriptionUID(subscriptionUID).
			SetGroupUID(groupUID).
			Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

func toNotificationSubscription(row *ent.NotificationSubscription, groupUIDs []string) bo.NotificationSubscription {
	if groupUIDs == nil {
		groupUIDs = []string{}
	}
	return bo.NotificationSubscription{
		UID: row.UID, NamespaceUID: row.NamespaceUID, Name: row.Name, Status: string(row.Status),
		Labels: row.Labels, LevelUIDs: row.LevelUids, DatasourceUIDs: row.DatasourceUids,
		DatasourceLevelUIDs: row.DatasourceLevelUids, RepeatIntervalSeconds: row.RepeatIntervalSeconds,
		GroupUIDs: groupUIDs, Remark: row.Remark, CreatedAt: row.CreatedAt,
	}
}
