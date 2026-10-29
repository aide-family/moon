package data

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertdictionary"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertrule"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertrulegroup"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertrulelevel"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/pkg/merr"
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
	driver, err := ruleDriver(in.Driver)
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
		SetDriver(driver).
		SetStatus(alertrule.StatusACTIVE).
		SetExpr(in.Expr).
		SetLabels(in.Labels).
		SetDatasourceUID("").
		SetDatasourceLabels(in.DatasourceLabels).
		SetSummary(in.Summary).
		SetDescription(in.Description).
		SetVariables(encodeRuleVariables(in.Variables)).
		SetProbeGroupLabels(in.ProbeGroupLabels).
		SetProbeIntervalSeconds(in.ProbeIntervalSeconds).
		SetProbeMetric(probeMetricValue(in.ProbeMetric)).
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
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetAlertRule(ctx, in.NamespaceUID, row.UID)
}

func (r *alertRuleRepo) UpdateAlertRule(ctx context.Context, in bo.AlertRule) (*bo.AlertRule, error) {
	driver, err := ruleDriver(in.Driver)
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
		SetDriver(driver).
		SetExpr(in.Expr).
		SetLabels(in.Labels).
		SetDatasourceUID("").
		SetDatasourceLabels(in.DatasourceLabels).
		SetSummary(in.Summary).
		SetDescription(in.Description).
		SetVariables(encodeRuleVariables(in.Variables)).
		SetProbeGroupLabels(in.ProbeGroupLabels).
		SetProbeIntervalSeconds(in.ProbeIntervalSeconds).
		SetProbeMetric(probeMetricValue(in.ProbeMetric)).
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
	if value == alertrule.StatusDISABLED {
		if err := recoverRuleFiring(ctx, r.client, r.history, r.log, namespaceUID, uid); err != nil {
			return nil, r.fail(err)
		}
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
	if err := recoverRuleFiring(ctx, r.client, r.history, r.log, namespaceUID, uid); err != nil {
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
	groups, err := r.activeGroupUIDs(ctx)
	if err != nil || len(groups) == 0 {
		return err
	}
	_, err = r.client.AlertRule.Update().Where(
		alertrule.DeletedAtIsNil(),
		alertrule.StatusEQ(alertrule.StatusACTIVE),
		alertrule.GroupUIDIn(groups...),
		alertrule.DriverNEQ(alertrule.DriverPROBE_HTTP),
		alertrule.Or(alertrule.NextEvalAtIsNil(), alertrule.NextEvalAtLTE(now)),
	).SetNextEvalAt(next).SetScheduledSlot(now).Save(ctx)
	return r.fail(err)
}

// ListProbeRulesDue 返回到期的探测规则。调度按每条规则自己的周期推进，
// 不能复用查询规则的 15 秒全局批次。
func (r *alertRuleRepo) ListProbeRulesDue(ctx context.Context, now time.Time) ([]bo.AlertRule, error) {
	groups, err := r.activeGroupUIDs(ctx)
	if err != nil || len(groups) == 0 {
		return nil, err
	}
	rows, err := r.client.AlertRule.Query().Where(
		alertrule.DeletedAtIsNil(),
		alertrule.StatusEQ(alertrule.StatusACTIVE),
		alertrule.GroupUIDIn(groups...),
		alertrule.DriverEQ(alertrule.DriverPROBE_HTTP),
		alertrule.Or(alertrule.NextEvalAtIsNil(), alertrule.NextEvalAtLTE(now)),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return r.withLevels(ctx, "", rows)
}

// SetProbeSchedule 写入本轮探测槽位与下一次探测时间。
func (r *alertRuleRepo) SetProbeSchedule(ctx context.Context, uid string, slot, next time.Time) error {
	_, err := r.client.AlertRule.Update().Where(
		alertrule.UIDEQ(uid),
		alertrule.DeletedAtIsNil(),
	).SetScheduledSlot(slot).SetNextEvalAt(next).Save(ctx)
	return r.fail(err)
}

// ListProbeRulesForClaim 返回当前窗口内可领取的探测规则：已写入本轮槽位，
// 且下一次探测时间还没到。执行器按 (target, slot) 幂等，重复领取不会重复探测。
func (r *alertRuleRepo) ListProbeRulesForClaim(ctx context.Context, now time.Time) ([]bo.AlertRule, error) {
	groups, err := r.activeGroupUIDs(ctx)
	if err != nil || len(groups) == 0 {
		return nil, err
	}
	rows, err := r.client.AlertRule.Query().Where(
		alertrule.DeletedAtIsNil(),
		alertrule.StatusEQ(alertrule.StatusACTIVE),
		alertrule.GroupUIDIn(groups...),
		alertrule.DriverEQ(alertrule.DriverPROBE_HTTP),
		alertrule.ScheduledSlotNotNil(),
		alertrule.NextEvalAtGT(now),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return r.withLevels(ctx, "", rows)
}

func (r *alertRuleRepo) ListRuleUIDs(ctx context.Context, namespaceUID, groupUID string) ([]string, error) {
	query := r.client.AlertRule.Query().Where(alertrule.NamespaceUIDEQ(namespaceUID), alertrule.DeletedAtIsNil())
	if groupUID != "" {
		query = query.Where(alertrule.GroupUIDEQ(groupUID))
	}
	rows, err := query.All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	uids := make([]string, 0, len(rows))
	for _, row := range rows {
		uids = append(uids, row.UID)
	}
	return uids, nil
}

func (r *alertRuleRepo) ListScheduled(ctx context.Context) ([]bo.AlertRule, error) {
	groups, err := r.activeGroupUIDs(ctx)
	if err != nil || len(groups) == 0 {
		return nil, err
	}
	rows, err := r.client.AlertRule.Query().Where(
		alertrule.DeletedAtIsNil(),
		alertrule.StatusEQ(alertrule.StatusACTIVE),
		alertrule.GroupUIDIn(groups...),
		alertrule.DriverNEQ(alertrule.DriverPROBE_HTTP),
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

func (r *alertRuleRepo) ClearScheduledSlot(ctx context.Context, ruleUID string, slot time.Time) error {
	_, err := r.client.AlertRule.Update().Where(
		alertrule.UIDEQ(ruleUID),
		alertrule.ScheduledSlotEQ(slot),
	).ClearScheduledSlot().Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	return nil
}

func (r *alertRuleRepo) activeGroupUIDs(ctx context.Context) ([]string, error) {
	rows, err := r.client.AlertRuleGroup.Query().Where(
		alertrulegroup.DeletedAtIsNil(),
		alertrulegroup.StatusEQ(alertrulegroup.StatusACTIVE),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	uids := make([]string, 0, len(rows))
	for _, row := range rows {
		uids = append(uids, row.UID)
	}
	return uids, nil
}

func (r *alertRuleRepo) dueRules(ctx context.Context, now time.Time) ([]*ent.AlertRule, error) {
	groups, err := r.activeGroupUIDs(ctx)
	if err != nil || len(groups) == 0 {
		return nil, err
	}
	rows, err := r.client.AlertRule.Query().Where(
		alertrule.DeletedAtIsNil(),
		alertrule.StatusEQ(alertrule.StatusACTIVE),
		alertrule.GroupUIDIn(groups...),
		alertrule.DriverNEQ(alertrule.DriverPROBE_HTTP),
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
		levels, err := r.client.AlertDictionary.Query().Where(alertdictionary.UIDIn(levelUIDs...), alertdictionary.DeletedAtIsNil()).All(ctx)
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
	items := make([]bo.AlertRule, 0, len(rows))
	for _, row := range rows {
		items = append(items, toRule(row, byRule[row.UID]))
	}
	return items, nil
}

func toGroup(row *ent.AlertRuleGroup) bo.AlertRuleGroup {
	return bo.AlertRuleGroup{UID: row.UID, NamespaceUID: row.NamespaceUID, Name: row.Name, Status: string(row.Status)}
}

// probeMetricValue 把观测指标归一成合法枚举。历史调用方与测试可能不传这个字段，
// 空值按 UP 处理，避免枚举校验把查询规则的写入挡掉。
func probeMetricValue(metric string) alertrule.ProbeMetric {
	switch alertrule.ProbeMetric(metric) {
	case alertrule.ProbeMetricUP, alertrule.ProbeMetricLATENCY_MS,
		alertrule.ProbeMetricSTATUS_CODE, alertrule.ProbeMetricCERT_DAYS:
		return alertrule.ProbeMetric(metric)
	default:
		return alertrule.ProbeMetricUP
	}
}

func toRule(row *ent.AlertRule, levels []bo.AlertRuleLevel) bo.AlertRule {
	return bo.AlertRule{
		UID: row.UID, NamespaceUID: row.NamespaceUID, GroupUID: row.GroupUID, Name: row.Name,
		Driver: string(row.Driver), Status: string(row.Status),
		Expr: row.Expr, Labels: row.Labels, DatasourceUID: row.DatasourceUID, DatasourceLabels: row.DatasourceLabels,
		Summary: row.Summary, Description: row.Description, NextEvalAt: row.NextEvalAt,
		ScheduledSlot: row.ScheduledSlot, Levels: levels, Variables: decodeRuleVariables(row.Variables),
		ProbeGroupLabels: row.ProbeGroupLabels, ProbeIntervalSeconds: row.ProbeIntervalSeconds,
		ProbeMetric: string(row.ProbeMetric),
	}
}

// ruleVariableJSON 是自定义查询变量的存储格式。键用小写，与设计文档约定的
// [{"name":"disk_free","expr":"...","description":"..."}] 一致。
type ruleVariableJSON struct {
	Name        string `json:"name"`
	Expr        string `json:"expr"`
	Description string `json:"description"`
}

// encodeRuleVariables 把变量列表编码成 JSON 文本。空列表写成 []，避免列值
// 为空字符串后 JSON 解析失败。字段都是字符串，编码不会失败，仍兜底返回 []
// 让写入路径不因为这一列报错。
func encodeRuleVariables(items []bo.AlertRuleVariable) string {
	rows := make([]ruleVariableJSON, 0, len(items))
	for _, item := range items {
		rows = append(rows, ruleVariableJSON{Name: item.Name, Expr: item.Expr, Description: item.Description})
	}
	encoded, err := json.Marshal(rows)
	if err != nil {
		return "[]"
	}
	return string(encoded)
}

// decodeRuleVariables 解析存储的变量 JSON。历史规则或损坏的配置都按没有变量
// 处理，读取规则不能因为这一列失败。
func decodeRuleVariables(raw string) []bo.AlertRuleVariable {
	rows := make([]ruleVariableJSON, 0)
	if err := json.Unmarshal([]byte(raw), &rows); err != nil {
		return []bo.AlertRuleVariable{}
	}
	items := make([]bo.AlertRuleVariable, 0, len(rows))
	for _, row := range rows {
		items = append(items, bo.AlertRuleVariable{Name: row.Name, Expr: row.Expr, Description: row.Description})
	}
	return items
}

func groupStatus(status string) (alertrulegroup.Status, error) {
	switch alertrulegroup.Status(status) {
	case alertrulegroup.StatusACTIVE, alertrulegroup.StatusDISABLED:
		return alertrulegroup.Status(status), nil
	default:
		return "", merr.Invalid("invalid request")
	}
}

func ruleDriver(driver string) (alertrule.Driver, error) {
	parsedDriver := alertrule.Driver(driver)
	switch parsedDriver {
	case alertrule.DriverMETRICS_PROMETHEUS, alertrule.DriverMETRICS_VICTORIA_METRICS, alertrule.DriverLOGS_ELASTICSEARCH, alertrule.DriverLOGS_LOKI, alertrule.DriverLOGS_VLOGS, alertrule.DriverTRACE_JAEGER:
		return parsedDriver, nil
	default:
		return "", merr.Invalid("invalid request")
	}
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
