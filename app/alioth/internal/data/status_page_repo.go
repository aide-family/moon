package data

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertevent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/namespace"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/proberesult"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/probesladaily"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/probetarget"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/probetargetgroup"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/statuspage"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	labelsutil "github.com/aide-family/moon/pkg/labels"
	"github.com/aide-family/moon/pkg/merr"
)

func (r *statusPageRepo) CreateStatusPage(ctx context.Context, in bo.StatusPage) (*bo.StatusPage, error) {
	token, err := newStatusPageToken()
	if err != nil {
		return nil, r.fail(err)
	}
	row, err := r.client.StatusPage.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetName(in.Name).
		SetTitle(in.Title).
		SetDescription(in.Description).
		SetToken(token).
		SetStatus(statuspage.StatusACTIVE).
		SetTargetGroupUids(in.TargetGroupUIDs).
		SetCreatedBy(in.CreatedBy).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return r.GetStatusPage(ctx, in.NamespaceUID, row.UID)
}

func (r *statusPageRepo) UpdateStatusPage(ctx context.Context, in bo.StatusPage) (*bo.StatusPage, error) {
	n, err := r.client.StatusPage.Update().
		Where(statuspage.UIDEQ(in.UID), statuspage.NamespaceUIDEQ(in.NamespaceUID), statuspage.DeletedAtIsNil()).
		SetName(in.Name).
		SetTitle(in.Title).
		SetDescription(in.Description).
		SetTargetGroupUids(in.TargetGroupUIDs).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("status page not found")
	}
	return r.GetStatusPage(ctx, in.NamespaceUID, in.UID)
}

func (r *statusPageRepo) UpdateStatusPageStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.StatusPage, error) {
	n, err := r.client.StatusPage.Update().
		Where(statuspage.UIDEQ(uid), statuspage.NamespaceUIDEQ(namespaceUID), statuspage.DeletedAtIsNil()).
		SetStatus(statuspage.Status(status)).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("status page not found")
	}
	return r.GetStatusPage(ctx, namespaceUID, uid)
}

func (r *statusPageRepo) RegenerateStatusPageToken(ctx context.Context, namespaceUID, uid string) (*bo.StatusPage, error) {
	token, err := newStatusPageToken()
	if err != nil {
		return nil, r.fail(err)
	}
	n, err := r.client.StatusPage.Update().
		Where(statuspage.UIDEQ(uid), statuspage.NamespaceUIDEQ(namespaceUID), statuspage.DeletedAtIsNil()).
		SetToken(token).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("status page not found")
	}
	return r.GetStatusPage(ctx, namespaceUID, uid)
}

func (r *statusPageRepo) DeleteStatusPage(ctx context.Context, namespaceUID, uid string) error {
	n, err := r.client.StatusPage.Update().
		Where(statuspage.UIDEQ(uid), statuspage.NamespaceUIDEQ(namespaceUID), statuspage.DeletedAtIsNil()).
		SetDeletedAt(time.Now().UTC()).
		Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("status page not found")
	}
	return nil
}

func (r *statusPageRepo) GetStatusPage(ctx context.Context, namespaceUID, uid string) (*bo.StatusPage, error) {
	row, err := r.client.StatusPage.Query().Where(
		statuspage.UIDEQ(uid), statuspage.NamespaceUIDEQ(namespaceUID), statuspage.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toStatusPage(row, true)
	return &item, nil
}

func (r *statusPageRepo) ListStatusPages(ctx context.Context, namespaceUID string, filter bo.StatusPageListFilter, page, pageSize int) ([]bo.StatusPage, int, error) {
	query := r.client.StatusPage.Query().Where(
		statuspage.NamespaceUIDEQ(namespaceUID), statuspage.DeletedAtIsNil(),
	)
	if filter.Name != "" {
		query = query.Where(statuspage.NameContainsFold(filter.Name))
	}
	if filter.Status != "" {
		query = query.Where(statuspage.StatusEQ(statuspage.Status(filter.Status)))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Desc(statuspage.FieldUpdatedAt)).
		Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.StatusPage, 0, len(rows))
	for _, row := range rows {
		items = append(items, toStatusPage(row, true))
	}
	return items, total, nil
}

// PublicStatusPage 把状态页需要的展示数据一次查完：组件状态、可用率与事件。
// 只返回名称与状态，不返回地址、标签和租户身份。
func (r *statusPageRepo) PublicStatusPage(ctx context.Context, token string, now time.Time) (*bo.PublicStatusPage, error) {
	page, err := r.client.StatusPage.Query().Where(
		statuspage.TokenEQ(token),
		statuspage.StatusEQ(statuspage.StatusACTIVE),
		statuspage.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, merr.NotFound("status page not found")
		}
		return nil, r.fail(err)
	}
	active, err := r.client.Namespace.Query().Where(
		namespace.UIDEQ(page.NamespaceUID),
		namespace.StatusEQ(namespace.StatusACTIVE),
		namespace.DeletedAtIsNil(),
	).Exist(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if !active {
		return nil, merr.NotFound("status page not found")
	}
	groups, err := r.client.ProbeTargetGroup.Query().Where(
		probetargetgroup.NamespaceUIDEQ(page.NamespaceUID),
		probetargetgroup.UIDIn(page.TargetGroupUids...),
		probetargetgroup.DeletedAtIsNil(),
	).Order(ent.Asc(probetargetgroup.FieldCreatedAt)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	result := &bo.PublicStatusPage{
		Title: page.Title, Description: page.Description, UpdatedAt: now,
	}
	if len(groups) == 0 {
		return result, nil
	}
	groupUIDs := make([]string, 0, len(groups))
	groupNames := make([]string, 0, len(groups))
	for _, group := range groups {
		groupUIDs = append(groupUIDs, group.UID)
		groupNames = append(groupNames, group.Name)
		component, err := r.publicStatusComponent(ctx, page.NamespaceUID, group, now)
		if err != nil {
			return nil, err
		}
		result.Components = append(result.Components, *component)
	}
	if err := r.publicStatusEvents(ctx, page.NamespaceUID, groupNames, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *statusPageRepo) publicStatusComponent(ctx context.Context, namespaceUID string, group *ent.ProbeTargetGroup, now time.Time) (*bo.PublicStatusComponent, error) {
	component := &bo.PublicStatusComponent{GroupUID: group.UID, Name: group.Name, Status: "UNKNOWN"}
	targets, err := r.client.ProbeTarget.Query().Where(
		probetarget.NamespaceUIDEQ(namespaceUID),
		probetarget.GroupUIDEQ(group.UID),
		probetarget.StatusEQ(probetarget.StatusACTIVE),
		probetarget.DeletedAtIsNil(),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	fault := false
	hasResult := false
	var lastChecked time.Time
	for _, target := range targets {
		latest, err := r.client.ProbeResult.Query().Where(
			proberesult.NamespaceUIDEQ(namespaceUID),
			proberesult.TargetUIDEQ(target.UID),
		).Order(ent.Desc(proberesult.FieldProbedAt)).First(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				continue
			}
			return nil, r.fail(err)
		}
		hasResult = true
		if latest.ProbedAt.After(lastChecked) {
			lastChecked = latest.ProbedAt
		}
		if !latest.Ok {
			fault = true
		}
	}
	if hasResult {
		component.LastCheckedAt = &lastChecked
		switch {
		case fault:
			component.Status = "FAULT"
		case now.Sub(lastChecked) > 10*time.Minute:
			component.Status = "UNKNOWN"
		default:
			component.Status = "OK"
		}
	}
	component.Availability24h, err = r.groupAvailability(ctx, namespaceUID, group.UID, now.Add(-24*time.Hour), now)
	if err != nil {
		return nil, err
	}
	component.Availability7d, err = r.groupAvailability(ctx, namespaceUID, group.UID, now.Add(-7*24*time.Hour), now)
	if err != nil {
		return nil, err
	}
	component.Availability30d, err = r.groupAvailability(ctx, namespaceUID, group.UID, now.Add(-30*24*time.Hour), now)
	if err != nil {
		return nil, err
	}
	return component, nil
}

// groupAvailability 从日聚合里汇总窗口内的可用率；没有数据时返回空字符串。
func (r *statusPageRepo) groupAvailability(ctx context.Context, namespaceUID, groupUID string, from, to time.Time) (string, error) {
	rows, err := r.client.ProbeSlaDaily.Query().Where(
		probesladaily.NamespaceUIDEQ(namespaceUID),
		probesladaily.GroupUIDEQ(groupUID),
		probesladaily.DayGTE(utcDay(from)),
		probesladaily.DayLTE(utcDay(to)),
	).All(ctx)
	if err != nil {
		return "", r.fail(err)
	}
	checks, okChecks := 0, 0
	for _, row := range rows {
		checks += row.Checks
		okChecks += row.OkChecks
	}
	return formatAvailability(okChecks, checks), nil
}

func (r *statusPageRepo) publicStatusEvents(ctx context.Context, namespaceUID string, groupNames []string, page *bo.PublicStatusPage) error {
	allowed := make(map[string]struct{}, len(groupNames))
	for _, name := range groupNames {
		allowed[name] = struct{}{}
	}
	rows, err := r.client.AlertEvent.Query().Where(
		alertevent.NamespaceUIDEQ(namespaceUID),
		alertevent.LabelsContains("probe_kind"),
	).Order(ent.Desc(alertevent.FieldStartsAt)).Limit(200).All(ctx)
	if err != nil {
		return r.fail(err)
	}
	recent := 0
	for _, row := range rows {
		labels, err := labelsutil.ParseObject(row.Labels)
		if err != nil {
			continue
		}
		if _, ok := allowed[labels[datasourceLabelKey]]; !ok {
			continue
		}
		event := bo.PublicStatusEvent{
			UID: row.UID, Title: row.Summary, Status: "RESOLVED", StartedAt: row.StartsAt, EndedAt: row.EndsAt,
		}
		if row.Status == alertevent.StatusFIRING {
			event.Status = "FIRING"
			page.ActiveEvents = append(page.ActiveEvents, event)
			continue
		}
		if recent < 10 {
			page.RecentEvents = append(page.RecentEvents, event)
			recent++
		}
	}
	return nil
}

// newStatusPageToken 生成 16 字节随机数的十六进制串，作为不可猜测的公开凭证。
func newStatusPageToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func toStatusPage(row *ent.StatusPage, withToken bool) bo.StatusPage {
	item := bo.StatusPage{
		UID: row.UID, NamespaceUID: row.NamespaceUID, Name: row.Name, Title: row.Title,
		Description: row.Description, Status: string(row.Status), TargetGroupUIDs: row.TargetGroupUids,
		CreatedBy: row.CreatedBy, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
	if withToken {
		item.Token = row.Token
	}
	return item
}

// utcDay 把时间截断到 UTC 自然日零点，与日聚合的 day 字段对齐。
func utcDay(value time.Time) time.Time {
	utc := value.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

// formatAvailability 把成功/总次数格式化为两位小数百分比；没有数据返回空串。
func formatAvailability(okChecks, checks int) string {
	if checks <= 0 {
		return ""
	}
	return strconv.FormatFloat(float64(okChecks)/float64(checks)*100, 'f', 2, 64)
}
