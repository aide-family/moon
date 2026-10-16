package eval

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/aide-family/Alioth/app/megrez/pkg/bo"
	"github.com/aide-family/Alioth/pkg/fingerprint"
)

// Hits 按级别条件的保存顺序计算命中。第一条满足的级别生效。
func Hits(rule bo.RuleSnapshot, samples []bo.Sample) []bo.EvalHit {
	hits := make([]bo.EvalHit, 0)
	for _, sample := range samples {
		for _, level := range rule.Levels {
			value, raw, ok := matchLevel(level, sample.Points)
			if !ok {
				continue
			}
			_ = value
			labels := sample.Labels
			if labels == nil {
				labels = map[string]string{}
			}
			encoded, err := json.Marshal(labels)
			if err != nil {
				encoded = []byte("{}")
			}
			hits = append(hits, bo.EvalHit{
				Fingerprint: fingerprint.Of(rule.UID, rule.DatasourceUID, labels),
				Labels:      string(encoded),
				Value:       raw,
				LevelUID:    level.LevelUID,
				Summary:     rule.Summary,
				Description: rule.Description,
			})
			break
		}
	}
	return hits
}

func matchLevel(level bo.Level, points []bo.Point) (float64, string, bool) {
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
	seconds := 15
	for _, level := range rule.Levels {
		if level.ForSeconds > seconds {
			seconds = level.ForSeconds
		}
	}
	end := now.UTC().Unix()
	return end - int64(seconds), end, 15
}

func metricsDriver(driver string) bool {
	return driver == "METRICS_PROMETHEUS" || driver == "METRICS_VICTORIA_METRICS"
}
