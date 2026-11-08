package inspection

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

const (
	runLease            = 35 * time.Minute
	maxSectionsRawBytes = 8 * 1024
	logSampleLimit      = 10
)

var levelPattern = regexp.MustCompile(`(?i)\b(error|warn|warning|info|debug)\b`)

// Schedule 扫描到期计划并生成执行记录。多副本靠 next_run_at 的乐观更新与
// (plan_uid, scheduled_at) 唯一索引去重。
func (gateway *Gateway) Schedule(ctx context.Context) error {
	now := time.Now().UTC()
	plans, err := gateway.repo.DueInspectionPlans(ctx, now, 20)
	if err != nil {
		return err
	}
	for _, plan := range plans {
		if plan.NextRunAt == nil {
			continue
		}
		scheduled := *plan.NextRunAt
		// 停机跨过多个周期时只补最近一次，按当前时间重算下次，避免连续补跑。
		next, err := nextRun(plan.Cron, plan.Timezone, now)
		if err != nil {
			gateway.log.Error(fmt.Sprintf("schedule inspection plan %s: %v", plan.UID, err))
			continue
		}
		advanced, err := gateway.repo.AdvanceInspectionPlan(ctx, plan.UID, scheduled, next, now)
		if err != nil {
			gateway.log.Error(fmt.Sprintf("advance inspection plan %s: %v", plan.UID, err))
			continue
		}
		if !advanced {
			continue
		}
		if _, err := gateway.repo.CreateInspectionRun(ctx, bo.InspectionRun{
			NamespaceUID: plan.NamespaceUID, PlanUID: plan.UID, PlanName: plan.Name,
			Trigger: "SCHEDULED", ScheduledAt: scheduled,
		}); err != nil {
			gateway.log.Error(fmt.Sprintf("create inspection run for %s: %v", plan.UID, err))
		}
	}
	return nil
}

// Trigger 手动触发一次巡检。计划已有执行中的记录时返回冲突。
func (gateway *Gateway) Trigger(ctx context.Context, uid string) (*bo.InspectionRun, error) {
	ns, err := gateway.access.Allow(ctx, perm.InspectionRun)
	if err != nil {
		return nil, err
	}
	plan, err := gateway.repo.GetInspectionPlan(ctx, ns, uid)
	if err != nil {
		return nil, gateway.access.Wrap(err)
	}
	active, err := gateway.repo.HasActiveInspectionRun(ctx, ns, uid)
	if err != nil {
		return nil, gateway.access.Wrap(err)
	}
	if active {
		return nil, merr.Conflict("inspection plan is running")
	}
	operator, _ := actor.UserID(ctx)
	run, err := gateway.repo.CreateInspectionRun(ctx, bo.InspectionRun{
		NamespaceUID: ns, PlanUID: plan.UID, PlanName: plan.Name,
		Trigger: "MANUAL", ScheduledAt: time.Now().UTC(), OperatorUID: operator,
	})
	if err != nil {
		return nil, gateway.access.Wrap(err)
	}
	if run == nil {
		return nil, merr.Conflict("inspection plan is running")
	}
	return run, nil
}

// ExecuteNext 领取一条待执行记录并同步执行。调度循环按 tick 调用它，
// 同一时刻只运行一个 run，避免后台查询占满控制面。
func (gateway *Gateway) ExecuteNext(ctx context.Context) error {
	now := time.Now().UTC()
	run, err := gateway.repo.ClaimInspectionRun(ctx, gateway.owner, now, now.Add(runLease))
	if err != nil || run == nil {
		return err
	}
	gateway.execute(ctx, *run)
	return nil
}

func (gateway *Gateway) execute(ctx context.Context, run bo.InspectionRun) {
	plan, err := gateway.repo.GetInspectionPlan(ctx, run.NamespaceUID, run.PlanUID)
	if err != nil {
		gateway.finish(ctx, run, "FAILED", "inspection plan not found")
		return
	}
	windowEnd := run.ScheduledAt
	windowStart := windowEnd.Add(-maxWindow(plan.Items))
	if err := gateway.repo.StartInspectionRun(ctx, run.UID, windowStart, windowEnd); err != nil {
		gateway.log.Error(fmt.Sprintf("start inspection run %s: %v", run.UID, err))
	}
	runCtx, cancel := context.WithTimeout(ctx, time.Duration(plan.TimeoutSeconds)*time.Second)
	defer cancel()

	sections := make([]bo.InspectionSection, 0, len(plan.Items))
	for _, item := range plan.Items {
		if !item.Enabled {
			continue
		}
		if runCtx.Err() != nil {
			sections = append(sections, bo.InspectionSection{
				Title: item.Title, Kind: item.Kind, Status: "FAILED", Error: "inspection timed out",
			})
			continue
		}
		sections = append(sections, gateway.executeSection(runCtx, *plan, item, windowStart, windowEnd))
	}

	total, ok, warn, failed := 0, 0, 0, 0
	for _, section := range sections {
		total++
		switch section.Status {
		case "OK":
			ok++
		case "WARN":
			warn++
		default:
			failed++
		}
	}
	status := "SUCCEEDED"
	if failed == total && total > 0 {
		status = "FAILED"
	} else if warn > 0 || failed > 0 {
		status = "PARTIAL"
	}
	started := run.ScheduledAt
	if run.StartedAt != nil {
		started = *run.StartedAt
	}
	summary, _ := json.Marshal(bo.InspectionSummary{
		Total: total, OK: ok, Warn: warn, Failed: failed, Abnormal: warn + failed,
		Duration: int(time.Since(started).Milliseconds()),
	})
	result, _ := json.Marshal(bo.InspectionResult{Sections: sections})
	run.ItemTotal, run.ItemOK, run.ItemWarn, run.ItemFailed = total, ok, warn, failed
	run.Summary, run.Result = string(summary), string(result)
	finished := time.Now().UTC()
	run.FinishedAt = &finished
	if err := gateway.persistReports(ctx, *plan, run, sections); err != nil {
		status, run.Error = "FAILED", "render report: "+err.Error()
		gateway.log.Error(fmt.Sprintf("render inspection run %s: %v", run.UID, err))
	}
	gateway.finish(ctx, run, status, run.Error)
}

func (gateway *Gateway) finish(ctx context.Context, run bo.InspectionRun, status, message string) {
	run.Status, run.Error = status, message
	if run.FinishedAt == nil {
		finished := time.Now().UTC()
		run.FinishedAt = &finished
	}
	if err := gateway.repo.FinishInspectionRun(context.WithoutCancel(ctx), run); err != nil {
		gateway.log.Error(fmt.Sprintf("finish inspection run %s: %v", run.UID, err))
	}
}

func (gateway *Gateway) executeSection(ctx context.Context, plan bo.InspectionPlan, item bo.InspectionItem, windowStart, windowEnd time.Time) bo.InspectionSection {
	section := bo.InspectionSection{Title: item.Title, Kind: item.Kind, Status: "OK"}
	switch item.Kind {
	case "NOTE":
		section.Note = item.Query
	case "METRICS_QUERY":
		gateway.executeMetrics(ctx, plan, item, windowStart, windowEnd, &section)
	case "LOGS_QUERY":
		gateway.executeLogs(ctx, plan, item, windowStart, windowEnd, &section)
	case "PROBE_TARGET":
		gateway.executeProbes(ctx, plan, item, windowStart, windowEnd, &section)
	case "ALERT_EVENTS":
		gateway.executeAlerts(ctx, plan, item, windowStart, windowEnd, &section)
	default:
		section.Status, section.Error = "FAILED", "unsupported item kind"
	}
	return section
}

func (gateway *Gateway) executeMetrics(ctx context.Context, plan bo.InspectionPlan, item bo.InspectionItem, windowStart, windowEnd time.Time, section *bo.InspectionSection) {
	raw, err := gateway.datasources.QueryRangeForNamespace(
		ctx, plan.NamespaceUID, item.DatasourceUID, item.Query,
		windowStart.Unix(), windowEnd.Unix(), int64(item.StepSeconds),
	)
	if err != nil {
		section.Status, section.Error = "FAILED", err.Error()
		return
	}
	series, table, metrics := parsePrometheusRange(raw)
	section.Series, section.Table, section.Metrics = series, table, metrics
	section.Note = fmt.Sprintf("窗口 %s → %s，%d 条序列", windowStart.Local().Format("2006-01-02 15:04"), windowEnd.Local().Format("2006-01-02 15:04"), len(series))
	if len(series) == 0 {
		section.Note += "；窗口内没有数据"
	}
}

func (gateway *Gateway) executeLogs(ctx context.Context, plan bo.InspectionPlan, item bo.InspectionItem, windowStart, windowEnd time.Time, section *bo.InspectionSection) {
	raw, err := gateway.datasources.QueryLogsForNamespace(
		ctx, plan.NamespaceUID, item.DatasourceUID, item.Query,
		windowStart.Unix(), windowEnd.Unix(), 50, "backward",
	)
	if err != nil {
		section.Status, section.Error = "FAILED", err.Error()
		return
	}
	logs, total := parseLogSamples(raw)
	section.Logs = logs
	if len(logs) == 0 {
		section.Raw = truncateRaw(raw)
		section.Note = fmt.Sprintf("已返回原始日志响应（%d 字节），未解析出样本", len(raw))
	} else {
		section.Note = fmt.Sprintf("窗口内命中 %d 条日志，展示最近 %d 条", total, len(logs))
	}
}

func (gateway *Gateway) executeProbes(ctx context.Context, plan bo.InspectionPlan, item bo.InspectionItem, windowStart, windowEnd time.Time, section *bo.InspectionSection) {
	var params struct {
		GroupUIDs         []string `json:"group_uids"`
		AvailabilityBelow float64  `json:"availability_below"`
	}
	if err := json.Unmarshal([]byte(item.Params), &params); err != nil || len(params.GroupUIDs) == 0 {
		section.Status, section.Error = "FAILED", "probe item requires group_uids"
		return
	}
	table := &bo.InspectionTable{Columns: []string{"目标组", "目标", "检查次数", "可用率", "平均延迟", "状态"}}
	warnings := 0
	for _, groupUID := range params.GroupUIDs {
		group, err := gateway.probes.GetProbeTargetGroup(ctx, plan.NamespaceUID, groupUID)
		if err != nil {
			section.Status, section.Error = "FAILED", err.Error()
			return
		}
		targets, err := gateway.probes.ListProbeTargets(ctx, plan.NamespaceUID, groupUID)
		if err != nil {
			section.Status, section.Error = "FAILED", err.Error()
			return
		}
		targetUIDs := make([]string, 0, len(targets))
		for _, target := range targets {
			targetUIDs = append(targetUIDs, target.UID)
		}
		results, err := gateway.probes.ListProbeResults(ctx, plan.NamespaceUID, targetUIDs, windowStart, windowEnd)
		if err != nil {
			section.Status, section.Error = "FAILED", err.Error()
			return
		}
		byTarget := make(map[string][]bo.ProbeResult, len(targetUIDs))
		for _, result := range results {
			byTarget[result.TargetUID] = append(byTarget[result.TargetUID], result)
		}
		for _, target := range targets {
			checks, ok, latency := probeStats(byTarget[target.UID])
			availability := 100.0
			if checks > 0 {
				availability = float64(ok) / float64(checks) * 100
			}
			status := "正常"
			if checks == 0 {
				status = "无数据"
			} else if availability < params.AvailabilityBelow || ok < checks {
				status = "异常"
				warnings++
			}
			table.Rows = append(table.Rows, []string{
				group.Name, target.Spec, strconv.Itoa(checks),
				fmt.Sprintf("%.2f%%", availability), fmt.Sprintf("%dms", latency), status,
			})
		}
	}
	section.Table = table
	if warnings > 0 {
		section.Status = "WARN"
	}
	section.Note = fmt.Sprintf("探测目标组 %d 个，异常目标 %d 个", len(params.GroupUIDs), warnings)
}

func (gateway *Gateway) executeAlerts(ctx context.Context, plan bo.InspectionPlan, item bo.InspectionItem, windowStart, windowEnd time.Time, section *bo.InspectionSection) {
	var params struct {
		Status     string   `json:"status"`
		WarnLevels []string `json:"warn_levels"`
		Limit      int      `json:"limit"`
	}
	_ = json.Unmarshal([]byte(item.Params), &params)
	if params.Status == "" {
		params.Status = "FIRING"
	}
	if params.Limit <= 0 || params.Limit > 200 {
		params.Limit = 100
	}
	events, _, err := gateway.events.ListEvents(ctx, plan.NamespaceUID, bo.AlertEventListFilter{
		EventStatus: params.Status, StartsAtFrom: &windowStart, StartsAtTo: &windowEnd,
	}, 1, params.Limit)
	if err != nil {
		section.Status, section.Error = "FAILED", err.Error()
		return
	}
	warnSet := make(map[string]struct{}, len(params.WarnLevels))
	for _, level := range params.WarnLevels {
		warnSet[strings.ToUpper(strings.TrimSpace(level))] = struct{}{}
	}
	rows := make([]bo.InspectionAlertRow, 0, len(events))
	warned := 0
	for _, event := range events {
		rows = append(rows, bo.InspectionAlertRow{
			Title: event.Summary, Level: event.LevelName, Status: event.Status,
			StartedAt: event.StartsAt.Local().Format("2006-01-02 15:04"),
		})
		if _, ok := warnSet[strings.ToUpper(event.LevelName)]; ok {
			warned++
		}
	}
	section.Alerts = rows
	if warned > 0 {
		section.Status = "WARN"
	}
	section.Note = fmt.Sprintf("窗口内 %s 事件 %d 条", params.Status, len(rows))
}

func maxWindow(items []bo.InspectionItem) time.Duration {
	longest := time.Hour
	for _, item := range items {
		if item.Enabled && time.Duration(item.WindowSeconds)*time.Second > longest {
			longest = time.Duration(item.WindowSeconds) * time.Second
		}
	}
	return longest
}

func probeStats(results []bo.ProbeResult) (checks, ok, avgLatency int) {
	if len(results) == 0 {
		return 0, 0, 0
	}
	total := 0
	for _, result := range results {
		total += result.DurationMS
		if result.OK {
			ok++
		}
	}
	return len(results), ok, total / len(results)
}

type promPayload struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string   `json:"metric"`
			Value  []json.RawMessage   `json:"value"`
			Values [][]json.RawMessage `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

func parsePrometheusRange(raw string) ([]bo.InspectionSeries, *bo.InspectionTable, []bo.InspectionMetric) {
	var payload promPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil || payload.Status != "success" {
		return nil, nil, nil
	}
	series := make([]bo.InspectionSeries, 0, len(payload.Data.Result))
	table := &bo.InspectionTable{Columns: []string{"序列", "最新值", "样本数"}}
	metrics := make([]bo.InspectionMetric, 0, 4)
	for _, item := range payload.Data.Result {
		name := promSeriesName(item.Metric)
		points := make([]bo.InspectionPoint, 0, len(item.Values))
		if len(item.Value) == 2 {
			if point, ok := promPoint(item.Value); ok {
				points = append(points, point)
			}
		}
		for _, value := range item.Values {
			if point, ok := promPoint(value); ok {
				points = append(points, point)
			}
		}
		if len(points) == 0 {
			continue
		}
		series = append(series, bo.InspectionSeries{Name: name, Points: points})
		latest := points[len(points)-1].Value
		table.Rows = append(table.Rows, []string{name, strconv.FormatFloat(latest, 'f', -1, 64), strconv.Itoa(len(points))})
		if len(metrics) < 4 {
			metrics = append(metrics, bo.InspectionMetric{Name: name, Value: strconv.FormatFloat(latest, 'f', -1, 64)})
		}
	}
	if len(table.Rows) == 0 {
		return series, nil, metrics
	}
	return series, table, metrics
}

func promPoint(raw []json.RawMessage) (bo.InspectionPoint, bool) {
	if len(raw) != 2 {
		return bo.InspectionPoint{}, false
	}
	var timestamp float64
	if err := json.Unmarshal(raw[0], &timestamp); err != nil {
		return bo.InspectionPoint{}, false
	}
	var valueText string
	if err := json.Unmarshal(raw[1], &valueText); err != nil {
		return bo.InspectionPoint{}, false
	}
	value, err := strconv.ParseFloat(valueText, 64)
	if err != nil {
		return bo.InspectionPoint{}, false
	}
	return bo.InspectionPoint{Timestamp: int64(timestamp), Value: value}, true
}

func promSeriesName(metric map[string]string) string {
	name := metric["__name__"]
	keys := make([]string, 0, len(metric))
	for key := range metric {
		if key != "__name__" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+metric[key])
	}
	if name == "" {
		name = strings.Join(parts, ",")
	} else if len(parts) > 0 {
		name += "{" + strings.Join(parts, ",") + "}"
	}
	return name
}

type lokiPayload struct {
	Data struct {
		Result []struct {
			Stream map[string]string `json:"stream"`
			Values [][]string        `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

func parseLogSamples(raw string) ([]bo.InspectionLogSample, int) {
	var payload lokiPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil || len(payload.Data.Result) == 0 {
		return nil, 0
	}
	total := 0
	samples := make([]bo.InspectionLogSample, 0, logSampleLimit)
	for _, stream := range payload.Data.Result {
		for _, value := range stream.Values {
			total++
			if len(samples) >= logSampleLimit || len(value) != 2 {
				continue
			}
			message := value[1]
			if len(message) > 400 {
				message = message[:400]
			}
			samples = append(samples, bo.InspectionLogSample{Time: value[0], Level: logLevel(message), Message: message})
		}
	}
	return samples, total
}

func logLevel(message string) string {
	if match := levelPattern.FindStringSubmatch(message); len(match) == 2 {
		return strings.ToUpper(match[1])
	}
	return "INFO"
}

func truncateRaw(raw string) string {
	if len(raw) <= maxSectionsRawBytes {
		return raw
	}
	return raw[:maxSectionsRawBytes]
}
