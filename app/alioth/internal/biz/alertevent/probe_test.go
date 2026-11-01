package alertevent

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/alertjudge"
)

// TestProbeLevelWindow 固定判定窗口：持续类要覆盖 for_seconds 加一个周期，
// 其余只看最近一个周期。窗口取大了会让刚恢复的目标继续命中。
func TestProbeLevelWindow(t *testing.T) {
	interval := 15 * time.Second
	cases := []struct {
		level bo.AlertRuleLevel
		want  time.Duration
	}{
		{bo.AlertRuleLevel{Sample: "FOR", ForSeconds: 0}, interval},
		{bo.AlertRuleLevel{Sample: "FOR", ForSeconds: 30}, 45 * time.Second},
		{bo.AlertRuleLevel{Sample: "MAX", ForSeconds: 300}, interval},
		{bo.AlertRuleLevel{Sample: "MIN"}, interval},
	}
	for _, item := range cases {
		if got := probeLevelWindow(item.level, interval); got != item.want {
			t.Fatalf("probeLevelWindow(%+v) = %v, want %v", item.level, got, item.want)
		}
	}
}

// TestFilterProbePoints 只保留窗口内的点，避免旧的成功点把持续条件一直否掉。
func TestFilterProbePoints(t *testing.T) {
	slot := time.Date(2026, 10, 29, 5, 39, 0, 0, time.UTC)
	points := []alertjudge.Point{
		{Value: 1, At: slot.Add(-30 * time.Second)},
		{Value: 0, At: slot.Add(-15 * time.Second)},
		{Value: 0, At: slot},
	}
	windowed := filterProbePoints(points, slot.Add(-15*time.Second))
	if len(windowed) != 2 {
		t.Fatalf("expected 2 points in window, got %d", len(windowed))
	}
	level := alertjudge.Level{Condition: "EQ", Sample: "FOR", Threshold: "0"}
	if _, _, ok := alertjudge.Match(level, windowed, 15); !ok {
		t.Fatal("two consecutive failures should hit")
	}
	if _, _, ok := alertjudge.Match(level, points, 15); ok {
		t.Fatal("window containing an old success should not hit")
	}
}

// TestRenderProbeAlertUsesTemplateEngine 固定探测事件的摘要/描述渲染：与查询规则
// 共用 alerttmpl，支持 $value、$labels、$externalURL，不能把模板原样写进事件。
func contextBackground() context.Context { return context.Background() }

func TestRenderProbeAlertUsesTemplateEngine(t *testing.T) {
	rule := bo.AlertRule{UID: "rule-1", NamespaceUID: "ns-1", Name: "玉衡服务探测监控", Driver: bo.ProbeDriverHTTP}
	target := bo.ProbeTarget{UID: "target-1", Spec: "http://127.0.0.1:8002/health"}
	labels := map[string]string{"probe_kind": "HTTP", "probe_target": target.Spec}
	encoded, err := json.Marshal(labels)
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 10, 29, 19, 0, 0, 0, time.UTC)

	summary := renderProbeAlert(contextBackground(), rule, target, "玉衡服务探测监控  {{$value}}", 12.5, labels, string(encoded), at)
	if strings.Contains(summary, "{{") {
		t.Fatalf("summary still contains a raw template: %q", summary)
	}
	if !strings.Contains(summary, "12.5") {
		t.Fatalf("summary should expand $value, got %q", summary)
	}

	// $labels 输出 JSON，且包含事件层才注入的 datasource / alert_level。
	renderLabels, renderJSON := probeRenderLabels(labels, "玉衡服务本地探测", "P0")
	description := renderProbeAlert(contextBackground(), rule, target, "探测 {{$labels}} 值 {{$value}} 地址 {{$externalURL}}", 7, renderLabels, renderJSON, at)
	if strings.Contains(description, "{{") {
		t.Fatalf("description still contains a raw template: %q", description)
	}
	if strings.Contains(description, "map[") {
		t.Fatalf("$labels should render as JSON, got %q", description)
	}
	for _, want := range []string{"probe_kind", "玉衡服务本地探测", `"alert_level":"P0"`, "http://127.0.0.1:8002/health"} {
		if !strings.Contains(description, want) {
			t.Fatalf("description should expand $labels and $externalURL, got %q", description)
		}
	}

	// 键索引仍然可用。
	single := renderProbeAlert(contextBackground(), rule, target, "{{$labels.datasource}}/{{$labels.alert_level}}", 0, renderLabels, renderJSON, at)
	if single != "玉衡服务本地探测/P0" {
		t.Fatalf("label index should work, got %q", single)
	}

	// 不含模板的文案原样返回，避免无谓的模板开销。
	if got := renderProbeAlert(contextBackground(), rule, target, "纯文本", 1, labels, string(encoded), at); got != "纯文本" {
		t.Fatalf("plain summary should be returned as-is, got %q", got)
	}
}

// TestProbeLabelsFollowDriver 探测事件的 probe_kind 由规则驱动映射：三种探测
// 类型分别是 HTTP / TCP / ICMP，其余标签合并顺序不变。
func TestProbeLabelsFollowDriver(t *testing.T) {
	cases := []struct {
		driver string
		kind   string
	}{
		{bo.ProbeDriverHTTP, "HTTP"},
		{bo.ProbeDriverTCP, "TCP"},
		{bo.ProbeDriverICMP, "ICMP"},
	}
	for _, tc := range cases {
		t.Run(tc.kind, func(t *testing.T) {
			rule := bo.AlertRule{UID: "rule-1", Driver: tc.driver, Labels: `{"team":"platform"}`}
			group := bo.ProbeTargetGroup{UID: "group-1", Name: "核心接口组", Labels: `{"env":"prod"}`}
			target := bo.ProbeTarget{UID: "target-1", Spec: "spec-" + tc.kind, Labels: `{"instance":"a"}`}
			labels, encoded := probeLabels(rule, group, target)
			if labels[probeKindLabel] != tc.kind {
				t.Fatalf("probe_kind = %q, want %q", labels[probeKindLabel], tc.kind)
			}
			if labels[probeTargetLabel] != target.Spec {
				t.Fatalf("probe_target = %q, want %q", labels[probeTargetLabel], target.Spec)
			}
			if labels[probeTargetUIDLabel] != target.UID {
				t.Fatalf("probe_target_uid = %q, want %q", labels[probeTargetUIDLabel], target.UID)
			}
			for key, want := range map[string]string{"env": "prod", "team": "platform", "instance": "a"} {
				if labels[key] != want {
					t.Fatalf("label %q = %q, want %q", key, labels[key], want)
				}
			}
			if !strings.Contains(encoded, `"probe_kind":"`+tc.kind+`"`) {
				t.Fatalf("encoded labels should contain probe_kind, got %s", encoded)
			}
		})
	}
}

// TestProbeExternalURLFollowsDriver 非 HTTP 目标也能作为 externalURL：TCP 的
// host:port 不能被 url.Parse 当成 scheme，ICMP 的 IPv6 字面量也不能被转义。
func TestProbeExternalURLFollowsDriver(t *testing.T) {
	cases := []struct {
		driver string
		spec   string
		want   string
	}{
		{bo.ProbeDriverHTTP, "https://example.com/health", "https://example.com/health"},
		{bo.ProbeDriverTCP, "127.0.0.1:8002", "127.0.0.1:8002"},
		{bo.ProbeDriverTCP, "[::1]:8002", "[::1]:8002"},
		{bo.ProbeDriverICMP, "example.com", "example.com"},
		{bo.ProbeDriverICMP, "::1", "::1"},
	}
	for _, tc := range cases {
		if got := probeExternalURL(tc.driver, tc.spec).String(); got != tc.want {
			t.Fatalf("probeExternalURL(%q, %q).String() = %q, want %q", tc.driver, tc.spec, got, tc.want)
		}
	}
}

// TestRenderProbeAlertNonHTTPExternalURL TCP / ICMP 的 $externalURL 必须渲染成
// 目标原文，否则通知里会显示空值或 "./" 前缀。
func TestRenderProbeAlertNonHTTPExternalURL(t *testing.T) {
	at := time.Date(2026, 10, 29, 19, 0, 0, 0, time.UTC)
	cases := []struct {
		driver string
		spec   string
		kind   string
	}{
		{bo.ProbeDriverTCP, "127.0.0.1:8002", "TCP"},
		{bo.ProbeDriverICMP, "::1", "ICMP"},
	}
	for _, tc := range cases {
		rule := bo.AlertRule{UID: "rule-1", Driver: tc.driver, Name: "探测监控"}
		target := bo.ProbeTarget{UID: "target-1", Spec: tc.spec}
		labels, encoded := probeLabels(rule, bo.ProbeTargetGroup{}, target)
		summary := renderProbeAlert(contextBackground(), rule, target, "探测 {{$labels.probe_kind}} {{$externalURL}}", 1, labels, encoded, at)
		want := "探测 " + tc.kind + " " + tc.spec
		if summary != want {
			t.Fatalf("summary = %q, want %q", summary, want)
		}
	}
}
