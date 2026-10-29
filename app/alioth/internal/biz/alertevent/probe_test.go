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

	description := renderProbeAlert(contextBackground(), rule, target, "探测 {{$labels}} 值 {{$value}} 地址 {{$externalURL}}", 7, labels, string(encoded), at)
	if strings.Contains(description, "{{") {
		t.Fatalf("description still contains a raw template: %q", description)
	}
	if !strings.Contains(description, "probe_kind") || !strings.Contains(description, "http://127.0.0.1:8002/health") {
		t.Fatalf("description should expand $labels and $externalURL, got %q", description)
	}

	// 不含模板的文案原样返回，避免无谓的模板开销。
	if got := renderProbeAlert(contextBackground(), rule, target, "纯文本", 1, labels, string(encoded), at); got != "纯文本" {
		t.Fatalf("plain summary should be returned as-is, got %q", got)
	}
}
