package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aide-family/Alioth/app/megrez/pkg/bo"
	"github.com/aide-family/Alioth/pkg/alerttmpl"
	"github.com/aide-family/Alioth/pkg/fingerprint"
)

// TemplateQuery executes an instant query for Prometheus' query template function.
type TemplateQuery = alerttmpl.QueryFunc

// Hits 按级别条件的保存顺序计算命中。第一条满足的级别生效。
func Hits(rule bo.RuleSnapshot, samples []bo.Sample) []bo.EvalHit {
	return HitsWithQuery(context.Background(), rule, samples, rule.EvalSlot, nil)
}

// HitsWithQuery evaluates hits and expands templates with Prometheus' alert template semantics.
func HitsWithQuery(ctx context.Context, rule bo.RuleSnapshot, samples []bo.Sample, at time.Time, query TemplateQuery) []bo.EvalHit {
	if at.IsZero() {
		at = time.Now()
	}

	hits := make([]bo.EvalHit, 0)
	for _, sample := range samples {
		for _, level := range rule.Levels {
			value, raw, ok := matchLevel(level, sample.Points, sample.Step)
			if !ok {
				continue
			}
			labels := mergeRuleLabels(rule.Labels, sample.Labels)
			encoded, err := json.Marshal(labels)
			if err != nil {
				encoded = []byte("{}")
			}
			hits = append(hits, bo.EvalHit{
				Fingerprint: fingerprint.Of(rule.UID, rule.DatasourceUID, labels),
				Labels:      string(encoded),
				Value:       raw,
				LevelUID:    level.LevelUID,
				Summary:     renderAlert(ctx, rule, rule.Summary, value, labels, string(encoded), at, query),
				Description: renderAlert(ctx, rule, rule.Description, value, labels, string(encoded), at, query),
			})
			break
		}
	}
	return hits
}

// matchLevel 按级别条件比较采样点。step 是数据源的时间桶秒数，只有
// VictoriaLogs stats 这类按桶返回的数据才非 0；FOR 必须保证桶连续，
// 否则窗口中间缺失的桶会被首末点跨度掩盖，形成误告警。
func matchLevel(level bo.Level, points []bo.Point, step int64) (float64, string, bool) {
	if len(points) == 0 {
		return 0, "", false
	}
	switch level.Sample {
	case "MAX":
		best := points[0]
		for _, point := range points[1:] {
			if point.Value > best.Value {
				best = point
			}
		}
		if compare(level.Condition, best.Value, level.Threshold) {
			return best.Value, best.Raw, true
		}
	case "MIN":
		best := points[0]
		for _, point := range points[1:] {
			if point.Value < best.Value {
				best = point
			}
		}
		if compare(level.Condition, best.Value, level.Threshold) {
			return best.Value, best.Raw, true
		}
	case "FOR":
		for _, point := range points {
			if !compare(level.Condition, point.Value, level.Threshold) {
				return 0, "", false
			}
		}
		if step > 0 {
			gap := time.Duration(step) * time.Second
			for index := 1; index < len(points); index++ {
				if points[index].At.Sub(points[index-1].At) > gap {
					return 0, "", false
				}
			}
		}
		span := points[len(points)-1].At.Sub(points[0].At)
		if level.ForSeconds > 0 && span < time.Duration(level.ForSeconds)*time.Second {
			return 0, "", false
		}
		last := points[len(points)-1]
		return last.Value, last.Raw, true
	}
	return 0, "", false
}

func compare(condition string, value float64, threshold string) bool {
	switch condition {
	case "EQ":
		n, ok := atof(threshold)
		return ok && value == n
	case "NE":
		n, ok := atof(threshold)
		return ok && value != n
	case "GT":
		n, ok := atof(threshold)
		return ok && value > n
	case "GTE":
		n, ok := atof(threshold)
		return ok && value >= n
	case "LT":
		n, ok := atof(threshold)
		return ok && value < n
	case "LTE":
		n, ok := atof(threshold)
		return ok && value <= n
	case "IN":
		return inList(value, threshold)
	case "NOT_IN":
		return !inList(value, threshold)
	default:
		return false
	}
}

func inList(value float64, threshold string) bool {
	for _, part := range strings.Split(threshold, ",") {
		n, ok := atof(strings.TrimSpace(part))
		if ok && value == n {
			return true
		}
	}
	return false
}

func atof(value string) (float64, bool) {
	n, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || math.IsNaN(n) {
		return 0, false
	}
	return n, true
}

func window(rule bo.RuleSnapshot, now time.Time) (int64, int64, int64) {
	const step = 15
	seconds := step
	hasFor := false
	for _, level := range rule.Levels {
		if level.Sample == "FOR" && level.ForSeconds > 0 {
			hasFor = true
		}
		if level.ForSeconds > seconds {
			seconds = level.ForSeconds
		}
	}
	end := now.UTC().Unix()
	if rule.Driver == "LOGS_VLOGS" {
		// stats_query_range 按 step 对齐时间桶。结束时间落在桶中间时，最后一个
		// 桶只统计了部分日志，count 偏低会让 FOR / MAX 误判；先对齐到桶边界。
		end -= end % step
		// 起始时间也要对齐，避免第一个桶只覆盖部分时间。
		if remainder := seconds % step; remainder != 0 {
			seconds += step - remainder
		}
		// [start, end) 按 step 切出的桶数比闭区间少一个：ForSeconds 的跨度会
		// 差一个步长。多查一个桶，让 FOR 的跨度覆盖完整窗口。
		if hasFor {
			seconds += step
		}
		return end - int64(seconds), end, step
	}
	return end - int64(seconds), end, step
}

// mergeRuleLabels 把规则上配置的标签并入采样标签。采样标签优先，规则标签只补缺，
// 这样规则标签修改后 fingerprint 会变化，旧事件恢复、新事件按新标签建立。
func mergeRuleLabels(raw string, sample map[string]string) map[string]string {
	labels := make(map[string]string, len(sample))
	for key, value := range sample {
		labels[key] = value
	}
	if strings.TrimSpace(raw) == "" {
		return labels
	}
	extra := map[string]string{}
	if err := json.Unmarshal([]byte(raw), &extra); err != nil {
		return labels
	}
	for key, value := range extra {
		if _, ok := labels[key]; !ok {
			labels[key] = value
		}
	}
	return labels
}

// renderAlert 用 pkg/alerttmpl 展开摘要和描述。数据、快捷变量和函数与通知
// 模板共用一份实现，保证 {{ $labels }}、{{ $value }}、{{ . }} 在两处一致。
func renderAlert(ctx context.Context, rule bo.RuleSnapshot, raw string, value float64, labels map[string]string, labelsJSON string, at time.Time, query alerttmpl.QueryFunc) string {
	if raw == "" || !strings.Contains(raw, "{{") {
		return raw
	}
	externalURL, _ := url.Parse(rule.Endpoint)
	if externalURL == nil {
		externalURL = &url.URL{}
	}
	externalURL.User = nil
	data := alerttmpl.New(labels, labelsJSON, value, externalURL.String(), nil)
	result, err := alerttmpl.Render(ctx, raw, data, "__alert_"+rule.Name, at, query, externalURL)
	if err != nil {
		return fmt.Sprintf("<error expanding template: %s>", err)
	}
	return result
}

func metricsDriver(driver string) bool {
	return driver == "METRICS_PROMETHEUS" || driver == "METRICS_VICTORIA_METRICS"
}

// numericDriver 返回会产出数值采样的驱动。Loki 的数值型 LogQL 与
// VictoriaLogs 的 stats 查询各有独立查询路径，但都复用同一套阈值判定。
func numericDriver(driver string) bool {
	return metricsDriver(driver) || driver == "LOGS_LOKI" || driver == "LOGS_VLOGS"
}
