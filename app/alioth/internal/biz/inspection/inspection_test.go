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

func TestParsePrometheusRange(t *testing.T) {
	raw := `{"status":"success","data":{"resultType":"matrix","result":[{"metric":{"__name__":"up","job":"api"},"values":[[1700000000,"1"],[1700000060,"0.5"]]}]}}`
	series, table, metrics := parsePrometheusRange(raw)
	if len(series) != 1 || series[0].Name != "up{job=api}" || len(series[0].Points) != 2 {
		t.Fatalf("unexpected series: %+v", series)
	}
	if table == nil || len(table.Rows) != 1 || table.Rows[0][1] != "0.5" {
		t.Fatalf("unexpected table: %+v", table)
	}
	if len(metrics) != 1 || metrics[0].Value != "0.5" {
		t.Fatalf("unexpected metrics: %+v", metrics)
	}
}

func TestParseLogSamples(t *testing.T) {
	raw := `{"data":{"result":[{"stream":{"app":"gateway"},"values":[["1700000000000000000","ERROR upstream timeout"],["1700000001000000000","INFO ok"]]}]}}`
	samples, total := parseLogSamples(raw)
	if total != 2 || len(samples) != 2 {
		t.Fatalf("samples=%d total=%d", len(samples), total)
	}
	if samples[0].Level != "ERROR" || samples[1].Level != "INFO" {
		t.Fatalf("unexpected levels: %+v", samples)
	}
}

func TestProbeStats(t *testing.T) {
	checks, ok, latency := probeStats([]bo.ProbeResult{
		{OK: true, DurationMS: 100}, {OK: false, DurationMS: 200}, {OK: true, DurationMS: 300},
	})
	if checks != 3 || ok != 2 || latency != 200 {
		t.Fatalf("checks=%d ok=%d latency=%d", checks, ok, latency)
	}
}

func TestRenderBody(t *testing.T) {
	data := bo.InspectionReportData{
		Title:   "核心业务巡检",
		Tenant:  bo.ReportTenant{Name: "星环科技"},
		Plan:    bo.ReportPlan{Name: "核心业务巡检"},
		Summary: bo.InspectionSummary{Total: 4, OK: 3, Warn: 1, Abnormal: 1},
	}
	html, err := renderBody("HTML", `<h1>{{.Title}}</h1><p>{{.Plan.Name}} · {{.Summary.Total}} 项</p>`, data)
	if err != nil {
		t.Fatalf("render html: %v", err)
	}
	if !strings.Contains(html, "核心业务巡检") || !strings.Contains(html, "4 项") {
		t.Fatalf("unexpected html: %s", html)
	}
	markdown, err := renderBody("MARKDOWN", "# {{.Title}}\n\n{{.Summary.OK}} 正常", data)
	if err != nil {
		t.Fatalf("render markdown: %v", err)
	}
	if !strings.Contains(markdown, "3 正常") {
		t.Fatalf("unexpected markdown: %s", markdown)
	}
	if _, err := renderBody("HTML", `{{if .Title}}`, data); err == nil {
		t.Fatal("broken template should fail")
	}
}

func TestRenderTitle(t *testing.T) {
	data := bo.InspectionReportData{
		Plan: bo.ReportPlan{Name: "核心业务巡检"},
		Run:  bo.ReportRun{FinishedAt: time.Date(2026, 11, 7, 9, 0, 0, 0, time.Local)},
	}
	title, err := renderTitle(`{{.Plan.Name}} · {{.Run.FinishedAt | date "2006-01-02 15:04"}}`, data)
	if err != nil {
		t.Fatalf("render title: %v", err)
	}
	if title != "核心业务巡检 · 2026-11-07 09:00" {
		t.Fatalf("title = %s", title)
	}
	def, err := renderTitle("", data)
	if err != nil || !strings.HasPrefix(def, "核心业务巡检 · 2026-11-07 09:00") {
		t.Fatalf("default title = %s err=%v", def, err)
	}
}

func TestShareToken(t *testing.T) {
	token, err := newShareToken()
	if err != nil {
		t.Fatalf("new token: %v", err)
	}
	if len(token) != 43 {
		t.Fatalf("token length = %d", len(token))
	}
	other, _ := newShareToken()
	if token == other {
		t.Fatal("share tokens must be unique")
	}
	if hashShareToken(token) == token {
		t.Fatal("token hash must not equal the token")
	}
	if hashShareToken(token) != hashShareToken(token) {
		t.Fatal("token hash must be deterministic")
	}
}

func TestRenderMarkdown(t *testing.T) {
	got := renderMarkdown("## 结论\n\n**正常**，共 4 项。")
	if !strings.Contains(got, "<h3>结论</h3>") || !strings.Contains(got, "<strong>正常</strong>") {
		t.Fatalf("unexpected markdown html: %s", got)
	}
}

func TestNormalizeTemplateAllowsReportFunctions(t *testing.T) {
	in := bo.InspectionTemplate{
		Name:         "带函数的模板",
		HTMLBody:     `<p>{{.Run.FinishedAt | date "2006-01-02 15:04"}}</p>`,
		MarkdownBody: `# {{.Plan.Name}}`,
	}
	if err := normalizeTemplate(&in); err != nil {
		t.Fatalf("report functions should be available while validating: %v", err)
	}
}
