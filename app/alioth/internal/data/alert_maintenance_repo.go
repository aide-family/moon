package data

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertevent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertmaintenancewindow"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	labelsutil "github.com/aide-family/moon/pkg/labels"
	"github.com/aide-family/moon/pkg/merr"
)

// maintenanceMatcher 是维护窗口匹配条件的落库形态。
type maintenanceMatcher struct {
	Labels         map[string]string `json:"labels"`
	LevelUIDs      []string          `json:"level_uids"`
	DatasourceUIDs []string          `json:"datasource_uids"`
	RuleUIDs       []string          `json:"rule_uids"`
}

func matcherOf(window bo.AlertMaintenanceWindow) maintenanceMatcher {
	return maintenanceMatcher{
		Labels:         parseLabelMatcher(window.Labels),
		LevelUIDs:      window.LevelUIDs,
		DatasourceUIDs: window.DatasourceUIDs,
		RuleUIDs:       window.RuleUIDs,
	}
}

// parseLabelMatcher 把标签匹配器 JSON 解析成 map；非法或空值按空匹配器处理。
func parseLabelMatcher(raw string) map[string]string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]string{}
	}
	var matcher map[string]string
	if err := json.Unmarshal([]byte(raw), &matcher); err != nil {
		return map[string]string{}
	}
	return matcher
}

func encodeMatcher(window bo.AlertMaintenanceWindow) string {
	payload, err := json.Marshal(matcherOf(window))
	if err != nil {
		return "{}"
	}
	return string(payload)
}

func decodeMatcher(row *ent.AlertMaintenanceWindow) bo.AlertMaintenanceWindow {
	item := bo.AlertMaintenanceWindow{
		UID: row.UID, NamespaceUID: row.NamespaceUID, Name: row.Name, Description: row.Description,
		StartsAt: row.StartsAt, EndsAt: row.EndsAt, Status: string(row.Status),
		CreatedBy: row.CreatedBy, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
	var matcher maintenanceMatcher
	if err := json.Unmarshal([]byte(row.Matcher), &matcher); err != nil {
		return item
	}
	if len(matcher.Labels) > 0 {
		payload, _ := json.Marshal(matcher.Labels)
		item.Labels = string(payload)
	} else {
		item.Labels = "{}"
	}
	item.LevelUIDs = matcher.LevelUIDs
	item.DatasourceUIDs = matcher.DatasourceUIDs
	item.RuleUIDs = matcher.RuleUIDs
	return item
}

// CreateWindow 保存维护窗口；窗口当前生效时顺带抑制已触发且匹配的事件。
func (r *alertMaintenanceRepo) CreateWindow(ctx context.Context, in bo.AlertMaintenanceWindow) (*bo.AlertMaintenanceWindow, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	row, err := tx.AlertMaintenanceWindow.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetName(in.Name).
		SetDescription(in.Description).
		SetMatcher(encodeMatcher(in)).
		SetStartsAt(in.StartsAt).
		SetEndsAt(in.EndsAt).
		SetStatus(alertmaintenancewindow.StatusACTIVE).
		SetCreatedBy(in.CreatedBy).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if err := applyWindowSuppression(ctx, tx, decodeMatcher(row), time.Now().UTC()); err != nil {
		return nil, r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetWindow(ctx, in.NamespaceUID, row.UID)
}

// UpdateWindow 更新配置；先释放旧配置设置的抑制，再按新配置重新抑制。
func (r *alertMaintenanceRepo) UpdateWindow(ctx context.Context, in bo.AlertMaintenanceWindow) (*bo.AlertMaintenanceWindow, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	previous, err := tx.AlertMaintenanceWindow.Query().Where(
		alertmaintenancewindow.UIDEQ(in.UID),
		alertmaintenancewindow.NamespaceUIDEQ(in.NamespaceUID),
		alertmaintenancewindow.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, merr.NotFound("maintenance window not found")
		}
		return nil, r.fail(err)
	}
	if err := releaseWindowSuppression(ctx, tx, decodeMatcher(previous), ""); err != nil {
		return nil, r.fail(err)
	}
	if _, err := tx.AlertMaintenanceWindow.Update().
		Where(alertmaintenancewindow.IDEQ(previous.ID)).
		SetName(in.Name).
		SetDescription(in.Description).
		SetMatcher(encodeMatcher(in)).
		SetStartsAt(in.StartsAt).
		SetEndsAt(in.EndsAt).
		Save(ctx); err != nil {
		return nil, r.fail(err)
	}
	updated, err := tx.AlertMaintenanceWindow.Get(ctx, previous.ID)
	if err != nil {
		return nil, r.fail(err)
	}
	if err := applyWindowSuppression(ctx, tx, decodeMatcher(updated), time.Now().UTC()); err != nil {
		return nil, r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetWindow(ctx, in.NamespaceUID, in.UID)
}

// UpdateWindowStatus 启停窗口。停用时立即解除它设置的抑制。
func (r *alertMaintenanceRepo) UpdateWindowStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.AlertMaintenanceWindow, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	row, err := tx.AlertMaintenanceWindow.Query().Where(
		alertmaintenancewindow.UIDEQ(uid),
		alertmaintenancewindow.NamespaceUIDEQ(namespaceUID),
		alertmaintenancewindow.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, merr.NotFound("maintenance window not found")
		}
		return nil, r.fail(err)
	}
	if _, err := tx.AlertMaintenanceWindow.Update().
		Where(alertmaintenancewindow.IDEQ(row.ID)).
		SetStatus(alertmaintenancewindow.Status(status)).
		Save(ctx); err != nil {
		return nil, r.fail(err)
	}
	window := decodeMatcher(row)
	window.Status = status
	if status == "DISABLED" {
		if err := releaseWindowSuppression(ctx, tx, window, "maintenance window disabled"); err != nil {
			return nil, r.fail(err)
		}
	} else if status == "ACTIVE" {
		if err := applyWindowSuppression(ctx, tx, window, time.Now().UTC()); err != nil {
			return nil, r.fail(err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetWindow(ctx, namespaceUID, uid)
}

// DeleteWindow 软删除窗口并解除它设置的抑制。
func (r *alertMaintenanceRepo) DeleteWindow(ctx context.Context, namespaceUID, uid string) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	row, err := tx.AlertMaintenanceWindow.Query().Where(
		alertmaintenancewindow.UIDEQ(uid),
		alertmaintenancewindow.NamespaceUIDEQ(namespaceUID),
		alertmaintenancewindow.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return merr.NotFound("maintenance window not found")
		}
		return r.fail(err)
	}
	if err := releaseWindowSuppression(ctx, tx, decodeMatcher(row), "maintenance window deleted"); err != nil {
		return r.fail(err)
	}
	if _, err := tx.AlertMaintenanceWindow.Update().
		Where(alertmaintenancewindow.IDEQ(row.ID)).
		SetDeletedAt(time.Now().UTC()).
		Save(ctx); err != nil {
		return r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return r.fail(err)
	}
	return nil
}

func (r *alertMaintenanceRepo) GetWindow(ctx context.Context, namespaceUID, uid string) (*bo.AlertMaintenanceWindow, error) {
	row, err := r.client.AlertMaintenanceWindow.Query().Where(
		alertmaintenancewindow.UIDEQ(uid),
		alertmaintenancewindow.NamespaceUIDEQ(namespaceUID),
		alertmaintenancewindow.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := decodeMatcher(row)
	return &item, nil
}

func (r *alertMaintenanceRepo) ListWindows(ctx context.Context, namespaceUID string, filter bo.AlertMaintenanceWindowFilter, page, pageSize int) ([]bo.AlertMaintenanceWindow, int, error) {
	query := r.client.AlertMaintenanceWindow.Query().Where(
		alertmaintenancewindow.NamespaceUIDEQ(namespaceUID),
		alertmaintenancewindow.DeletedAtIsNil(),
	)
	if filter.Name != "" {
		query = query.Where(alertmaintenancewindow.NameContainsFold(filter.Name))
	}
	if filter.Status != "" {
		query = query.Where(alertmaintenancewindow.StatusEQ(alertmaintenancewindow.Status(filter.Status)))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Desc(alertmaintenancewindow.FieldStartsAt)).
		Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.AlertMaintenanceWindow, 0, len(rows))
	for _, row := range rows {
		items = append(items, decodeMatcher(row))
	}
	return items, total, nil
}

// ListActiveWindows 返回在给定时刻生效的启用窗口，供评测写入判断是否静默。
func (r *alertMaintenanceRepo) ListActiveWindows(ctx context.Context, namespaceUID string, at time.Time) ([]bo.AlertMaintenanceWindow, error) {
	rows, err := r.client.AlertMaintenanceWindow.Query().Where(
		alertmaintenancewindow.NamespaceUIDEQ(namespaceUID),
		alertmaintenancewindow.DeletedAtIsNil(),
		alertmaintenancewindow.StatusEQ(alertmaintenancewindow.StatusACTIVE),
		alertmaintenancewindow.StartsAtLTE(at),
		alertmaintenancewindow.EndsAtGT(at),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]bo.AlertMaintenanceWindow, 0, len(rows))
	for _, row := range rows {
		items = append(items, decodeMatcher(row))
	}
	return items, nil
}

// matchActiveWindowTx 在写入事件的事务里找命中的生效窗口。评测与探测回写共用。
func matchActiveWindowTx(ctx context.Context, tx *ent.Tx, namespaceUID, labels string, levelUID, datasourceUID, ruleUID string, at time.Time) (*bo.AlertMaintenanceWindow, error) {
	rows, err := tx.AlertMaintenanceWindow.Query().Where(
		alertmaintenancewindow.NamespaceUIDEQ(namespaceUID),
		alertmaintenancewindow.DeletedAtIsNil(),
		alertmaintenancewindow.StatusEQ(alertmaintenancewindow.StatusACTIVE),
		alertmaintenancewindow.StartsAtLTE(at),
		alertmaintenancewindow.EndsAtGT(at),
	).All(ctx)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		window := decodeMatcher(row)
		if windowMatchesEvent(window, labels, levelUID, datasourceUID, ruleUID) {
			return &window, nil
		}
	}
	return nil, nil
}

// windowMatchesEvent 判断一条事件是否被维护窗口匹配。条件之间是 AND，空条件不限制。
func windowMatchesEvent(window bo.AlertMaintenanceWindow, labels string, levelUID, datasourceUID, ruleUID string) bool {
	if len(window.LevelUIDs) > 0 && !containsValue(window.LevelUIDs, levelUID) {
		return false
	}
	if len(window.DatasourceUIDs) > 0 && !containsValue(window.DatasourceUIDs, datasourceUID) {
		return false
	}
	if len(window.RuleUIDs) > 0 && !containsValue(window.RuleUIDs, ruleUID) {
		return false
	}
	matcher := parseLabelMatcher(window.Labels)
	if len(matcher) == 0 {
		return true
	}
	actual, err := labelsutil.ParseObject(labels)
	if err != nil {
		return false
	}
	for key, pattern := range matcher {
		value, ok := actual[key]
		if !ok || !regexp.MustCompile("^(?:"+pattern+")$").MatchString(value) {
			return false
		}
	}
	return true
}

func containsValue(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

// applyWindowSuppression 把当前生效窗口应用到已 FIRING 且匹配的事件上。
// 只处理尚未被抑制的事件，已手动抑制的事件不覆盖。
func applyWindowSuppression(ctx context.Context, tx *ent.Tx, window bo.AlertMaintenanceWindow, now time.Time) error {
	if window.Status != "ACTIVE" || !now.Before(window.EndsAt) || now.Before(window.StartsAt) {
		return nil
	}
	rows, err := tx.AlertEvent.Query().Where(
		alertevent.NamespaceUIDEQ(window.NamespaceUID),
		alertevent.StatusEQ(alertevent.StatusFIRING),
		alertevent.SuppressedEQ(false),
	).All(ctx)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if !windowMatchesEvent(window, row.Labels, row.LevelUID, row.DatasourceUID, row.RuleUID) {
			continue
		}
		if _, err := tx.AlertEvent.Update().
			Where(alertevent.IDEQ(row.ID), alertevent.SuppressedEQ(false)).
			SetSuppressed(true).
			SetSuppressedUntil(window.EndsAt).
			Save(ctx); err != nil {
			return err
		}
		if err := appendEventAction(ctx, tx, window.NamespaceUID, row.UID, actionMaintenance, "",
			"maintenance window "+window.Name+" until "+window.EndsAt.UTC().Format(time.RFC3339)); err != nil {
			return err
		}
	}
	return nil
}

// releaseWindowSuppression 解除由该窗口设置的抑制。窗口结束时间用于区分
// 手动抑制与其他窗口的抑制：只有 suppressed_until 恰好等于窗口结束时间的事件被解除。
func releaseWindowSuppression(ctx context.Context, tx *ent.Tx, window bo.AlertMaintenanceWindow, detail string) error {
	rows, err := tx.AlertEvent.Query().Where(
		alertevent.NamespaceUIDEQ(window.NamespaceUID),
		alertevent.StatusEQ(alertevent.StatusFIRING),
		alertevent.SuppressedEQ(true),
		alertevent.SuppressedUntilEQ(window.EndsAt),
	).All(ctx)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if _, err := tx.AlertEvent.Update().
			Where(alertevent.IDEQ(row.ID)).
			SetSuppressed(false).
			ClearSuppressedUntil().
			Save(ctx); err != nil {
			return err
		}
		message := detail
		if message == "" {
			message = "maintenance window updated"
		}
		if err := appendEventAction(ctx, tx, window.NamespaceUID, row.UID, actionMaintenance, "", message); err != nil {
			return err
		}
	}
	return nil
}
