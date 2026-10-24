package data

import (
	"context"
	"encoding/json"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertdictionary"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertevalbatch"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertevent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alerteventfiring"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alerthistoryevent"
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
	// 分区由 partition.Maintainer 定时预创建：写路径不再执行 DDL，避免 catalog 锁和写延迟。
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
	recovered, err := r.applyHits(ctx, tx, rule, report)
	if err != nil {
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
	// 恢复事件在事务提交后归档到租户历史库。归档失败时保留主库行，
	// 历史查询仍能在合并结果里看到它，不会因为分库暂时不可用而丢历史。
	r.archiveEvents(ctx, rule.NamespaceUID, recovered)
	return nil
}

func (r *alertEventRepo) applyHits(ctx context.Context, tx *ent.Tx, rule bo.AlertRule, report bo.EvalReport) ([]bo.AlertEvent, error) {
	rows, err := tx.AlertEventFiring.Query().Where(
		alerteventfiring.NamespaceUIDEQ(rule.NamespaceUID),
		alerteventfiring.RuleUIDEQ(rule.UID),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	current := make(map[string]firingEvent, len(rows))
	for _, row := range rows {
		event, err := tx.AlertEvent.Query().Where(
			alertevent.UIDEQ(row.EventUID),
			alertevent.NamespaceUIDEQ(rule.NamespaceUID),
		).Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				// 事件已经归档但 firing 占位没删干净时，先清掉孤儿行，
				// 否则这条规则之后的每轮评测都会卡在这里。
				if _, deleteErr := tx.AlertEventFiring.Delete().Where(alerteventfiring.IDEQ(row.ID)).Exec(ctx); deleteErr != nil {
					return nil, r.fail(deleteErr)
				}
				continue
			}
			return nil, r.fail(err)
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
	recovered := make([]bo.AlertEvent, 0)
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
				return nil, r.fail(err)
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
			return nil, r.fail(err)
		}
		if err := tx.AlertEventFiring.Create().
			SetNamespaceUID(rule.NamespaceUID).
			SetFingerprint(fingerprint).
			SetEventUID(eventUID).
			SetRuleUID(rule.UID).
			Exec(ctx); err != nil {
			return nil, r.fail(err)
		}
		if err := dispatchEventNotifications(ctx, tx, notifyEvent{
			NamespaceUID: rule.NamespaceUID, EventUID: eventUID, RuleUID: rule.UID,
			DatasourceUID: report.DatasourceUID, LevelUID: hit.LevelUID, Phase: "FIRING",
			Summary: textOr(hit.Summary, rule.Summary), Description: textOr(hit.Description, rule.Description),
			Labels: labels, Value: hit.Value, ExternalURL: source.Endpoint, Fingerprint: fingerprint,
		}); err != nil {
			return nil, err
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
			return nil, r.fail(err)
		}
		if _, err := tx.AlertEventFiring.Delete().Where(alerteventfiring.IDEQ(ref.firing.ID)).Exec(ctx); err != nil {
			return nil, r.fail(err)
		}
		recoveredItem := toEvent(ref.event, now)
		recoveredItem.Status = string(alertevent.StatusRECOVERED)
		recoveredItem.EndsAt = &now
		recoveredItem.EvalSlot = report.EvalSlot
		recovered = append(recovered, recoveredItem)
		if err := dispatchEventNotifications(ctx, tx, notifyEvent{
			NamespaceUID: rule.NamespaceUID, EventUID: ref.event.UID, RuleUID: rule.UID,
			DatasourceUID: ref.event.DatasourceUID, LevelUID: ref.event.LevelUID, Phase: "RECOVERED",
			Summary: ref.event.Summary, Description: ref.event.Description, Labels: ref.event.Labels,
			Value: ref.event.SampleValue, ExternalURL: source.Endpoint, Fingerprint: fingerprint,
		}); err != nil {
			return nil, err
		}
	}
	return recovered, nil
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
			if ent.IsNotFound(err) {
				_, _ = r.client.AlertEventFiring.Delete().Where(alerteventfiring.IDEQ(row.ID)).Exec(ctx)
				continue
			}
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
	if filter.FilterByEvents && len(filter.EventUIDs) == 0 {
		return nil, 0, nil
	}
	if filter.View == "HISTORY" {
		return r.listHistoryEvents(ctx, namespaceUID, filter, page, pageSize)
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
	if filter.FilterByEvents && len(filter.EventUIDs) == 0 {
		return 0, nil
	}
	if filter.View == "HISTORY" {
		items, err := r.historyItems(ctx, namespaceUID, filter)
		if err != nil {
			return 0, err
		}
		return len(items), nil
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

// ListEventUIDsForUser 返回当前成员实际能收到通知的事件。
// 这里按 NotificationTask 的事件标识判断，而不是先取规则再放行整个规则：
// 同一规则下可能只有部分事件命中用户的订阅，不能因为规则曾投递给该用户，
// 就把该规则的所有实时事件都算成「我的告警」。
func (r *alertEventRepo) ListEventUIDsForUser(ctx context.Context, namespaceUID, memberUID string, tokens []string) ([]string, error) {
	groups, err := r.client.NotificationGroup.Query().Where(
		notificationgroup.NamespaceUIDEQ(namespaceUID),
		notificationgroup.DeletedAtIsNil(),
		notificationgroup.StatusEQ(notificationgroup.StatusACTIVE),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	groupUIDs := make([]string, 0, len(groups))
	for _, group := range groups {
		if groupMatchesUser(group, memberUID, tokens) {
			groupUIDs = append(groupUIDs, group.UID)
		}
	}

	conditions := make([]predicate.NotificationTask, 0, len(groupUIDs)+len(tokens))
	if len(groupUIDs) > 0 {
		conditions = append(conditions, notificationtask.GroupUIDIn(groupUIDs...))
	}
	for _, token := range tokens {
		if strings.TrimSpace(token) != "" {
			conditions = append(conditions, notificationtask.ReceiversContainsFold(token))
		}
	}
	if len(conditions) == 0 {
		return nil, nil
	}
	tasks, err := r.client.NotificationTask.Query().Where(
		notificationtask.NamespaceUIDEQ(namespaceUID),
		notificationtask.EventUIDNEQ(""),
		notificationtask.Or(conditions...),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	groupSet := make(map[string]struct{}, len(groupUIDs))
	for _, uid := range groupUIDs {
		groupSet[uid] = struct{}{}
	}
	seen := make(map[string]struct{}, len(tasks))
	uids := make([]string, 0, len(tasks))
	for _, task := range tasks {
		if task.EventUID == "" {
			continue
		}
		_, byGroup := groupSet[task.GroupUID]
		// SQL 的 ContainsFold 只是粗筛，接收方是逗号分隔列表，必须再按完整项比较，
		// 避免 alice@example.com 命中 alice@example.com.cn。
		if !byGroup && !receiversContain(task.Receivers, tokens) {
			continue
		}
		if _, ok := seen[task.EventUID]; ok {
			continue
		}
		seen[task.EventUID] = struct{}{}
		uids = append(uids, task.EventUID)
	}
	sort.Strings(uids)
	return uids, nil
}

// groupMatchesUser 判断当前成员是否是通知组的接收人。
// 同时检查新字段和旧版字段，保证升级前后创建的通知组都能识别。
func groupMatchesUser(group *ent.NotificationGroup, memberUID string, tokens []string) bool {
	if memberUID != "" &&
		(containsString(group.MemberUids, memberUID) ||
			containsString(group.EmailMemberUids, memberUID) ||
			containsString(group.SmsMemberUids, memberUID)) {
		return true
	}
	return receiversContain(group.Receivers, tokens) ||
		receiversContain(group.EmailReceivers, tokens) ||
		receiversContain(group.SmsReceivers, tokens)
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
	if filter.FilterByEvents {
		where = append(where, alertevent.UIDIn(filter.EventUIDs...))
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
		if ent.IsNotFound(err) {
			return r.getHistoryEvent(ctx, namespaceUID, uid)
		}
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
	recovered := make([]bo.AlertEvent, 0, len(rows))
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
		item := toEvent(row, now)
		item.Status = string(alertevent.StatusRECOVERED_BY_MANUAL)
		item.EndsAt = &now
		recovered = append(recovered, item)
	}
	if err := tx.Commit(); err != nil {
		return r.fail(err)
	}
	r.archiveEvents(ctx, namespaceUID, recovered)
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
	// 已归档的历史事件不在主库。归档失败时主库行会保留，所以两处相加后
	// 还要扣掉「主库和归档索引同时存在」的部分，避免统计重复。
	if r.history != nil {
		// 历史库暂时不可用时只降级为「主库恢复数」，实时告警页不应因此整页报错。
		if archived, err := r.history.count(ctx, namespaceUID); err != nil {
			r.log.Error("count alert history: " + err.Error())
		} else if overlap, err := r.archivedHotCount(ctx, namespaceUID); err != nil {
			r.log.Error("count archived hot alerts: " + err.Error())
		} else {
			stats.Recovered += archived - overlap
			if stats.Recovered < 0 {
				stats.Recovered = 0
			}
		}
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
	now := time.Now().UTC()
	items := make([]bo.AlertEvent, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEvent(row, now))
	}
	return r.nameEvents(ctx, items)
}

// nameEvents 给事件补规则、数据源、级别和介入人名称。热数据和历史分片共用。
func (r *alertEventRepo) nameEvents(ctx context.Context, items []bo.AlertEvent) ([]bo.AlertEvent, error) {
	seen := map[string]struct{}{}
	uids := make([]string, 0)
	datasourceUIDs := make([]string, 0)
	seenDatasources := map[string]struct{}{}
	ruleUIDs := make([]string, 0)
	seenRules := map[string]struct{}{}
	levelUIDs := make([]string, 0)
	seenLevels := map[string]struct{}{}
	for _, row := range items {
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
	sample, err := labelsutil.ParseObject(eventLabels)
	if err != nil {
		// 无法解析的旧标签保持原样，不因附加元数据把事件写坏。
		return eventLabels
	}
	merged := make(map[string]string, len(sample))
	for key, value := range sample {
		merged[key] = value
	}
	if sourceLabels, err := labelsutil.ParseObject(source.Labels); err == nil {
		for key, value := range sourceLabels {
			if _, ok := merged[key]; !ok {
				merged[key] = value
			}
		}
	}
	// 元数据键每次命中都重算：数据源改名或改等级后，事件标签不能继续保留旧值。
	// 采样标签同名的键仍然优先，不被元数据覆盖。
	for key, value := range map[string]string{
		datasourceLabelKey:      source.Name,
		alertLevelLabelKey:      levelNames[levelUID],
		datasourceLevelLabelKey: levelNames[source.LevelUID],
	} {
		if _, fromSample := sample[key]; fromSample {
			continue
		}
		if value == "" {
			delete(merged, key)
			continue
		}
		merged[key] = value
	}
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
	recovered := make([]bo.AlertEvent, 0, len(rows))
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
		item := toEvent(row, now)
		item.Status = string(alertevent.StatusRECOVERED)
		item.EndsAt = &now
		recovered = append(recovered, item)
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
	r.archiveEvents(ctx, namespaceUID, recovered)
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
func recoverRuleFiring(ctx context.Context, client *ent.Client, history *historyStore, log *slog.Logger, namespaceUID, ruleUID string) error {
	rows, err := client.AlertEventFiring.Query().Where(
		alerteventfiring.NamespaceUIDEQ(namespaceUID),
		alerteventfiring.RuleUIDEQ(ruleUID),
	).All(ctx)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	recovered := make([]bo.AlertEvent, 0, len(rows))
	for _, row := range rows {
		event, err := client.AlertEvent.Query().Where(
			alertevent.UIDEQ(row.EventUID),
			alertevent.NamespaceUIDEQ(namespaceUID),
		).Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				_, _ = client.AlertEventFiring.Delete().Where(alerteventfiring.IDEQ(row.ID)).Exec(ctx)
				continue
			}
			return err
		}
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
		item := toEvent(event, now)
		item.Status = string(alertevent.StatusRECOVERED)
		item.EndsAt = &now
		recovered = append(recovered, item)
	}
	archiveRecoveredItems(ctx, client, history, log, namespaceUID, recovered)
	return nil
}

// archiveEvents 把恢复事件写到租户历史库，再删掉主库行并记录路由索引。
// 任何一步失败都保留主库行：历史列表会合并主库兜底，宁可多留一条也不丢历史。
func (r *alertEventRepo) archiveEvents(ctx context.Context, namespaceUID string, items []bo.AlertEvent) {
	archiveRecoveredItems(ctx, r.client, r.history, r.log, namespaceUID, items)
}

// archiveRecoveredItems 与具体仓储无关，规则状态切换的直接恢复路径也能复用。
func archiveRecoveredItems(ctx context.Context, client *ent.Client, history *historyStore, log *slog.Logger, namespaceUID string, items []bo.AlertEvent) {
	if history == nil || len(items) == 0 {
		return
	}
	for _, item := range items {
		// 回填会一次处理多个租户，必须用事件自己的租户标识写索引和删除主库行。
		eventNamespace := item.NamespaceUID
		if eventNamespace == "" {
			eventNamespace = namespaceUID
			item.NamespaceUID = eventNamespace
		}
		database, table, week, err := history.archive(ctx, item)
		if err != nil {
			log.Error("archive alert event " + item.UID + ": " + err.Error())
			continue
		}
		if err := recordHistoryIndex(ctx, client, eventNamespace, item.UID, database, table, week); err != nil {
			log.Error("index alert event " + item.UID + ": " + err.Error())
			continue
		}
		if _, err := client.AlertEvent.Delete().Where(
			alertevent.NamespaceUIDEQ(eventNamespace),
			alertevent.UIDEQ(item.UID),
		).Exec(ctx); err != nil {
			log.Error("delete archived alert event " + item.UID + ": " + err.Error())
		}
		if _, err := client.AlertEventFiring.Delete().Where(
			alerteventfiring.NamespaceUIDEQ(eventNamespace),
			alerteventfiring.EventUIDEQ(item.UID),
		).Exec(ctx); err != nil {
			log.Error("delete archived firing " + item.UID + ": " + err.Error())
		}
	}
}

// recordHistoryIndex 记录事件所在租户库和周表，详情按 UID 直接定位，不用扫描所有周表。
func recordHistoryIndex(ctx context.Context, client *ent.Client, namespaceUID, eventUID, database, table string, week time.Time) error {
	existing, err := client.AlertHistoryEvent.Query().Where(alerthistoryevent.EventUIDEQ(eventUID)).Only(ctx)
	if err == nil {
		_, err = client.AlertHistoryEvent.Update().Where(alerthistoryevent.IDEQ(existing.ID)).
			SetDatabaseName(database).
			SetTableName(table).
			SetWeekStart(week).
			Save(ctx)
		return err
	}
	if !ent.IsNotFound(err) {
		return err
	}
	_, err = client.AlertHistoryEvent.Create().
		SetUID(id.New()).
		SetNamespaceUID(namespaceUID).
		SetEventUID(eventUID).
		SetDatabaseName(database).
		SetTableName(table).
		SetWeekStart(week).
		Save(ctx)
	if err != nil && ent.IsConstraintError(err) {
		return nil
	}
	return err
}

// archivedHotCount 统计仍留在主库、但已经写入历史索引的恢复事件，用于统计去重。
func (r *alertEventRepo) archivedHotCount(ctx context.Context, namespaceUID string) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, `SELECT count(*) FROM alert_history_events i
		JOIN alert_events e ON e.uid = i.event_uid
		WHERE i.namespace_uid = $1 AND e.namespace_uid = $1 AND e.status <> 'FIRING'`, namespaceUID).Scan(&count); err != nil {
		return 0, r.fail(err)
	}
	return count, nil
}

// historyItems 合并主库里尚未归档的历史行（旧数据、归档失败兜底）和租户历史库的周表。
func (r *alertEventRepo) historyItems(ctx context.Context, namespaceUID string, filter bo.AlertEventListFilter) ([]bo.AlertEvent, error) {
	now := time.Now().UTC()
	rows, err := r.client.AlertEvent.Query().Where(
		alertevent.NamespaceUIDEQ(namespaceUID),
		alertevent.Or(
			alertevent.StatusNEQ(alertevent.StatusFIRING),
			suppressedPredicate(now),
		),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]bo.AlertEvent, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEvent(row, now))
	}
	if r.history != nil {
		archived, err := r.history.list(ctx, namespaceUID)
		if err != nil {
			// 分库不可用时退回主库历史行，页面仍能展示升级前和归档失败的数据。
			r.log.Error("list alert history: " + err.Error())
			if requiresHistoryNames(filter) {
				items, err = r.nameEvents(ctx, items)
				if err != nil {
					return nil, err
				}
			}
			return filterHistoryItems(items, filter)
		}
		// 主库行优先：归档成功但删除主库失败时，主库状态仍是最新的。
		index := make(map[string]struct{}, len(items))
		for _, item := range items {
			index[item.UID] = struct{}{}
		}
		for _, item := range archived {
			if _, ok := index[item.UID]; ok {
				continue
			}
			items = append(items, item)
		}
	}
	if requiresHistoryNames(filter) {
		items, err = r.nameEvents(ctx, items)
		if err != nil {
			return nil, err
		}
	}
	return filterHistoryItems(items, filter)
}

func requiresHistoryNames(filter bo.AlertEventListFilter) bool {
	return strings.TrimSpace(filter.RuleName) != "" || strings.TrimSpace(filter.DatasourceName) != ""
}

func filterHistoryItems(items []bo.AlertEvent, filter bo.AlertEventListFilter) ([]bo.AlertEvent, error) {
	var matchers map[string]string
	if filter.LabelsJSON != "" && filter.LabelsJSON != "{}" {
		parsed, err := labelsutil.ParseObject(filter.LabelsJSON)
		if err != nil {
			return nil, err
		}
		matchers = parsed
	}
	name := strings.ToLower(strings.TrimSpace(filter.Name))
	ruleName := strings.ToLower(strings.TrimSpace(filter.RuleName))
	datasourceName := strings.ToLower(strings.TrimSpace(filter.DatasourceName))
	out := make([]bo.AlertEvent, 0, len(items))
	for _, item := range items {
		if name != "" && !strings.Contains(strings.ToLower(item.Summary), name) && !strings.Contains(strings.ToLower(item.Fingerprint), name) {
			continue
		}
		if filter.LevelUID != "" && item.LevelUID != filter.LevelUID {
			continue
		}
		if ruleName != "" && !strings.Contains(strings.ToLower(item.RuleName), ruleName) {
			continue
		}
		if datasourceName != "" && !strings.Contains(strings.ToLower(item.DatasourceName), datasourceName) {
			continue
		}
		if filter.StartsAtFrom != nil && item.StartsAt.Before(*filter.StartsAtFrom) {
			continue
		}
		if filter.StartsAtTo != nil && !item.StartsAt.Before(*filter.StartsAtTo) {
			continue
		}
		if filter.EventStatus != "" && item.Status != filter.EventStatus {
			continue
		}
		if filter.Suppression == "SUPPRESSED" && !item.Suppressed {
			continue
		}
		if filter.Suppression == "UNSUPPRESSED" && item.Suppressed {
			continue
		}
		if filter.FilterByRules {
			allowed := false
			for _, uid := range filter.RuleUIDs {
				if uid == item.RuleUID {
					allowed = true
					break
				}
			}
			if !allowed {
				continue
			}
		}
		if filter.FilterByEvents {
			allowed := false
			for _, uid := range filter.EventUIDs {
				if uid == item.UID {
					allowed = true
					break
				}
			}
			if !allowed {
				continue
			}
		}
		if len(matchers) > 0 {
			labels, err := labelsutil.ParseObject(item.Labels)
			if err != nil || !labelsutil.CompileMatchers(matchers)(labels) {
				continue
			}
		}
		out = append(out, item)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].UID > out[j].UID
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out, nil
}

func (r *alertEventRepo) listHistoryEvents(ctx context.Context, namespaceUID string, filter bo.AlertEventListFilter, page, pageSize int) ([]bo.AlertEvent, int, error) {
	items, err := r.historyItems(ctx, namespaceUID, filter)
	if err != nil {
		return nil, 0, err
	}
	total := len(items)
	start := (page - 1) * pageSize
	if start >= total {
		return []bo.AlertEvent{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	pageItems := items[start:end]
	if !requiresHistoryNames(filter) {
		pageItems, err = r.nameEvents(ctx, pageItems)
		if err != nil {
			return nil, 0, err
		}
	}
	return pageItems, total, nil
}

func (r *alertEventRepo) getHistoryEvent(ctx context.Context, namespaceUID, uid string) (*bo.AlertEvent, error) {
	if r.history == nil {
		return nil, merr.NotFound("not found")
	}
	index, err := r.client.AlertHistoryEvent.Query().Where(
		alerthistoryevent.NamespaceUIDEQ(namespaceUID),
		alerthistoryevent.EventUIDEQ(uid),
	).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, merr.NotFound("not found")
		}
		return nil, r.fail(err)
	}
	item, err := r.history.get(ctx, namespaceUID, index.DatabaseName, index.TableName, uid)
	if err != nil {
		return nil, err
	}
	items, err := r.nameEvents(ctx, []bo.AlertEvent{*item})
	if err != nil {
		return nil, err
	}
	return &items[0], nil
}

// BackfillHistory 把主库里遗留的已恢复事件归档到租户历史库。新恢复的事件走
// archiveEvents，这里是升级旧数据和归档失败重试的兜底。
func (r *alertEventRepo) BackfillHistory(ctx context.Context, limit int) (int, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := r.client.AlertEvent.Query().Where(
		alertevent.StatusNEQ(alertevent.StatusFIRING),
	).Order(ent.Asc(alertevent.FieldCreatedAt)).Limit(limit).All(ctx)
	if err != nil {
		return 0, r.fail(err)
	}
	now := time.Now().UTC()
	items := make([]bo.AlertEvent, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEvent(row, now))
	}
	r.archiveEvents(ctx, "", items)
	return len(items), nil
}
