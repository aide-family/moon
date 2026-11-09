package inspection

import (
	"strings"
	"testing"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

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
