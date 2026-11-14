package inspection

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
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
		run.Status, run.Error = "FAILED", "render report: "+err.Error()
		gateway.log.Error(fmt.Sprintf("render inspection run %s: %v", run.UID, err))
		gateway.finish(ctx, run, run.Status, run.Error)
		return
	}
	gateway.finish(ctx, run, status, "")
	gateway.deliver(ctx, *plan, run)
}

// deliver 按发送条件把报告摘要与链接投递到通知组。投递失败只记日志，不影响报告。
func (gateway *Gateway) deliver(ctx context.Context, plan bo.InspectionPlan, run bo.InspectionRun) {
	if len(plan.NotificationGroupUIDs) == 0 {
		return
	}
	if plan.NotifyPolicy == "ON_ABNORMAL" && run.ItemWarn+run.ItemFailed == 0 {
		return
	}
	subject, body := reportNotice(plan, run)
	created, err := gateway.notifications.DispatchInspectionReport(ctx, run.NamespaceUID, run.UID, subject, body, plan.NotificationGroupUIDs)
	if err != nil {
		gateway.log.Error(fmt.Sprintf("deliver inspection run %s: %v", run.UID, err))
		return
	}
	if created > 0 {
		if err := gateway.repo.MarkInspectionRunNotified(context.WithoutCancel(ctx), run.UID, time.Now().UTC()); err != nil {
			gateway.log.Error(fmt.Sprintf("mark inspection run %s notified: %v", run.UID, err))
		}
	}
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
	params := parseMetricsParams(item.Params)
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
	for index := range section.Metrics {
		section.Metrics[index].Unit = params.Unit
	}
	section.Note = fmt.Sprintf("窗口 %s → %s，%d 条序列", windowStart.Local().Format("2006-01-02 15:04"), windowEnd.Local().Format("2006-01-02 15:04"), len(series))
	if len(series) == 0 {
		section.Note += "；窗口内没有数据"
		if params.EmptyIsWarn {
			section.Status = "WARN"
			section.Note += "（已按配置标记警告）"
		}
		return
	}
	if status := metricThresholdStatus(series, params.Thresholds); status != "" {
		section.Status = status
		if status == "OK" {
			section.Note += "；命中正常阈值"
		} else {
			section.Note += fmt.Sprintf("；命中阈值，标记 %s", status)
		}
	}
}

// metricThresholdStatus 对每条序列的最新值评估阈值。状态按严重度取最高：
// FAILED > WARN > OK；OK 阈值用于显式声明「命中即正常」，不会覆盖异常。
func metricThresholdStatus(series []bo.InspectionSeries, thresholds []metricThreshold) string {
	status := ""
	for _, threshold := range thresholds {
		for _, item := range series {
			if len(item.Points) == 0 {
				continue
			}
			value := item.Points[len(item.Points)-1].Value
			matched := false
			switch threshold.Op {
			case ">":
				matched = value > threshold.Value
			case ">=":
				matched = value >= threshold.Value
			case "<":
				matched = value < threshold.Value
			case "<=":
				matched = value <= threshold.Value
			case "==":
				matched = math.Abs(value-threshold.Value) < 1e-9
			case "!=":
				matched = math.Abs(value-threshold.Value) >= 1e-9
			}
			if !matched {
				continue
			}
			switch threshold.Status {
			case "FAILED":
				return "FAILED"
			case "WARN":
				status = "WARN"
			case "OK":
				if status == "" {
					status = "OK"
				}
			}
		}
	}
	return status
}

func (gateway *Gateway) executeLogs(ctx context.Context, plan bo.InspectionPlan, item bo.InspectionItem, windowStart, windowEnd time.Time, section *bo.InspectionSection) {
	params := parseLogsParams(item.Params)
	raw, err := gateway.datasources.QueryLogsForNamespace(
		ctx, plan.NamespaceUID, item.DatasourceUID, item.Query,
		windowStart.Unix(), windowEnd.Unix(), int32(params.Limit), strings.ToLower(params.Direction),
	)
	if err != nil {
		section.Status, section.Error = "FAILED", err.Error()
		return
	}
	logs, total := parseLogSamples(raw)
	if len(params.Levels) > 0 {
		allowed := make(map[string]struct{}, len(params.Levels))
		for _, level := range params.Levels {
			allowed[strings.ToUpper(level)] = struct{}{}
		}
		filtered := make([]bo.InspectionLogSample, 0, len(logs))
		for _, sample := range logs {
			if _, ok := allowed[strings.ToUpper(sample.Level)]; ok {
				filtered = append(filtered, sample)
			}
		}
		logs = filtered
	}
	section.Logs = logs
	if len(logs) == 0 {
		section.Raw = truncateRaw(raw)
		section.Note = fmt.Sprintf("已返回原始日志响应（%d 字节），未解析出匹配样本", len(raw))
	} else {
		section.Note = fmt.Sprintf("窗口内命中 %d 条日志，展示 %d 条", total, len(logs))
	}
	if params.WarnAbove > 0 && total >= params.WarnAbove {
		section.Status = "WARN"
		section.Note += fmt.Sprintf("；命中数达到告警阈值 %d", params.WarnAbove)
	}
}

func (gateway *Gateway) executeProbes(ctx context.Context, plan bo.InspectionPlan, item bo.InspectionItem, windowStart, windowEnd time.Time, section *bo.InspectionSection) {
	params := parseProbeParams(item.Params)
	if len(params.GroupUIDs) == 0 {
		section.Status, section.Error = "FAILED", "probe item requires group_uids"
		return
	}
	table := &bo.InspectionTable{Columns: []string{"目标组", "目标", "检查次数", "可用率", "平均延迟", "P95", "证书剩余", "状态"}}
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
			summary := summarizeProbe(byTarget[target.UID])
			availability := 0.0
			if summary.checks > 0 {
				availability = float64(summary.ok) / float64(summary.checks) * 100
			}
			status := probeStatus(params, summary, availability)
			if status == "异常" {
				warnings++
			}
			table.Rows = append(table.Rows, []string{
				group.Name, target.Spec, strconv.Itoa(summary.checks),
				fmt.Sprintf("%.2f%%", availability), fmt.Sprintf("%dms", summary.avgLatency),
				fmt.Sprintf("%dms", summary.p95Latency), probeCertText(summary.certDays), status,
			})
		}
	}
	section.Table = table
	if warnings > 0 {
		section.Status = "WARN"
	}
	section.Note = fmt.Sprintf("探测目标组 %d 个，按%s评估，异常目标 %d 个", len(params.GroupUIDs), probeMetricLabel(params.Metric), warnings)
}

func probeStatus(params probeParams, summary probeSummary, availability float64) string {
	if summary.checks == 0 {
		return "无数据"
	}
	switch params.Metric {
	case "LATENCY":
		if summary.avgLatency > params.LatencyAboveMS {
			return "异常"
		}
	case "P95":
		if summary.p95Latency > params.LatencyAboveMS {
			return "异常"
		}
	case "CERT_DAYS":
		if summary.certDays <= 0 {
			return "无数据"
		}
		if summary.certDays < params.CertDaysBelow {
			return "异常"
		}
	default:
		warnOnDown := params.WarnOnDown == nil || *params.WarnOnDown
		if availability < params.AvailabilityBelow || (warnOnDown && summary.ok < summary.checks) {
			return "异常"
		}
	}
	return "正常"
}

func probeCertText(days int) string {
	if days <= 0 {
		return "—"
	}
	return fmt.Sprintf("%d天", days)
}

func probeMetricLabel(metric string) string {
	switch metric {
	case "LATENCY":
		return "平均延迟"
	case "P95":
		return "P95 延迟"
	case "CERT_DAYS":
		return "证书剩余天数"
	default:
		return "可用率"
	}
}

func (gateway *Gateway) executeAlerts(ctx context.Context, plan bo.InspectionPlan, item bo.InspectionItem, windowStart, windowEnd time.Time, section *bo.InspectionSection) {
	params := parseAlertParams(item.Params)
	fetchLimit := params.Limit
	if len(params.LevelUIDs) > 0 || params.Labels != "" {
		fetchLimit = 200
	}
	status := params.Status
	if status == "ALL" {
		status = ""
	}
	events, _, err := gateway.events.ListEvents(ctx, plan.NamespaceUID, bo.AlertEventListFilter{
		EventStatus: status, StartsAtFrom: &windowStart, StartsAtTo: &windowEnd,
		LabelsJSON: params.Labels,
	}, 1, fetchLimit)
	if err != nil {
		section.Status, section.Error = "FAILED", err.Error()
		return
	}
	levelSet := make(map[string]struct{}, len(params.LevelUIDs))
	for _, level := range params.LevelUIDs {
		levelSet[level] = struct{}{}
	}
	warnSet := make(map[string]struct{}, len(params.WarnLevels))
	for _, level := range params.WarnLevels {
		warnSet[strings.ToUpper(strings.TrimSpace(level))] = struct{}{}
	}
	rows := make([]bo.InspectionAlertRow, 0, params.Limit)
	warned := 0
	for _, event := range events {
		if len(levelSet) > 0 {
			if _, ok := levelSet[event.LevelUID]; !ok {
				continue
			}
		}
		rows = append(rows, bo.InspectionAlertRow{
			Title: event.Summary, Level: event.LevelName, Status: event.Status,
			StartedAt: event.StartsAt.Local().Format("2006-01-02 15:04"),
		})
		if _, ok := warnSet[strings.ToUpper(event.LevelName)]; ok {
			warned++
		} else if _, ok := warnSet[strings.ToUpper(event.LevelUID)]; ok {
			warned++
		}
		if len(rows) >= params.Limit {
			break
		}
	}
	section.Alerts = rows
	if warned > 0 {
		section.Status = "WARN"
	}
	section.Note = fmt.Sprintf("窗口内 %s 事件 %d 条", alertStatusLabel(params.Status), len(rows))
}

func alertStatusLabel(status string) string {
	if status == "ALL" {
		return "全部"
	}
	return "未恢复"
}

// reportNotice 生成投递摘要与报告链接。链接前缀来自 ALIOTH_CONSOLE_BASE_URL。
func reportNotice(plan bo.InspectionPlan, run bo.InspectionRun) (string, string) {
	finished := run.ScheduledAt
	if run.FinishedAt != nil {
		finished = *run.FinishedAt
	}
	subject := fmt.Sprintf("[巡检] %s · %s", plan.Name, finished.Local().Format("2006-01-02 15:04"))
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("ALIOTH_CONSOLE_BASE_URL")), "/")
	body := fmt.Sprintf(
		"巡检：%s\n时间：%s\n结果：%d 项，%d 正常，%d 警告，%d 失败",
		plan.Name, finished.Local().Format("2006-01-02 15:04"),
		run.ItemTotal, run.ItemOK, run.ItemWarn, run.ItemFailed,
	)
	if baseURL != "" {
		body += "\n报告：" + baseURL + "/inspection/runs/" + run.UID
	} else {
		body += "\n报告：请登录 Alioth 查看执行 " + run.UID
	}
	return subject, body
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

type probeSummary struct {
	checks     int
	ok         int
	avgLatency int
	p95Latency int
	certDays   int
}

func summarizeProbe(results []bo.ProbeResult) probeSummary {
	checks, ok, avgLatency := probeStats(results)
	summary := probeSummary{checks: checks, ok: ok, avgLatency: avgLatency}
	if len(results) == 0 {
		return summary
	}
	durations := make([]int, 0, len(results))
	for _, result := range results {
		durations = append(durations, result.DurationMS)
		if result.Value != "" {
			if days, err := strconv.ParseFloat(strings.TrimSpace(result.Value), 64); err == nil && days > 0 {
				value := int(days)
				if summary.certDays == 0 || value < summary.certDays {
					summary.certDays = value
				}
			}
		}
	}
	sort.Ints(durations)
	index := int(math.Ceil(float64(len(durations))*0.95)) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(durations) {
		index = len(durations) - 1
	}
	summary.p95Latency = durations[index]
	return summary
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
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, 0
	}
	// Loki: data.result[].values。
	var loki lokiPayload
	if err := json.Unmarshal([]byte(raw), &loki); err == nil && len(loki.Data.Result) > 0 {
		total := 0
		samples := make([]bo.InspectionLogSample, 0, logSampleLimit)
		for _, stream := range loki.Data.Result {
			for _, value := range stream.Values {
				total++
				if len(samples) >= logSampleLimit || len(value) != 2 {
					continue
				}
				samples = append(samples, logSample(value[0], value[1]))
			}
		}
		return samples, total
	}
	// Elasticsearch: hits.hits[]._source。
	var es struct {
		Hits struct {
			Total any `json:"total"`
			Hits  []struct {
				Source map[string]any `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.Unmarshal([]byte(raw), &es); err == nil && len(es.Hits.Hits) > 0 {
		total := len(es.Hits.Hits)
		if object, ok := es.Hits.Total.(map[string]any); ok {
			if value, ok := object["value"].(float64); ok {
				total = int(value)
			}
		}
		samples := make([]bo.InspectionLogSample, 0, logSampleLimit)
		for _, hit := range es.Hits.Hits {
			if len(samples) >= logSampleLimit {
				break
			}
			samples = append(samples, logMapSample(hit.Source))
		}
		return samples, total
	}
	// VictoriaLogs 与其它驱动：JSON 行或纯文本行。
	lines := strings.Split(raw, "\n")
	samples := make([]bo.InspectionLogSample, 0, logSampleLimit)
	total := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		total++
		if len(samples) >= logSampleLimit {
			continue
		}
		var object map[string]any
		if json.Unmarshal([]byte(line), &object) == nil {
			samples = append(samples, logMapSample(object))
			continue
		}
		samples = append(samples, logSample("", line))
	}
	return samples, total
}

func logSample(timestamp, message string) bo.InspectionLogSample {
	if len(message) > 400 {
		message = message[:400]
	}
	return bo.InspectionLogSample{Time: timestamp, Level: logLevel(message), Message: message}
}

func logMapSample(object map[string]any) bo.InspectionLogSample {
	message := stringField(object, "_msg", "message", "msg", "log", "body")
	timestamp := stringField(object, "_time", "@timestamp", "timestamp", "time")
	level := strings.ToUpper(stringField(object, "level", "severity", "_level"))
	if level == "" {
		level = logLevel(message)
	}
	if len(message) > 400 {
		message = message[:400]
	}
	return bo.InspectionLogSample{Time: timestamp, Level: level, Message: message}
}

func stringField(object map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := object[key]; ok {
			switch typed := value.(type) {
			case string:
				return typed
			case float64:
				return strconv.FormatFloat(typed, 'f', -1, 64)
			}
		}
	}
	return ""
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
