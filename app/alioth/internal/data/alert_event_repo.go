package data

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertdictionary"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertevalbatch"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertevent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alerteventfiring"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertrule"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/datasource"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationgroup"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationtask"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/predicate"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/user"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/id"
	labelsutil "github.com/aide-family/Alioth/pkg/labels"
	"github.com/aide-family/Alioth/pkg/merr"
)

func (r *alertEventRepo) ApplyReport(ctx context.Context, rule bo.AlertRule, report bo.EvalReport) error {
	// 分区由 partition.Usecase 定时预创建：写路径不再执行 DDL，避免 catalog 锁和写延迟。
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()

	dsUID := report.DatasourceUID
	if dsUID == "" {
		dsUID = rule.DatasourceUID
	}
	recorded, err := tx.AlertEvalBatch.Query().Where(
		alertevalbatch.RuleUIDEQ(report.RuleUID),
		alertevalbatch.DatasourceUIDEQ(dsUID),
		alertevalbatch.EvalSlotEQ(report.EvalSlot),
	).Exist(ctx)
	if err != nil {
		return r.fail(err)
	}
	if recorded {
		return nil
	}
	report.DatasourceUID = dsUID
	// 跳过评测也要走恢复判断：规则从指标改成日志/链路后，旧的 FIRING 不能永久留在实时列表里。
	if err := r.applyHits(ctx, tx, rule, report); err != nil {
		return err
	}
	if err := tx.AlertEvalBatch.Create().
		SetRuleUID(report.RuleUID).
		SetDatasourceUID(dsUID).
		SetEvalSlot(report.EvalSlot).
		Exec(ctx); err != nil {
		if ent.IsConstraintError(err) {
			return nil
		}
		return r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return r.fail(err)
	}
	return nil
}

func (r *alertEventRepo) applyHits(ctx context.Context, tx *ent.Tx, rule bo.AlertRule, report bo.EvalReport) error {
	rows, err := tx.AlertEventFiring.Query().Where(
		alerteventfiring.NamespaceUIDEQ(rule.NamespaceUID),
		alerteventfiring.RuleUIDEQ(rule.UID),
	).All(ctx)
	if err != nil {
		return r.fail(err)
	}
	current := make(map[string]firingEvent, len(rows))
	for _, row := range rows {
		event, err := tx.AlertEvent.Query().Where(
			alertevent.UIDEQ(row.EventUID),
			alertevent.NamespaceUIDEQ(rule.NamespaceUID),
		).Only(ctx)
		if err != nil {
			return r.fail(err)
		}
		if event.DatasourceUID != report.DatasourceUID {
			continue
		}
		current[row.Fingerprint] = firingEvent{firing: row, event: event}
	}
	hits := make(map[string]bo.EvalHit, len(report.Hits))
	for _, hit := range report.Hits {
		if hit.Fingerprint == "" {
			continue
		}
		hits[hit.Fingerprint] = hit
	}
	now := time.Now().UTC()
	source := r.datasourceMeta(ctx, tx, report.DatasourceUID)
	// 级别名称同时用于事件标签和通知；一次取回级别字典，避免每条命中都查一次。
	levelUIDs := make([]string, 0, len(hits)+1)
	seenLevels := map[string]struct{}{}
	for _, uid := range append([]string{source.LevelUID}, hitLevelUIDs(hits)...) {
		if uid == "" {
			continue
		}
		if _, ok := seenLevels[uid]; ok {
			continue
		}
		seenLevels[uid] = struct{}{}
		levelUIDs = append(levelUIDs, uid)
	}
	levelNames := r.levelNames(ctx, levelUIDs)
	for fingerprint, hit := range hits {
		labels := mergeEventLabels(hit.Labels, source, hit.LevelUID, levelNames)
		if ref, ok := current[fingerprint]; ok {
			update := tx.AlertEvent.Update().Where(
				alertevent.UIDEQ(ref.event.UID),
				alertevent.NamespaceUIDEQ(rule.NamespaceUID),
				alertevent.StatusEQ(alertevent.StatusFIRING),
			).
				SetLabels(labels).
				SetSummary(textOr(hit.Summary, rule.Summary)).
				SetDescription(textOr(hit.Description, rule.Description)).
				SetSampleValue(hit.Value).
				SetLevelUID(hit.LevelUID).
				SetEvalSlot(report.EvalSlot)
			// 抑制已到期又再次命中：清掉抑制状态，让事件回到实时列表。
			if ref.event.Suppressed && !suppressedAt(true, ref.event.SuppressedUntil, now) {
				update.SetSuppressed(false).ClearSuppressedUntil()
			}
			if _, err := update.Save(ctx); err != nil {
				return r.fail(err)
			}
			continue
		}
		eventUID := id.New()
		if err := tx.AlertEvent.Create().
			SetUID(eventUID).
			SetNamespaceUID(rule.NamespaceUID).
			SetRuleUID(rule.UID).
			SetDatasourceUID(report.DatasourceUID).
			SetFingerprint(fingerprint).
			SetStatus(alertevent.StatusFIRING).
			SetExpr(rule.Expr).
			SetLabels(labels).
			SetSummary(textOr(hit.Summary, rule.Summary)).
			SetDescription(textOr(hit.Description, rule.Description)).
			SetSampleValue(hit.Value).
			SetLevelUID(hit.LevelUID).
			SetStartsAt(now).
			SetEvalSlot(report.EvalSlot).
			SetSuppressed(false).
			Exec(ctx); err != nil {
			return r.fail(err)
		}
		if err := tx.AlertEventFiring.Create().
			SetNamespaceUID(rule.NamespaceUID).
			SetFingerprint(fingerprint).
			SetEventUID(eventUID).
			SetRuleUID(rule.UID).
			Exec(ctx); err != nil {
			return r.fail(err)
		}
		if err := dispatchEventNotifications(ctx, tx, notifyEvent{
			NamespaceUID: rule.NamespaceUID, EventUID: eventUID, RuleUID: rule.UID,
			DatasourceUID: report.DatasourceUID, LevelUID: hit.LevelUID, Phase: "FIRING",
			Summary: textOr(hit.Summary, rule.Summary), Description: textOr(hit.Description, rule.Description),
			Labels: labels, Value: hit.Value, ExternalURL: source.Endpoint, Fingerprint: fingerprint,
		}); err != nil {
			return err
		}
	}
	for fingerprint, ref := range current {
		if _, ok := hits[fingerprint]; ok {
			continue
		}
		if _, err := tx.AlertEvent.Update().Where(
			alertevent.UIDEQ(ref.event.UID),
			alertevent.NamespaceUIDEQ(rule.NamespaceUID),
			alertevent.StatusEQ(alertevent.StatusFIRING),
		).SetStatus(alertevent.StatusRECOVERED).SetEndsAt(now).SetEvalSlot(report.EvalSlot).Save(ctx); err != nil {
			return r.fail(err)
		}
		if _, err := tx.AlertEventFiring.Delete().Where(alerteventfiring.IDEQ(ref.firing.ID)).Exec(ctx); err != nil {
			return r.fail(err)
		}
		if err := dispatchEventNotifications(ctx, tx, notifyEvent{
			NamespaceUID: rule.NamespaceUID, EventUID: ref.event.UID, RuleUID: rule.UID,
			DatasourceUID: ref.event.DatasourceUID, LevelUID: ref.event.LevelUID, Phase: "RECOVERED",
			Summary: ref.event.Summary, Description: ref.event.Description, Labels: ref.event.Labels,
			Value: ref.event.SampleValue, ExternalURL: source.Endpoint, Fingerprint: fingerprint,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *alertEventRepo) HasBatch(ctx context.Context, ruleUID, datasourceUID string, slot time.Time) (bool, error) {
	ok, err := r.client.AlertEvalBatch.Query().Where(
		alertevalbatch.RuleUIDEQ(ruleUID),
		alertevalbatch.DatasourceUIDEQ(datasourceUID),
		alertevalbatch.EvalSlotEQ(slot),
	).Exist(ctx)
	if err != nil {
		return false, r.fail(err)
	}
	return ok, nil
}

func (r *alertEventRepo) ListFiringFingerprints(ctx context.Context, namespaceUID, ruleUID, datasourceUID string) ([]string, error) {
	rows, err := r.client.AlertEventFiring.Query().Where(
		alerteventfiring.NamespaceUIDEQ(namespaceUID),
		alerteventfiring.RuleUIDEQ(ruleUID),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]string, 0, len(rows))
	for _, row := range rows {
		event, err := r.client.AlertEvent.Query().Where(
			alertevent.UIDEQ(row.EventUID),
			alertevent.NamespaceUIDEQ(namespaceUID),
		).Only(ctx)
		if err != nil {
			return nil, r.fail(err)
		}
		if event.DatasourceUID != datasourceUID {
			continue
		}
		items = append(items, row.Fingerprint)
	}
	return items, nil
}

func (r *alertEventRepo) AllBatchesRecorded(ctx context.Context, ruleUID string, datasourceUIDs []string, slot time.Time) (bool, error) {
	if len(datasourceUIDs) == 0 {
		return true, nil
	}
	for _, dsUID := range datasourceUIDs {
		ok, err := r.HasBatch(ctx, ruleUID, dsUID, slot)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

func (r *alertEventRepo) ListEvents(ctx context.Context, namespaceUID string, filter bo.AlertEventListFilter, page, pageSize int) ([]bo.AlertEvent, int, error) {
	if filter.FilterByRules && len(filter.RuleUIDs) == 0 {
		return nil, 0, nil
	}
	where, matchers, err := eventWhere(namespaceUID, filter, time.Now().UTC())
	if err != nil {
		return nil, 0, r.fail(err)
	}
	if len(matchers) > 0 {
		return r.listEventsByMatchers(ctx, where, matchers, page, pageSize)
	}
	total, err := r.client.AlertEvent.Query().Where(where...).Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := r.client.AlertEvent.Query().Where(where...).
		Order(ent.Desc(alertevent.FieldCreatedAt)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items, err := r.namedEvents(ctx, rows)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// listEventsByMatchers 在 Go 侧完成标签正则匹配。PostgreSQL 的 ARE 和 Go 的
// RE2 在 \b、\w、\Q 等语法上行为不同；告警页面必须和订阅、数据源使用同一个
// 匹配器实现。这里先只取 id 和 labels 做筛选与分页，再按 id 读取整页事件。
func (r *alertEventRepo) listEventsByMatchers(ctx context.Context, where []predicate.AlertEvent, matchers map[string]string, page, pageSize int) ([]bo.AlertEvent, int, error) {
	rows, err := r.client.AlertEvent.Query().Where(where...).
		Select(alertevent.FieldID, alertevent.FieldLabels).
		Order(ent.Desc(alertevent.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	ids := make([]int, 0, len(rows))
	matches := labelsutil.CompileMatchers(matchers)
	for _, row := range rows {
		labels, err := labelsutil.ParseObject(row.Labels)
		if err != nil {
			continue
		}
		if matches(labels) {
			ids = append(ids, row.ID)
		}
	}
	total := len(ids)
	start := (page - 1) * pageSize
	if start >= total {
		return []bo.AlertEvent{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	pageRows, err := r.client.AlertEvent.Query().Where(alertevent.IDIn(ids[start:end]...)).
		Order(ent.Desc(alertevent.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items, err := r.namedEvents(ctx, pageRows)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// CountEvents 只统计符合条件的告警数量，不读取和排序列表数据。
// 带标签匹配时只读取 labels 列，在 Go 侧完成正则筛选。
func (r *alertEventRepo) CountEvents(ctx context.Context, namespaceUID string, filter bo.AlertEventListFilter) (int, error) {
	if filter.FilterByRules && len(filter.RuleUIDs) == 0 {
		return 0, nil
	}
	where, matchers, err := eventWhere(namespaceUID, filter, time.Now().UTC())
	if err != nil {
		return 0, r.fail(err)
	}
	if len(matchers) > 0 {
		rows, err := r.client.AlertEvent.Query().Where(where...).
			Select(alertevent.FieldLabels).
			All(ctx)
		if err != nil {
			return 0, r.fail(err)
		}
		total := 0
		matches := labelsutil.CompileMatchers(matchers)
		for _, row := range rows {
			labels, err := labelsutil.ParseObject(row.Labels)
			if err != nil {
				continue
			}
			if matches(labels) {
				total++
			}
		}
		return total, nil
	}
	total, err := r.client.AlertEvent.Query().Where(where...).Count(ctx)
	if err != nil {
		return 0, r.fail(err)
	}
	return total, nil
}

// ListRuleUIDsForReceivers 返回当前成员能收到通知的规则。
// 规则不再直接绑定通知组，成员关系由订阅下的通知组决定，任务表记录实际路由结果。
func (r *alertEventRepo) ListRuleUIDsForReceivers(ctx context.Context, namespaceUID, memberUID string, tokens []string) ([]string, error) {
	groups, err := r.client.NotificationGroup.Query().Where(
		notificationgroup.NamespaceUIDEQ(namespaceUID),
		notificationgroup.DeletedAtIsNil(),
		notificationgroup.StatusEQ(notificationgroup.StatusACTIVE),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	matched := make([]string, 0)
	for _, group := range groups {
		if (memberUID != "" && containsString(group.MemberUids, memberUID)) || receiversContain(group.Receivers, tokens) {
			matched = append(matched, group.UID)
		}
	}
	if len(matched) == 0 {
		return nil, nil
	}
	tasks, err := r.client.NotificationTask.Query().Where(
		notificationtask.NamespaceUIDEQ(namespaceUID),
		notificationtask.GroupUIDIn(matched...),
		notificationtask.EventUIDNEQ(""),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	seen := make(map[string]struct{}, len(tasks))
	uids := make([]string, 0, len(tasks))
	for _, task := range tasks {
		if task.RuleUID == "" {
			continue
		}
		if _, ok := seen[task.RuleUID]; ok {
			continue
		}
		seen[task.RuleUID] = struct{}{}
		uids = append(uids, task.RuleUID)
	}
	return uids, nil
}

func receiversContain(receivers string, tokens []string) bool {
	parts := strings.Split(receivers, ",")
	for _, part := range parts {
		value := strings.ToLower(strings.TrimSpace(part))
		if value == "" {
			continue
		}
		for _, token := range tokens {
			if value == strings.ToLower(strings.TrimSpace(token)) {
				return true
			}
		}
	}
	return false
}

func eventWhere(namespaceUID string, filter bo.AlertEventListFilter, now time.Time) ([]predicate.AlertEvent, map[string]string, error) {
	where := []predicate.AlertEvent{alertevent.NamespaceUIDEQ(namespaceUID)}
	if filter.View == "HISTORY" {
		// 抑制到期的 FIRING 回到实时列表，不再算历史。
		where = append(where, alertevent.Or(
			alertevent.StatusNEQ(alertevent.StatusFIRING),
			suppressedPredicate(now),
		))
	} else {
		where = append(where, alertevent.StatusEQ(alertevent.StatusFIRING), alertevent.Or(
			alertevent.SuppressedEQ(false),
			alertevent.SuppressedUntilLTE(now),
		))
	}
	if filter.Name != "" {
		where = append(where, alertevent.Or(
			alertevent.SummaryContainsFold(filter.Name),
			alertevent.FingerprintContainsFold(filter.Name),
		))
	}
	if filter.LevelUID != "" {
		where = append(where, alertevent.LevelUIDEQ(filter.LevelUID))
	}
	if filter.FilterByRules {
		where = append(where, alertevent.RuleUIDIn(filter.RuleUIDs...))
	}
	matchers := map[string]string{}
	if filter.LabelsJSON != "" && filter.LabelsJSON != "{}" {
		parsed, err := labelsutil.ParseObject(filter.LabelsJSON)
		if err != nil {
			return nil, nil, err
		}
		matchers = parsed
	}
	return where, matchers, nil
}

func (r *alertEventRepo) GetEvent(ctx context.Context, namespaceUID, uid string) (*bo.AlertEvent, error) {
	row, err := r.client.AlertEvent.Query().Where(
		alertevent.NamespaceUIDEQ(namespaceUID),
		alertevent.UIDEQ(uid),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items, err := r.namedEvents(ctx, []*ent.AlertEvent{row})
	if err != nil {
		return nil, err
	}
	return &items[0], nil
}

func (r *alertEventRepo) Intervene(ctx context.Context, namespaceUID string, uids []string, userUID string) error {
	if len(uids) == 0 {
		return nil
	}
	_, err := r.client.AlertEvent.Update().Where(
		alertevent.NamespaceUIDEQ(namespaceUID),
		alertevent.UIDIn(uids...),
		alertevent.StatusEQ(alertevent.StatusFIRING),
	).SetIntervenedBy(userUID).SetIntervenedAt(time.Now().UTC()).Save(ctx)
	return r.fail(err)
}

// Suppress 抑制一条实时告警。until 为空时清掉到期时间，语义是永久抑制。
func (r *alertEventRepo) Suppress(ctx context.Context, namespaceUID, uid string, until *time.Time) error {
	update := r.client.AlertEvent.Update().Where(
		alertevent.NamespaceUIDEQ(namespaceUID),
		alertevent.UIDEQ(uid),
		alertevent.StatusEQ(alertevent.StatusFIRING),
	).SetSuppressed(true)
	if until == nil {
		update = update.ClearSuppressedUntil()
	} else {
		update = update.SetSuppressedUntil(*until)
	}
	n, err := update.Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("not found")
	}
	return nil
}

func (r *alertEventRepo) RecoverManual(ctx context.Context, namespaceUID string, uids []string) error {
	if len(uids) == 0 {
		return nil
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	rows, err := tx.AlertEvent.Query().Where(
		alertevent.NamespaceUIDEQ(namespaceUID),
		alertevent.UIDIn(uids...),
		alertevent.StatusEQ(alertevent.StatusFIRING),
	).All(ctx)
	if err != nil {
		return r.fail(err)
	}
	now := time.Now().UTC()
	for _, row := range rows {
		if _, err := tx.AlertEvent.Update().Where(
			alertevent.UIDEQ(row.UID),
			alertevent.NamespaceUIDEQ(namespaceUID),
			alertevent.StatusEQ(alertevent.StatusFIRING),
		).SetStatus(alertevent.StatusRECOVERED_BY_MANUAL).SetEndsAt(now).Save(ctx); err != nil {
			return r.fail(err)
		}
		if _, err := tx.AlertEventFiring.Delete().Where(alerteventfiring.EventUIDEQ(row.UID)).Exec(ctx); err != nil {
			return r.fail(err)
		}
	}
	if err := tx.Commit(); err != nil {
		return r.fail(err)
	}
	return nil
}

func (r *alertEventRepo) Stats(ctx context.Context, namespaceUID string) (bo.AlertStats, error) {
	var stats bo.AlertStats
	var err error
	now := time.Now().UTC()
	stats.Firing, err = r.client.AlertEvent.Query().Where(
		alertevent.NamespaceUIDEQ(namespaceUID),
		alertevent.StatusEQ(alertevent.StatusFIRING),
		alertevent.Or(
			alertevent.SuppressedEQ(false),
			alertevent.SuppressedUntilLTE(now),
		),
	).Count(ctx)
	if err != nil {
		return stats, r.fail(err)
	}
	stats.Suppressed, err = r.client.AlertEvent.Query().Where(
		alertevent.NamespaceUIDEQ(namespaceUID),
		alertevent.StatusEQ(alertevent.StatusFIRING),
		suppressedPredicate(now),
	).Count(ctx)
	if err != nil {
		return stats, r.fail(err)
	}
	stats.Recovered, err = r.client.AlertEvent.Query().Where(
		alertevent.NamespaceUIDEQ(namespaceUID),
		alertevent.StatusNEQ(alertevent.StatusFIRING),
	).Count(ctx)
	if err != nil {
		return stats, r.fail(err)
	}
	return stats, nil
}

func toEvent(row *ent.AlertEvent, now time.Time) bo.AlertEvent {
	return bo.AlertEvent{
		UID: row.UID, NamespaceUID: row.NamespaceUID, RuleUID: row.RuleUID, DatasourceUID: row.DatasourceUID,
		Fingerprint: row.Fingerprint, Status: string(row.Status), Expr: row.Expr, Labels: row.Labels,
		Summary: row.Summary, Description: row.Description, SampleValue: row.SampleValue, LevelUID: row.LevelUID,
		StartsAt: row.StartsAt, EndsAt: row.EndsAt, EvalSlot: row.EvalSlot,
		Suppressed: suppressedAt(row.Suppressed, row.SuppressedUntil, now), SuppressedUntil: row.SuppressedUntil,
		IntervenedBy: row.IntervenedBy, IntervenedAt: row.IntervenedAt, CreatedAt: row.CreatedAt,
	}
}

// suppressedAt 判断事件在给定时间是否仍在抑制中。没有到期时间的旧数据保持持续抑制。
func suppressedAt(suppressed bool, until *time.Time, now time.Time) bool {
	if !suppressed {
		return false
	}
	return until == nil || until.After(now)
}

// suppressedPredicate 是数据库侧的「仍在抑制中」条件，和 suppressedAt 语义一致。
// 抑制到期时间为空的旧数据必须继续算抑制，避免升级后历史抑制全部失效。
func suppressedPredicate(now time.Time) predicate.AlertEvent {
	return alertevent.And(
		alertevent.SuppressedEQ(true),
		alertevent.Or(
			alertevent.SuppressedUntilIsNil(),
			alertevent.SuppressedUntilGT(now),
		),
	)
}

func (r *alertEventRepo) namedEvents(ctx context.Context, rows []*ent.AlertEvent) ([]bo.AlertEvent, error) {
	items := make([]bo.AlertEvent, 0, len(rows))
	seen := map[string]struct{}{}
	uids := make([]string, 0)
	datasourceUIDs := make([]string, 0)
	seenDatasources := map[string]struct{}{}
	ruleUIDs := make([]string, 0)
	seenRules := map[string]struct{}{}
	levelUIDs := make([]string, 0)
	seenLevels := map[string]struct{}{}
	now := time.Now().UTC()
	for _, row := range rows {
		items = append(items, toEvent(row, now))
		// 软删除的数据源也要显示名称，历史事件才不会只剩一个 UID。
		if row.DatasourceUID != "" {
			if _, ok := seenDatasources[row.DatasourceUID]; !ok {
				seenDatasources[row.DatasourceUID] = struct{}{}
				datasourceUIDs = append(datasourceUIDs, row.DatasourceUID)
			}
		}
		// 规则和级别同样按 UID 补名称，详情页直接显示名称并跳转，不用再发请求。
		if row.RuleUID != "" {
			if _, ok := seenRules[row.RuleUID]; !ok {
				seenRules[row.RuleUID] = struct{}{}
				ruleUIDs = append(ruleUIDs, row.RuleUID)
			}
		}
		if row.LevelUID != "" {
			if _, ok := seenLevels[row.LevelUID]; !ok {
				seenLevels[row.LevelUID] = struct{}{}
				levelUIDs = append(levelUIDs, row.LevelUID)
			}
		}
		if row.IntervenedBy == "" {
			continue
		}
		if _, ok := seen[row.IntervenedBy]; ok {
			continue
		}
		seen[row.IntervenedBy] = struct{}{}
		uids = append(uids, row.IntervenedBy)
	}
	sources := r.datasourceMetas(ctx, datasourceUIDs)
	// 数据源等级名称也要补进标签和详情，与告警等级一起取字典。
	for _, source := range sources {
		if source.LevelUID == "" {
			continue
		}
		if _, ok := seenLevels[source.LevelUID]; ok {
			continue
		}
		seenLevels[source.LevelUID] = struct{}{}
		levelUIDs = append(levelUIDs, source.LevelUID)
	}
	rules := r.ruleNames(ctx, ruleUIDs)
	levels := r.levelNames(ctx, levelUIDs)
	for i := range items {
		source := sources[items[i].DatasourceUID]
		items[i].DatasourceName = source.Name
		items[i].RuleName, items[i].RuleGroupUID = rules[items[i].RuleUID].Name, rules[items[i].RuleUID].GroupUID
		items[i].LevelName = levels[items[i].LevelUID]
		// 旧事件没有名称标签，读取时补上，详情标签与写入后的新事件保持一致。
		items[i].Labels = backfillEventLabels(items[i].Labels, source.Name, items[i].LevelName, levels[source.LevelUID])
	}
	if len(uids) == 0 {
		return items, nil
	}
	users, err := r.client.User.Query().Where(user.UIDIn(uids...)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	userNames := make(map[string]string, len(users))
	for _, account := range users {
		userNames[account.UID] = accountName(account)
	}
	for i := range items {
		items[i].IntervenedName = userNames[items[i].IntervenedBy]
	}
	return items, nil
}

// datasourceNameInfo 是事件展示需要的数据源名称和等级标识。
type datasourceNameInfo struct {
	Name     string
	LevelUID string
}

// datasourceMetas 一次取出事件用到的数据源名称和等级。查询不排除软删除，
// 历史事件仍然能显示当时的数据源名称。
func (r *alertEventRepo) datasourceMetas(ctx context.Context, uids []string) map[string]datasourceNameInfo {
	metas := map[string]datasourceNameInfo{}
	if len(uids) == 0 {
		return metas
	}
	rows, err := r.client.Datasource.Query().Where(datasource.UIDIn(uids...)).All(ctx)
	if err != nil {
		r.log.Error("datasource names: " + err.Error())
		return metas
	}
	for _, row := range rows {
		metas[row.UID] = datasourceNameInfo{Name: row.Name, LevelUID: row.LevelUID}
	}
	return metas
}

// ruleNameInfo 是事件详情展示规则名称和跳转规则组所需的字段。
type ruleNameInfo struct {
	Name     string
	GroupUID string
}

// ruleNames 一次取出事件用到的规则名称。查询不排除软删除，
// 规则删除后历史事件仍能显示名称和所属规则组。
func (r *alertEventRepo) ruleNames(ctx context.Context, uids []string) map[string]ruleNameInfo {
	infos := map[string]ruleNameInfo{}
	if len(uids) == 0 {
		return infos
	}
	rows, err := r.client.AlertRule.Query().Where(alertrule.UIDIn(uids...)).All(ctx)
	if err != nil {
		r.log.Error("rule names: " + err.Error())
		return infos
	}
	for _, row := range rows {
		infos[row.UID] = ruleNameInfo{Name: row.Name, GroupUID: row.GroupUID}
	}
	return infos
}

// levelNames 一次取出事件用到的告警级别名称。查询不排除软删除，
// 级别被删除后历史事件仍能显示名称。
func (r *alertEventRepo) levelNames(ctx context.Context, uids []string) map[string]string {
	names := map[string]string{}
	if len(uids) == 0 {
		return names
	}
	rows, err := r.client.AlertDictionary.Query().Where(alertdictionary.UIDIn(uids...)).All(ctx)
	if err != nil {
		r.log.Error("level names: " + err.Error())
		return names
	}
	for _, row := range rows {
		names[row.UID] = row.Name
	}
	return names
}

// datasourceMetaInfo 是写事件时需要的数据源信息。
type datasourceMetaInfo struct {
	Name     string
	Labels   string
	Endpoint string
	LevelUID string
}

// firingEvent 是当前 FIRING 占位和它对应的事件快照。
type firingEvent struct {
	firing *ent.AlertEventFiring
	event  *ent.AlertEvent
}

// hitLevelUIDs 返回本轮命中使用的级别标识。
func hitLevelUIDs(hits map[string]bo.EvalHit) []string {
	uids := make([]string, 0, len(hits))
	for _, hit := range hits {
		if hit.LevelUID != "" {
			uids = append(uids, hit.LevelUID)
		}
	}
	return uids
}

// datasourceMeta 一次取出写事件需要的数据源名称、标签和地址。
// 查询不排除软删除：数据源刚被删除时，回写仍应带上当时的标签。
func (r *alertEventRepo) datasourceMeta(ctx context.Context, tx *ent.Tx, uid string) datasourceMetaInfo {
	if uid == "" {
		return datasourceMetaInfo{}
	}
	row, err := tx.Datasource.Query().Where(datasource.UIDEQ(uid)).Only(ctx)
	if err != nil {
		return datasourceMetaInfo{}
	}
	return datasourceMetaInfo{Name: row.Name, Labels: row.Labels, Endpoint: row.Endpoint, LevelUID: row.LevelUID}
}

// 事件标签里保留数据源、告警等级和数据源等级名称的键。
const (
	datasourceLabelKey      = "datasource"
	alertLevelLabelKey      = "alert_level"
	datasourceLevelLabelKey = "datasource_level"
)

// mergeEventLabels 把数据源名称、数据源自身标签、告警等级名称和数据源等级名称
// 附加到事件标签上。指标标签优先：同名键保留采样标签，元数据只补缺。
func mergeEventLabels(eventLabels string, source datasourceMetaInfo, levelUID string, levelNames map[string]string) string {
	merged, err := labelsutil.ParseObject(eventLabels)
	if err != nil {
		// 无法解析的旧标签保持原样，不因附加元数据把事件写坏。
		return eventLabels
	}
	if sourceLabels, err := labelsutil.ParseObject(source.Labels); err == nil {
		for key, value := range sourceLabels {
			if _, ok := merged[key]; !ok {
				merged[key] = value
			}
		}
	}
	extras := map[string]string{
		datasourceLabelKey:      source.Name,
		alertLevelLabelKey:      levelNames[levelUID],
		datasourceLevelLabelKey: levelNames[source.LevelUID],
	}
	fillLabels(merged, extras)
	out, err := json.Marshal(merged)
	if err != nil {
		return eventLabels
	}
	return string(out)
}

// fillLabels 只补缺失的键，已有标签永远优先。
func fillLabels(labels map[string]string, extras map[string]string) {
	for key, value := range extras {
		if value == "" {
			continue
		}
		if _, ok := labels[key]; !ok {
			labels[key] = value
		}
	}
}

// backfillEventLabels 给历史事件补上展示用的名称标签；写入时已经带上的不覆盖。
func backfillEventLabels(eventLabels, datasourceName, levelName, datasourceLevelName string) string {
	merged, err := labelsutil.ParseObject(eventLabels)
	if err != nil {
		return eventLabels
	}
	fillLabels(merged, map[string]string{
		datasourceLabelKey:      datasourceName,
		alertLevelLabelKey:      levelName,
		datasourceLevelLabelKey: datasourceLevelName,
	})
	out, err := json.Marshal(merged)
	if err != nil {
		return eventLabels
	}
	return string(out)
}

func accountName(account *ent.User) string {
	if account.Nickname != "" {
		return account.Nickname
	}
	if account.Name != "" {
		return account.Name
	}
	return account.Email
}

func (r *alertEventRepo) RecoverSource(ctx context.Context, namespaceUID string, ruleUIDs, datasourceUIDs, levelUIDs []string) error {
	if len(ruleUIDs) == 0 && len(datasourceUIDs) == 0 && len(levelUIDs) == 0 {
		return nil
	}
	where := []predicate.AlertEvent{
		alertevent.NamespaceUIDEQ(namespaceUID),
		alertevent.StatusEQ(alertevent.StatusFIRING),
	}
	if len(ruleUIDs) > 0 {
		where = append(where, alertevent.RuleUIDIn(ruleUIDs...))
	}
	if len(datasourceUIDs) > 0 {
		where = append(where, alertevent.DatasourceUIDIn(datasourceUIDs...))
	}
	if len(levelUIDs) > 0 {
		where = append(where, alertevent.LevelUIDIn(levelUIDs...))
	}
	rows, err := r.client.AlertEvent.Query().Where(where...).All(ctx)
	if err != nil {
		return r.fail(err)
	}
	if len(rows) == 0 {
		return nil
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	now := time.Now().UTC()
	for _, row := range rows {
		if _, err := tx.AlertEvent.Update().Where(
			alertevent.UIDEQ(row.UID),
			alertevent.NamespaceUIDEQ(namespaceUID),
			alertevent.StatusEQ(alertevent.StatusFIRING),
		).SetStatus(alertevent.StatusRECOVERED).SetEndsAt(now).Save(ctx); err != nil {
			return r.fail(err)
		}
		if _, err := tx.AlertEventFiring.Delete().Where(
			alerteventfiring.NamespaceUIDEQ(namespaceUID),
			alerteventfiring.EventUIDEQ(row.UID),
		).Exec(ctx); err != nil {
			return r.fail(err)
		}
		if err := dispatchEventNotifications(ctx, tx, notifyEvent{
			NamespaceUID: namespaceUID, EventUID: row.UID, RuleUID: row.RuleUID,
			DatasourceUID: row.DatasourceUID, LevelUID: row.LevelUID, Phase: "RECOVERED",
			Summary: row.Summary, Description: row.Description, Labels: row.Labels,
			Value: row.SampleValue, ExternalURL: r.datasourceEndpoint(ctx, tx, row.DatasourceUID),
			Fingerprint: row.Fingerprint,
		}); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return r.fail(err)
	}
	return nil
}

// datasourceEndpoint 返回数据源地址，作为模板里的 $externalURL，和 megrez 评测时一致。
func (r *alertEventRepo) datasourceEndpoint(ctx context.Context, tx *ent.Tx, datasourceUID string) string {
	if datasourceUID == "" {
		return ""
	}
	row, err := tx.Datasource.Query().Where(
		datasource.UIDEQ(datasourceUID),
		datasource.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return ""
	}
	return row.Endpoint
}

func textOr(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

// recoverRuleFiring 把某条规则当前所有 FIRING 事件转为恢复，并清空 firing 标记。
// 停用或删除规则后不会再评测，必须在这里主动收口，否则实时告警会永久挂着。
func recoverRuleFiring(ctx context.Context, client *ent.Client, namespaceUID, ruleUID string) error {
	rows, err := client.AlertEventFiring.Query().Where(
		alerteventfiring.NamespaceUIDEQ(namespaceUID),
		alerteventfiring.RuleUIDEQ(ruleUID),
	).All(ctx)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	for _, row := range rows {
		if _, err := client.AlertEvent.Update().Where(
			alertevent.UIDEQ(row.EventUID),
			alertevent.NamespaceUIDEQ(namespaceUID),
			alertevent.StatusEQ(alertevent.StatusFIRING),
		).SetStatus(alertevent.StatusRECOVERED).SetEndsAt(now).Save(ctx); err != nil {
			return err
		}
		if _, err := client.AlertEventFiring.Delete().Where(alerteventfiring.IDEQ(row.ID)).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}
