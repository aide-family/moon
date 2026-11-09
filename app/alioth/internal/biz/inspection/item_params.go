package inspection

import (
	"bytes"
	"encoding/json"
	"math"
	"strings"

	"github.com/aide-family/moon/pkg/merr"
)

// metricsParams 是指标检查项的类型专属配置。
type metricsParams struct {
	Unit        string            `json:"unit,omitempty"`
	EmptyIsWarn bool              `json:"empty_is_warn,omitempty"`
	Thresholds  []metricThreshold `json:"thresholds,omitempty"`
}

type metricThreshold struct {
	Op     string  `json:"op"`
	Value  float64 `json:"value"`
	Status string  `json:"status"`
}

// logsParams 是日志检查项的类型专属配置。
type logsParams struct {
	Levels    []string `json:"levels,omitempty"`
	Limit     int      `json:"limit,omitempty"`
	Direction string   `json:"direction,omitempty"`
	WarnAbove int      `json:"warn_above,omitempty"`
}

// probeParams 是探测检查项的类型专属配置。
type probeParams struct {
	GroupUIDs         []string `json:"group_uids"`
	Metric            string   `json:"metric,omitempty"`
	AvailabilityBelow float64  `json:"availability_below,omitempty"`
	LatencyAboveMS    int      `json:"latency_above_ms,omitempty"`
	CertDaysBelow     int      `json:"cert_days_below,omitempty"`
	WarnOnDown        *bool    `json:"warn_on_down,omitempty"`
}

// alertParams 是告警检查项的类型专属配置。
type alertParams struct {
	Status     string   `json:"status,omitempty"`
	LevelUIDs  []string `json:"level_uids,omitempty"`
	Labels     string   `json:"labels,omitempty"`
	Limit      int      `json:"limit,omitempty"`
	WarnLevels []string `json:"warn_levels,omitempty"`
}

var (
	thresholdOps   = map[string]struct{}{">": {}, ">=": {}, "<": {}, "<=": {}, "==": {}, "!=": {}}
	thresholdState = map[string]struct{}{"WARN": {}, "FAILED": {}}
	logLevels      = map[string]struct{}{"ERROR": {}, "WARN": {}, "INFO": {}, "DEBUG": {}}
	logDirections  = map[string]struct{}{"BACKWARD": {}, "FORWARD": {}}
	probeMetrics   = map[string]struct{}{"AVAILABILITY": {}, "LATENCY": {}, "P95": {}, "CERT_DAYS": {}}
	alertStatuses  = map[string]struct{}{"FIRING": {}, "ALL": {}}
)

// decodeParams 严格解析类型参数：未知键直接拒绝，避免界面写错字段却静默不生效。
func decodeParams(raw string, target any) error {
	decoder := json.NewDecoder(bytes.NewBufferString(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return merr.Invalid("invalid item params: " + err.Error())
	}
	if decoder.More() {
		return merr.Invalid("invalid item params: trailing content")
	}
	return nil
}

// normalizeItemParams 补齐默认值、规范化枚举，并返回可落库的紧凑 JSON。
func normalizeItemParams(kind, raw string) (string, error) {
	switch kind {
	case "METRICS_QUERY":
		var params metricsParams
		if err := decodeParams(raw, &params); err != nil {
			return "", err
		}
		params.Unit = strings.TrimSpace(params.Unit)
		if len([]rune(params.Unit)) > 16 {
			return "", merr.Invalid("metric unit is too long")
		}
		if len(params.Thresholds) > 5 {
			return "", merr.Invalid("too many metric thresholds")
		}
		for index := range params.Thresholds {
			threshold := &params.Thresholds[index]
			if _, ok := thresholdOps[threshold.Op]; !ok {
				return "", merr.Invalid("invalid metric threshold operator")
			}
			if math.IsNaN(threshold.Value) || math.IsInf(threshold.Value, 0) {
				return "", merr.Invalid("invalid metric threshold value")
			}
			threshold.Status = strings.ToUpper(strings.TrimSpace(threshold.Status))
			if _, ok := thresholdState[threshold.Status]; !ok {
				return "", merr.Invalid("invalid metric threshold status")
			}
		}
		return marshalParams(params)

	case "LOGS_QUERY":
		var params logsParams
		if err := decodeParams(raw, &params); err != nil {
			return "", err
		}
		if params.Limit == 0 {
			params.Limit = 50
		}
		if params.Limit < 20 || params.Limit > 200 {
			return "", merr.Invalid("log limit must be between 20 and 200")
		}
		params.Direction = strings.ToUpper(strings.TrimSpace(params.Direction))
		if params.Direction == "" {
			params.Direction = "BACKWARD"
		}
		if _, ok := logDirections[params.Direction]; !ok {
			return "", merr.Invalid("invalid log direction")
		}
		levels := make([]string, 0, len(params.Levels))
		for _, level := range params.Levels {
			level = strings.ToUpper(strings.TrimSpace(level))
			if level == "" {
				continue
			}
			if _, ok := logLevels[level]; !ok {
				return "", merr.Invalid("invalid log level")
			}
			levels = uniqueStrings(append(levels, level))
		}
		params.Levels = levels
		if params.WarnAbove < 0 || params.WarnAbove > 1_000_000 {
			return "", merr.Invalid("log warn threshold is out of range")
		}
		return marshalParams(params)

	case "PROBE_TARGET":
		var params probeParams
		if err := decodeParams(raw, &params); err != nil {
			return "", err
		}
		params.GroupUIDs = uniqueStrings(params.GroupUIDs)
		if len(params.GroupUIDs) == 0 {
			return "", merr.Invalid("probe item requires group_uids")
		}
		params.Metric = strings.ToUpper(strings.TrimSpace(params.Metric))
		if params.Metric == "" {
			params.Metric = "AVAILABILITY"
		}
		if _, ok := probeMetrics[params.Metric]; !ok {
			return "", merr.Invalid("invalid probe metric")
		}
		if params.AvailabilityBelow == 0 {
			params.AvailabilityBelow = 99.5
		}
		if params.AvailabilityBelow < 0 || params.AvailabilityBelow > 100 {
			return "", merr.Invalid("availability threshold must be between 0 and 100")
		}
		if params.LatencyAboveMS < 0 || params.LatencyAboveMS > 600_000 {
			return "", merr.Invalid("latency threshold is out of range")
		}
		if (params.Metric == "LATENCY" || params.Metric == "P95") && params.LatencyAboveMS == 0 {
			params.LatencyAboveMS = 500
		}
		if params.CertDaysBelow < 0 || params.CertDaysBelow > 3650 {
			return "", merr.Invalid("certificate threshold is out of range")
		}
		if params.Metric == "CERT_DAYS" && params.CertDaysBelow == 0 {
			params.CertDaysBelow = 14
		}
		warnOnDown := true
		if params.WarnOnDown != nil {
			warnOnDown = *params.WarnOnDown
		}
		params.WarnOnDown = &warnOnDown
		return marshalParams(params)

	case "ALERT_EVENTS":
		var params alertParams
		if err := decodeParams(raw, &params); err != nil {
			return "", err
		}
		params.Status = strings.ToUpper(strings.TrimSpace(params.Status))
		if params.Status == "" {
			params.Status = "FIRING"
		}
		if _, ok := alertStatuses[params.Status]; !ok {
			return "", merr.Invalid("invalid alert status")
		}
		if params.Limit == 0 {
			params.Limit = 50
		}
		if params.Limit < 20 || params.Limit > 200 {
			return "", merr.Invalid("alert limit must be between 20 and 200")
		}
		params.LevelUIDs = uniqueStrings(params.LevelUIDs)
		warnLevels := make([]string, 0, len(params.WarnLevels))
		for _, level := range params.WarnLevels {
			level = strings.ToUpper(strings.TrimSpace(level))
			if level != "" {
				warnLevels = uniqueStrings(append(warnLevels, level))
			}
		}
		params.WarnLevels = warnLevels
		labels, err := normalizeAlertLabels(params.Labels)
		if err != nil {
			return "", err
		}
		params.Labels = labels
		return marshalParams(params)

	case "NOTE":
		if strings.TrimSpace(raw) != "" && strings.TrimSpace(raw) != "{}" {
			return "", merr.Invalid("note item does not accept params")
		}
		return "{}", nil
	default:
		return "", merr.Invalid("invalid inspection item kind")
	}
}

func marshalParams(value any) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", merr.Invalid("invalid item params")
	}
	return string(raw), nil
}

// normalizeAlertLabels 接受 JSON 对象或 service=order,env=prod 的简写，统一存 JSON。
func normalizeAlertLabels(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	if strings.HasPrefix(raw, "{") {
		var labels map[string]string
		if err := json.Unmarshal([]byte(raw), &labels); err != nil {
			return "", merr.Invalid("invalid alert labels")
		}
		encoded, err := json.Marshal(labels)
		if err != nil {
			return "", merr.Invalid("invalid alert labels")
		}
		return string(encoded), nil
	}
	labels := make(map[string]string)
	for _, pair := range strings.Split(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		key, value, ok := strings.Cut(pair, "=")
		if !ok {
			return "", merr.Invalid("invalid alert labels")
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "" {
			return "", merr.Invalid("invalid alert labels")
		}
		labels[key] = value
	}
	encoded, err := json.Marshal(labels)
	if err != nil {
		return "", merr.Invalid("invalid alert labels")
	}
	return string(encoded), nil
}

func parseMetricsParams(raw string) metricsParams {
	var params metricsParams
	_ = json.Unmarshal([]byte(raw), &params)
	return params
}

func parseLogsParams(raw string) logsParams {
	var params logsParams
	_ = json.Unmarshal([]byte(raw), &params)
	if params.Limit == 0 {
		params.Limit = 50
	}
	if params.Direction == "" {
		params.Direction = "BACKWARD"
	}
	return params
}

func parseProbeParams(raw string) probeParams {
	var params probeParams
	_ = json.Unmarshal([]byte(raw), &params)
	if params.Metric == "" {
		params.Metric = "AVAILABILITY"
	}
	if params.AvailabilityBelow == 0 {
		params.AvailabilityBelow = 99.5
	}
	if params.WarnOnDown == nil {
		warnOnDown := true
		params.WarnOnDown = &warnOnDown
	}
	return params
}

func parseAlertParams(raw string) alertParams {
	var params alertParams
	_ = json.Unmarshal([]byte(raw), &params)
	if params.Status == "" {
		params.Status = "FIRING"
	}
	if params.Limit == 0 {
		params.Limit = 50
	}
	return params
}
