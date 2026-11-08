package inspection

import (
	"strings"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

func TestNextRun(t *testing.T) {
	from := time.Date(2026, 11, 8, 10, 30, 0, 0, time.UTC)
	next, err := nextRun("0 9 * * *", "Asia/Shanghai", from)
	if err != nil {
		t.Fatalf("nextRun: %v", err)
	}
	if got := next.In(time.FixedZone("CST", 8*3600)).Format("2006-01-02 15:04"); got != "2026-11-09 09:00" {
		t.Fatalf("next run = %s", got)
	}
}

func TestNextRunRejectsInvalid(t *testing.T) {
	from := time.Now()
	for _, expression := range []string{"", "0 9 * *", "not-a-cron"} {
		if _, err := nextRun(expression, "Asia/Shanghai", from); err == nil {
			t.Fatalf("cron %q should be rejected", expression)
		}
	}
	if _, err := nextRun("0 9 * * *", "Mars/Olympus", from); err == nil {
		t.Fatal("invalid timezone should be rejected")
	}
	// 每分钟执行会把控制面打满，最短间隔要求 5 分钟。
	if _, err := nextRun("* * * * *", "Asia/Shanghai", from); err == nil {
		t.Fatal("too frequent schedule should be rejected")
	}
}

func TestNormalizeItem(t *testing.T) {
	item := bo.InspectionItem{Kind: "metrics_query", Title: " 核心指标 ", DatasourceUID: "ds-1", Query: "up"}
	if err := normalizeItem(&item, 3); err != nil {
		t.Fatalf("normalize item: %v", err)
	}
	if item.Kind != "METRICS_QUERY" || item.Title != "核心指标" || item.Sort != 3 {
		t.Fatalf("unexpected normalized item: %+v", item)
	}
	if item.WindowSeconds != 3600 || item.StepSeconds != 60 || item.Display != "TABLE" {
		t.Fatalf("defaults not applied: %+v", item)
	}

	note := bo.InspectionItem{Kind: "NOTE", Title: "说明", Query: "## 背景"}
	if err := normalizeItem(&note, 0); err != nil {
		t.Fatalf("normalize note: %v", err)
	}
	if note.Display != "TEXT" {
		t.Fatalf("note display = %s", note.Display)
	}

	missingSource := bo.InspectionItem{Kind: "LOGS_QUERY", Title: "日志", Query: "{}"}
	if err := normalizeItem(&missingSource, 0); err == nil {
		t.Fatal("logs item without datasource should be rejected")
	}
	badParams := bo.InspectionItem{Kind: "PROBE_TARGET", Title: "探测", Params: "[]"}
	if err := normalizeItem(&badParams, 0); err == nil {
		t.Fatal("non-object params should be rejected")
	}
}

func TestNormalizedFormats(t *testing.T) {
	got := normalizedFormats([]string{"html", " HTML ", "markdown", "pdf"})
	if strings.Join(got, ",") != "HTML,MARKDOWN" {
		t.Fatalf("formats = %v", got)
	}
}

func TestNormalizeTemplate(t *testing.T) {
	item := bo.InspectionTemplate{
		Name:         " 每日报告 ",
		HTMLBody:     `<h1>{{.Plan.Name}}</h1>`,
		MarkdownBody: "# {{.Plan.Name}}",
	}
	if err := normalizeTemplate(&item); err != nil {
		t.Fatalf("normalize template: %v", err)
	}
	if item.Name != "每日报告" || item.Status != "ACTIVE" {
		t.Fatalf("unexpected template: %+v", item)
	}

	broken := bo.InspectionTemplate{Name: "坏模板", HTMLBody: `{{if .Plan}}`}
	if err := normalizeTemplate(&broken); err == nil {
		t.Fatal("broken html template should be rejected")
	}
	empty := bo.InspectionTemplate{Name: "空模板"}
	if err := normalizeTemplate(&empty); err == nil {
		t.Fatal("template without bodies should be rejected")
	}
}
