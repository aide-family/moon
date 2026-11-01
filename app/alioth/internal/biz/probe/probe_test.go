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

// TestNormalizeGroupPerDriver 目标归一化按驱动区分：TCP 只收 host:port，
// ICMP 只收 host 或 IP，同一组内按归一化后的地址去重。
func TestNormalizeGroupPerDriver(t *testing.T) {
	cases := []struct {
		name   string
		driver string
		specs  []string
		want   []string
	}{
		{"tcp", bo.ProbeDriverTCP, []string{"Example.COM:443", "[::1]:8002"}, []string{"example.com:443", "[::1]:8002"}},
		{"icmp", bo.ProbeDriverICMP, []string{"Example.COM", "[::1]", "127.0.0.1"}, []string{"example.com", "::1", "127.0.0.1"}},
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
// 校验（method、期望状态码、受控头），TCP / ICMP 只校验 timeout_ms。
func TestNormalizeGroupRequestValidationPerDriver(t *testing.T) {
	// HTTP 的只对 HTTP 有意义的字段在 TCP / ICMP 上不参与校验，可以带着保存。
	cases := []struct {
		driver string
		spec   string
	}{
		{bo.ProbeDriverTCP, "127.0.0.1:8002"},
		{bo.ProbeDriverICMP, "127.0.0.1"},
	}
	for _, tc := range cases {
		group := probeTestGroup(tc.driver, `{"method":"POST","timeout_ms":1500}`, tc.spec)
		if err := normalizeGroup(&group); err != nil {
			t.Fatalf("%s should skip HTTP-only request validation: %v", tc.driver, err)
		}
		// TCP / ICMP 仍然只校验超时范围。
		short := probeTestGroup(tc.driver, `{"timeout_ms":50}`, tc.spec)
		requireInvalidError(t, normalizeGroup(&short))
	}
	// HTTP 的完整校验不变：非 GET/HEAD 方法仍然拒绝。
	group := probeTestGroup(bo.ProbeDriverHTTP, `{"method":"POST"}`, "http://example.com/health")
	requireInvalidError(t, normalizeGroup(&group))
}
