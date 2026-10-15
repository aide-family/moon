package data

import (
	"context"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertevalbatch"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertevent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alerteventfiring"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertrule"
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

	recorded, err := tx.AlertEvalBatch.Query().Where(
		alertevalbatch.RuleUIDEQ(report.RuleUID),
		alertevalbatch.EvalSlotEQ(report.EvalSlot),
	).Exist(ctx)
	if err != nil {
		return r.fail(err)
	}
	if recorded {
		return nil
	}
	if !report.Skipped {
		if err := r.applyHits(ctx, tx, rule, report); err != nil {
			return err
		}
	}
	if err := tx.AlertEvalBatch.Create().SetRuleUID(report.RuleUID).SetEvalSlot(report.EvalSlot).Exec(ctx); err != nil {
		if ent.IsConstraintError(err) {
			return nil
		}
		return r.fail(err)
	}
	if _, err := tx.AlertRule.Update().Where(
		alertrule.UIDEQ(rule.UID),
		alertrule.ScheduledSlotEQ(report.EvalSlot),
	).ClearScheduledSlot().Save(ctx); err != nil {
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
			SetDatasourceUID(rule.DatasourceUID).
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
	}
	for fingerprint, row := range current {
		if _, ok := hits[fingerprint]; ok {
			continue
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
	}
	return nil
}

func (r *alertEventRepo) HasBatch(ctx context.Context, ruleUID string, slot time.Time) (bool, error) {
	ok, err := r.client.AlertEvalBatch.Query().Where(
		alertevalbatch.RuleUIDEQ(ruleUID),
		alertevalbatch.EvalSlotEQ(slot),
	).Exist(ctx)
	if err != nil {
		return false, r.fail(err)
	}
	return ok, nil
}

func (r *alertEventRepo) ListFiringFingerprints(ctx context.Context, namespaceUID, ruleUID string) ([]string, error) {
	rows, err := r.client.AlertEventFiring.Query().Where(
		alerteventfiring.NamespaceUIDEQ(namespaceUID),
		alerteventfiring.RuleUIDEQ(ruleUID),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]string, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.Fingerprint)
	}
	return items, nil
}

func (r *alertEventRepo) ListEvents(ctx context.Context, namespaceUID, view, name string, page, pageSize int) ([]bo.AlertEvent, int, error) {
	where := eventWhere(namespaceUID, view, name)
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

func eventWhere(namespaceUID, view, name string) []predicate.AlertEvent {
	where := []predicate.AlertEvent{alertevent.NamespaceUIDEQ(namespaceUID)}
	if view == "HISTORY" {
		where = append(where, alertevent.Or(
			alertevent.StatusNEQ(alertevent.StatusFIRING),
			alertevent.SuppressedEQ(true),
		))
	} else {
		where = append(where, alertevent.StatusEQ(alertevent.StatusFIRING), alertevent.SuppressedEQ(false))
	}
	if name != "" {
		where = append(where, alertevent.Or(
			alertevent.SummaryContainsFold(name),
			alertevent.FingerprintContainsFold(name),
		))
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
