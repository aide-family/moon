package strutil

import (
	"strings"
	"testing"
)

const feishuRichTextAlertTemplate = `{{/* 飞书富文本告警模板 - 适用于 Prometheus Alertmanager */}}
{{- define "feishu.richtext.alert" -}}
{
  "msg_type": "post",
  "content": {
    "post": {
      "zh_cn": {
        "title": {{ printf "%q" (printf "🚨 %s" (.Labels.alertname | default "未知告警")) }},
        "content": [
          [
            {"tag": "text", "text": "【告警状态】"},
            {"tag": "text", "text": {{ printf "%q" (if eq (.Labels.alertstate | default "unknown") "firing" "🔴 触发中" "🟢 已恢复") }} }
          ],
          [
            {"tag": "text", "text": "【严重等级】"},
            {"tag": "text", "text": {{ printf "%q" (printf "%s | %s" (.Labels.severity | default "S1") (.Labels.level_name | default "unknown")) }} }
          ],
          [
            {"tag": "text", "text": "【告警实例】"},
            {"tag": "text", "text": {{ printf "%q" (.Labels.instance | default "unknown") }} }
          ],
          [
            {"tag": "text", "text": "【数据源】"},
            {"tag": "text", "text": {{ printf "%q" (.Labels.datasource_name | default "unknown") }} }
          ],
          [
            {"tag": "text", "text": "【策略分组】"},
            {"tag": "text", "text": {{ printf "%q" (.Labels.strategy_group_name | default "unknown") }} }
          ],
          [
            {"tag": "text", "text": "【策略名称】"},
            {"tag": "text", "text": {{ printf "%q" (.Labels.strategy_name | default "unknown") }} }
          ],
          [
            {"tag": "text", "text": "【事件指纹】"},
            {"tag": "text", "text": {{ printf "%q" (.Labels.fingerprint | default "unknown") }} }
          ]
          {{- if .Labels.alert_event_uid }}
          ,
          [
            {"tag": "text", "text": "\n"},
            {"tag": "a", "text": "👉 点击查看详情", "href": {{ printf "%q" (printf "http://your-alert-dashboard?event_uid=%s" .Labels.alert_event_uid) }} }
          ]
          {{- end }}
        ]
      }
    }
  }
}
{{- end }}`

func TestNormalizeTextTemplateUnwrapDefine(t *testing.T) {
	normalized := NormalizeTextTemplate(feishuRichTextAlertTemplate)
	if strings.Contains(normalized, "define ") {
		t.Fatalf("define block should be removed: %s", normalized)
	}
	if !strings.Contains(normalized, "{{- if .Labels.alert_event_uid }}") {
		t.Fatalf("inner if block should remain: %s", normalized)
	}
	if strings.Contains(normalized, "(if eq ") {
		t.Fatalf("inline if eq should be rewritten: %s", normalized)
	}
	if !strings.Contains(normalized, "(ternary ") {
		t.Fatalf("expected ternary rewrite: %s", normalized)
	}
}

func TestUnwrapTemplateDefineBlockKeepsInnerEnd(t *testing.T) {
	tmpl := `{{- define "x" -}}before{{- if .A }}mid{{- end }}after{{- end }}`
	got := unwrapTemplateDefineBlock(tmpl)
	want := `before{{- if .A }}mid{{- end }}after`
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestExecuteFeishuRichTextAlertTemplate(t *testing.T) {
	data := map[string]any{
		"Labels": map[string]any{
			"alertname":           "HighCPU",
			"alertstate":          "firing",
			"severity":            "critical",
			"level_name":          "P1",
			"instance":            "host-a",
			"datasource_name":     "prom-main",
			"strategy_group_name": "infra",
			"strategy_name":     "cpu-rule",
			"fingerprint":         "fp-1",
			"alert_event_uid":     "12345",
		},
	}
	out, err := ExecuteTextTemplate(feishuRichTextAlertTemplate, data)
	if err != nil {
		t.Fatalf("ExecuteTextTemplate failed: %v", err)
	}
	if !strings.Contains(out, "HighCPU") {
		t.Fatalf("expected alert name in output: %s", out)
	}
	if !strings.Contains(out, "🔴 触发中") {
		t.Fatalf("expected firing status in output: %s", out)
	}
	if !strings.Contains(out, "event_uid=12345") {
		t.Fatalf("expected detail link in output: %s", out)
	}
}

func TestExecuteFeishuRichTextAlertTemplateEmptyLabels(t *testing.T) {
	data := map[string]any{
		"Labels": map[string]any{
			"alert_event_uid": "",
			"alertname":       "",
			"alertstate":      "",
			"severity":        "",
			"level_name":      "",
			"instance":        "",
			"datasource_name": "",
			"strategy_group_name": "",
			"strategy_name":   "",
			"fingerprint":     "",
		},
	}
	out, err := ExecuteTextTemplate(feishuRichTextAlertTemplate, data)
	if err != nil {
		t.Fatalf("ExecuteTextTemplate failed: %v", err)
	}
	if !strings.Contains(out, "未知告警") {
		t.Fatalf("expected default alert name in output: %s", out)
	}
	if !strings.Contains(out, "🟢 已恢复") {
		t.Fatalf("expected recovered status in output: %s", out)
	}
	if strings.Contains(out, "alert_event_uid") {
		t.Fatalf("detail link should be omitted when event uid is empty: %s", out)
	}
}

func TestReplaceInlineIfEqExpressions(t *testing.T) {
	in := `(if eq (.Labels.alertstate | default "unknown") "firing" "🔴 触发中" "🟢 已恢复")`
	out := replaceInlineIfEqExpressions(in)
	want := `(ternary "🔴 触发中" "🟢 已恢复" (eq (.Labels.alertstate | default "unknown") "firing"))`
	if out != want {
		t.Fatalf("got %q, want %q", out, want)
	}
}
