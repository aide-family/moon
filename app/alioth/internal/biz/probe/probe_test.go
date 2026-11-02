package probe

import (
	"strings"
	"testing"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

// probeTestGroup 构造一个字段齐全、可直接交给 normalizeGroup 的目标组。
func probeTestGroup(driver, request string, specs ...string) bo.ProbeTargetGroup {
	group := bo.ProbeTargetGroup{
		Name:           "核心接口组",
		Driver:         driver,
		Labels:         `{"env":"prod"}`,
		ExecutorLabels: "{}",
		Request:        request,
	}
	for _, spec := range specs {
		group.Targets = append(group.Targets, bo.ProbeTarget{Spec: spec, Status: "ACTIVE"})
	}
	return group
}

// requireInvalidError 断言错误是客户端可读的非法参数错误，而不是服务端错误。
func requireInvalidError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected invalid error, got nil")
	}
	if !merr.IsInvalid(err) {
		t.Fatalf("expected invalid error, got %v", err)
	}
}

// TestNormalizeGroupDefaultsToHTTP 不传驱动的旧客户端仍然按 HTTP 处理，
// URL 归一化与请求参数补默认值的行为不变。
func TestNormalizeGroupDefaultsToHTTP(t *testing.T) {
	group := probeTestGroup("", "", "HTTP://Example.COM:80/health")
	if err := normalizeGroup(&group); err != nil {
		t.Fatalf("normalizeGroup: %v", err)
	}
	if group.Driver != bo.ProbeDriverHTTP {
		t.Fatalf("driver = %q, want %q", group.Driver, bo.ProbeDriverHTTP)
	}
	if group.Targets[0].Spec != "http://example.com/health" {
		t.Fatalf("target spec = %q, want %q", group.Targets[0].Spec, "http://example.com/health")
	}
	if !strings.Contains(group.Request, `"timeout_ms":3000`) {
		t.Fatalf("request should be normalized with default timeout, got %s", group.Request)
	}
	if group.Labels != `{"env":"prod"}` {
		t.Fatalf("group labels = %q, want kept", group.Labels)
	}
}

// TestNormalizeGroupPerDriver 目标归一化按驱动区分：TCP / TLS 只收 host:port，
// ICMP 只收 host 或 IP，DNS 只收域名，同一组内按归一化后的地址去重。
func TestNormalizeGroupPerDriver(t *testing.T) {
	cases := []struct {
		name   string
		driver string
		specs  []string
		want   []string
	}{
		{"tcp", bo.ProbeDriverTCP, []string{"Example.COM:443", "[::1]:8002"}, []string{"example.com:443", "[::1]:8002"}},
		{"icmp", bo.ProbeDriverICMP, []string{"Example.COM", "[::1]", "127.0.0.1"}, []string{"example.com", "::1", "127.0.0.1"}},
		{"dns", bo.ProbeDriverDNS, []string{"Example.COM.", "api.example.org"}, []string{"example.com", "api.example.org"}},
		{"tls", bo.ProbeDriverTLS, []string{"Example.COM:443", "[::1]:8002"}, []string{"example.com:443", "[::1]:8002"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			group := probeTestGroup(tc.driver, "", tc.specs...)
			if err := normalizeGroup(&group); err != nil {
				t.Fatalf("normalizeGroup: %v", err)
			}
			if group.Driver != tc.driver {
				t.Fatalf("driver = %q, want %q", group.Driver, tc.driver)
			}
			for i, want := range tc.want {
				if group.Targets[i].Spec != want {
					t.Fatalf("targets[%d] = %q, want %q", i, group.Targets[i].Spec, want)
				}
				if group.Targets[i].Sort != i {
					t.Fatalf("targets[%d].Sort = %d, want %d", i, group.Targets[i].Sort, i)
				}
			}
		})
	}
}

// TestNormalizeGroupRejectsWrongTargetFormat 目标格式与驱动不匹配时拒绝，
// 避免把 HTTP 地址配成 TCP 目标或把 host:port 配成 ICMP 目标。
func TestNormalizeGroupRejectsWrongTargetFormat(t *testing.T) {
	cases := []struct {
		name   string
		driver string
		spec   string
	}{
		{"tcp scheme", bo.ProbeDriverTCP, "https://example.com:443/health"},
		{"tcp missing port", bo.ProbeDriverTCP, "example.com"},
		{"tcp reserved host", bo.ProbeDriverTCP, "169.254.169.254:80"},
		{"icmp port", bo.ProbeDriverICMP, "example.com:443"},
		{"icmp scheme", bo.ProbeDriverICMP, "icmp://example.com"},
		{"icmp reserved host", bo.ProbeDriverICMP, "169.254.169.254"},
		{"icmp path", bo.ProbeDriverICMP, "example.com/health"},
		{"dns scheme", bo.ProbeDriverDNS, "https://example.com"},
		{"dns port", bo.ProbeDriverDNS, "example.com:53"},
		{"dns ip literal", bo.ProbeDriverDNS, "127.0.0.1"},
		{"dns path", bo.ProbeDriverDNS, "example.com/resolve"},
		{"dns invalid label", bo.ProbeDriverDNS, "-example.com"},
		{"tls scheme", bo.ProbeDriverTLS, "https://example.com:443"},
		{"tls missing port", bo.ProbeDriverTLS, "example.com"},
		{"tls reserved host", bo.ProbeDriverTLS, "169.254.169.254:443"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			group := probeTestGroup(tc.driver, "", tc.spec)
			requireInvalidError(t, normalizeGroup(&group))
		})
	}
}

// TestNormalizeGroupRejectsDuplicateTargets 组内地址按归一化结果去重：
// 大小写或 IPv6 写法不同但归一化后相同的地址必须被拒绝。
func TestNormalizeGroupRejectsDuplicateTargets(t *testing.T) {
	cases := []struct {
		name   string
		driver string
		specs  []string
	}{
		{"http", bo.ProbeDriverHTTP, []string{"http://example.com:80/health", "http://EXAMPLE.com/health"}},
		{"tcp", bo.ProbeDriverTCP, []string{"Example.COM:443", "example.com:443"}},
		{"icmp", bo.ProbeDriverICMP, []string{"[::1]", "::1"}},
		{"dns", bo.ProbeDriverDNS, []string{"Example.COM.", "example.com"}},
		{"tls", bo.ProbeDriverTLS, []string{"Example.COM:443", "example.com:443"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			group := probeTestGroup(tc.driver, "", tc.specs...)
			requireInvalidError(t, normalizeGroup(&group))
		})
	}
}

// TestNormalizeGroupRejectsUnsupportedDriver 未实现的探测驱动不能保存目标组。
func TestNormalizeGroupRejectsUnsupportedDriver(t *testing.T) {
	for _, driver := range []string{"PROBE_UDP", "METRICS_PROMETHEUS", "probe_tcp"} {
		group := probeTestGroup(driver, "", "127.0.0.1:8002")
		requireInvalidError(t, normalizeGroup(&group))
	}
}

// TestNormalizeGroupRequestValidationPerDriver 请求参数按驱动校验：HTTP 走完整
// 校验（含响应头断言与响应体正则），TCP 只校验超时，ICMP / DNS / TLS 各校验
// 自己的字段；不适用字段对其它驱动不参与校验，可以带着保存。
func TestNormalizeGroupRequestValidationPerDriver(t *testing.T) {
	// HTTP 的只对 HTTP 有意义的字段在其它驱动上不参与校验，可以带着保存。
	cases := []struct {
		driver string
		spec   string
	}{
		{bo.ProbeDriverTCP, "127.0.0.1:8002"},
		{bo.ProbeDriverICMP, "127.0.0.1"},
		{bo.ProbeDriverDNS, "example.com"},
		{bo.ProbeDriverTLS, "example.com:443"},
	}
	for _, tc := range cases {
		group := probeTestGroup(tc.driver, `{"method":"POST","timeout_ms":1500}`, tc.spec)
		if err := normalizeGroup(&group); err != nil {
			t.Fatalf("%s should skip HTTP-only request validation: %v", tc.driver, err)
		}
		// 所有驱动都仍然校验超时范围。
		short := probeTestGroup(tc.driver, `{"timeout_ms":50}`, tc.spec)
		requireInvalidError(t, normalizeGroup(&short))
	}

	invalid := []struct {
		name    string
		driver  string
		spec    string
		request string
	}{
		{"http method", bo.ProbeDriverHTTP, "http://example.com/health", `{"method":"POST"}`},
		{"http body regex", bo.ProbeDriverHTTP, "http://example.com/health", `{"body_regex":"("}`},
		{"http expect header regex", bo.ProbeDriverHTTP, "http://example.com/health", `{"expect_headers":{"X-Test":"("}}`},
		{"icmp count", bo.ProbeDriverICMP, "127.0.0.1", `{"icmp_count":11}`},
		{"icmp packet size", bo.ProbeDriverICMP, "127.0.0.1", `{"icmp_packet_size":1401}`},
		{"dns record type", bo.ProbeDriverDNS, "example.com", `{"dns_record_type":"CAA"}`},
		{"dns expect values", bo.ProbeDriverDNS, "example.com", `{"dns_expect_values":"("}`},
		{"dns server", bo.ProbeDriverDNS, "example.com", `{"dns_server":"8.8.8.8"}`},
		{"tls sni", bo.ProbeDriverTLS, "example.com:443", `{"tls_server_name":"https://api.example.com"}`},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			group := probeTestGroup(tc.driver, tc.request, tc.spec)
			requireInvalidError(t, normalizeGroup(&group))
		})
	}
}

// TestNormalizeGroupPersistsDriverDefaults 保存时按驱动把默认参数写进入库请求，
// 执行器读到的就是补齐后的值：ICMP 次数与 payload、DNS 记录类型。
func TestNormalizeGroupPersistsDriverDefaults(t *testing.T) {
	icmp := probeTestGroup(bo.ProbeDriverICMP, "", "127.0.0.1")
	if err := normalizeGroup(&icmp); err != nil {
		t.Fatalf("normalizeGroup icmp: %v", err)
	}
	for _, want := range []string{`"icmp_count":1`, `"icmp_packet_size":56`} {
		if !strings.Contains(icmp.Request, want) {
			t.Fatalf("icmp request should persist default %s, got %s", want, icmp.Request)
		}
	}

	dns := probeTestGroup(bo.ProbeDriverDNS, "", "example.com")
	if err := normalizeGroup(&dns); err != nil {
		t.Fatalf("normalizeGroup dns: %v", err)
	}
	if !strings.Contains(dns.Request, `"dns_record_type":"A"`) {
		t.Fatalf("dns request should persist default record type, got %s", dns.Request)
	}
}
