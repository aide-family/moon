package bo

import (
	"reflect"
	"testing"
)

// TestProbeDrivers 固定驱动清单与顺序：套餐勾选页、规则列表与调度查询都按这个
// 顺序展示或过滤，新增探测类型时只改 ProbeDrivers。
func TestProbeDrivers(t *testing.T) {
	want := []string{"PROBE_HTTP", "PROBE_TCP", "PROBE_ICMP"}
	if got := ProbeDrivers(); !reflect.DeepEqual(got, want) {
		t.Fatalf("ProbeDrivers() = %v, want %v", got, want)
	}
}

// TestIsProbeDriver 三种探测驱动都算探测类型，数据源驱动与未知驱动不算。
func TestIsProbeDriver(t *testing.T) {
	for _, driver := range ProbeDrivers() {
		if !IsProbeDriver(driver) {
			t.Fatalf("IsProbeDriver(%q) = false, want true", driver)
		}
	}
	for _, driver := range []string{"", "METRICS_PROMETHEUS", "LOGS_LOKI", "PROBE_UDP", "probe_http"} {
		if IsProbeDriver(driver) {
			t.Fatalf("IsProbeDriver(%q) = true, want false", driver)
		}
	}
}

// TestProbeKindOf 事件标签 probe_kind 取值由驱动映射，未知驱动不能凭空造出类型。
func TestProbeKindOf(t *testing.T) {
	cases := map[string]string{
		ProbeDriverHTTP: "HTTP",
		ProbeDriverTCP:  "TCP",
		ProbeDriverICMP: "ICMP",
		"":              "",
		"PROBE_UDP":     "",
	}
	for driver, want := range cases {
		if got := ProbeKindOf(driver); got != want {
			t.Fatalf("ProbeKindOf(%q) = %q, want %q", driver, got, want)
		}
	}
}

// TestValidProbeMetric 观测指标按驱动校验：HTTP 四个指标都可用，TCP / ICMP
// 没有状态码与证书，只允许 UP 与 LATENCY_MS。
func TestValidProbeMetric(t *testing.T) {
	allMetrics := []string{ProbeMetricUp, ProbeMetricLatencyMS, ProbeMetricStatusCode, ProbeMetricCertDays}
	for _, metric := range allMetrics {
		if !ValidProbeMetric(ProbeDriverHTTP, metric) {
			t.Fatalf("ValidProbeMetric(HTTP, %q) = false, want true", metric)
		}
	}
	for _, driver := range []string{ProbeDriverTCP, ProbeDriverICMP} {
		for _, metric := range []string{ProbeMetricUp, ProbeMetricLatencyMS} {
			if !ValidProbeMetric(driver, metric) {
				t.Fatalf("ValidProbeMetric(%q, %q) = false, want true", driver, metric)
			}
		}
		for _, metric := range []string{ProbeMetricStatusCode, ProbeMetricCertDays} {
			if ValidProbeMetric(driver, metric) {
				t.Fatalf("ValidProbeMetric(%q, %q) = true, want false", driver, metric)
			}
		}
	}
	for _, driver := range []string{"", "METRICS_PROMETHEUS", "PROBE_UDP"} {
		for _, metric := range append(allMetrics, "") {
			if ValidProbeMetric(driver, metric) {
				t.Fatalf("ValidProbeMetric(%q, %q) = true, want false", driver, metric)
			}
		}
	}
}

// TestAlertRuleIsProbeRule 驱动决定规则类型：三种探测驱动都是探测规则，
// 其余驱动仍按查询规则处理。
func TestAlertRuleIsProbeRule(t *testing.T) {
	for _, driver := range ProbeDrivers() {
		if !(AlertRule{Driver: driver}).IsProbeRule() {
			t.Fatalf("AlertRule{Driver: %q}.IsProbeRule() = false, want true", driver)
		}
	}
	for _, driver := range []string{"", "METRICS_PROMETHEUS", "LOGS_LOKI", "TRACE_JAEGER", "PROBE_UDP"} {
		if (AlertRule{Driver: driver}).IsProbeRule() {
			t.Fatalf("AlertRule{Driver: %q}.IsProbeRule() = true, want false", driver)
		}
	}
}
