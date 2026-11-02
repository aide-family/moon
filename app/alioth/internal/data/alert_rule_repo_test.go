package data

import (
	"testing"

	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertrule"
)

// TestRuleDriverAcceptsProbeDrivers 探测规则的驱动白名单包含三种探测类型，
// 未知驱动仍然拒绝；HTTP 行为保持不变。
func TestRuleDriverAcceptsProbeDrivers(t *testing.T) {
	for _, driver := range []string{"PROBE_HTTP", "PROBE_TCP", "PROBE_ICMP", "PROBE_DNS", "PROBE_TLS"} {
		got, err := ruleDriver(driver)
		if err != nil {
			t.Fatalf("ruleDriver(%q): %v", driver, err)
		}
		if string(got) != driver {
			t.Fatalf("ruleDriver(%q) = %q, want %q", driver, got, driver)
		}
	}
	for _, driver := range []string{"", "PROBE_UDP", "unknown", "probe_tcp"} {
		if _, err := ruleDriver(driver); err == nil {
			t.Fatalf("ruleDriver(%q) should fail", driver)
		}
	}
}

// TestProbeRuleDrivers 调度与领取使用的探测驱动集合必须覆盖全部探测类型，
// 顺序与 bo.ProbeDrivers 一致，避免 ent 枚举与业务常量漂移。
func TestProbeRuleDrivers(t *testing.T) {
	got := probeRuleDrivers()
	want := []alertrule.Driver{
		alertrule.DriverPROBE_HTTP, alertrule.DriverPROBE_TCP, alertrule.DriverPROBE_ICMP,
		alertrule.DriverPROBE_DNS, alertrule.DriverPROBE_TLS,
	}
	if len(got) != len(want) {
		t.Fatalf("probeRuleDrivers() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("probeRuleDrivers()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestProbeMetricValueKeepsEveryCatalogMetric 防止 ent 枚举漏加新指标时把
// DNS_MS / PACKET_LOSS / ANSWER_COUNT 之类的指标静默回落成 UP。
func TestProbeMetricValueKeepsEveryCatalogMetric(t *testing.T) {
	metrics := []string{
		"UP", "LATENCY_MS", "STATUS_CODE", "CERT_DAYS",
		"DNS_MS", "TCP_CONNECT_MS", "TLS_MS", "TTFB_MS",
		"PACKET_LOSS", "JITTER_MS", "MIN_RTT_MS", "MAX_RTT_MS", "ANSWER_COUNT",
	}
	for _, metric := range metrics {
		if got := string(probeMetricValue(metric)); got != metric {
			t.Fatalf("probeMetricValue(%q) = %q, want kept", metric, got)
		}
	}
	if got := string(probeMetricValue("UNKNOWN")); got != "UP" {
		t.Fatalf("probeMetricValue(UNKNOWN) = %q, want UP fallback", got)
	}
}
