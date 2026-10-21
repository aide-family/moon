package alerttmpl

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestNewKeepsPrometheusAndMessageFields(t *testing.T) {
	labels := map[string]string{"job": "api"}
	data := New(labels, `{"job":"api"}`, 3, "http://prometheus.example", map[string]any{
		"phase": "FIRING", "Value": "overridden",
	})
	if data["Value"] != 3.0 {
		t.Fatalf("reserved key overridden: %#v", data["Value"])
	}
	if data["labels"] != `{"job":"api"}` || data["phase"] != "FIRING" {
		t.Fatalf("data = %#v", data)
	}
	externalURL, _ := url.Parse("http://prometheus.example")
	got, err := Render(context.Background(), "{{ . }} {{ $labels.job }} {{ $value }} {{ .phase }}", data, "t", time.Now(), nil, externalURL)
	if err != nil {
		t.Fatal(err)
	}
	if got != "3 api 3 FIRING" {
		t.Fatalf("render = %q", got)
	}
}

// 历史任务参数只有小写字段，FromParams 要补齐结构化字段。
func TestFromParamsFillsStructuredFields(t *testing.T) {
	data := FromParams(map[string]any{
		"summary": "cpu", "labels": `{"job":"api"}`, "value": "7.5",
		"externalURL": "http://prometheus.example",
	})
	got, err := Render(context.Background(), "{{ .summary }} {{ $labels.job }} {{ . }} {{ $externalURL }}", data, "t", time.Now(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != "cpu api 7.5 http://prometheus.example" {
		t.Fatalf("render = %q", got)
	}
}

func TestRenderSupportsQuery(t *testing.T) {
	query := func(context.Context, string, time.Time) ([]Sample, error) {
		return []Sample{{Labels: map[string]string{"instance": "target"}, Value: 9}}, nil
	}
	data := New(map[string]string{"instance": "local"}, `{}`, 1, "", nil)
	got, err := Render(context.Background(),
		`{{ with query "up" }}{{ with first . }}{{ .Labels.instance }}={{ value . }}{{ end }}{{ end }}`,
		data, "t", time.Now(), query, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != "target=9" {
		t.Fatalf("query render = %q", got)
	}
}

func TestRenderReportsBrokenTemplate(t *testing.T) {
	if _, err := Render(context.Background(), "{{ .Labels ", New(nil, "{}", 0, "", nil), "t", time.Now(), nil, nil); err == nil {
		t.Fatal("broken template accepted")
	}
}

func TestValidateAcceptsPrometheusSyntax(t *testing.T) {
	if err := Validate("{{ $labels.job }} {{ humanize $value }} {{ . }}", "t"); err != nil {
		t.Fatalf("prometheus template rejected: %v", err)
	}
	if err := Validate("{{ $labels.job ", "t"); err == nil || !strings.Contains(err.Error(), "template") {
		t.Fatalf("broken template accepted: %v", err)
	}
}
