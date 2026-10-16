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
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertrulenotificationgroup"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationgroup"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/notificationtask"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/predicate"
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
	if !report.Skipped {
		if err := r.applyHits(ctx, tx, rule, report); err != nil {
			return err
		}
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
		if err := createNotifyTasks(ctx, tx, rule, eventUID, notificationtask.PhaseFIRING, textOr(hit.Summary, rule.Summary), textOr(hit.Description, rule.Description), hit.Labels, fingerprint); err != nil {
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
		if err := createNotifyTasks(ctx, tx, rule, row.EventUID, notificationtask.PhaseRECOVERED, event.Summary, event.Description, event.Labels, fingerprint); err != nil {
			return err
		}
	}
	return nil
}

func createNotifyTasks(ctx context.Context, tx *ent.Tx, rule bo.AlertRule, eventUID string, phase notificationtask.Phase, summary, description, labels, fingerprint string) error {
	links, err := tx.AlertRuleNotificationGroup.Query().Where(
		alertrulenotificationgroup.RuleUIDEQ(rule.UID),
		alertrulenotificationgroup.NamespaceUIDEQ(rule.NamespaceUID),
		alertrulenotificationgroup.DeletedAtIsNil(),
	).All(ctx)
	if err != nil {
		return err
	}
	if len(links) == 0 {
		return nil
	}
	groupUIDs := make([]string, 0, len(links))
	for _, link := range links {
		groupUIDs = append(groupUIDs, link.GroupUID)
	}
	groups, err := tx.NotificationGroup.Query().Where(
		notificationgroup.UIDIn(groupUIDs...),
		notificationgroup.NamespaceUIDEQ(rule.NamespaceUID),
		notificationgroup.StatusEQ(notificationgroup.StatusACTIVE),
		notificationgroup.DeletedAtIsNil(),
	).All(ctx)
	if err != nil {
		return err
	}
	params := paramsJSON(eventUID, string(phase), summary, description, labels, fingerprint)
	for _, group := range groups {
		err := tx.NotificationTask.Create().
			SetUID(id.New()).
			SetNamespaceUID(rule.NamespaceUID).
			SetEventUID(eventUID).
			SetPhase(phase).
			SetStatus(notificationtask.StatusPENDING).
			SetRuleUID(rule.UID).
			SetGroupUID(group.UID).
			SetChannelUID(group.ChannelUID).
			SetTemplateUID(group.TemplateUID).
			SetReceivers(group.Receivers).
			SetParams(params).
			SetRetriesLeft(3).
			Exec(ctx)
		if err != nil {
			if ent.IsConstraintError(err) {
				continue
			}
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
	items := make([]bo.AlertEvent, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEvent(row))
	}
	return items, total, nil
}

func (r *alertEventRepo) ListRuleUIDsForReceivers(ctx context.Context, namespaceUID string, tokens []string) ([]string, error) {
	if len(tokens) == 0 {
		return nil, nil
	}
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
		if receiversContain(group.Receivers, tokens) {
			matched = append(matched, group.UID)
		}
	}
	if len(matched) == 0 {
		return nil, nil
	}
	links, err := r.client.AlertRuleNotificationGroup.Query().Where(
		alertrulenotificationgroup.NamespaceUIDEQ(namespaceUID),
		alertrulenotificationgroup.DeletedAtIsNil(),
		alertrulenotificationgroup.GroupUIDIn(matched...),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	seen := make(map[string]struct{}, len(links))
	uids := make([]string, 0, len(links))
	for _, link := range links {
		if _, ok := seen[link.RuleUID]; ok {
			continue
		}
		seen[link.RuleUID] = struct{}{}
		uids = append(uids, link.RuleUID)
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
	item := toEvent(row)
	return &item, nil
}

func (r *alertEventRepo) Intervene(ctx context.Context, namespaceUID string, uids []string, userUID string) error {
	if len(uids) == 0 {
		return nil
	}
	_, err := r.client.AlertEvent.Update().Where(
		alertevent.NamespaceUIDEQ(namespaceUID),
		alertevent.UIDIn(uids...),
		alertevent.StatusEQ(alertevent.StatusFIRING),
	).SetIntervenedBy(userUID).Save(ctx)
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
		IntervenedBy: row.IntervenedBy, CreatedAt: row.CreatedAt,
	}
}

func textOr(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
