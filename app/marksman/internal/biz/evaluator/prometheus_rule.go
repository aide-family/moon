package evaluator

import (
	"fmt"
	"maps"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/strutil"
	"github.com/aide-family/magicbox/strutil/cnst"
	"github.com/prometheus/common/model"

	"github.com/aide-family/marksman/internal/biz/bo"
)

const (
	annotationSummary     = cnst.AnnotationKeySummary
	annotationDescription = cnst.AnnotationKeyDescription

	alertStateFiring = "firing"
)

var (
	prometheusLabelsVarPattern = regexp.MustCompile(`\{\{\s*-?\s*\$labels\.([a-zA-Z_][a-zA-Z0-9_]*)\s*\}\}`)
	prometheusValueVarPattern  = regexp.MustCompile(`\{\{\s*-?\s*\$value\s*\}\}`)
)

// prometheusAlertRule mirrors a Prometheus alerting rule (evaluation-time only).
type prometheusAlertRule struct {
	AlertName   string
	Expr        string
	For         time.Duration
	Labels      map[string]string
	Annotations map[string]string
}

// buildPrometheusAlertRule converts stored strategy data into a Prometheus-compatible alerting rule.
func buildPrometheusAlertRule(info *bo.EvaluateMetricStrategyBo) *prometheusAlertRule {
	if info == nil {
		return nil
	}

	window := time.Duration(info.GetDurationSec()) * time.Second
	comparisonExpr := buildComparisonExpr(info.GetExpr(), info.GetCondition(), info.GetValues())
	expr := buildModeExpr(comparisonExpr, info.GetMode(), info.GetValues(), window)

	return &prometheusAlertRule{
		AlertName: info.GetStrategyName(),
		Expr:      expr,
		For:       ruleForDuration(info.GetMode(), window),
		Labels:    buildRuleLabels(info),
		Annotations: map[string]string{
			annotationSummary:     info.GetSummary(),
			annotationDescription: info.GetDescription(),
		},
	}
}

func buildComparisonExpr(baseExpr string, cond enum.ConditionMetric, values []float64) string {
	base := strings.TrimSpace(baseExpr)
	if base == "" {
		return ""
	}
	if cond == enum.ConditionMetric_CONDITION_METRIC_UNKNOWN || len(values) == 0 {
		return base
	}

	switch cond {
	case enum.ConditionMetric_CONDITION_METRIC_EQ:
		return fmt.Sprintf("(%s) == %g", base, values[0])
	case enum.ConditionMetric_CONDITION_METRIC_NE:
		return fmt.Sprintf("(%s) != %g", base, values[0])
	case enum.ConditionMetric_CONDITION_METRIC_GT:
		return fmt.Sprintf("(%s) > %g", base, values[0])
	case enum.ConditionMetric_CONDITION_METRIC_GTE:
		return fmt.Sprintf("(%s) >= %g", base, values[0])
	case enum.ConditionMetric_CONDITION_METRIC_LT:
		return fmt.Sprintf("(%s) < %g", base, values[0])
	case enum.ConditionMetric_CONDITION_METRIC_LTE:
		return fmt.Sprintf("(%s) <= %g", base, values[0])
	case enum.ConditionMetric_CONDITION_METRIC_IN:
		if len(values) < 2 {
			return base
		}
		return fmt.Sprintf("((%s) >= %g and (%s) <= %g)", base, values[0], base, values[1])
	case enum.ConditionMetric_CONDITION_METRIC_NOT_IN:
		if len(values) < 2 {
			return base
		}
		return fmt.Sprintf("((%s) < %g or (%s) > %g)", base, values[0], base, values[1])
	default:
		return base
	}
}

func buildModeExpr(comparisonExpr string, mode enum.SampleMode, values []float64, window time.Duration) string {
	if comparisonExpr == "" {
		return ""
	}
	switch mode {
	case enum.SampleMode_SAMPLE_MODE_MAX:
		n := legacySampleCount(values)
		return fmt.Sprintf(`count_over_time((%s)[%s:]) > %d`, comparisonExpr, promDurationString(window), n)
	case enum.SampleMode_SAMPLE_MODE_MIN:
		n := legacySampleCount(values)
		if n <= 0 {
			n = 1
		}
		return fmt.Sprintf(`count_over_time((%s)[%s:]) >= %d`, comparisonExpr, promDurationString(window), n)
	default:
		return comparisonExpr
	}
}

func ruleForDuration(mode enum.SampleMode, window time.Duration) time.Duration {
	switch mode {
	case enum.SampleMode_SAMPLE_MODE_MAX, enum.SampleMode_SAMPLE_MODE_MIN:
		return 0
	default:
		return window
	}
}

func legacySampleCount(values []float64) int {
	if len(values) <= 1 {
		return 0
	}
	n := int(values[1])
	if n < 0 {
		return 0
	}
	return n
}

func buildRuleLabels(info *bo.EvaluateMetricStrategyBo) map[string]string {
	labels := map[string]string{
		cnst.LabelAlertName:           info.GetStrategyName(),
		cnst.LabelSeverity:            info.GetLevelName(),
		cnst.LabelNamespaceUID:        strconv.FormatInt(info.GetNamespaceUID().Int64(), 10),
		cnst.LabelStrategyGroupUID:    strconv.FormatInt(info.GetStrategyGroupUID().Int64(), 10),
		cnst.LabelStrategyGroupName:   info.GetStrategyGroupName(),
		cnst.LabelStrategyUID:         strconv.FormatInt(info.GetStrategyUID().Int64(), 10),
		cnst.LabelLevelUID:            strconv.FormatInt(info.GetLevelUID().Int64(), 10),
		cnst.LabelDatasourceUID:       strconv.FormatInt(info.GetDatasourceUID().Int64(), 10),
		cnst.LabelDatasourceName:      info.GetDatasourceName(),
		cnst.LabelDatasourceLevelName: info.GetDatasourceLevelName(),
	}
	for k, v := range info.GetLabels() {
		if bo.IsReservedAlertSystemLabelKey(k) {
			continue
		}
		labels[k] = v
	}
	return labels
}

func promDurationString(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}
	if d%time.Second == 0 {
		return fmt.Sprintf("%ds", int(d/time.Second))
	}
	return d.String()
}

func sampleStreamLabels(series *model.SampleStream) map[string]string {
	if series == nil {
		return map[string]string{}
	}
	labels := make(map[string]string, len(series.Metric))
	for k, v := range series.Metric {
		labels[string(k)] = string(v)
	}
	return labels
}

// mergePrometheusAlertLabels merges query labels with rule labels; rule labels take precedence.
func mergePrometheusAlertLabels(seriesLabels, ruleLabels map[string]string) map[string]string {
	merged := maps.Clone(seriesLabels)
	for k, v := range ruleLabels {
		merged[k] = v
	}
	merged[cnst.LabelAlertState] = alertStateFiring
	return merged
}

func expandRuleLabels(ruleLabels, seriesLabels map[string]string) map[string]string {
	expanded := make(map[string]string, len(ruleLabels))
	for k, v := range ruleLabels {
		expanded[k] = executePrometheusTemplate(v, seriesLabels, 0)
	}
	return expanded
}

func executePrometheusTemplate(tmpl string, labels map[string]string, value float64) string {
	if tmpl == "" {
		return ""
	}
	normalized := normalizePrometheusTemplate(tmpl)
	data := &prometheusTemplateData{
		Labels: labels,
		Value:  value,
	}
	out, err := strutil.ExecuteTextTemplate(normalized, data)
	if err != nil {
		return tmpl
	}
	return out
}

type prometheusTemplateData struct {
	Labels map[string]string
	Value  float64
}

func normalizePrometheusTemplate(tmpl string) string {
	out := prometheusLabelsVarPattern.ReplaceAllString(tmpl, `{{ index .Labels "$1" }}`)
	out = prometheusValueVarPattern.ReplaceAllString(out, `{{ .Value }}`)
	return out
}

func prometheusAlertFingerprint(labels map[string]string) string {
	filtered := bo.FilterLabelsForAlertFingerprint(labels)
	ls := make(model.LabelSet, len(filtered))
	for k, v := range filtered {
		ls[model.LabelName(k)] = model.LabelValue(v)
	}
	return ls.Fingerprint().String()
}

type firingSeries struct {
	labels map[string]string
	value  float64
}

func collectFiringSeries(matrix model.Matrix, forDur, evalInterval time.Duration, end time.Time) []firingSeries {
	if len(matrix) == 0 {
		return nil
	}
	out := make([]firingSeries, 0, len(matrix))
	for _, series := range matrix {
		if ok, value := isSeriesFiring(series, forDur, evalInterval, end); ok {
			out = append(out, firingSeries{
				labels: sampleStreamLabels(series),
				value:  value,
			})
		}
	}
	return out
}

func isSeriesFiring(series *model.SampleStream, forDur, evalInterval time.Duration, end time.Time) (bool, float64) {
	if series == nil || len(series.Values) == 0 {
		return false, 0
	}

	last := series.Values[len(series.Values)-1]
	lastVal := float64(last.Value)
	if math.IsNaN(lastVal) {
		return false, 0
	}
	if forDur <= 0 {
		return true, lastVal
	}

	cutoff := end.Add(-forDur)
	activeStart := time.Time{}
	for i := len(series.Values) - 1; i >= 0; i-- {
		p := series.Values[i]
		ts := p.Timestamp.Time()
		if ts.After(end) {
			continue
		}
		val := float64(p.Value)
		if math.IsNaN(val) {
			break
		}
		activeStart = ts
		if ts.Before(cutoff) {
			break
		}
	}
	if activeStart.IsZero() {
		return false, 0
	}
	if end.Sub(activeStart)+evalInterval/2 < forDur {
		return false, 0
	}
	return true, lastVal
}
