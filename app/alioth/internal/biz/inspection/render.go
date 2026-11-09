package inspection

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	htmltmpl "html/template"
	"strconv"
	"strings"
	texttmpl "text/template"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

const (
	renderTimeout = 10 * time.Second
	maxReportSize = 2 * 1024 * 1024
	reportRetain  = 90 * 24 * time.Hour
)

// persistReports 按计划绑定的模板格式渲染检查项结果并落库。
// 渲染失败会让整次执行标记为 FAILED，避免投递半份报告。
func (gateway *Gateway) persistReports(ctx context.Context, plan bo.InspectionPlan, run bo.InspectionRun, sections []bo.InspectionSection) error {
	tenantName := ""
	if ns, err := gateway.namespaces.GetNamespace(ctx, run.NamespaceUID); err == nil && ns != nil {
		tenantName = ns.Name
	}
	data := reportData(plan, run, sections, tenantName)
	title, err := renderTitle(plan.TitleTemplate, data)
	if err != nil {
		return err
	}
	data.Title = title
	template, err := gateway.repo.GetInspectionTemplate(ctx, run.NamespaceUID, plan.TemplateUID)
	if err != nil {
		return gateway.access.Wrap(err)
	}
	format := template.Format
	body := template.HTMLBody
	if format == "MARKDOWN" {
		body = template.MarkdownBody
	}
	if strings.TrimSpace(body) == "" {
		return merr.Invalid("template body is required for " + format)
	}
	content, err := renderBody(format, body, data)
	if err != nil {
		return err
	}
	if len(content) > maxReportSize {
		return merr.Invalid("report exceeds 2 MB")
	}
	sum := sha256.Sum256([]byte(content))
	if _, err := gateway.repo.CreateInspectionReport(ctx, bo.InspectionReport{
		NamespaceUID: run.NamespaceUID, RunUID: run.UID, Format: format,
		Title: title, Content: content, Size: len(content), Checksum: hex.EncodeToString(sum[:]),
	}); err != nil {
		return err
	}
	return nil
}

func reportData(plan bo.InspectionPlan, run bo.InspectionRun, sections []bo.InspectionSection, tenantName string) bo.InspectionReportData {
	summary := bo.InspectionSummary{}
	_ = json.Unmarshal([]byte(run.Summary), &summary)
	abnormal := make([]bo.InspectionSection, 0, len(sections))
	for _, section := range sections {
		if section.Status != "OK" {
			abnormal = append(abnormal, section)
		}
	}
	started := run.ScheduledAt
	if run.StartedAt != nil {
		started = *run.StartedAt
	}
	finished := started
	if run.FinishedAt != nil {
		finished = *run.FinishedAt
	}
	windowStart, windowEnd := started, run.ScheduledAt
	if run.WindowStart != nil {
		windowStart = *run.WindowStart
	}
	if run.WindowEnd != nil {
		windowEnd = *run.WindowEnd
	}
	return bo.InspectionReportData{
		GeneratedAt: time.Now(),
		Tenant:      bo.ReportTenant{Name: tenantName},
		Plan:        bo.ReportPlan{Name: plan.Name, Description: plan.Description, Labels: plan.Labels},
		Run: bo.ReportRun{
			Trigger: run.Trigger, StartedAt: started, FinishedAt: finished,
			WindowStart: windowStart, WindowEnd: windowEnd,
			Duration: time.Since(started).Round(time.Second).String(), OperatorName: run.OperatorName,
		},
		Summary:     summary,
		Sections:    sections,
		Abnormal:    abnormal,
		HasAbnormal: len(abnormal) > 0,
	}
}

func renderTitle(raw string, data bo.InspectionReportData) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return data.Plan.Name + " · " + data.Run.FinishedAt.Local().Format("2006-01-02 15:04"), nil
	}
	tmpl, err := texttmpl.New("title").Funcs(reportFuncs()).Parse(raw)
	if err != nil {
		return "", merr.Invalid("invalid title template: " + err.Error())
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", merr.Invalid("render title failed: " + err.Error())
	}
	return buf.String(), nil
}

func renderBody(format, body string, data bo.InspectionReportData) (string, error) {
	type rendered struct {
		content string
		err     error
	}
	done := make(chan rendered, 1)
	go func() {
		var buf bytes.Buffer
		if format == "HTML" {
			tmpl, err := htmltmpl.New("report").Funcs(reportFuncs()).Parse(body)
			if err != nil {
				done <- rendered{err: merr.Invalid("invalid html template: " + err.Error())}
				return
			}
			if err := tmpl.Execute(&buf, data); err != nil {
				done <- rendered{err: merr.Invalid("render html failed: " + err.Error())}
				return
			}
		} else {
			tmpl, err := texttmpl.New("report").Funcs(reportFuncs()).Parse(body)
			if err != nil {
				done <- rendered{err: merr.Invalid("invalid markdown template: " + err.Error())}
				return
			}
			if err := tmpl.Execute(&buf, data); err != nil {
				done <- rendered{err: merr.Invalid("render markdown failed: " + err.Error())}
				return
			}
		}
		done <- rendered{content: buf.String()}
	}()
	select {
	case result := <-done:
		if result.err != nil {
			return "", result.err
		}
		return result.content, nil
	case <-time.After(renderTimeout):
		return "", merr.Invalid("report render timed out")
	}
}

func reportFuncs() map[string]any {
	return map[string]any{
		"date": func(layout string, value any) string {
			switch typed := value.(type) {
			case time.Time:
				return typed.Local().Format(layout)
			case *time.Time:
				if typed != nil {
					return typed.Local().Format(layout)
				}
			case string:
				return typed
			}
			return ""
		},
		"duration": func(value time.Duration) string { return value.Round(time.Second).String() },
		"percent":  func(value float64) string { return fmt.Sprintf("%.2f%%", value) },
		"default": func(fallback, value any) any {
			if value == nil || value == "" || value == 0 {
				return fallback
			}
			return value
		},
		"join": func(values []string, sep string) string { return strings.Join(values, sep) },
		"markdown": func(raw string) htmltmpl.HTML {
			return htmltmpl.HTML(renderMarkdown(raw))
		},
		"metricCards": func(metrics []bo.InspectionMetric) htmltmpl.HTML { return metricCards(metrics) },
		"tableHTML":   func(table *bo.InspectionTable) htmltmpl.HTML { return tableHTML(table) },
		"seriesChart": func(series []bo.InspectionSeries) htmltmpl.HTML { return seriesChart(series) },
		"logsHTML":    func(logs []bo.InspectionLogSample) htmltmpl.HTML { return logsHTML(logs) },
		"alertsHTML":  func(alerts []bo.InspectionAlertRow) htmltmpl.HTML { return alertsHTML(alerts) },
		"sectionHTML": func(section bo.InspectionSection) htmltmpl.HTML { return sectionHTML(section) },
		"json": func(value any) string {
			out, _ := json.Marshal(value)
			return string(out)
		},
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"mul": func(a, b int) int { return a * b },
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
	}
}

// renderMarkdown 是报告内说明字段的最小 Markdown 渲染：标题、加粗、行内代码与段落。
// 报告模板仍由 html/template 负责外层结构，这里只处理说明文本。
func renderMarkdown(raw string) string {
	lines := strings.Split(html.EscapeString(raw), "\n")
	var out strings.Builder
	paragraph := make([]string, 0, 2)
	flush := func() {
		if len(paragraph) == 0 {
			return
		}
		out.WriteString("<p style=\"margin:0 0 8px\">" + strings.Join(paragraph, " ") + "</p>")
		paragraph = paragraph[:0]
	}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "### "):
			flush()
			out.WriteString("<h4>" + inlineMarkdown(strings.TrimPrefix(trimmed, "### ")) + "</h4>")
		case strings.HasPrefix(trimmed, "## "):
			flush()
			out.WriteString("<h3>" + inlineMarkdown(strings.TrimPrefix(trimmed, "## ")) + "</h3>")
		case strings.HasPrefix(trimmed, "# "):
			flush()
			out.WriteString("<h2>" + inlineMarkdown(strings.TrimPrefix(trimmed, "# ")) + "</h2>")
		case trimmed == "":
			flush()
		default:
			paragraph = append(paragraph, inlineMarkdown(trimmed))
		}
	}
	flush()
	return out.String()
}

func inlineMarkdown(text string) string {
	parts := strings.Split(text, "**")
	for index := 1; index < len(parts); index += 2 {
		parts[index] = "<strong>" + parts[index] + "</strong>"
	}
	return strings.Join(parts, "")
}

// ListReports 返回一次执行的全部报告。
func (gateway *Gateway) ListReports(ctx context.Context, runUID string) ([]bo.InspectionReport, error) {
	ns, err := gateway.access.Allow(ctx, perm.InspectionRead)
	if err != nil {
		return nil, err
	}
	items, err := gateway.repo.ListInspectionReportsByRun(ctx, ns, runUID)
	return items, gateway.access.Wrap(err)
}

// GetReport 读取一份报告快照。
func (gateway *Gateway) GetReport(ctx context.Context, uid string) (*bo.InspectionReport, error) {
	ns, err := gateway.access.Allow(ctx, perm.InspectionRead)
	if err != nil {
		return nil, err
	}
	item, err := gateway.repo.GetInspectionReport(ctx, ns, uid)
	return item, gateway.access.Wrap(err)
}

// PruneReports 删除超过保留期的报告，由调度器每小时调用。
func (gateway *Gateway) PruneReports(ctx context.Context) (int, error) {
	return gateway.repo.PruneInspectionReports(ctx, time.Now().UTC().Add(-reportRetain))
}

// metricCards 渲染指标卡网格，供自定义 HTML 模板直接使用。
func metricCards(metrics []bo.InspectionMetric) htmltmpl.HTML {
	if len(metrics) == 0 {
		return ""
	}
	var out strings.Builder
	out.WriteString(`<div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(150px,1fr));gap:10px;margin:12px 0">`)
	for _, metric := range metrics {
		out.WriteString(`<div style="border:1px solid #e2e8f0;border-radius:10px;padding:10px 12px;background:#fff">`)
		out.WriteString(`<div style="font-size:11px;color:#64748b;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">`)
		out.WriteString(html.EscapeString(metric.Name))
		out.WriteString(`</div><div style="font-family:ui-monospace,monospace;font-size:20px;font-weight:600;margin-top:2px">`)
		out.WriteString(html.EscapeString(metric.Value))
		if metric.Unit != "" {
			out.WriteString(`<span style="font-size:11px;color:#64748b;font-weight:400;margin-left:3px">`)
			out.WriteString(html.EscapeString(metric.Unit))
			out.WriteString(`</span>`)
		}
		out.WriteString(`</div></div>`)
	}
	out.WriteString(`</div>`)
	return htmltmpl.HTML(out.String())
}

// tableHTML 渲染检查项表格。
func tableHTML(table *bo.InspectionTable) htmltmpl.HTML {
	if table == nil || len(table.Columns) == 0 {
		return ""
	}
	var out strings.Builder
	out.WriteString(`<div style="overflow:auto;border:1px solid #e2e8f0;border-radius:10px;margin:12px 0"><table style="width:100%;border-collapse:collapse;font-size:12px">`)
	out.WriteString(`<thead><tr style="background:#f8fafc;color:#64748b">`)
	for _, column := range table.Columns {
		out.WriteString(`<th style="text-align:left;padding:7px 9px;font-weight:500;white-space:nowrap;border-bottom:1px solid #e2e8f0">`)
		out.WriteString(html.EscapeString(column))
		out.WriteString(`</th>`)
	}
	out.WriteString(`</tr></thead><tbody>`)
	for _, row := range table.Rows {
		out.WriteString(`<tr>`)
		for _, cell := range row {
			out.WriteString(`<td style="padding:7px 9px;border-bottom:1px solid #f1f5f9;vertical-align:top;word-break:break-word">`)
			out.WriteString(html.EscapeString(cell))
			out.WriteString(`</td>`)
		}
		out.WriteString(`</tr>`)
	}
	out.WriteString(`</tbody></table></div>`)
	return htmltmpl.HTML(out.String())
}

// seriesChart 渲染多序列趋势 SVG；模板不需要引入前端图表库。
func seriesChart(series []bo.InspectionSeries) htmltmpl.HTML {
	if len(series) == 0 {
		return ""
	}
	const width, height, padding = 720.0, 190.0, 26.0
	min, max := 0.0, 0.0
	first := true
	for _, item := range series {
		for _, point := range item.Points {
			if first {
				min, max, first = point.Value, point.Value, false
			} else {
				if point.Value < min {
					min = point.Value
				}
				if point.Value > max {
					max = point.Value
				}
			}
		}
	}
	if first {
		return ""
	}
	span := max - min
	if span == 0 {
		span = 1
	}
	colors := []string{"#1677ff", "#db7a0e", "#047857", "#7c3aed", "#dc2626", "#0891b2"}
	var out strings.Builder
	out.WriteString(`<div style="border:1px solid #e2e8f0;border-radius:10px;padding:12px;margin:12px 0;background:#fff">`)
	out.WriteString(`<svg viewBox="0 0 720 190" style="width:100%;height:auto;display:block" role="img">`)
	for line := 0; line <= 4; line++ {
		y := padding + (height-2*padding)*float64(line)/4
		out.WriteString(`<line x1="26" y1="` + formatFloat(y) + `" x2="694" y2="` + formatFloat(y) + `" stroke="#f1f5f9" stroke-width="1"/>`)
	}
	limit := len(series)
	if limit > 6 {
		limit = 6
	}
	for index := 0; index < limit; index++ {
		item := series[index]
		if len(item.Points) < 2 {
			continue
		}
		out.WriteString(`<path d="`)
		for pointIndex, point := range item.Points {
			x := padding + (width-2*padding)*float64(pointIndex)/float64(len(item.Points)-1)
			y := height - padding - (point.Value-min)/span*(height-2*padding)
			if pointIndex == 0 {
				out.WriteString("M")
			} else {
				out.WriteString(" L")
			}
			out.WriteString(formatFloat(x) + " " + formatFloat(y))
		}
		out.WriteString(`" fill="none" stroke="` + colors[index%len(colors)] + `" stroke-width="2" stroke-linejoin="round"/>`)
		out.WriteString(`<text x="26" y="` + formatFloat(16+float64(index)*14) + `" fill="` + colors[index%len(colors)] + `" font-size="10">` + html.EscapeString(truncateText(item.Name, 56)) + `</text>`)
	}
	out.WriteString(`</svg></div>`)
	return htmltmpl.HTML(out.String())
}

// logsHTML 渲染日志样本列表。
func logsHTML(logs []bo.InspectionLogSample) htmltmpl.HTML {
	if len(logs) == 0 {
		return ""
	}
	var out strings.Builder
	out.WriteString(`<div style="border:1px solid #e2e8f0;border-radius:10px;margin:12px 0;overflow:hidden;font-family:ui-monospace,monospace;font-size:11px">`)
	for _, log := range logs {
		color := "#64748b"
		if log.Level == "ERROR" {
			color = "#dc2626"
		} else if log.Level == "WARN" {
			color = "#db7a0e"
		}
		out.WriteString(`<div style="display:grid;grid-template-columns:90px 52px minmax(0,1fr);gap:8px;padding:6px 9px;border-bottom:1px solid #f1f5f9">`)
		out.WriteString(`<span style="color:#94a3b8">` + html.EscapeString(log.Time) + `</span>`)
		out.WriteString(`<span style="color:` + color + `;font-weight:600">` + html.EscapeString(log.Level) + `</span>`)
		out.WriteString(`<span style="word-break:break-word">` + html.EscapeString(log.Message) + `</span></div>`)
	}
	out.WriteString(`</div>`)
	return htmltmpl.HTML(out.String())
}

// alertsHTML 渲染告警事件表。
func alertsHTML(alerts []bo.InspectionAlertRow) htmltmpl.HTML {
	if len(alerts) == 0 {
		return ""
	}
	var out strings.Builder
	out.WriteString(`<div style="overflow:auto;border:1px solid #e2e8f0;border-radius:10px;margin:12px 0"><table style="width:100%;border-collapse:collapse;font-size:12px">`)
	out.WriteString(`<thead><tr style="background:#f8fafc;color:#64748b"><th style="text-align:left;padding:7px 9px">摘要</th><th style="text-align:left;padding:7px 9px">级别</th><th style="text-align:left;padding:7px 9px">状态</th><th style="text-align:left;padding:7px 9px">开始</th></tr></thead><tbody>`)
	for _, alert := range alerts {
		out.WriteString(`<tr><td style="padding:7px 9px;border-bottom:1px solid #f1f5f9">` + html.EscapeString(alert.Title) + `</td><td style="padding:7px 9px;border-bottom:1px solid #f1f5f9">` + html.EscapeString(alert.Level) + `</td><td style="padding:7px 9px;border-bottom:1px solid #f1f5f9">` + html.EscapeString(alert.Status) + `</td><td style="padding:7px 9px;border-bottom:1px solid #f1f5f9;white-space:nowrap">` + html.EscapeString(alert.StartedAt) + `</td></tr>`)
	}
	out.WriteString(`</tbody></table></div>`)
	return htmltmpl.HTML(out.String())
}

// sectionHTML 把一个检查项的全部结果按类型渲染，模板可逐章节使用。
func sectionHTML(section bo.InspectionSection) htmltmpl.HTML {
	var out strings.Builder
	color := "#047857"
	if section.Status == "WARN" {
		color = "#db7a0e"
	} else if section.Status == "FAILED" {
		color = "#dc2626"
	}
	out.WriteString(`<section style="margin:26px 0;padding-top:6px">`)
	out.WriteString(`<div style="display:flex;align-items:center;gap:10px;border-bottom:1px solid #e2e8f0;padding-bottom:8px"><h2 style="margin:0;font-size:17px">` + html.EscapeString(section.Title) + `</h2><span style="margin-left:auto;font-size:11px;color:` + color + `;border:1px solid ` + color + `33;background:` + color + `14;padding:2px 8px;border-radius:999px">` + html.EscapeString(section.Status) + `</span></div>`)
	if section.Error != "" {
		out.WriteString(`<p style="color:#dc2626;font-size:12px">` + html.EscapeString(section.Error) + `</p>`)
	} else if section.Note != "" {
		out.WriteString(`<div style="font-size:12px;color:#475569;margin-top:8px">` + renderMarkdown(section.Note) + `</div>`)
	}
	out.WriteString(string(metricCards(section.Metrics)))
	out.WriteString(string(seriesChart(section.Series)))
	out.WriteString(string(tableHTML(section.Table)))
	out.WriteString(string(logsHTML(section.Logs)))
	out.WriteString(string(alertsHTML(section.Alerts)))
	if section.Raw != "" {
		out.WriteString(`<details style="font-size:11px;color:#64748b"><summary>原始响应</summary><pre style="white-space:pre-wrap;word-break:break-all;background:#f8fafc;padding:8px;border-radius:8px">` + html.EscapeString(truncateText(section.Raw, 4000)) + `</pre></details>`)
	}
	out.WriteString(`</section>`)
	return htmltmpl.HTML(out.String())
}

func formatFloat(value float64) string { return strconv.FormatFloat(value, 'f', 1, 64) }

func truncateText(value string, max int) string {
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max]) + "…"
}
