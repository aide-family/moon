package bo_test

import (
	_ "embed"
	"testing"
	"time"

	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/pkg/strategy"
	"github.com/aide-family/moon/pkg/util/hash"
	"github.com/aide-family/moon/pkg/util/times"
)

//go:embed alarm.tmpl
var formatterStr string

func TestAlertFormatter(t *testing.T) {
	now := time.Now()
	alarmInfo := &bo.AlertBo{
		Status: "resolved",
		Labels: &strategy.Labels{
			strategy.MetricInstance: "localhost",
			strategy.MetricAlert:    "test_alert",
			"endpoint":              "127.0.0.1",
			"job":                   "test",
			"severity":              "critical",
			"app":                   "moon",
		},
		Annotations: &strategy.Annotations{
			strategy.MetricSummary:     "test hook template summary",
			strategy.MetricDescription: "test hook template description",
		},
		StartsAt:     now.Add(-time.Minute * 5).Format(times.ParseLayout),
		EndsAt:       now.Format(times.ParseLayout),
		GeneratorURL: "https://github.com/aide-family/moon",
		Fingerprint:  hash.MD5(now.String()),
	}

	t.Log(strategy.Formatter(formatterStr, alarmInfo))
}
