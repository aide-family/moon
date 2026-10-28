package data

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationgroup"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationoncall"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationoncallgroup"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/app/alioth/pkg/oncall"
	"github.com/aide-family/moon/pkg/merr"
)

// onCallBinding 是写入时已经校验过的通知组绑定，windows 是规范化后的 JSON。
type onCallBinding struct {
	GroupUID    string
	TimeWindows string
}

func (r *notificationRepo) CreateOnCall(ctx context.Context, in bo.NotificationOnCall) (*bo.NotificationOnCall, error) {
	timezone, bindings, err := normalizeOnCallInput(in)
	if err != nil {
		return nil, err
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	row, err := tx.NotificationOnCall.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetName(in.Name).
		SetStatus(notificationoncall.StatusACTIVE).
		SetLabels(in.Labels).
		SetLevelUids(in.LevelUIDs).
		SetDatasourceUids(in.DatasourceUIDs).
		SetDatasourceLevelUids(in.DatasourceLevelUIDs).
		SetRepeatIntervalSeconds(in.RepeatIntervalSeconds).
		SetTimezone(timezone).
		SetRemark(in.Remark).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if err := replaceOnCallGroups(ctx, tx, in.NamespaceUID, row.UID, bindings); err != nil {
		return nil, r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetOnCall(ctx, in.NamespaceUID, row.UID)
}

func (r *notificationRepo) UpdateOnCall(ctx context.Context, in bo.NotificationOnCall) (*bo.NotificationOnCall, error) {
	timezone, bindings, err := normalizeOnCallInput(in)
	if err != nil {
		return nil, err
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	n, err := tx.NotificationOnCall.Update().Where(
		notificationoncall.NamespaceUIDEQ(in.NamespaceUID),
		notificationoncall.UIDEQ(in.UID),
		notificationoncall.DeletedAtIsNil(),
	).
		SetName(in.Name).
		SetLabels(in.Labels).
		SetLevelUids(in.LevelUIDs).
		SetDatasourceUids(in.DatasourceUIDs).
		SetDatasourceLevelUids(in.DatasourceLevelUIDs).
		SetRepeatIntervalSeconds(in.RepeatIntervalSeconds).
		SetTimezone(timezone).
		SetRemark(in.Remark).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	if err := replaceOnCallGroups(ctx, tx, in.NamespaceUID, in.UID, bindings); err != nil {
		return nil, r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetOnCall(ctx, in.NamespaceUID, in.UID)
}

func (r *notificationRepo) UpdateOnCallStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.NotificationOnCall, error) {
	value := notificationoncall.StatusACTIVE
	if status == "DISABLED" {
		value = notificationoncall.StatusDISABLED
	} else if status != "ACTIVE" {
		return nil, merr.Invalid("invalid request")
	}
	n, err := r.client.NotificationOnCall.Update().Where(
		notificationoncall.NamespaceUIDEQ(namespaceUID),
		notificationoncall.UIDEQ(uid),
		notificationoncall.DeletedAtIsNil(),
	).SetStatus(value).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	return r.GetOnCall(ctx, namespaceUID, uid)
}

func (r *notificationRepo) DeleteOnCall(ctx context.Context, namespaceUID, uid string) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	now := time.Now()
	if _, err := tx.NotificationOnCallGroup.Update().Where(
		notificationoncallgroup.NamespaceUIDEQ(namespaceUID),
		notificationoncallgroup.OncallUIDEQ(uid),
		notificationoncallgroup.DeletedAtIsNil(),
	).SetDeletedAt(now).Save(ctx); err != nil {
		return r.fail(err)
	}
	n, err := tx.NotificationOnCall.Update().Where(
		notificationoncall.NamespaceUIDEQ(namespaceUID),
		notificationoncall.UIDEQ(uid),
		notificationoncall.DeletedAtIsNil(),
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

func (r *notificationRepo) GetOnCall(ctx context.Context, namespaceUID, uid string) (*bo.NotificationOnCall, error) {
	row, err := r.client.NotificationOnCall.Query().Where(
		notificationoncall.NamespaceUIDEQ(namespaceUID),
		notificationoncall.UIDEQ(uid),
		notificationoncall.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	links, err := r.client.NotificationOnCallGroup.Query().Where(
		notificationoncallgroup.NamespaceUIDEQ(namespaceUID),
		notificationoncallgroup.OncallUIDEQ(uid),
		notificationoncallgroup.DeletedAtIsNil(),
	).Order(ent.Asc(notificationoncallgroup.FieldUID)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	groupNames, err := r.onCallGroupNames(ctx, namespaceUID, onCallLinkGroupUIDs(links))
	if err != nil {
		return nil, err
	}
	item, err := toNotificationOnCall(row, links, groupNames)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *notificationRepo) ListOnCalls(ctx context.Context, namespaceUID string, page, pageSize int, name string) ([]bo.NotificationOnCall, int, error) {
	query := r.client.NotificationOnCall.Query().Where(
		notificationoncall.NamespaceUIDEQ(namespaceUID),
		notificationoncall.DeletedAtIsNil(),
	)
	if name != "" {
		query = query.Where(notificationoncall.NameContainsFold(name))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Asc(notificationoncall.FieldUID)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	if len(rows) == 0 {
		return []bo.NotificationOnCall{}, total, nil
	}
	uids := make([]string, len(rows))
	for i, row := range rows {
		uids[i] = row.UID
	}
	links, err := r.client.NotificationOnCallGroup.Query().Where(
		notificationoncallgroup.NamespaceUIDEQ(namespaceUID),
		notificationoncallgroup.OncallUIDIn(uids...),
		notificationoncallgroup.DeletedAtIsNil(),
	).Order(ent.Asc(notificationoncallgroup.FieldUID)).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	groupNames, err := r.onCallGroupNames(ctx, namespaceUID, onCallLinkGroupUIDs(links))
	if err != nil {
		return nil, 0, err
	}
	linksByOnCall := make(map[string][]*ent.NotificationOnCallGroup, len(rows))
	for _, link := range links {
		linksByOnCall[link.OncallUID] = append(linksByOnCall[link.OncallUID], link)
	}
	items := make([]bo.NotificationOnCall, 0, len(rows))
	for _, row := range rows {
		item, err := toNotificationOnCall(row, linksByOnCall[row.UID], groupNames)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *item)
	}
	return items, total, nil
}

// normalizeOnCallInput 校验时区和每个绑定的时段，并把窗口编码成保存形态。
// 同一通知组在一次请求里出现多次时合并为一项，避免写入重复绑定。
func normalizeOnCallInput(in bo.NotificationOnCall) (string, []onCallBinding, error) {
	timezone, err := oncall.NormalizeTimezone(in.Timezone)
	if err != nil {
		return "", nil, err
	}
	bindings := make([]onCallBinding, 0, len(in.Groups))
	seen := make(map[string]struct{}, len(in.Groups))
	for _, group := range in.Groups {
		if group.GroupUID == "" {
			continue
		}
		if _, ok := seen[group.GroupUID]; ok {
			continue
		}
		specs := make([]oncall.Spec, 0, len(group.Windows))
		for _, window := range group.Windows {
			specs = append(specs, oncall.Spec{Weekdays: window.Weekdays, Start: window.Start, End: window.End})
		}
		raw, err := oncall.Marshal(specs)
		if err != nil {
			return "", nil, err
		}
		seen[group.GroupUID] = struct{}{}
		bindings = append(bindings, onCallBinding{GroupUID: group.GroupUID, TimeWindows: raw})
	}
	return timezone, bindings, nil
}

// replaceOnCallGroups 软删除旧绑定后写入新绑定。旧行保留审计，重新绑定同一通知组不会撞唯一约束。
func replaceOnCallGroups(ctx context.Context, tx *ent.Tx, namespaceUID, onCallUID string, bindings []onCallBinding) error {
	if _, err := tx.NotificationOnCallGroup.Update().Where(
		notificationoncallgroup.NamespaceUIDEQ(namespaceUID),
		notificationoncallgroup.OncallUIDEQ(onCallUID),
		notificationoncallgroup.DeletedAtIsNil(),
	).SetDeletedAt(time.Now()).Save(ctx); err != nil {
		return err
	}
	for _, binding := range bindings {
		if err := tx.NotificationOnCallGroup.Create().
			SetUID(id.New()).
			SetNamespaceUID(namespaceUID).
			SetOncallUID(onCallUID).
			SetGroupUID(binding.GroupUID).
			SetTimeWindows(binding.TimeWindows).
			Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

// onCallGroupNames 读取绑定通知组的名称。通知组软删除后名称仍要显示，方便识别历史绑定。
func (r *notificationRepo) onCallGroupNames(ctx context.Context, namespaceUID string, groupUIDs []string) (map[string]string, error) {
	names := map[string]string{}
	if len(groupUIDs) == 0 {
		return names, nil
	}
	rows, err := r.client.NotificationGroup.Query().Where(
		notificationgroup.NamespaceUIDEQ(namespaceUID),
		notificationgroup.UIDIn(uniqueStrings(groupUIDs)...),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	for _, row := range rows {
		names[row.UID] = row.Name
	}
	return names, nil
}

func onCallLinkGroupUIDs(links []*ent.NotificationOnCallGroup) []string {
	uids := make([]string, 0, len(links))
	for _, link := range links {
		uids = append(uids, link.GroupUID)
	}
	return uids
}

func toNotificationOnCall(row *ent.NotificationOnCall, links []*ent.NotificationOnCallGroup, groupNames map[string]string) (*bo.NotificationOnCall, error) {
	groups := make([]bo.NotificationOnCallGroup, 0, len(links))
	for _, link := range links {
		specs, err := oncall.Unmarshal(link.TimeWindows)
		if err != nil {
			return nil, err
		}
		windows := make([]bo.NotificationOnCallWindow, 0, len(specs))
		for _, spec := range specs {
			windows = append(windows, bo.NotificationOnCallWindow{
				Weekdays: spec.Weekdays, Start: spec.Start, End: spec.End,
			})
		}
		groups = append(groups, bo.NotificationOnCallGroup{
			GroupUID: link.GroupUID, GroupName: groupNames[link.GroupUID], Windows: windows,
		})
	}
	return &bo.NotificationOnCall{
		UID: row.UID, NamespaceUID: row.NamespaceUID, Name: row.Name, Status: string(row.Status),
		Labels: row.Labels, LevelUIDs: row.LevelUids, DatasourceUIDs: row.DatasourceUids,
		DatasourceLevelUIDs: row.DatasourceLevelUids, RepeatIntervalSeconds: row.RepeatIntervalSeconds,
		Timezone: row.Timezone, Groups: groups, Remark: row.Remark, CreatedAt: row.CreatedAt,
	}, nil
}
