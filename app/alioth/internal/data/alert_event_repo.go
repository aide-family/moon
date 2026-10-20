package data

import (
	"context"
	"strings"
	"time"

	entsql "entgo.io/ent/dialect/sql"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertevalbatch"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertevent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alerteventfiring"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationgroup"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationtask"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/predicate"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/user"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/id"
	"github.com/aide-family/Alioth/pkg/merr"
)

func (r *alertEventRepo) ApplyReport(ctx context.Context, rule bo.AlertRule, report bo.EvalReport) error {
	if err := ensureAlertEventWeeks(ctx, r.db, time.Now().UTC()); err != nil {
		return err
	}
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
	current := make(map[string]*ent.AlertEventFiring, len(rows))
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
		current[row.Fingerprint] = row
	}
	hits := make(map[string]bo.EvalHit, len(report.Hits))
	for _, hit := range report.Hits {
		if hit.Fingerprint == "" {
			continue
		}
		hits[hit.Fingerprint] = hit
	}
	now := time.Now().UTC()
	for fingerprint, hit := range hits {
		if row, ok := current[fingerprint]; ok {
			if _, err := tx.AlertEvent.Update().Where(
				alertevent.UIDEQ(row.EventUID),
				alertevent.NamespaceUIDEQ(rule.NamespaceUID),
				alertevent.StatusEQ(alertevent.StatusFIRING),
			).
				SetLabels(hit.Labels).
				SetSummary(textOr(hit.Summary, rule.Summary)).
				SetDescription(textOr(hit.Description, rule.Description)).
				SetSampleValue(hit.Value).
				SetLevelUID(hit.LevelUID).
				SetEvalSlot(report.EvalSlot).
				Save(ctx); err != nil {
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
			SetLabels(hit.Labels).
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
			Labels: hit.Labels, Fingerprint: fingerprint,
		}); err != nil {
			return err
		}
	}
	for fingerprint, row := range current {
		if _, ok := hits[fingerprint]; ok {
			continue
		}
		event, err := tx.AlertEvent.Query().Where(alertevent.UIDEQ(row.EventUID)).Only(ctx)
		if err != nil {
			return r.fail(err)
		}
		if _, err := tx.AlertEvent.Update().Where(
			alertevent.UIDEQ(row.EventUID),
			alertevent.NamespaceUIDEQ(rule.NamespaceUID),
			alertevent.StatusEQ(alertevent.StatusFIRING),
		).SetStatus(alertevent.StatusRECOVERED).SetEndsAt(now).SetEvalSlot(report.EvalSlot).Save(ctx); err != nil {
			return r.fail(err)
		}
		if _, err := tx.AlertEventFiring.Delete().Where(alerteventfiring.IDEQ(row.ID)).Exec(ctx); err != nil {
			return r.fail(err)
		}
		if err := dispatchEventNotifications(ctx, tx, notifyEvent{
			NamespaceUID: rule.NamespaceUID, EventUID: row.EventUID, RuleUID: rule.UID,
			DatasourceUID: event.DatasourceUID, LevelUID: event.LevelUID, Phase: "RECOVERED",
			Summary: event.Summary, Description: event.Description, Labels: event.Labels, Fingerprint: fingerprint,
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
	where := eventWhere(namespaceUID, filter)
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

func eventWhere(namespaceUID string, filter bo.AlertEventListFilter) []predicate.AlertEvent {
	where := []predicate.AlertEvent{alertevent.NamespaceUIDEQ(namespaceUID)}
	if filter.View == "HISTORY" {
		where = append(where, alertevent.Or(
			alertevent.StatusNEQ(alertevent.StatusFIRING),
			alertevent.SuppressedEQ(true),
		))
	} else {
		where = append(where, alertevent.StatusEQ(alertevent.StatusFIRING), alertevent.SuppressedEQ(false))
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
	if filter.LabelsJSON != "" && filter.LabelsJSON != "{}" {
		matchers := filter.LabelsJSON
		where = append(where, predicate.AlertEvent(func(s *entsql.Selector) {
			s.Where(entsql.P(func(b *entsql.Builder) {
				b.WriteString("labels::jsonb @> ")
				b.Arg(matchers)
				b.WriteString("::jsonb")
			}))
		}))
	}
	return where
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

func (r *alertEventRepo) Suppress(ctx context.Context, namespaceUID, uid string) error {
	n, err := r.client.AlertEvent.Update().Where(
		alertevent.NamespaceUIDEQ(namespaceUID),
		alertevent.UIDEQ(uid),
		alertevent.StatusEQ(alertevent.StatusFIRING),
	).SetSuppressed(true).Save(ctx)
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
	stats.Firing, err = r.client.AlertEvent.Query().Where(
		alertevent.NamespaceUIDEQ(namespaceUID),
		alertevent.StatusEQ(alertevent.StatusFIRING),
		alertevent.SuppressedEQ(false),
	).Count(ctx)
	if err != nil {
		return stats, r.fail(err)
	}
	stats.Suppressed, err = r.client.AlertEvent.Query().Where(
		alertevent.NamespaceUIDEQ(namespaceUID),
		alertevent.StatusEQ(alertevent.StatusFIRING),
		alertevent.SuppressedEQ(true),
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

func toEvent(row *ent.AlertEvent) bo.AlertEvent {
	return bo.AlertEvent{
		UID: row.UID, NamespaceUID: row.NamespaceUID, RuleUID: row.RuleUID, DatasourceUID: row.DatasourceUID,
		Fingerprint: row.Fingerprint, Status: string(row.Status), Expr: row.Expr, Labels: row.Labels,
		Summary: row.Summary, Description: row.Description, SampleValue: row.SampleValue, LevelUID: row.LevelUID,
		StartsAt: row.StartsAt, EndsAt: row.EndsAt, EvalSlot: row.EvalSlot, Suppressed: row.Suppressed,
		IntervenedBy: row.IntervenedBy, IntervenedAt: row.IntervenedAt, CreatedAt: row.CreatedAt,
	}
}

func (r *alertEventRepo) namedEvents(ctx context.Context, rows []*ent.AlertEvent) ([]bo.AlertEvent, error) {
	items := make([]bo.AlertEvent, 0, len(rows))
	seen := map[string]struct{}{}
	uids := make([]string, 0)
	for _, row := range rows {
		items = append(items, toEvent(row))
		if row.IntervenedBy == "" {
			continue
		}
		if _, ok := seen[row.IntervenedBy]; ok {
			continue
		}
		seen[row.IntervenedBy] = struct{}{}
		uids = append(uids, row.IntervenedBy)
	}
	if len(uids) == 0 {
		return items, nil
	}
	users, err := r.client.User.Query().Where(user.UIDIn(uids...)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	names := make(map[string]string, len(users))
	for _, account := range users {
		names[account.UID] = accountName(account)
	}
	for i := range items {
		items[i].IntervenedName = names[items[i].IntervenedBy]
	}
	return items, nil
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
			Summary: row.Summary, Description: row.Description, Labels: row.Labels, Fingerprint: row.Fingerprint,
		}); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return r.fail(err)
	}
	return nil
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
