package probe

import (
	"context"
	"testing"
)

// TestRunICMPLoopback 打本机回环。容器里没给 ICMP 权限时跳过，
// 不把「环境没权限」当成探测逻辑失败。
func TestRunICMPLoopback(t *testing.T) {
	request := DefaultRequest()
	request.TimeoutMS = 2000
	outcome := Run(context.Background(), DriverICMP, "127.0.0.1", request)
	if outcome.ErrorKind == ErrorPermission {
		t.Skipf("ICMP privileges unavailable: %s", outcome.ErrorMessage)
	}
	if !outcome.OK {
		t.Fatalf("loopback ICMP should succeed, got %+v", outcome)
	}
	if outcome.ResolvedIP != "127.0.0.1" {
		t.Fatalf("ResolvedIP=%q, want 127.0.0.1", outcome.ResolvedIP)
	}
	if outcome.Value("UP") != 1 || outcome.Value("LATENCY_MS") != float64(outcome.DurationMS) {
		t.Fatalf("unexpected values: UP=%v LATENCY_MS=%v DurationMS=%d", outcome.Value("UP"), outcome.Value("LATENCY_MS"), outcome.DurationMS)
	}
	// ICMP 没有状态码与证书，观测值固定为 0。
	if outcome.Value("STATUS_CODE") != 0 || outcome.Value("CERT_DAYS") != 0 {
		t.Fatalf("unexpected HTTP-only values: %+v", outcome)
	}
}

// TestRunICMPUnresolvable 校验解析失败会被记成 dns，且不会 panic。
func TestRunICMPUnresolvable(t *testing.T) {
	request := DefaultRequest()
	request.TimeoutMS = 500
	outcome := Run(context.Background(), DriverICMP, "probe.invalid", request)
	if outcome.OK {
		t.Fatalf("unresolvable target should fail, got %+v", outcome)
	}
	if outcome.ErrorKind != ErrorDNS && outcome.ErrorKind != ErrorTimeout {
		t.Fatalf("unexpected error kind %q (%s)", outcome.ErrorKind, outcome.ErrorMessage)
	}
}
