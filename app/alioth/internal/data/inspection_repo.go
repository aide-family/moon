package data

import (
	"context"
	stdSQL "database/sql"
	"errors"
	"time"

	"entgo.io/ent/dialect/sql"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/inspectionitem"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/inspectionplan"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/inspectionrun"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/inspectiontemplate"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/pkg/merr"
)

func (r *inspectionRepo) CreateInspectionPlan(ctx context.Context, in bo.InspectionPlan) (*bo.InspectionPlan, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()

	uid := id.New()
	build := tx.InspectionPlan.Create().
		SetUID(uid).
		SetNamespaceUID(in.NamespaceUID).
		SetName(in.Name).
		SetDescription(in.Description).
		SetStatus(inspectionplan.Status(in.Status)).
		SetCron(in.Cron).
		SetTimezone(in.Timezone).
		SetTimeoutSeconds(in.TimeoutSeconds).
		SetTemplateUID(in.TemplateUID).
		SetOutputFormats(in.OutputFormats).
		SetTitleTemplate(in.TitleTemplate).
		SetNotificationGroupUids(in.NotificationGroupUIDs).
		SetNotifyPolicy(inspectionplan.NotifyPolicy(in.NotifyPolicy)).
		SetLabels(in.Labels).
		SetRemark(in.Remark).
		SetCreatedBy(in.CreatedBy)
	if in.NextRunAt != nil {
		build.SetNextRunAt(*in.NextRunAt)
	}
	if _, err := build.Save(ctx); err != nil {
		return nil, r.fail(err)
	}
	if err := r.createInspectionItems(ctx, tx, in.NamespaceUID, uid, in.Items); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetInspectionPlan(ctx, in.NamespaceUID, uid)
}

func (r *inspectionRepo) UpdateInspectionPlan(ctx context.Context, in bo.InspectionPlan) (*bo.InspectionPlan, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()

	update := tx.InspectionPlan.Update().
		Where(
			inspectionplan.UIDEQ(in.UID),
			inspectionplan.NamespaceUIDEQ(in.NamespaceUID),
			inspectionplan.DeletedAtIsNil(),
		).
		SetName(in.Name).
		SetDescription(in.Description).
		SetStatus(inspectionplan.Status(in.Status)).
		SetCron(in.Cron).
		SetTimezone(in.Timezone).
		SetTimeoutSeconds(in.TimeoutSeconds).
		SetTemplateUID(in.TemplateUID).
		SetOutputFormats(in.OutputFormats).
		SetTitleTemplate(in.TitleTemplate).
		SetNotificationGroupUids(in.NotificationGroupUIDs).
		SetNotifyPolicy(inspectionplan.NotifyPolicy(in.NotifyPolicy)).
		SetLabels(in.Labels).
		SetRemark(in.Remark)
	if in.NextRunAt != nil {
		update.SetNextRunAt(*in.NextRunAt)
	} else {
		update.ClearNextRunAt()
	}
	n, err := update.Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("inspection plan not found")
	}
	// 检查项随计划整体保存：先清空再写入，避免前端漏传导致旧项残留。
	if _, err := tx.InspectionItem.Delete().
		Where(
			inspectionitem.PlanUIDEQ(in.UID),
			inspectionitem.NamespaceUIDEQ(in.NamespaceUID),
		).Exec(ctx); err != nil {
		return nil, r.fail(err)
	}
	if err := r.createInspectionItems(ctx, tx, in.NamespaceUID, in.UID, in.Items); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetInspectionPlan(ctx, in.NamespaceUID, in.UID)
}

func (r *inspectionRepo) createInspectionItems(ctx context.Context, tx *ent.Tx, namespaceUID, planUID string, items []bo.InspectionItem) error {
	for index, item := range items {
		build := tx.InspectionItem.Create().
			SetUID(id.New()).
			SetNamespaceUID(namespaceUID).
			SetPlanUID(planUID).
			SetKind(inspectionitem.Kind(item.Kind)).
			SetTitle(item.Title).
			SetSort(index).
			SetDatasourceUID(item.DatasourceUID).
			SetQuery(item.Query).
			SetWindowSeconds(item.WindowSeconds).
			SetStepSeconds(item.StepSeconds).
			SetDisplay(inspectionitem.Display(item.Display)).
			SetParams(item.Params).
			SetEnabled(item.Enabled)
		if _, err := build.Save(ctx); err != nil {
			return r.fail(err)
		}
	}
	return nil
}

func (r *inspectionRepo) UpdateInspectionPlanStatus(ctx context.Context, namespaceUID, uid, status string, nextRunAt *time.Time) (*bo.InspectionPlan, error) {
	update := r.client.InspectionPlan.Update().
		Where(
			inspectionplan.UIDEQ(uid),
			inspectionplan.NamespaceUIDEQ(namespaceUID),
			inspectionplan.DeletedAtIsNil(),
		).
		SetStatus(inspectionplan.Status(status))
	if nextRunAt != nil {
		update.SetNextRunAt(*nextRunAt)
	} else {
		update.ClearNextRunAt()
	}
	n, err := update.Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("inspection plan not found")
	}
	return r.GetInspectionPlan(ctx, namespaceUID, uid)
}

func (r *inspectionRepo) DeleteInspectionPlan(ctx context.Context, namespaceUID, uid string) error {
	now := time.Now().UTC()
	n, err := r.client.InspectionPlan.Update().
		Where(
			inspectionplan.UIDEQ(uid),
			inspectionplan.NamespaceUIDEQ(namespaceUID),
			inspectionplan.DeletedAtIsNil(),
		).
		SetDeletedAt(now).
		Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("inspection plan not found")
	}
	if _, err := r.client.InspectionItem.Update().
		Where(
			inspectionitem.PlanUIDEQ(uid),
			inspectionitem.NamespaceUIDEQ(namespaceUID),
			inspectionitem.DeletedAtIsNil(),
		).
		SetDeletedAt(now).
		Save(ctx); err != nil {
		return r.fail(err)
	}
	return nil
}

func (r *inspectionRepo) GetInspectionPlan(ctx context.Context, namespaceUID, uid string) (*bo.InspectionPlan, error) {
	row, err := r.client.InspectionPlan.Query().Where(
		inspectionplan.UIDEQ(uid),
		inspectionplan.NamespaceUIDEQ(namespaceUID),
		inspectionplan.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items, err := r.client.InspectionItem.Query().Where(
		inspectionitem.PlanUIDEQ(uid),
		inspectionitem.NamespaceUIDEQ(namespaceUID),
		inspectionitem.DeletedAtIsNil(),
	).Order(inspectionitem.BySort()).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	plan := toInspectionPlan(row)
	plan.Items = make([]bo.InspectionItem, 0, len(items))
	for _, item := range items {
		plan.Items = append(plan.Items, toInspectionItem(item))
	}
	return &plan, nil
}

func (r *inspectionRepo) ListInspectionPlans(ctx context.Context, namespaceUID string, filter bo.InspectionPlanListFilter, page, pageSize int) ([]bo.InspectionPlan, int, error) {
	query := r.client.InspectionPlan.Query().Where(
		inspectionplan.NamespaceUIDEQ(namespaceUID),
		inspectionplan.DeletedAtIsNil(),
	)
	if filter.Name != "" {
		query = query.Where(inspectionplan.NameContainsFold(filter.Name))
	}
	if filter.Status != "" {
		query = query.Where(inspectionplan.StatusEQ(inspectionplan.Status(filter.Status)))
	}
	if filter.TemplateUID != "" {
		query = query.Where(inspectionplan.TemplateUIDEQ(filter.TemplateUID))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Desc(inspectionplan.FieldUpdatedAt)).
		Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	if len(rows) == 0 {
		return []bo.InspectionPlan{}, total, nil
	}
	planUIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		planUIDs = append(planUIDs, row.UID)
	}
	items, err := r.client.InspectionItem.Query().Where(
		inspectionitem.PlanUIDIn(planUIDs...),
		inspectionitem.NamespaceUIDEQ(namespaceUID),
		inspectionitem.DeletedAtIsNil(),
	).Order(inspectionitem.BySort()).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	byPlan := make(map[string][]bo.InspectionItem, len(rows))
	for _, item := range items {
		byPlan[item.PlanUID] = append(byPlan[item.PlanUID], toInspectionItem(item))
	}
	plans := make([]bo.InspectionPlan, 0, len(rows))
	for _, row := range rows {
		plan := toInspectionPlan(row)
		plan.Items = byPlan[row.UID]
		if plan.Items == nil {
			plan.Items = []bo.InspectionItem{}
		}
		plans = append(plans, plan)
	}
	return plans, total, nil
}

func (r *inspectionRepo) CreateInspectionTemplate(ctx context.Context, in bo.InspectionTemplate) (*bo.InspectionTemplate, error) {
	row, err := r.client.InspectionTemplate.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetName(in.Name).
		SetDescription(in.Description).
		SetHTMLBody(in.HTMLBody).
		SetMarkdownBody(in.MarkdownBody).
		SetIsBuiltin(in.IsBuiltin).
		SetStatus(inspectiontemplate.Status(in.Status)).
		SetCreatedBy(in.CreatedBy).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return r.GetInspectionTemplate(ctx, in.NamespaceUID, row.UID)
}

func (r *inspectionRepo) UpdateInspectionTemplate(ctx context.Context, in bo.InspectionTemplate) (*bo.InspectionTemplate, error) {
	n, err := r.client.InspectionTemplate.Update().
		Where(
			inspectiontemplate.UIDEQ(in.UID),
			inspectiontemplate.NamespaceUIDEQ(in.NamespaceUID),
			inspectiontemplate.IsBuiltinEQ(false),
			inspectiontemplate.DeletedAtIsNil(),
		).
		SetName(in.Name).
		SetDescription(in.Description).
		SetHTMLBody(in.HTMLBody).
		SetMarkdownBody(in.MarkdownBody).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("inspection template not found")
	}
	return r.GetInspectionTemplate(ctx, in.NamespaceUID, in.UID)
}

func (r *inspectionRepo) DeleteInspectionTemplate(ctx context.Context, namespaceUID, uid string) error {
	n, err := r.client.InspectionTemplate.Update().
		Where(
			inspectiontemplate.UIDEQ(uid),
			inspectiontemplate.NamespaceUIDEQ(namespaceUID),
			inspectiontemplate.IsBuiltinEQ(false),
			inspectiontemplate.DeletedAtIsNil(),
		).
		SetDeletedAt(time.Now().UTC()).
		Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("inspection template not found")
	}
	return nil
}

func (r *inspectionRepo) GetInspectionTemplate(ctx context.Context, namespaceUID, uid string) (*bo.InspectionTemplate, error) {
	row, err := r.client.InspectionTemplate.Query().Where(
		inspectiontemplate.UIDEQ(uid),
		inspectiontemplate.DeletedAtIsNil(),
		inspectiontemplate.Or(
			inspectiontemplate.NamespaceUIDEQ(namespaceUID),
			inspectiontemplate.NamespaceUIDEQ(""),
		),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toInspectionTemplate(row)
	return &item, nil
}

func (r *inspectionRepo) ListInspectionTemplates(ctx context.Context, namespaceUID, name string, page, pageSize int) ([]bo.InspectionTemplate, int, error) {
	query := r.client.InspectionTemplate.Query().Where(
		inspectiontemplate.DeletedAtIsNil(),
		inspectiontemplate.Or(
			inspectiontemplate.NamespaceUIDEQ(namespaceUID),
			inspectiontemplate.NamespaceUIDEQ(""),
		),
	)
	if name != "" {
		query = query.Where(inspectiontemplate.NameContainsFold(name))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(
		inspectiontemplate.ByIsBuiltin(sql.OrderDesc()),
		ent.Desc(inspectiontemplate.FieldUpdatedAt),
	).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.InspectionTemplate, 0, len(rows))
	for _, row := range rows {
		items = append(items, toInspectionTemplate(row))
	}
	return items, total, nil
}

func toInspectionPlan(row *ent.InspectionPlan) bo.InspectionPlan {
	return bo.InspectionPlan{
		UID: row.UID, NamespaceUID: row.NamespaceUID, Name: row.Name, Description: row.Description,
		Status: string(row.Status), Cron: row.Cron, Timezone: row.Timezone,
		TimeoutSeconds: row.TimeoutSeconds, TemplateUID: row.TemplateUID,
		OutputFormats: row.OutputFormats, TitleTemplate: row.TitleTemplate,
		NotificationGroupUIDs: row.NotificationGroupUids, NotifyPolicy: string(row.NotifyPolicy),
		Labels: row.Labels, Remark: row.Remark, NextRunAt: row.NextRunAt, LastRunAt: row.LastRunAt,
		CreatedBy: row.CreatedBy, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func toInspectionItem(row *ent.InspectionItem) bo.InspectionItem {
	return bo.InspectionItem{
		UID: row.UID, NamespaceUID: row.NamespaceUID, PlanUID: row.PlanUID,
		Kind: string(row.Kind), Title: row.Title, Sort: row.Sort,
		DatasourceUID: row.DatasourceUID, Query: row.Query,
		WindowSeconds: row.WindowSeconds, StepSeconds: row.StepSeconds,
		Display: string(row.Display), Params: row.Params, Enabled: row.Enabled,
	}
}

func toInspectionTemplate(row *ent.InspectionTemplate) bo.InspectionTemplate {
	return bo.InspectionTemplate{
		UID: row.UID, NamespaceUID: row.NamespaceUID, Name: row.Name, Description: row.Description,
		HTMLBody: row.HTMLBody, MarkdownBody: row.MarkdownBody,
		IsBuiltin: row.IsBuiltin, Status: string(row.Status),
		CreatedBy: row.CreatedBy, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

// DueInspectionPlans 返回已到执行时间且启用中的计划。调度器随后用
// AdvanceInspectionPlan 的乐观更新决定由哪个副本推进。
func (r *inspectionRepo) DueInspectionPlans(ctx context.Context, now time.Time, limit int) ([]bo.InspectionPlan, error) {
	rows, err := r.client.InspectionPlan.Query().Where(
		inspectionplan.StatusEQ(inspectionplan.StatusACTIVE),
		inspectionplan.DeletedAtIsNil(),
		inspectionplan.NextRunAtNotNil(),
		inspectionplan.NextRunAtLTE(now),
	).Order(inspectionplan.ByNextRunAt()).Limit(limit).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	plans := make([]bo.InspectionPlan, 0, len(rows))
	for _, row := range rows {
		plan, err := r.GetInspectionPlan(ctx, row.NamespaceUID, row.UID)
		if err != nil {
			return nil, err
		}
		plans = append(plans, *plan)
	}
	return plans, nil
}

func (r *inspectionRepo) AdvanceInspectionPlan(ctx context.Context, planUID string, expectedNext, next, lastRunAt time.Time) (bool, error) {
	n, err := r.client.InspectionPlan.Update().
		Where(
			inspectionplan.UIDEQ(planUID),
			inspectionplan.StatusEQ(inspectionplan.StatusACTIVE),
			inspectionplan.DeletedAtIsNil(),
			inspectionplan.NextRunAtEQ(expectedNext),
		).
		SetNextRunAt(next).
		SetLastRunAt(lastRunAt).
		Save(ctx)
	if err != nil {
		return false, r.fail(err)
	}
	return n > 0, nil
}

func (r *inspectionRepo) CreateInspectionRun(ctx context.Context, in bo.InspectionRun) (*bo.InspectionRun, error) {
	build := r.client.InspectionRun.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetPlanUID(in.PlanUID).
		SetPlanName(in.PlanName).
		SetTrigger(inspectionrun.Trigger(in.Trigger)).
		SetStatus(inspectionrun.StatusPENDING).
		SetScheduledAt(in.ScheduledAt).
		SetOperatorUID(in.OperatorUID).
		SetOperatorName(in.OperatorName)
	row, err := build.Save(ctx)
	if err != nil {
		// 同一计划同一计划时刻只允许一条记录，多副本重复调度时静默忽略。
		if ent.IsConstraintError(err) {
			return nil, nil
		}
		return nil, r.fail(err)
	}
	return r.GetInspectionRun(ctx, row.NamespaceUID, row.UID)
}

func (r *inspectionRepo) HasActiveInspectionRun(ctx context.Context, namespaceUID, planUID string) (bool, error) {
	ok, err := r.client.InspectionRun.Query().Where(
		inspectionrun.NamespaceUIDEQ(namespaceUID),
		inspectionrun.PlanUIDEQ(planUID),
		inspectionrun.StatusIn(inspectionrun.StatusPENDING, inspectionrun.StatusRUNNING),
	).Exist(ctx)
	if err != nil {
		return false, r.fail(err)
	}
	return ok, nil
}

func (r *inspectionRepo) ClaimInspectionRun(ctx context.Context, owner string, now, leaseUntil time.Time) (*bo.InspectionRun, error) {
	const claimSQL = `
UPDATE inspection_runs
SET status = 'RUNNING', started_at = $1, lease_owner = $2, lease_until = $3, updated_at = $1
WHERE uid = (
  SELECT uid FROM inspection_runs
  WHERE status = 'PENDING'
  ORDER BY scheduled_at ASC
  FOR UPDATE SKIP LOCKED
  LIMIT 1
)
RETURNING uid, namespace_uid`
	var uid, namespaceUID string
	err := r.db.QueryRowContext(ctx, claimSQL, now, owner, leaseUntil).Scan(&uid, &namespaceUID)
	if err != nil {
		if errors.Is(err, stdSQL.ErrNoRows) {
			return nil, nil
		}
		return nil, r.fail(err)
	}
	return r.GetInspectionRun(ctx, namespaceUID, uid)
}

func (r *inspectionRepo) StartInspectionRun(ctx context.Context, uid string, windowStart, windowEnd time.Time) error {
	_, err := r.client.InspectionRun.Update().
		Where(inspectionrun.UIDEQ(uid)).
		SetWindowStart(windowStart).
		SetWindowEnd(windowEnd).
		Save(ctx)
	return r.fail(err)
}

func (r *inspectionRepo) FinishInspectionRun(ctx context.Context, in bo.InspectionRun) error {
	update := r.client.InspectionRun.Update().
		Where(inspectionrun.UIDEQ(in.UID)).
		SetStatus(inspectionrun.Status(in.Status)).
		SetItemTotal(in.ItemTotal).
		SetItemOk(in.ItemOK).
		SetItemWarn(in.ItemWarn).
		SetItemFailed(in.ItemFailed).
		SetSummary(in.Summary).
		SetResult(in.Result).
		SetError(in.Error).
		SetLeaseOwner("")
	if in.FinishedAt != nil {
		update.SetFinishedAt(*in.FinishedAt)
	}
	update.ClearLeaseUntil()
	if in.NotifiedAt != nil {
		update.SetNotifiedAt(*in.NotifiedAt)
	}
	n, err := update.Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("inspection run not found")
	}
	return nil
}

func (r *inspectionRepo) GetInspectionRun(ctx context.Context, namespaceUID, uid string) (*bo.InspectionRun, error) {
	row, err := r.client.InspectionRun.Query().Where(
		inspectionrun.UIDEQ(uid),
		inspectionrun.NamespaceUIDEQ(namespaceUID),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toInspectionRun(row)
	return &item, nil
}

func (r *inspectionRepo) ListInspectionRuns(ctx context.Context, namespaceUID string, filter bo.InspectionRunListFilter, page, pageSize int) ([]bo.InspectionRun, int, error) {
	query := r.client.InspectionRun.Query().Where(inspectionrun.NamespaceUIDEQ(namespaceUID))
	if filter.PlanUID != "" {
		query = query.Where(inspectionrun.PlanUIDEQ(filter.PlanUID))
	}
	if filter.Status != "" {
		query = query.Where(inspectionrun.StatusEQ(inspectionrun.Status(filter.Status)))
	}
	if filter.Trigger != "" {
		query = query.Where(inspectionrun.TriggerEQ(inspectionrun.Trigger(filter.Trigger)))
	}
	if filter.From != nil {
		query = query.Where(inspectionrun.ScheduledAtGTE(*filter.From))
	}
	if filter.To != nil {
		query = query.Where(inspectionrun.ScheduledAtLT(*filter.To))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Desc(inspectionrun.FieldScheduledAt)).
		Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.InspectionRun, 0, len(rows))
	for _, row := range rows {
		items = append(items, toInspectionRun(row))
	}
	return items, total, nil
}

func toInspectionRun(row *ent.InspectionRun) bo.InspectionRun {
	return bo.InspectionRun{
		UID: row.UID, NamespaceUID: row.NamespaceUID, PlanUID: row.PlanUID, PlanName: row.PlanName,
		Trigger: string(row.Trigger), Status: string(row.Status), ScheduledAt: row.ScheduledAt,
		StartedAt: row.StartedAt, FinishedAt: row.FinishedAt, WindowStart: row.WindowStart, WindowEnd: row.WindowEnd,
		ItemTotal: row.ItemTotal, ItemOK: row.ItemOk, ItemWarn: row.ItemWarn, ItemFailed: row.ItemFailed,
		Summary: row.Summary, Result: row.Result, Error: row.Error,
		OperatorUID: row.OperatorUID, OperatorName: row.OperatorName,
		NotifiedAt: row.NotifiedAt, LeaseOwner: row.LeaseOwner, LeaseUntil: row.LeaseUntil,
		CreatedAt: row.CreatedAt,
	}
}
