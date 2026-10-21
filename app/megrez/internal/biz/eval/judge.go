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
	"github.com/aide-family/Alioth/pkg/fingerprint"
	"github.com/prometheus/common/model"
	promlabels "github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql"
	promtemplate "github.com/prometheus/prometheus/template"
)

// TemplateQuery executes an instant query for Prometheus' query template function.
type TemplateQuery func(context.Context, bo.RuleSnapshot, string, time.Time) ([]bo.Sample, error)

// Hits 按级别条件的保存顺序计算命中。第一条满足的级别生效。
func Hits(rule bo.RuleSnapshot, samples []bo.Sample) []bo.EvalHit {
	return HitsWithQuery(context.Background(), rule, samples, rule.EvalSlot, nil)
}

// HitsWithQuery evaluates hits and expands templates with Prometheus' alert template semantics.
func HitsWithQuery(ctx context.Context, rule bo.RuleSnapshot, samples []bo.Sample, at time.Time, query TemplateQuery) []bo.EvalHit {
	if at.IsZero() {
		at = time.Now()
	}
	var queryFunc promtemplate.QueryFunc
	if query != nil {
		queryFunc = func(ctx context.Context, expr string, queryAt time.Time) (promql.Vector, error) {
			queried, err := query(ctx, rule, expr, queryAt)
			if err != nil {
				return nil, err
			}
			vector := make(promql.Vector, 0, len(queried))
			for _, item := range queried {
				if len(item.Points) == 0 {
					continue
				}
				point := item.Points[len(item.Points)-1]
				vector = append(vector, promql.Sample{Metric: promlabels.FromMap(item.Labels), F: point.Value})
			}
			return vector, nil
		}
	}

	hits := make([]bo.EvalHit, 0)
	for _, sample := range samples {
		for _, level := range rule.Levels {
			value, raw, ok := matchLevel(level, sample.Points)
			if !ok {
				continue
			}
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
				Summary:     renderAlert(ctx, rule, rule.Summary, value, labels, string(encoded), at, queryFunc),
				Description: renderAlert(ctx, rule, rule.Description, value, labels, string(encoded), at, queryFunc),
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

// alertTemplateData mirrors Prometheus' AlertTemplateData and retains lowercase
// accessors for Alioth notification-style templates. Lowercase labels is the same
// JSON string stored on the event; native $labels/.Labels remains a map.
type alertTemplateData map[string]any

func (a alertTemplateData) String() string {
	return fmt.Sprint(struct {
		Labels         map[string]string
		ExternalLabels map[string]string
		ExternalURL    string
		Value          float64
	}{
		Labels: a["Labels"].(map[string]string), ExternalLabels: a["ExternalLabels"].(map[string]string),
		ExternalURL: a["ExternalURL"].(string), Value: a["Value"].(float64),
	})
}

func renderAlert(ctx context.Context, rule bo.RuleSnapshot, raw string, value float64, labels map[string]string, labelsJSON string, at time.Time, queryFunc promtemplate.QueryFunc) string {
	if raw == "" || !strings.Contains(raw, "{{") {
		return raw
	}
	externalURL, _ := url.Parse(rule.Endpoint)
	if externalURL == nil {
		externalURL = &url.URL{}
	}
	externalURL.User = nil
	externalLabels := map[string]string{}
	data := alertTemplateData{
		"Labels": labels, "ExternalLabels": externalLabels, "ExternalURL": externalURL.String(), "Value": value,
		"labels": labelsJSON, "externalLabels": externalLabels, "externalURL": externalURL.String(), "value": value,
	}
	defs := "{{$labels := .Labels}}{{$externalLabels := .ExternalLabels}}{{$externalURL := .ExternalURL}}{{$value := .Value}}"
	expander := promtemplate.NewTemplateExpander(
		ctx, defs+raw, "__alert_"+rule.Name, data,
		model.Time(at.UnixMilli()), queryFunc, externalURL, nil,
	)
	result, err := expander.Expand()
	if err != nil {
		return fmt.Sprintf("<error expanding template: %s>", err)
	}
	return result
}

func metricsDriver(driver string) bool {
	return driver == "METRICS_PROMETHEUS" || driver == "METRICS_VICTORIA_METRICS"
}
