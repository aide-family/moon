package inspection

import (
	"strings"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

func TestSeriesChartRendersAxes(t *testing.T) {
	start := time.Date(2026, 11, 14, 9, 0, 0, 0, time.Local)
	points := []bo.InspectionPoint{
		{Timestamp: start.Unix(), Value: 0},
		{Timestamp: start.Add(10 * time.Minute).Unix(), Value: 50},
		{Timestamp: start.Add(20 * time.Minute).Unix(), Value: 100},
	}
	output := string(seriesChart([]bo.InspectionSeries{{Name: "node-a", Points: points}}))
	for _, want := range []string{
		`text-anchor="end"`, "100", "50.0", "0.00",
		start.Format("01-02 15:04"), start.Add(20 * time.Minute).Format("01-02 15:04"),
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("series chart missing axis label %q: %s", want, output)
		}
	}
}

func TestReportHelperFuncsRegistered(t *testing.T) {
	funcs := reportFuncs()
	for _, name := range []string{"metricCards", "tableHTML", "seriesChart", "logsHTML", "alertsHTML", "sectionHTML"} {
		if _, ok := funcs[name]; !ok {
			t.Fatalf("template helper %q is not registered", name)
		}
	}
}

func TestSectionHTMLRendersAllInspectionData(t *testing.T) {
	section := bo.InspectionSection{
		Title:   "综合检查",
		Kind:    "METRICS_QUERY",
		Status:  "WARN",
		Note:    "窗口内命中 12 条",
		Metrics: []bo.InspectionMetric{{Name: "P95", Value: "1.8", Unit: "s"}},
		Series:  []bo.InspectionSeries{{Name: "up", Points: []bo.InspectionPoint{{Value: 1}, {Value: 0}}}},
		Table:   &bo.InspectionTable{Columns: []string{"目标", "状态"}, Rows: [][]string{{"api", "正常"}}},
		Logs:    []bo.InspectionLogSample{{Level: "ERROR", Message: "timeout"}},
		Alerts:  []bo.InspectionAlertRow{{Title: "P1 告警", Level: "P1"}},
	}
	output := string(sectionHTML(section))
	for _, want := range []string{"综合检查", "P95", "<svg", "api", "timeout", "P1 告警", "WARN"} {
		if !strings.Contains(output, want) {
			t.Fatalf("section html missing %q: %s", want, output)
		}
	}
}
