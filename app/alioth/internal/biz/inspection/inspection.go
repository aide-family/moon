package inspection

import (
	"context"
	"encoding/json"
	htmltmpl "html/template"
	"log/slog"
	"strings"
	texttmpl "text/template"
	"time"

	"github.com/google/wire"
	robfigcron "github.com/robfig/cron/v3"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/biz/datasource"
	"github.com/aide-family/moon/app/alioth/internal/biz/notification"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/app/alioth/pkg/page"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

var ProviderSet = wire.NewSet(NewGateway)

const (
	maxNameRunes  = 64
	maxItems      = 50
	minWindowSecs = 60
	maxWindowSecs = 30 * 24 * 60 * 60
	maxStepSecs   = 24 * 60 * 60
	defaultTZ     = "Asia/Shanghai"
)

var (
	planStatuses   = map[string]struct{}{"ACTIVE": {}, "PAUSED": {}}
	outputFormats  = map[string]struct{}{"HTML": {}, "MARKDOWN": {}}
	notifyPolicies = map[string]struct{}{"ALWAYS": {}, "ON_ABNORMAL": {}}
	itemKinds      = map[string]struct{}{
		"METRICS_QUERY": {}, "LOGS_QUERY": {}, "PROBE_TARGET": {}, "ALERT_EVENTS": {}, "NOTE": {},
	}
	itemDisplays = map[string]struct{}{"METRIC": {}, "TABLE": {}, "TREND": {}, "LOGS": {}, "TEXT": {}}
	cronParser   = robfigcron.NewParser(
		robfigcron.Minute | robfigcron.Hour | robfigcron.Dom | robfigcron.Month | robfigcron.Dow,
	)
)

// Gateway 管理巡检计划与报告模板。调度与执行在后续里程碑接入。
type Gateway struct {
	repo          repository.InspectionRepo
	datasources   *datasource.Gateway
	probes        repository.ProbeTargetRepo
	events        repository.AlertEventRepo
	namespaces    repository.NamespaceRepo
	notifications *notification.Center
	access        access.Checker
	log           *slog.Logger
	owner         string
}

func NewGateway(
	repo repository.InspectionRepo,
	datasources *datasource.Gateway,
	probes repository.ProbeTargetRepo,
	events repository.AlertEventRepo,
	namespaces repository.NamespaceRepo,
	notifications *notification.Center,
	checker access.Checker,
	logger *slog.Logger,
) *Gateway {
	return &Gateway{
		repo: repo, datasources: datasources, probes: probes, events: events,
		namespaces: namespaces, notifications: notifications, access: checker, log: logger, owner: id.New(),
	}
}

// CreatePlan 校验并创建巡检计划，检查项随计划一起保存。
func (gateway *Gateway) CreatePlan(ctx context.Context, in bo.InspectionPlan) (*bo.InspectionPlan, error) {
	ns, err := gateway.access.Allow(ctx, perm.InspectionWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	in.CreatedBy, _ = actor.UserID(ctx)
	if err := gateway.normalizePlan(ctx, ns, &in); err != nil {
		return nil, err
	}
	item, err := gateway.repo.CreateInspectionPlan(ctx, in)
	return item, gateway.access.Wrap(err)
}

// UpdatePlan 更新计划与检查项；检查项按整体替换。
func (gateway *Gateway) UpdatePlan(ctx context.Context, in bo.InspectionPlan) (*bo.InspectionPlan, error) {
	ns, err := gateway.access.Allow(ctx, perm.InspectionWrite)
	if err != nil {
		return nil, err
	}
	current, err := gateway.repo.GetInspectionPlan(ctx, ns, in.UID)
	if err != nil {
		return nil, gateway.access.Wrap(err)
	}
	in.NamespaceUID = ns
	in.CreatedBy = current.CreatedBy
	in.LastRunAt = current.LastRunAt
	if err := gateway.normalizePlan(ctx, ns, &in); err != nil {
		return nil, err
	}
	item, err := gateway.repo.UpdateInspectionPlan(ctx, in)
	return item, gateway.access.Wrap(err)
}

// UpdatePlanStatus 启停计划；启用时重算下次执行，暂停时清空调度。
func (gateway *Gateway) UpdatePlanStatus(ctx context.Context, uid, status string) (*bo.InspectionPlan, error) {
	ns, err := gateway.access.Allow(ctx, perm.InspectionWrite)
	if err != nil {
		return nil, err
	}
	status = strings.ToUpper(strings.TrimSpace(status))
	if _, ok := planStatuses[status]; !ok {
		return nil, merr.Invalid("invalid status")
	}
	current, err := gateway.repo.GetInspectionPlan(ctx, ns, uid)
	if err != nil {
		return nil, gateway.access.Wrap(err)
	}
	var next *time.Time
	if status == "ACTIVE" {
		value, err := nextRun(current.Cron, current.Timezone, time.Now())
		if err != nil {
			return nil, err
		}
		next = &value
	}
	item, err := gateway.repo.UpdateInspectionPlanStatus(ctx, ns, uid, status, next)
	return item, gateway.access.Wrap(err)
}

// DeletePlan 软删除计划。
func (gateway *Gateway) DeletePlan(ctx context.Context, uid string) error {
	ns, err := gateway.access.Allow(ctx, perm.InspectionDelete)
	if err != nil {
		return err
	}
	return gateway.access.Wrap(gateway.repo.DeleteInspectionPlan(ctx, ns, uid))
}

// GetPlan 读取计划详情。
func (gateway *Gateway) GetPlan(ctx context.Context, uid string) (*bo.InspectionPlan, error) {
	ns, err := gateway.access.Allow(ctx, perm.InspectionRead)
	if err != nil {
		return nil, err
	}
	item, err := gateway.repo.GetInspectionPlan(ctx, ns, uid)
	return item, gateway.access.Wrap(err)
}

// TestItem 用真实数据执行一个检查项配置；不落库、不产生告警、不投递。
func (gateway *Gateway) TestItem(ctx context.Context, in bo.InspectionItem) (*bo.InspectionSection, int64, error) {
	ns, err := gateway.access.Allow(ctx, perm.InspectionRead)
	if err != nil {
		return nil, 0, err
	}
	in.NamespaceUID = ns
	in.Enabled = true
	if err := normalizeItem(&in, 0); err != nil {
		return nil, 0, err
	}
	plan := bo.InspectionPlan{NamespaceUID: ns}
	end := time.Now().UTC()
	start := end.Add(-time.Duration(in.WindowSeconds) * time.Second)
	runCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	started := time.Now()
	section := gateway.executeSection(runCtx, plan, in, start, end)
	return &section, time.Since(started).Milliseconds(), nil
}

// ListPlans 分页列出巡检计划。
func (gateway *Gateway) ListPlans(ctx context.Context, pageNo, pageSize int32, filter bo.InspectionPlanListFilter) ([]bo.InspectionPlan, int, int, int, error) {
	ns, err := gateway.access.Allow(ctx, perm.InspectionRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	filter.Name = strings.TrimSpace(filter.Name)
	filter.Status = strings.ToUpper(strings.TrimSpace(filter.Status))
	items, total, err := gateway.repo.ListInspectionPlans(ctx, ns, filter, p, s)
	return items, total, p, s, gateway.access.Wrap(err)
}

// CreateTemplate 校验并创建报告模板。
func (gateway *Gateway) CreateTemplate(ctx context.Context, in bo.InspectionTemplate) (*bo.InspectionTemplate, error) {
	ns, err := gateway.access.Allow(ctx, perm.InspectionTemplateWrite)
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	in.IsBuiltin = false
	createdBy, _ := actor.UserID(ctx)
	in.CreatedBy = createdBy
	if err := normalizeTemplate(&in); err != nil {
		return nil, err
	}
	item, err := gateway.repo.CreateInspectionTemplate(ctx, in)
	return item, gateway.access.Wrap(err)
}

// UpdateTemplate 更新报告模板；内置模板只读。
func (gateway *Gateway) UpdateTemplate(ctx context.Context, in bo.InspectionTemplate) (*bo.InspectionTemplate, error) {
	ns, err := gateway.access.Allow(ctx, perm.InspectionTemplateWrite)
	if err != nil {
		return nil, err
	}
	current, err := gateway.repo.GetInspectionTemplate(ctx, ns, in.UID)
	if err != nil {
		return nil, gateway.access.Wrap(err)
	}
	if current.IsBuiltin {
		return nil, merr.Invalid("builtin template is read-only")
	}
	in.NamespaceUID = ns
	in.IsBuiltin = false
	if err := normalizeTemplate(&in); err != nil {
		return nil, err
	}
	item, err := gateway.repo.UpdateInspectionTemplate(ctx, in)
	return item, gateway.access.Wrap(err)
}

// DeleteTemplate 删除租户模板；内置模板不可删除。
func (gateway *Gateway) DeleteTemplate(ctx context.Context, uid string) error {
	ns, err := gateway.access.Allow(ctx, perm.InspectionTemplateDelete)
	if err != nil {
		return err
	}
	current, err := gateway.repo.GetInspectionTemplate(ctx, ns, uid)
	if err != nil {
		return gateway.access.Wrap(err)
	}
	if current.IsBuiltin {
		return merr.Invalid("builtin template is read-only")
	}
	return gateway.access.Wrap(gateway.repo.DeleteInspectionTemplate(ctx, ns, uid))
}

// GetTemplate 读取模板，内置模板对所有租户可见。
func (gateway *Gateway) GetTemplate(ctx context.Context, uid string) (*bo.InspectionTemplate, error) {
	ns, err := gateway.access.Allow(ctx, perm.InspectionTemplateRead)
	if err != nil {
		return nil, err
	}
	item, err := gateway.repo.GetInspectionTemplate(ctx, ns, uid)
	return item, gateway.access.Wrap(err)
}

// ListTemplates 分页列出模板，内置模板排在前面。
func (gateway *Gateway) ListTemplates(ctx context.Context, pageNo, pageSize int32, name string) ([]bo.InspectionTemplate, int, int, int, error) {
	ns, err := gateway.access.Allow(ctx, perm.InspectionTemplateRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := gateway.repo.ListInspectionTemplates(ctx, ns, strings.TrimSpace(name), p, s)
	return items, total, p, s, gateway.access.Wrap(err)
}

func (gateway *Gateway) normalizePlan(ctx context.Context, namespaceUID string, in *bo.InspectionPlan) error {
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)
	in.Cron = strings.TrimSpace(in.Cron)
	in.Timezone = strings.TrimSpace(in.Timezone)
	in.Remark = strings.TrimSpace(in.Remark)
	if in.Name == "" {
		return merr.Invalid("name is required")
	}
	if in.TimeoutSeconds == 0 {
		in.TimeoutSeconds = 300
	}
	if in.TimeoutSeconds < 30 || in.TimeoutSeconds > 1800 {
		return merr.Invalid("timeout seconds must be between 30 and 1800")
	}
	if in.Timezone == "" {
		in.Timezone = defaultTZ
	}
	if _, err := time.LoadLocation(in.Timezone); err != nil {
		return merr.Invalid("invalid timezone")
	}
	if in.Cron == "" {
		return merr.Invalid("cron is required")
	}
	next, err := nextRun(in.Cron, in.Timezone, time.Now())
	if err != nil {
		return err
	}
	in.Status = strings.ToUpper(strings.TrimSpace(in.Status))
	if in.Status == "" {
		in.Status = "ACTIVE"
	}
	if _, ok := planStatuses[in.Status]; !ok {
		return merr.Invalid("invalid status")
	}
	if in.Status == "ACTIVE" {
		in.NextRunAt = &next
	} else {
		in.NextRunAt = nil
	}

	in.OutputFormats = normalizedFormats(in.OutputFormats)
	if len(in.OutputFormats) == 0 {
		return merr.Invalid("at least one output format is required")
	}
	in.NotifyPolicy = strings.ToUpper(strings.TrimSpace(in.NotifyPolicy))
	if in.NotifyPolicy == "" {
		in.NotifyPolicy = "ON_ABNORMAL"
	}
	if _, ok := notifyPolicies[in.NotifyPolicy]; !ok {
		return merr.Invalid("invalid notify policy")
	}
	in.NotificationGroupUIDs = uniqueStrings(in.NotificationGroupUIDs)
	if in.Labels == "" {
		in.Labels = "{}"
	}
	if err := validateJSONObject(in.Labels, "labels"); err != nil {
		return err
	}
	if in.TemplateUID == "" {
		return merr.Invalid("template is required")
	}
	if _, err := gateway.repo.GetInspectionTemplate(ctx, namespaceUID, in.TemplateUID); err != nil {
		return gateway.access.Wrap(err)
	}
	if len(in.Items) == 0 {
		return merr.Invalid("at least one inspection item is required")
	}
	if len(in.Items) > maxItems {
		return merr.Invalid("too many inspection items")
	}
	for index := range in.Items {
		if err := normalizeItem(&in.Items[index], index); err != nil {
			return err
		}
	}
	return nil
}

func normalizeItem(in *bo.InspectionItem, index int) error {
	in.Kind = strings.ToUpper(strings.TrimSpace(in.Kind))
	in.Title = strings.TrimSpace(in.Title)
	in.DatasourceUID = strings.TrimSpace(in.DatasourceUID)
	in.Display = strings.ToUpper(strings.TrimSpace(in.Display))
	in.Params = strings.TrimSpace(in.Params)
	if _, ok := itemKinds[in.Kind]; !ok {
		return merr.Invalid("invalid inspection item kind")
	}
	if in.Title == "" {
		return merr.Invalid("inspection item title is required")
	}
	if in.WindowSeconds == 0 {
		in.WindowSeconds = 3600
	}
	if in.WindowSeconds < minWindowSecs || in.WindowSeconds > maxWindowSecs {
		return merr.Invalid("inspection item window is out of range")
	}
	if in.StepSeconds == 0 {
		in.StepSeconds = 60
	}
	if in.StepSeconds < 5 || in.StepSeconds > maxStepSecs {
		return merr.Invalid("inspection item step is out of range")
	}
	if in.Display == "" {
		switch in.Kind {
		case "NOTE":
			in.Display = "TEXT"
		case "LOGS_QUERY":
			in.Display = "LOGS"
		default:
			in.Display = "TABLE"
		}
	}
	if _, ok := itemDisplays[in.Display]; !ok {
		return merr.Invalid("invalid inspection item display")
	}
	if !displayAllowed(in.Kind, in.Display) {
		return merr.Invalid("inspection item display does not match its kind")
	}
	if in.Params == "" {
		in.Params = "{}"
	}
	if err := validateJSONObject(in.Params, "params"); err != nil {
		return err
	}
	normalizedParams, err := normalizeItemParams(in.Kind, in.Params)
	if err != nil {
		return err
	}
	in.Params = normalizedParams
	switch in.Kind {
	case "METRICS_QUERY", "LOGS_QUERY":
		if in.DatasourceUID == "" {
			return merr.Invalid("datasource is required")
		}
		if strings.TrimSpace(in.Query) == "" {
			return merr.Invalid("query is required")
		}
	case "NOTE":
		if strings.TrimSpace(in.Query) == "" {
			return merr.Invalid("note content is required")
		}
	}
	in.Sort = index
	return nil
}

func normalizeTemplate(in *bo.InspectionTemplate) error {
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)
	if in.Name == "" {
		return merr.Invalid("name is required")
	}
	if in.HTMLBody == "" && in.MarkdownBody == "" {
		return merr.Invalid("at least one template body is required")
	}
	if in.Status == "" {
		in.Status = "ACTIVE"
	}
	if in.HTMLBody != "" {
		if _, err := htmltmpl.New(in.Name).Funcs(reportFuncs()).Parse(in.HTMLBody); err != nil {
			return merr.Invalid("invalid html template: " + err.Error())
		}
	}
	if in.MarkdownBody != "" {
		if _, err := texttmpl.New(in.Name).Funcs(reportFuncs()).Parse(in.MarkdownBody); err != nil {
			return merr.Invalid("invalid markdown template: " + err.Error())
		}
	}
	return nil
}

func displayAllowed(kind, display string) bool {
	switch kind {
	case "METRICS_QUERY":
		return display == "METRIC" || display == "TABLE" || display == "TREND"
	case "LOGS_QUERY":
		return display == "LOGS" || display == "TABLE"
	case "PROBE_TARGET", "ALERT_EVENTS":
		return display == "TABLE"
	case "NOTE":
		return display == "TEXT"
	default:
		return false
	}
}

func normalizedFormats(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		format := strings.ToUpper(strings.TrimSpace(value))
		if _, ok := outputFormats[format]; !ok {
			continue
		}
		if _, ok := seen[format]; ok {
			continue
		}
		seen[format] = struct{}{}
		out = append(out, format)
	}
	return out
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func validateJSONObject(raw, field string) error {
	var object map[string]any
	if err := json.Unmarshal([]byte(raw), &object); err != nil {
		return merr.Invalid("invalid " + field + " json")
	}
	return nil
}

// nextRun 解析 5 段 cron，返回下一次执行时间。最短调度间隔 5 分钟，防止误配把控制面打满。
func nextRun(expression, timezone string, from time.Time) (time.Time, error) {
	schedule, err := cronParser.Parse(strings.TrimSpace(expression))
	if err != nil {
		return time.Time{}, merr.Invalid("invalid cron expression")
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, merr.Invalid("invalid timezone")
	}
	first := schedule.Next(from.In(location))
	if first.IsZero() {
		return time.Time{}, merr.Invalid("cron has no next run")
	}
	second := schedule.Next(first)
	if second.IsZero() || second.Sub(first) < 5*time.Minute {
		return time.Time{}, merr.Invalid("inspection schedule must be at least 5 minutes apart")
	}
	return first.UTC(), nil
}

// GetRun 读取执行记录详情。
func (gateway *Gateway) GetRun(ctx context.Context, uid string) (*bo.InspectionRun, error) {
	ns, err := gateway.access.Allow(ctx, perm.InspectionRead)
	if err != nil {
		return nil, err
	}
	item, err := gateway.repo.GetInspectionRun(ctx, ns, uid)
	return item, gateway.access.Wrap(err)
}

// ListRuns 分页列出执行记录。
func (gateway *Gateway) ListRuns(ctx context.Context, pageNo, pageSize int32, filter bo.InspectionRunListFilter) ([]bo.InspectionRun, int, int, int, error) {
	ns, err := gateway.access.Allow(ctx, perm.InspectionRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	if filter.From != nil && filter.To != nil {
		if filter.To.Before(*filter.From) {
			return nil, 0, 0, 0, merr.Invalid("invalid time range")
		}
		if filter.To.Sub(*filter.From) > 30*24*time.Hour {
			return nil, 0, 0, 0, merr.Invalid("inspection run range must be within 30 days")
		}
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := gateway.repo.ListInspectionRuns(ctx, ns, filter, p, s)
	return items, total, p, s, gateway.access.Wrap(err)
}

// ResendReport 把指定报告重新入队投递。使用独立来源标识，避免被
// (event_uid, phase, group_uid, channel_uid) 唯一索引挡住。
func (gateway *Gateway) ResendReport(ctx context.Context, uid string) (int, error) {
	ns, err := gateway.access.Allow(ctx, perm.InspectionRun)
	if err != nil {
		return 0, err
	}
	report, err := gateway.repo.GetInspectionReport(ctx, ns, uid)
	if err != nil {
		return 0, gateway.access.Wrap(err)
	}
	run, err := gateway.repo.GetInspectionRun(ctx, ns, report.RunUID)
	if err != nil {
		return 0, gateway.access.Wrap(err)
	}
	plan, err := gateway.repo.GetInspectionPlan(ctx, ns, run.PlanUID)
	if err != nil {
		return 0, gateway.access.Wrap(err)
	}
	if len(plan.NotificationGroupUIDs) == 0 {
		return 0, merr.Invalid("inspection plan has no notification groups")
	}
	subject, body := reportNotice(*plan, *run)
	source := run.UID + ":resend:" + id.New()
	return gateway.notifications.DispatchInspectionReport(ctx, ns, source, subject, body, plan.NotificationGroupUIDs)
}
