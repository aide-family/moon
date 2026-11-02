package probe

import (
	"context"
	"testing"
	"time"
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

// TestRunICMPMultiplePackets 打本机回环验证多包统计。无 ICMP 权限时跳过。
func TestRunICMPMultiplePackets(t *testing.T) {
	request := DefaultRequest()
	request.TimeoutMS = 3000
	request.ICMPCount = 3
	request.ICMPPacketSize = 64
	outcome := Run(context.Background(), DriverICMP, "127.0.0.1", request)
	if outcome.ErrorKind == ErrorPermission {
		t.Skipf("ICMP privileges unavailable: %s", outcome.ErrorMessage)
	}
	if !outcome.OK {
		t.Fatalf("loopback ICMP should succeed, got %+v", outcome)
	}
	if outcome.PacketLoss < 0 || outcome.PacketLoss > 100 {
		t.Fatalf("packet loss out of range: %d", outcome.PacketLoss)
	}
	if outcome.MinRTTMS > outcome.DurationMS || outcome.DurationMS > outcome.MaxRTTMS {
		t.Fatalf("min/avg/max out of order: %d/%d/%d", outcome.MinRTTMS, outcome.DurationMS, outcome.MaxRTTMS)
	}
	if outcome.Value("PACKET_LOSS") != float64(outcome.PacketLoss) ||
		outcome.Value("JITTER_MS") != float64(outcome.JitterMS) ||
		outcome.Value("MIN_RTT_MS") != float64(outcome.MinRTTMS) ||
		outcome.Value("MAX_RTT_MS") != float64(outcome.MaxRTTMS) {
		t.Fatalf("unexpected ICMP metric values: %+v", outcome)
	}
	if outcome.ResolvedIP != "127.0.0.1" {
		t.Fatalf("ResolvedIP=%q", outcome.ResolvedIP)
	}
}

func TestICMPStats(t *testing.T) {
	avg, jitter, minRTT, maxRTT, loss := icmpStats(3, []time.Duration{
		10 * time.Millisecond, 20 * time.Millisecond, 30 * time.Millisecond,
	})
	if avg != 20 || jitter != 10 || minRTT != 10 || maxRTT != 30 || loss != 0 {
		t.Fatalf("full reply stats: avg=%d jitter=%d min=%d max=%d loss=%d", avg, jitter, minRTT, maxRTT, loss)
	}

	avg, jitter, minRTT, maxRTT, loss = icmpStats(4, []time.Duration{
		10 * time.Millisecond, 20 * time.Millisecond,
	})
	if loss != 50 || avg != 15 || jitter != 10 || minRTT != 10 || maxRTT != 20 {
		t.Fatalf("partial reply stats: avg=%d jitter=%d min=%d max=%d loss=%d", avg, jitter, minRTT, maxRTT, loss)
	}

	_, _, _, _, loss = icmpStats(2, nil)
	if loss != 100 {
		t.Fatalf("all lost loss=%d, want 100", loss)
	}

	// 单包没有相邻回包，抖动为 0。
	avg, jitter, _, _, _ = icmpStats(1, []time.Duration{time.Second})
	if avg != 1000 || jitter != 0 {
		t.Fatalf("single reply stats: avg=%d jitter=%d", avg, jitter)
	}

	// 丢包率四舍五入：3 包只回 1 包记 67。
	if _, _, _, _, loss = icmpStats(3, []time.Duration{time.Millisecond}); loss != 67 {
		t.Fatalf("loss rounding=%d, want 67", loss)
	}
	if avg, jitter, minRTT, maxRTT, loss = icmpStats(0, nil); avg != 0 || jitter != 0 || minRTT != 0 || maxRTT != 0 || loss != 0 {
		t.Fatalf("zero sent stats should be zero: %d %d %d %d %d", avg, jitter, minRTT, maxRTT, loss)
	}
}
