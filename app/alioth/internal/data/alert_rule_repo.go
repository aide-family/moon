package data

import (
	"context"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertlevel"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertrule"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertrulegroup"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertrulelevel"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/alertrulenotificationgroup"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/id"
	"github.com/aide-family/Alioth/pkg/merr"
)

func (r *alertRuleRepo) CreateAlertRuleGroup(ctx context.Context, in bo.AlertRuleGroup) (*bo.AlertRuleGroup, error) {
	row, err := r.client.AlertRuleGroup.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetName(in.Name).
		SetStatus(alertrulegroup.StatusACTIVE).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toGroup(row)
	return &item, nil
}

func (r *alertRuleRepo) UpdateAlertRuleGroup(ctx context.Context, in bo.AlertRuleGroup) (*bo.AlertRuleGroup, error) {
	n, err := r.client.AlertRuleGroup.Update().Where(
		alertrulegroup.NamespaceUIDEQ(in.NamespaceUID),
		alertrulegroup.UIDEQ(in.UID),
		alertrulegroup.DeletedAtIsNil(),
	).SetName(in.Name).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	return r.GetAlertRuleGroup(ctx, in.NamespaceUID, in.UID)
}

func (r *alertRuleRepo) UpdateAlertRuleGroupStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.AlertRuleGroup, error) {
	value, err := groupStatus(status)
	if err != nil {
		return nil, err
	}
	n, err := r.client.AlertRuleGroup.Update().Where(
		alertrulegroup.NamespaceUIDEQ(namespaceUID),
		alertrulegroup.UIDEQ(uid),
		alertrulegroup.DeletedAtIsNil(),
	).SetStatus(value).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	return r.GetAlertRuleGroup(ctx, namespaceUID, uid)
}

func (r *alertRuleRepo) DeleteAlertRuleGroup(ctx context.Context, namespaceUID, uid string) error {
	used, err := r.client.AlertRule.Query().Where(
		alertrule.NamespaceUIDEQ(namespaceUID),
		alertrule.GroupUIDEQ(uid),
		alertrule.DeletedAtIsNil(),
	).Exist(ctx)
	if err != nil {
		return r.fail(err)
	}
	if used {
		return merr.Conflict("group still has rules")
	}
	n, err := r.client.AlertRuleGroup.Update().Where(
		alertrulegroup.NamespaceUIDEQ(namespaceUID),
		alertrulegroup.UIDEQ(uid),
		alertrulegroup.DeletedAtIsNil(),
	).SetDeletedAt(time.Now()).Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("not found")
	}
	return nil
}

func (r *alertRuleRepo) GetAlertRuleGroup(ctx context.Context, namespaceUID, uid string) (*bo.AlertRuleGroup, error) {
	row, err := r.client.AlertRuleGroup.Query().Where(
		alertrulegroup.NamespaceUIDEQ(namespaceUID),
		alertrulegroup.UIDEQ(uid),
		alertrulegroup.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toGroup(row)
	return &item, nil
}

func (r *alertRuleRepo) ListAlertRuleGroups(ctx context.Context, namespaceUID string, page, pageSize int, name string) ([]bo.AlertRuleGroup, int, error) {
	query := r.client.AlertRuleGroup.Query().Where(alertrulegroup.NamespaceUIDEQ(namespaceUID), alertrulegroup.DeletedAtIsNil())
	if name != "" {
		query = query.Where(alertrulegroup.NameContainsFold(name))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Asc(alertrulegroup.FieldUID)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.AlertRuleGroup, 0, len(rows))
	for _, row := range rows {
		items = append(items, toGroup(row))
	}
	return items, total, nil
}

func (r *alertRuleRepo) CreateAlertRule(ctx context.Context, in bo.AlertRule) (*bo.AlertRule, error) {
	kind, driver, err := ruleEnums(in.Type, in.Driver)
	if err != nil {
		return nil, err
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	row, err := tx.AlertRule.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetGroupUID(in.GroupUID).
		SetName(in.Name).
		SetType(kind).
		SetDriver(driver).
		SetStatus(alertrule.StatusACTIVE).
		SetExpr(in.Expr).
		SetLabels(in.Labels).
		SetDatasourceUID(in.DatasourceUID).
		SetSummary(in.Summary).
		SetDescription(in.Description).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if err := insertLevels(ctx, tx, in.NamespaceUID, row.UID, in.Levels); err != nil {
		if merr.IsInvalid(err) {
			return nil, err
		}
		return nil, r.fail(err)
	}
	if err := replaceRuleGroups(ctx, tx, in.NamespaceUID, row.UID, in.NotificationGroupUIDs); err != nil {
		return nil, r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetAlertRule(ctx, in.NamespaceUID, row.UID)
}

func (r *alertRuleRepo) UpdateAlertRule(ctx context.Context, in bo.AlertRule) (*bo.AlertRule, error) {
	kind, driver, err := ruleEnums(in.Type, in.Driver)
	if err != nil {
		return nil, err
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	n, err := tx.AlertRule.Update().Where(
		alertrule.NamespaceUIDEQ(in.NamespaceUID),
		alertrule.UIDEQ(in.UID),
		alertrule.DeletedAtIsNil(),
	).
		SetGroupUID(in.GroupUID).
		SetName(in.Name).
		SetType(kind).
		SetDriver(driver).
		SetExpr(in.Expr).
		SetLabels(in.Labels).
		SetDatasourceUID(in.DatasourceUID).
		SetSummary(in.Summary).
		SetDescription(in.Description).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	if _, err := tx.AlertRuleLevel.Update().Where(
		alertrulelevel.NamespaceUIDEQ(in.NamespaceUID),
		alertrulelevel.RuleUIDEQ(in.UID),
		alertrulelevel.DeletedAtIsNil(),
	).SetDeletedAt(time.Now()).Save(ctx); err != nil {
		return nil, r.fail(err)
	}
	if err := insertLevels(ctx, tx, in.NamespaceUID, in.UID, in.Levels); err != nil {
		if merr.IsInvalid(err) {
			return nil, err
		}
		return nil, r.fail(err)
	}
	if err := replaceRuleGroups(ctx, tx, in.NamespaceUID, in.UID, in.NotificationGroupUIDs); err != nil {
		return nil, r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetAlertRule(ctx, in.NamespaceUID, in.UID)
}

func (r *alertRuleRepo) UpdateAlertRuleStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.AlertRule, error) {
	value := alertrule.StatusACTIVE
	if status == "DISABLED" {
		value = alertrule.StatusDISABLED
	} else if status != "ACTIVE" {
		return nil, merr.Invalid("invalid request")
	}
	n, err := r.client.AlertRule.Update().Where(
		alertrule.NamespaceUIDEQ(namespaceUID),
		alertrule.UIDEQ(uid),
		alertrule.DeletedAtIsNil(),
	).SetStatus(value).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	return r.GetAlertRule(ctx, namespaceUID, uid)
}

func (r *alertRuleRepo) DeleteAlertRule(ctx context.Context, namespaceUID, uid string) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.AlertRuleLevel.Update().Where(
		alertrulelevel.NamespaceUIDEQ(namespaceUID),
		alertrulelevel.RuleUIDEQ(uid),
		alertrulelevel.DeletedAtIsNil(),
	).SetDeletedAt(time.Now()).Save(ctx); err != nil {
		return r.fail(err)
	}
	n, err := tx.AlertRule.Update().Where(
		alertrule.NamespaceUIDEQ(namespaceUID),
		alertrule.UIDEQ(uid),
		alertrule.DeletedAtIsNil(),
	).SetDeletedAt(time.Now()).Save(ctx)
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

func (r *alertRuleRepo) GetAlertRule(ctx context.Context, namespaceUID, uid string) (*bo.AlertRule, error) {
	row, err := r.client.AlertRule.Query().Where(
		alertrule.NamespaceUIDEQ(namespaceUID),
		alertrule.UIDEQ(uid),
		alertrule.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items, err := r.withLevels(ctx, namespaceUID, []*ent.AlertRule{row})
	if err != nil {
		return nil, err
	}
	return &items[0], nil
}

func (r *alertRuleRepo) ListAlertRules(ctx context.Context, namespaceUID string, page, pageSize int, groupUID, name string) ([]bo.AlertRule, int, error) {
	query := r.client.AlertRule.Query().Where(alertrule.NamespaceUIDEQ(namespaceUID), alertrule.DeletedAtIsNil())
	if groupUID != "" {
		query = query.Where(alertrule.GroupUIDEQ(groupUID))
	}
	if name != "" {
		query = query.Where(alertrule.NameContainsFold(name))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Asc(alertrule.FieldUID)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items, err := r.withLevels(ctx, namespaceUID, rows)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *alertRuleRepo) ListEvalCandidates(ctx context.Context, now time.Time) ([]bo.AlertRule, error) {
	rows, err := r.dueRules(ctx, now)
	if err != nil {
		return nil, err
	}
	return r.withLevels(ctx, "", rows)
}

func (r *alertRuleRepo) ScheduleDue(ctx context.Context, now, next time.Time) error {
	_, err := r.client.AlertRule.Update().Where(
		alertrule.DeletedAtIsNil(),
		alertrule.StatusEQ(alertrule.StatusACTIVE),
		alertrule.Or(alertrule.NextEvalAtIsNil(), alertrule.NextEvalAtLTE(now)),
	).SetNextEvalAt(next).SetScheduledSlot(now).Save(ctx)
	return r.fail(err)
}

func (r *alertRuleRepo) ListScheduled(ctx context.Context) ([]bo.AlertRule, error) {
	rows, err := r.client.AlertRule.Query().Where(
		alertrule.DeletedAtIsNil(),
		alertrule.StatusEQ(alertrule.StatusACTIVE),
		alertrule.ScheduledSlotNotNil(),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return r.withLevels(ctx, "", rows)
}

func (r *alertRuleRepo) GetAlertRuleByUID(ctx context.Context, uid string) (*bo.AlertRule, error) {
	row, err := r.client.AlertRule.Query().Where(alertrule.UIDEQ(uid), alertrule.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items, err := r.withLevels(ctx, row.NamespaceUID, []*ent.AlertRule{row})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, merr.NotFound("not found")
	}
	return &items[0], nil
}

func (r *alertRuleRepo) dueRules(ctx context.Context, now time.Time) ([]*ent.AlertRule, error) {
	rows, err := r.client.AlertRule.Query().Where(
		alertrule.DeletedAtIsNil(),
		alertrule.StatusEQ(alertrule.StatusACTIVE),
		alertrule.Or(alertrule.NextEvalAtIsNil(), alertrule.NextEvalAtLTE(now)),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return rows, nil
}

func insertLevels(ctx context.Context, tx *ent.Tx, namespaceUID, ruleUID string, levels []bo.AlertRuleLevel) error {
	for _, level := range levels {
		condition, sample, err := levelEnums(level.Condition, level.Sample)
		if err != nil {
			return err
		}
		if _, err := tx.AlertRuleLevel.Create().
			SetUID(id.New()).
			SetNamespaceUID(namespaceUID).
			SetRuleUID(ruleUID).
			SetCondition(condition).
			SetSample(sample).
			SetThreshold(level.Threshold).
			SetForSeconds(level.ForSeconds).
			SetLevelUID(level.LevelUID).
			Save(ctx); err != nil {
			return err
		}
	}
	return nil
}

func replaceRuleGroups(ctx context.Context, tx *ent.Tx, namespaceUID, ruleUID string, groupUIDs []string) error {
	if _, err := tx.AlertRuleNotificationGroup.Update().Where(
		alertrulenotificationgroup.NamespaceUIDEQ(namespaceUID),
		alertrulenotificationgroup.RuleUIDEQ(ruleUID),
		alertrulenotificationgroup.DeletedAtIsNil(),
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
		if err := tx.AlertRuleNotificationGroup.Create().
			SetUID(id.New()).
			SetNamespaceUID(namespaceUID).
			SetRuleUID(ruleUID).
			SetGroupUID(groupUID).
			Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (r *alertRuleRepo) withLevels(ctx context.Context, namespaceUID string, rows []*ent.AlertRule) ([]bo.AlertRule, error) {
	if len(rows) == 0 {
		return []bo.AlertRule{}, nil
	}
	uids := make([]string, len(rows))
	for i, row := range rows {
		uids[i] = row.UID
	}
	query := r.client.AlertRuleLevel.Query().Where(
		alertrulelevel.RuleUIDIn(uids...),
		alertrulelevel.DeletedAtIsNil(),
	)
	if namespaceUID != "" {
		query = query.Where(alertrulelevel.NamespaceUIDEQ(namespaceUID))
	}
	links, err := query.All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	levelUIDs := make([]string, 0, len(links))
	for _, link := range links {
		levelUIDs = append(levelUIDs, link.LevelUID)
	}
	names := map[string]string{}
	if len(levelUIDs) > 0 {
		levels, err := r.client.AlertLevel.Query().Where(alertlevel.UIDIn(levelUIDs...), alertlevel.DeletedAtIsNil()).All(ctx)
		if err != nil {
			return nil, r.fail(err)
		}
		for _, level := range levels {
			names[level.UID] = level.Name
		}
	}
	byRule := map[string][]bo.AlertRuleLevel{}
	for _, link := range links {
		byRule[link.RuleUID] = append(byRule[link.RuleUID], bo.AlertRuleLevel{
			UID: link.UID, Condition: link.Condition, Sample: string(link.Sample),
			Threshold: link.Threshold, ForSeconds: link.ForSeconds, LevelUID: link.LevelUID, LevelName: names[link.LevelUID],
		})
	}
	groupQuery := r.client.AlertRuleNotificationGroup.Query().Where(
		alertrulenotificationgroup.RuleUIDIn(uids...),
		alertrulenotificationgroup.DeletedAtIsNil(),
	)
	if namespaceUID != "" {
		groupQuery = groupQuery.Where(alertrulenotificationgroup.NamespaceUIDEQ(namespaceUID))
	}
	groupLinks, err := groupQuery.All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	groupsByRule := map[string][]string{}
	for _, link := range groupLinks {
		groupsByRule[link.RuleUID] = append(groupsByRule[link.RuleUID], link.GroupUID)
	}
	items := make([]bo.AlertRule, 0, len(rows))
	for _, row := range rows {
		items = append(items, toRule(row, byRule[row.UID], groupsByRule[row.UID]))
	}
	return items, nil
}

func toGroup(row *ent.AlertRuleGroup) bo.AlertRuleGroup {
	return bo.AlertRuleGroup{UID: row.UID, NamespaceUID: row.NamespaceUID, Name: row.Name, Status: string(row.Status)}
}

func toRule(row *ent.AlertRule, levels []bo.AlertRuleLevel, groupUIDs []string) bo.AlertRule {
	if groupUIDs == nil {
		groupUIDs = []string{}
	}
	return bo.AlertRule{
		UID: row.UID, NamespaceUID: row.NamespaceUID, GroupUID: row.GroupUID, Name: row.Name,
		Type: string(row.Type), Driver: string(row.Driver), Status: string(row.Status),
		Expr: row.Expr, Labels: row.Labels, DatasourceUID: row.DatasourceUID,
		Summary: row.Summary, Description: row.Description, NextEvalAt: row.NextEvalAt,
		ScheduledSlot: row.ScheduledSlot, Levels: levels, NotificationGroupUIDs: groupUIDs,
	}
}

func groupStatus(status string) (alertrulegroup.Status, error) {
	switch alertrulegroup.Status(status) {
	case alertrulegroup.StatusACTIVE, alertrulegroup.StatusDISABLED:
		return alertrulegroup.Status(status), nil
	default:
		return "", merr.Invalid("invalid request")
	}
}

func ruleEnums(kind, driver string) (alertrule.Type, alertrule.Driver, error) {
	parsedType := alertrule.Type(kind)
	switch parsedType {
	case alertrule.TypeMETRICS, alertrule.TypeLOGS, alertrule.TypeTRACE:
	default:
		return "", "", merr.Invalid("invalid request")
	}
	parsedDriver := alertrule.Driver(driver)
	switch parsedDriver {
	case alertrule.DriverMETRICS_PROMETHEUS, alertrule.DriverMETRICS_VICTORIA_METRICS, alertrule.DriverLOGS_ELASTICSEARCH, alertrule.DriverTRACE_JAEGER:
	default:
		return "", "", merr.Invalid("invalid request")
	}
	return parsedType, parsedDriver, nil
}

func levelEnums(condition, sample string) (string, alertrulelevel.Sample, error) {
	switch condition {
	case "EQ", "NE", "GT", "GTE", "LT", "LTE", "IN", "NOT_IN":
	default:
		return "", "", merr.Invalid("invalid request")
	}
	parsedSample := alertrulelevel.Sample(sample)
	switch parsedSample {
	case alertrulelevel.SampleFOR, alertrulelevel.SampleMAX, alertrulelevel.SampleMIN:
	default:
		return "", "", merr.Invalid("invalid request")
	}
	return condition, parsedSample, nil
}
