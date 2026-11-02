package bo

import (
	"reflect"
	"testing"
)

// TestProbeDrivers 固定驱动清单与顺序：套餐勾选页、规则列表与调度查询都按这个
// 顺序展示或过滤，新增探测类型时只改 ProbeDrivers。
func TestProbeDrivers(t *testing.T) {
	want := []string{"PROBE_HTTP", "PROBE_TCP", "PROBE_ICMP", "PROBE_DNS", "PROBE_TLS"}
	if got := ProbeDrivers(); !reflect.DeepEqual(got, want) {
		t.Fatalf("ProbeDrivers() = %v, want %v", got, want)
	}
}

// TestIsProbeDriver 五种探测驱动都算探测类型，数据源驱动与未知驱动不算。
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
		ProbeDriverDNS:  "DNS",
		ProbeDriverTLS:  "TLS",
		"":              "",
		"PROBE_UDP":     "",
	}
	for driver, want := range cases {
		if got := ProbeKindOf(driver); got != want {
			t.Fatalf("ProbeKindOf(%q) = %q, want %q", driver, got, want)
		}
	}
}

// TestValidProbeMetric 观测指标按驱动校验，指标目录以 M27 文档为准：
// HTTP 有状态码、证书与耗时分解；TCP 有解析与建连；ICMP 有丢包与 RTT 统计；
// DNS 有解析与答案数；TLS 有解析、建连、握手与证书。
func TestValidProbeMetric(t *testing.T) {
	allowedByDriver := map[string][]string{
		ProbeDriverHTTP: {
			ProbeMetricUp, ProbeMetricLatencyMS, ProbeMetricStatusCode, ProbeMetricCertDays,
			ProbeMetricDNSMS, ProbeMetricTCPConnectMS, ProbeMetricTLSMS, ProbeMetricTTFBMS,
		},
		ProbeDriverTCP:  {ProbeMetricUp, ProbeMetricLatencyMS, ProbeMetricDNSMS, ProbeMetricTCPConnectMS},
		ProbeDriverICMP: {ProbeMetricUp, ProbeMetricLatencyMS, ProbeMetricPacketLoss, ProbeMetricJitterMS, ProbeMetricMinRTTMS, ProbeMetricMaxRTTMS},
		ProbeDriverDNS:  {ProbeMetricUp, ProbeMetricLatencyMS, ProbeMetricDNSMS, ProbeMetricAnswerCount},
		ProbeDriverTLS:  {ProbeMetricUp, ProbeMetricLatencyMS, ProbeMetricDNSMS, ProbeMetricTCPConnectMS, ProbeMetricTLSMS, ProbeMetricCertDays},
	}
	allMetrics := []string{
		ProbeMetricUp, ProbeMetricLatencyMS, ProbeMetricStatusCode, ProbeMetricCertDays,
		ProbeMetricDNSMS, ProbeMetricTCPConnectMS, ProbeMetricTLSMS, ProbeMetricTTFBMS,
		ProbeMetricPacketLoss, ProbeMetricJitterMS, ProbeMetricMinRTTMS, ProbeMetricMaxRTTMS, ProbeMetricAnswerCount,
	}
	for _, driver := range ProbeDrivers() {
		allowed := allowedByDriver[driver]
		if len(allowed) == 0 {
			t.Fatalf("driver %q has no expected metric list", driver)
		}
		expectations := make(map[string]bool, len(allowed))
		for _, metric := range allowed {
			expectations[metric] = true
		}
		for _, metric := range allMetrics {
			if got := ValidProbeMetric(driver, metric); got != expectations[metric] {
				t.Fatalf("ValidProbeMetric(%q, %q) = %v, want %v", driver, metric, got, expectations[metric])
			}
		}
		if ValidProbeMetric(driver, "") {
			t.Fatalf("ValidProbeMetric(%q, \"\") = true, want false", driver)
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

// TestAlertRuleIsProbeRule 驱动决定规则类型：五种探测驱动都是探测规则，
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
