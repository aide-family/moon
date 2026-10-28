package notify

import (
	"context"
	"testing"

	"github.com/aide-family/moon/app/merak/pkg/bo"
)

func TestDeliverReturnsRenderedContentForAudit(t *testing.T) {
	result := Deliver(context.Background(), bo.TaskSnapshot{
		ChannelType:     "UNSUPPORTED",
		ChannelApp:      "OTHER",
		TemplateType:    "EMAIL",
		TemplateSubject: "Alert: {{.name}}",
		TemplateBody:    "Host {{.name}} is {{.state}}",
		Params:          `{"name":"node-1","state":"down"}`,
	})
	if result.OK {
		t.Fatal("unsupported channel unexpectedly delivered")
	}
	if result.Subject != "Alert: node-1" || result.Content != "Host node-1 is down" {
		t.Fatalf("rendered delivery content = subject %q, content %q", result.Subject, result.Content)
	}
	if result.Detail == "" {
		t.Fatal("failed delivery has no detail")
	}
}

// 通知模板和告警摘要、描述共用 pkg/alerttmpl，Prometheus 写法必须能渲染。
func TestRenderSupportsPrometheusTemplate(t *testing.T) {
	params := `{"phase":"FIRING","summary":"CPU 高","description":"cpu 95",` +
		`"labels":"{\"job\":\"api\",\"instance\":\"10.0.0.1:9100\"}",` +
		`"Labels":{"job":"api","instance":"10.0.0.1:9100"},"Value":95,"value":95,` +
		`"externalURL":"http://prometheus.example","ExternalURL":"http://prometheus.example"}`
	cases := map[string]string{
		"【{{ .phase }}】{{ $labels.job }} {{ .summary }}":    "【FIRING】api CPU 高",
		"{{ $labels.instance }}":                            "10.0.0.1:9100",
		"value: {{ . }} {{ .Value }} {{ humanize $value }}": "value: 95 95 95",
		"{{ $externalURL }}":                                "http://prometheus.example",
	}
	for raw, want := range cases {
		got, err := render(raw, params)
		if err != nil {
			t.Fatalf("render %q: %v", raw, err)
		}
		if got != want {
			t.Fatalf("render %q = %q, want %q", raw, got, want)
		}
	}
}

// 系统预设模板的字段和输出格式保持不变。
func TestRenderKeepsPresetFields(t *testing.T) {
	params := `{"phase":"FIRING","summary":"测试告警","description":"测试告警, value: 1",` +
		`"labels":"{\"__name__\":\"up\",\"job\":\"api\"}","eventUid":"evt-1"}`
	got, err := render("【{{.phase}}】{{.summary}}\n{{.description}}\n标签：{{.labels}}", params)
	if err != nil {
		t.Fatal(err)
	}
	want := "【FIRING】测试告警\n测试告警, value: 1\n标签：{\"__name__\":\"up\",\"job\":\"api\"}"
	if got != want {
		t.Fatalf("preset render = %q, want %q", got, want)
	}
}

// 历史任务只有小写字段，也要能取到结构化标签和采样值。
func TestRenderFillsLegacyParams(t *testing.T) {
	params := `{"phase":"RECOVERED","summary":"cpu","labels":"{\"job\":\"api\"}"}`
	got, err := render("{{ .phase }} {{ .summary }} {{ $labels.job }} {{ . }}", params)
	if err != nil {
		t.Fatal(err)
	}
	if got != "RECOVERED cpu api 0" {
		t.Fatalf("legacy render = %q", got)
	}
}
