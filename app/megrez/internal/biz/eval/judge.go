package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/aide-family/moon/app/megrez/pkg/bo"
	"github.com/aide-family/moon/pkg/alertjudge"
	"github.com/aide-family/moon/pkg/alerttmpl"
	"github.com/aide-family/moon/pkg/fingerprint"
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
			labelsJSON := string(encoded)
			// 变量只影响文案：先按命中标签求值，再注入摘要/描述模板；
			// fingerprint、级别判定和恢复路径都不使用变量。
			variables := resolveVariables(ctx, rule, value, labels, labelsJSON, at, query)
			hits = append(hits, bo.EvalHit{
				Fingerprint: fingerprint.Of(rule.UID, rule.DatasourceUID, labels),
				Labels:      labelsJSON,
				Value:       raw,
				LevelUID:    level.LevelUID,
				Summary:     renderAlert(ctx, rule, rule.Summary, value, labels, labelsJSON, at, query, variables),
				Description: renderAlert(ctx, rule, rule.Description, value, labels, labelsJSON, at, query, variables),
			})
			break
		}
	}
	return hits
}

// matchLevel 按级别条件比较采样点，比较逻辑与探测规则共用 pkg/alertjudge。
func matchLevel(level bo.Level, points []bo.Point, step int64) (float64, string, bool) {
	if len(points) == 0 {
		return 0, "", false
	}
	converted := make([]alertjudge.Point, 0, len(points))
	for _, point := range points {
		converted = append(converted, alertjudge.Point{Value: point.Value, At: point.At, Raw: point.Raw})
	}
	return alertjudge.Match(alertjudge.Level{
		Condition:  level.Condition,
		Sample:     level.Sample,
		Threshold:  level.Threshold,
		ForSeconds: level.ForSeconds,
	}, converted, step)
}

func window(rule bo.RuleSnapshot, now time.Time) (int64, int64, int64) {
	step := int64(15)
	hasFor := false
	shortestFor := int64(0)
	for _, level := range rule.Levels {
		if level.Sample == "FOR" && level.ForSeconds > 0 {
			hasFor = true
			forSeconds := int64(level.ForSeconds)
			if shortestFor == 0 || forSeconds < shortestFor {
				shortestFor = forSeconds
			}
		}
	}
	// bucketDriver 表示按时间桶返回数据的驱动：桶边界由 step 决定，窗口必须先对齐，
	// 否则最后一个桶只统计了部分数据，count 偏低会让 FOR / MAX 误判。
	bucketDriver := rule.Driver == "LOGS_VLOGS" || rule.Driver == "LOGS_ELASTICSEARCH"
	if bucketDriver && shortestFor > 0 && shortestFor < step {
		// 持续 10 秒的规则若仍按 15 秒切桶，至少需要两个满桶（30 秒）才能命中，
		// 与配置语义相差一个量级。短窗口改用与 FOR 相同的桶宽，两个桶即可覆盖。
		step = shortestFor
	}
	seconds := step
	for _, level := range rule.Levels {
		forSeconds := int64(level.ForSeconds)
		if forSeconds <= 0 {
			continue
		}
		// query_range 只在 start + k*step 上取点。窗口不是 step 整数倍时，首末
		// 采样点的跨度会小于 ForSeconds，FOR 永远无法满足；先向上取整到步长。
		if remainder := forSeconds % step; remainder != 0 {
			forSeconds += step - remainder
		}
		if forSeconds > seconds {
			seconds = forSeconds
		}
	}
	// VictoriaMetrics 默认 -search.latencyOffset=30s：结束时间贴近当前时刻、
	// 宽度不足一个偏移周期的区间会返回空矩阵。指标规则至少覆盖两个 step，
	// 保证每次评测都能拿到采样点，而不是只写一条空批次的幂等记录。
	if rule.Driver == "METRICS_VICTORIA_METRICS" && seconds < 2*step {
		seconds = 2 * step
	}
	end := now.UTC().Unix()
	if bucketDriver {
		// stats_query_range 和 date_histogram 都按 step 对齐时间桶。结束时间落在
		// 桶中间时最后一个桶只统计了部分数据；先对齐到桶边界。
		end -= end % step
		// [start, end) 按 step 切出的桶数比闭区间少一个：ForSeconds 的跨度会
		// 差一个步长。多查一个桶，让 FOR 的跨度覆盖完整窗口。
		if hasFor {
			seconds += step
		}
		return end - seconds, end, step
	}
	return end - seconds, end, step
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

// variableValues 是一次命中里自定义变量的求值结果。names 包含规则声明的
// 全部变量名，用于生成 $name 快捷变量；values/labels 只包含求值成功的变量，
// 查询失败或空结果时不写键，模板里显示为 <no value>。
type variableValues struct {
	names  []string
	values map[string]any
	labels map[string]map[string]string
}

// resolveVariables 只对摘要或描述里实际引用的变量求值：模板里出现 $name，
// 或整体引用了 .Vars（此时无法静态判断键，全部求值）。变量查询语句先用同一
// 份告警数据展开，从而继承主查询命中的 $labels/$value；再用评测期的即时查询
// 取第一条序列。任何一步失败都只丢弃该变量，不影响命中。
func resolveVariables(ctx context.Context, rule bo.RuleSnapshot, value float64, labels map[string]string, labelsJSON string, at time.Time, query TemplateQuery) *variableValues {
	if len(rule.Variables) == 0 {
		return nil
	}
	externalURL := alertExternalURL(rule)
	data := alerttmpl.NewData(labels, labelsJSON, value, externalURL.String(), nil)
	allReferenced := referencesVarTable(rule.Summary) || referencesVarTable(rule.Description)
	resolved := &variableValues{
		names:  make([]string, 0, len(rule.Variables)),
		values: make(map[string]any, len(rule.Variables)),
		labels: make(map[string]map[string]string, len(rule.Variables)),
	}
	for _, variable := range rule.Variables {
		name := strings.TrimSpace(variable.Name)
		if name == "" {
			continue
		}
		resolved.names = append(resolved.names, name)
		if !allReferenced && !referencesVariable(rule.Summary, name) && !referencesVariable(rule.Description, name) {
			continue
		}
		expr := variable.Expr
		if strings.Contains(expr, "{{") {
			rendered, err := alerttmpl.Render(ctx, expr, data, "__alert_var_"+rule.Name+"_"+name, at, nil, externalURL)
			if err != nil {
				continue
			}
			expr = rendered
		}
		expr = strings.TrimSpace(expr)
		if expr == "" || query == nil {
			continue
		}
		samples, err := query(ctx, expr, at)
		if err != nil || len(samples) == 0 {
			continue
		}
		resolved.values[name] = samples[0].Value
		resolved.labels[name] = samples[0].Labels
	}
	return resolved
}

// referencesVarTable 判断模板是否整体引用了变量表。
func referencesVarTable(raw string) bool {
	return strings.Contains(raw, ".Vars") || strings.Contains(raw, ".VarLabels")
}

// referencesVariable 判断模板里是否写了 $name。变量名后面必须是非标识符
// 字符，避免 $disk_free 误命中 $disk_free_ratio。
func referencesVariable(raw, name string) bool {
	if raw == "" || name == "" {
		return false
	}
	needle := "$" + name
	for offset := 0; offset < len(raw); {
		index := strings.Index(raw[offset:], needle)
		if index < 0 {
			return false
		}
		end := offset + index + len(needle)
		if end >= len(raw) || !isVariableNameByte(raw[end]) {
			return true
		}
		offset = end
	}
	return false
}

func isVariableNameByte(char byte) bool {
	return char == '_' ||
		(char >= '0' && char <= '9') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= 'a' && char <= 'z')
}

// alertExternalURL 解析规则配置的外部地址。凭据不参与模板展开。
func alertExternalURL(rule bo.RuleSnapshot) *url.URL {
	externalURL, _ := url.Parse(rule.Endpoint)
	if externalURL == nil {
		externalURL = &url.URL{}
	}
	externalURL.User = nil
	return externalURL
}

// renderAlert 用 pkg/alerttmpl 展开摘要和描述。数据、快捷变量和函数与通知
// 模板共用一份实现，保证 {{ $labels }}、{{ $value }}、{{ . }} 在两处一致。
// variables 为 nil 时走与通知完全相同的 Render 路径。
func renderAlert(ctx context.Context, rule bo.RuleSnapshot, raw string, value float64, labels map[string]string, labelsJSON string, at time.Time, query alerttmpl.QueryFunc, variables *variableValues) string {
	if raw == "" || !strings.Contains(raw, "{{") {
		return raw
	}
	externalURL := alertExternalURL(rule)
	var extra map[string]any
	var names []string
	if variables != nil {
		extra = map[string]any{"Vars": variables.values, "VarLabels": variables.labels}
		names = variables.names
	}
	data := alerttmpl.NewData(labels, labelsJSON, value, externalURL.String(), extra)
	result, err := alerttmpl.RenderWithVars(ctx, raw, data, "__alert_"+rule.Name, at, query, externalURL, names)
	if err != nil {
		return fmt.Sprintf("<error expanding template: %s>", err)
	}
	return result
}

func metricsDriver(driver string) bool {
	return driver == "METRICS_PROMETHEUS" || driver == "METRICS_VICTORIA_METRICS"
}

// numericDriver 返回会产出数值采样的驱动。Loki 的数值型 LogQL、VictoriaLogs 的
// stats 查询和 Elasticsearch 的时间桶计数各有独立查询路径，但都复用同一套阈值判定。
func numericDriver(driver string) bool {
	return metricsDriver(driver) || driver == "LOGS_LOKI" || driver == "LOGS_VLOGS" || driver == "LOGS_ELASTICSEARCH"
}
