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
	data := NewData(labels, `{"job":"api"}`, 3, "http://prometheus.example", map[string]any{
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
	data := NewData(map[string]string{"instance": "local"}, `{}`, 1, "", nil)
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
	if _, err := Render(context.Background(), "{{ .Labels ", NewData(nil, "{}", 0, "", nil), "t", time.Now(), nil, nil); err == nil {
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

// 声明变量后 {{ $disk_free }} 才是合法模板，渲染时直接读出 Vars/VarLabels。
func TestRenderWithVarsInjectsDeclaredVariables(t *testing.T) {
	data := NewData(map[string]string{"instance": "a"}, `{"instance":"a"}`, 91.5, "", map[string]any{
		"Vars":      map[string]any{"disk_free": 42.5},
		"VarLabels": map[string]map[string]string{"disk_free": {"mountpoint": "/data"}},
	})
	got, err := RenderWithVars(context.Background(),
		`{{ $disk_free }} {{ $disk_free | printf "%.1f" }} {{ $disk_free | humanize }} {{ index .Vars "disk_free" }} {{ index (index .VarLabels "disk_free") "mountpoint" }} {{ $labels.instance }} {{ $value }}`,
		data, "t", time.Now(), nil, nil, []string{"disk_free"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "42.5 42.5 42.5 42.5 /data a 91.5" {
		t.Fatalf("render with vars = %q", got)
	}
}

// 变量查询失败时 eval 不写 Vars 键，模板仍要能展开成 <no value>。
func TestRenderWithVarsToleratesMissingVariableResult(t *testing.T) {
	data := NewData(nil, "{}", 0, "", nil)
	got, err := RenderWithVars(context.Background(), `{{ $disk_free }}`, data, "t", time.Now(), nil, nil, []string{"disk_free"})
	if err != nil {
		t.Fatalf("missing variable result must not fail: %v", err)
	}
	if got != "<no value>" {
		t.Fatalf("missing variable render = %q", got)
	}
	if data["Vars"] == nil || data["VarLabels"] == nil {
		t.Fatalf("Vars/VarLabels not prepared: %#v", data)
	}
}

// 没有声明变量时 RenderWithVars 必须与 Render 完全一致，也不能改写数据。
func TestRenderWithVarsKeepsPlainRenderWithoutVars(t *testing.T) {
	raw := "{{ . }} {{ $labels.job }} {{ humanize1024 2048 }}"
	want, err := Render(context.Background(), raw,
		NewData(map[string]string{"job": "api"}, `{"job":"api"}`, 3, "http://prometheus.example", nil),
		"t", time.Now(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	withVars := NewData(map[string]string{"job": "api"}, `{"job":"api"}`, 3, "http://prometheus.example", nil)
	got, err := RenderWithVars(context.Background(), raw, withVars, "t", time.Now(), nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("RenderWithVars without vars = %q, want %q", got, want)
	}
	if _, ok := withVars["Vars"]; ok {
		t.Fatalf("RenderWithVars without vars touched data: %#v", withVars)
	}
}

func TestValidateWithVarsInjectsDeclaredNames(t *testing.T) {
	// 未声明变量时保存必须报未定义变量，声明后才能通过。
	if err := Validate("{{ $disk_free }}", "t"); err == nil {
		t.Fatal("undeclared variable accepted")
	}
	if err := ValidateWithVars("{{ $disk_free }}", "t", []string{"disk_free"}); err != nil {
		t.Fatalf("declared variable rejected: %v", err)
	}
	if err := ValidateWithVars("{{ $disk_free }}", "t", []string{"disk_total"}); err == nil {
		t.Fatal("missing declaration accepted")
	}
	// 非法名字不参与定义拼接，不能破坏模板语法。
	if err := ValidateWithVars("{{ $labels.job }}", "t", []string{`x" }}{{ .Value`, "1bad", "with-dash", ""}); err != nil {
		t.Fatalf("invalid variable names broke template: %v", err)
	}
	if err := ValidateWithVars("{{ $labels.job ", "t", []string{"disk_free"}); err == nil {
		t.Fatal("broken template accepted")
	}
}

// TestLabelsRenderAsJSON $labels 输出 JSON（与事件「标签」字段一致），
// 而不是 Go 的 map[...]；键索引仍然可用。
func TestLabelsRenderAsJSON(t *testing.T) {
	labels := map[string]string{"job": "api", "datasource": "核心接口组"}
	data := NewData(labels, `{"job":"api","datasource":"核心接口组"}`, 3, "http://api.example", nil)
	externalURL, _ := url.Parse("http://api.example")
	got, err := Render(context.Background(), "{{ $labels }}", data, "t", time.Now(), nil, externalURL)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "map[") {
		t.Fatalf("$labels should render as JSON, got %q", got)
	}
	if !strings.Contains(got, `"job":"api"`) || !strings.Contains(got, `"datasource":"核心接口组"`) {
		t.Fatalf("$labels should contain the labels JSON, got %q", got)
	}
	indexed, err := Render(context.Background(), "{{ $labels.job }}", data, "t", time.Now(), nil, externalURL)
	if err != nil {
		t.Fatal(err)
	}
	if indexed != "api" {
		t.Fatalf("label indexing should work, got %q", indexed)
	}

	// 通知侧走 FromParams：只有小写 labels 时也要补成同样的 JSON 输出。
	fromParams := FromParams(map[string]any{"labels": `{"job":"api"}`})
	rendered, err := Render(context.Background(), "{{ $labels }}", fromParams, "t", time.Now(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if rendered != `{"job":"api"}` {
		t.Fatalf("FromParams $labels = %q", rendered)
	}
}
