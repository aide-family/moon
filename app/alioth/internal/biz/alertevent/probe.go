package alertevent

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/alertjudge"
	"github.com/aide-family/moon/pkg/fingerprint"
	labelsutil "github.com/aide-family/moon/pkg/labels"
	"github.com/aide-family/moon/pkg/merr"
)

// 探测任务里携带的标签键。probe_target_uid 用于按目标收敛事件，
// probe_target / probe_kind 用于展示与订阅匹配；三者都是稳定值。
const (
	probeTargetUIDLabel = "probe_target_uid"
	probeTargetLabel    = "probe_target"
	probeKindLabel      = "probe_kind"
)

// defaultProbeInterval 是规则没写周期时的兜底值，与表单默认一致。
const defaultProbeInterval = 30 * time.Second

// ScheduleProbes 推进探测规则的下一次探测时间。探测周期由规则自己决定，
// 不能复用查询规则的 15 秒全局批次。
func (coordinator *Coordinator) ScheduleProbes(ctx context.Context) error {
	if coordinator.locks == nil {
		return merr.Server("lock failed")
	}
	hold, err := coordinator.locks.Acquire(ctx, ProbeScheduleLockKey, ProbeScheduleTTL)
	if err != nil {
		if isNotAcquired(err) {
			return nil
		}
		return err
	}
	defer func() { _ = hold.Release(context.Background()) }()

	now := time.Now().UTC().Truncate(time.Second)
	rules, err := coordinator.rules.ListProbeRulesDue(ctx, now)
	if err != nil {
		return err
	}
	for _, rule := range rules {
		interval := probeInterval(rule)
		slot := now.Truncate(interval)
		next := slot.Add(interval)
		if !next.After(now) {
			next = now.Add(interval)
		}
		if err := coordinator.rules.SetProbeSchedule(ctx, rule.UID, slot, next); err != nil {
			coordinator.log.Error("schedule probe: " + err.Error())
		}
	}
	return nil
}

// ClaimProbeTasks 返回本轮可执行的探测任务。只接受天权的 service key。
func (coordinator *Coordinator) ClaimProbeTasks(ctx context.Context) ([]bo.ProbeTask, error) {
	if err := coordinator.requireMegrez(ctx); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	rules, err := coordinator.rules.ListProbeRulesForClaim(ctx, now)
	if err != nil {
		return nil, err
	}
	tasks := make([]bo.ProbeTask, 0, len(rules))
	for _, rule := range rules {
		if rule.ScheduledSlot == nil {
			continue
		}
		groups, err := coordinator.probes.ListProbeTargetGroupsByMatchers(ctx, rule.NamespaceUID, rule.Driver, rule.ProbeGroupLabels)
		if err != nil {
			return nil, err
		}
		for _, group := range groups {
			targets := make([]bo.ProbeTarget, 0, len(group.Targets))
			for _, target := range group.Targets {
				if target.Status == "ACTIVE" {
					targets = append(targets, target)
				}
			}
			if len(targets) == 0 {
				continue
			}
			tasks = append(tasks, bo.ProbeTask{
				Rule: rule, Group: group, Targets: targets, Slot: *rule.ScheduledSlot,
			})
		}
	}
	return tasks, nil
}

// ReportProbeResults 落库一次回写，按目标判定并写事件。回写幂等：同一
// (target, slot) 重复提交不会新增记录，也不会重复触发事件。
func (coordinator *Coordinator) ReportProbeResults(ctx context.Context, in bo.ProbeReport) error {
	if err := coordinator.requireMegrez(ctx); err != nil {
		return err
	}
	if strings.TrimSpace(in.RuleUID) == "" || len(in.Results) == 0 {
		return merr.Invalid("invalid request")
	}
	rule, err := coordinator.rules.GetAlertRuleByUID(ctx, in.RuleUID)
	if err != nil {
		return err
	}
	if !rule.IsProbeRule() {
		return merr.Invalid("rule is not a probe rule")
	}
	groups, err := coordinator.probes.ListProbeTargetGroupsByMatchers(ctx, rule.NamespaceUID, rule.Driver, rule.ProbeGroupLabels)
	if err != nil {
		return err
	}
	targets := make(map[string]bo.ProbeTarget)
	groupByUID := make(map[string]bo.ProbeTargetGroup)
	for _, group := range groups {
		groupByUID[group.UID] = group
		for _, target := range group.Targets {
			targets[target.UID] = target
		}
	}

	probedAt := time.Now().UTC()
	records := make([]bo.ProbeResult, 0, len(in.Results))
	for _, item := range in.Results {
		target, ok := targets[item.TargetUID]
		if !ok {
			// 目标可能在执行期间被删除，直接丢弃这条结果。
			continue
		}
		group := groupByUID[target.GroupUID]
		records = append(records, bo.ProbeResult{
			NamespaceUID: rule.NamespaceUID,
			RuleUID:      rule.UID,
			GroupUID:     group.UID,
			TargetUID:    target.UID,
			TargetSpec:   target.Spec,
			Slot:         in.Slot,
			ProbedAt:     probedAt,
			OK:           item.OK,
			Value:        item.Value,
			DurationMS:   item.DurationMS,
			StatusCode:   item.StatusCode,
			ErrorKind:    item.ErrorKind,
			ErrorMessage: item.ErrorMessage,
			DNSMS:        item.DNSMS,
			TCPMS:        item.TCPMS,
			TLSMS:        item.TLSMS,
			TTFBMS:       item.TTFBMS,
			CertDays:     item.CertDays,
			ResolvedIP:   item.ResolvedIP,
			Executor:     item.Executor,
		})
	}
	if len(records) == 0 {
		return nil
	}
	if err := coordinator.probes.SaveProbeResults(ctx, records); err != nil {
		return err
	}
	for _, record := range records {
		target := targets[record.TargetUID]
		group := groupByUID[target.GroupUID]
		if err := coordinator.judgeProbeTarget(ctx, *rule, group, target, in.Slot); err != nil {
			coordinator.log.Error("judge probe target: " + err.Error())
		}
	}
	return nil
}

// judgeProbeTarget 取该目标在窗口内的结果序列，按级别条件判定并写事件。
// 点数不足一个窗口时不判定，避免刚建规则就误报。
func (coordinator *Coordinator) judgeProbeTarget(ctx context.Context, rule bo.AlertRule, group bo.ProbeTargetGroup, target bo.ProbeTarget, slot time.Time) error {
	interval := probeInterval(rule)
	window := probeWindow(rule, interval)
	records, err := coordinator.probes.ListProbeResults(ctx, rule.NamespaceUID, []string{target.UID}, slot.Add(-window), slot)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return nil
	}
	points := make([]alertjudge.Point, 0, len(records))
	for _, record := range records {
		value, ok := alertjudge.Atof(record.Value)
		if !ok {
			continue
		}
		points = append(points, alertjudge.Point{Value: value, At: record.Slot, Raw: record.Value})
	}
	if len(points) == 0 {
		return nil
	}
	hits := make([]bo.EvalHit, 0, 1)
	for _, level := range rule.Levels {
		value, raw, ok := alertjudge.Match(alertjudge.Level{
			Condition: level.Condition, Sample: level.Sample,
			Threshold: level.Threshold, ForSeconds: level.ForSeconds,
		}, points, int64(interval.Seconds()))
		if !ok {
			continue
		}
		labelMap, labels := probeLabels(rule, group, target)
		hits = append(hits, bo.EvalHit{
			Fingerprint: fingerprint.Of(rule.UID, target.UID, labelMap),
			Labels:      labels,
			Value:       raw,
			LevelUID:    level.LevelUID,
			Summary:     renderProbeTemplate(rule.Summary, target, group),
			Description: renderProbeTemplate(rule.Description, target, group),
		})
		_ = value
		break
	}
	// 没有命中也要调用：仍在触发的旧事件要按指纹恢复。
	_, err = coordinator.events.ApplyProbeHits(ctx, rule, group, target, slot, hits)
	return err
}

// probeLabels 生成探测事件的稳定标签集合：规则标签 + 目标组标签 + 目标标签，
// 再补目标标识。易变信息（耗时、状态码、错误原因）不进标签。
func probeLabels(rule bo.AlertRule, group bo.ProbeTargetGroup, target bo.ProbeTarget) (map[string]string, string) {
	merged := map[string]string{}
	for _, raw := range []string{group.Labels, target.Labels, rule.Labels} {
		parsed, err := labelsutil.ParseObject(raw)
		if err != nil {
			continue
		}
		for key, value := range parsed {
			merged[key] = value
		}
	}
	merged[probeKindLabel] = "HTTP"
	merged[probeTargetLabel] = target.Spec
	merged[probeTargetUIDLabel] = target.UID
	encoded, err := json.Marshal(merged)
	if err != nil {
		return merged, "{}"
	}
	return merged, string(encoded)
}

// renderProbeTemplate 只做最小替换：探测事件不引入自定义变量与查询函数，
// 模板里的 $labels.* 与 $value 直接取这次命中可用的值。
func renderProbeTemplate(raw string, target bo.ProbeTarget, group bo.ProbeTargetGroup) string {
	if strings.TrimSpace(raw) == "" {
		return raw
	}
	replacer := strings.NewReplacer(
		"{{ $labels.probe_target }}", target.Spec,
		"{{ $labels.probe_group }}", group.Name,
		"{{ $labels.datasource }}", group.Name,
	)
	return replacer.Replace(raw)
}

// probeInterval 返回规则的探测周期。
func probeInterval(rule bo.AlertRule) time.Duration {
	if rule.ProbeIntervalSeconds <= 0 {
		return defaultProbeInterval
	}
	return time.Duration(rule.ProbeIntervalSeconds) * time.Second
}

// probeWindow 是判定窗口：至少一个周期，且覆盖最长的持续秒数。
func probeWindow(rule bo.AlertRule, interval time.Duration) time.Duration {
	window := interval
	for _, level := range rule.Levels {
		if seconds := time.Duration(level.ForSeconds) * time.Second; seconds > window {
			window = seconds
		}
	}
	return window + interval
}

// isNotAcquired 判断锁是否被别人持有。
func isNotAcquired(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not acquired")
}

// pruneProbeWindow 是探测结果的保留窗口：判定只需要最近几天，更早的明细清理掉。
const pruneProbeWindow = 7 * 24 * time.Hour

// PruneProbeResults 清理超过保留期的探测记录，由调度器每小时调用一次。
func (coordinator *Coordinator) PruneProbeResults(ctx context.Context) (int, error) {
	return coordinator.probes.PruneProbeResults(ctx, time.Now().UTC().Add(-pruneProbeWindow), 200)
}
