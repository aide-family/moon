package inspection

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	htmltmpl "html/template"
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

// persistReports 把本次执行的检查项结果渲染成勾选的格式并落库。
// 任一格式渲染失败都会让整次执行标记为 FAILED，避免投递半份报告。
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
	for _, format := range plan.OutputFormats {
		body := ""
		if format == "HTML" {
			body = template.HTMLBody
		} else {
			body = template.MarkdownBody
		}
		if strings.TrimSpace(body) == "" {
			builtin, err := gateway.repo.FirstBuiltinInspectionTemplate(ctx, format)
			if err != nil {
				return err
			}
			if format == "HTML" {
				body = builtin.HTMLBody
			} else {
				body = builtin.MarkdownBody
			}
		}
		if strings.TrimSpace(body) == "" {
			return merr.Invalid("no " + format + " template body")
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
