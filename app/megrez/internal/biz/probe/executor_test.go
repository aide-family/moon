package probe

import (
	"bytes"
	"context"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aide-family/moon/app/megrez/pkg/bo"
	probesdk "github.com/aide-family/moon/pkg/probe"
)

// testLogger 返回写进内存的日志，便于断言「缺权限」的告警文案。
func testLogger() (*slog.Logger, *bytes.Buffer) {
	buffer := &bytes.Buffer{}
	return slog.New(slog.NewTextHandler(buffer, &slog.HandlerOptions{Level: slog.LevelWarn})), buffer
}

// TestExecuteTCP 覆盖 TCP 探测成功：本机监听端口可连，UP 取 1，
// LATENCY_MS 取耗时。
func TestExecuteTCP(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = listener.Close() }()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()

	logger, _ := testLogger()
	target := bo.ProbeTarget{UID: "tcp-ok", Spec: listener.Addr().String()}

	result, reportable := execute(context.Background(), logger, probesdk.DriverTCP, bo.ProbeGroup{}, target, "UP", 0)
	if !reportable {
		t.Fatalf("tcp success marked non-reportable: %+v", result)
	}
	if !result.OK || result.TargetUID != "tcp-ok" || result.Value != "1" || result.ErrorKind != "" {
		t.Fatalf("tcp success = %+v", result)
	}
	if result.DurationMS < 0 || result.TCPMS < 0 {
		t.Fatalf("tcp latency = %+v", result)
	}

	result, reportable = execute(context.Background(), logger, probesdk.DriverTCP, bo.ProbeGroup{}, target, "LATENCY_MS", 0)
	if !reportable || !result.OK {
		t.Fatalf("tcp latency probe = %+v reportable=%v", result, reportable)
	}
	latency, err := strconv.ParseFloat(result.Value, 64)
	if err != nil || latency != float64(result.DurationMS) {
		t.Fatalf("latency value = %q duration = %d err = %v", result.Value, result.DurationMS, err)
	}
}

// TestExecuteTCPConnectFailure 覆盖 TCP 连接失败：监听端口关闭后建连被拒，
// 结果按 connect 分类且仍然可回写（这是目标故障，不是执行器故障）。
func TestExecuteTCPConnectFailure(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	spec := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatalf("close listener: %v", err)
	}

	logger, _ := testLogger()
	result, reportable := execute(context.Background(), logger, probesdk.DriverTCP, bo.ProbeGroup{}, bo.ProbeTarget{UID: "tcp-fail", Spec: spec}, "UP", 0)
	if !reportable {
		t.Fatalf("tcp connect failure marked non-reportable: %+v", result)
	}
	if result.OK || result.Value != "0" || result.ErrorKind != probesdk.ErrorConnect || result.ErrorMessage == "" {
		t.Fatalf("tcp connect failure = %+v", result)
	}
}

// TestExecuteHTTP 覆盖 HTTP 探测不受影响：httptest 服务返回 200，状态码指标
// 取到实际状态码。
func TestExecuteHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger, _ := testLogger()
	group := bo.ProbeGroup{Request: bo.ProbeRequest{TimeoutMS: 2000}}
	result, reportable := execute(context.Background(), logger, probesdk.DriverHTTP, group, bo.ProbeTarget{UID: "http-ok", Spec: server.URL}, "STATUS_CODE", 0)
	if !reportable {
		t.Fatalf("http result marked non-reportable: %+v", result)
	}
	if !result.OK || result.StatusCode != http.StatusOK || result.Value != "200" {
		t.Fatalf("http result = %+v", result)
	}
}

// TestExecutePermissionNotReportable 覆盖缺权限分支：通过 runProbe 注入
// permission 结果，断言不回写、结果为空、日志提示 ping_group_range / NET_RAW，
// 同时覆盖驱动与目标透传、maxTimeoutMS 超时封顶。
func TestExecutePermissionNotReportable(t *testing.T) {
	original := runProbe
	t.Cleanup(func() { runProbe = original })

	var (
		gotDriver   string
		gotSpec     string
		gotDeadline time.Time
	)
	runProbe = func(ctx context.Context, driver, spec string, _ probesdk.Request) probesdk.Result {
		gotDriver, gotSpec = driver, spec
		gotDeadline, _ = ctx.Deadline()
		return probesdk.Result{ErrorKind: probesdk.ErrorPermission, ErrorMessage: "socket: permission denied"}
	}

	logger, buffer := testLogger()
	group := bo.ProbeGroup{Request: bo.ProbeRequest{TimeoutMS: 5000}}
	result, reportable := execute(context.Background(), logger, probesdk.DriverICMP, group, bo.ProbeTarget{UID: "icmp-no-perm", Spec: "127.0.0.1"}, "UP", 100)
	if reportable {
		t.Fatalf("permission result marked reportable: %+v", result)
	}
	if result != (bo.ProbeResult{}) {
		t.Fatalf("permission result should be empty, got %+v", result)
	}
	if gotDriver != probesdk.DriverICMP || gotSpec != "127.0.0.1" {
		t.Fatalf("driver/spec passed to run = %q/%q", gotDriver, gotSpec)
	}
	// 组请求 5 秒，执行器上限 100ms：ctx 必须按上限封顶。
	if remaining := time.Until(gotDeadline); remaining > 150*time.Millisecond {
		t.Fatalf("maxTimeoutMS not applied, remaining = %v", remaining)
	}
	logged := buffer.String()
	if !strings.Contains(logged, "ping_group_range") || !strings.Contains(logged, "NET_RAW") {
		t.Fatalf("permission warning missing privilege hint: %q", logged)
	}
	if !strings.Contains(logged, "127.0.0.1") {
		t.Fatalf("permission warning missing target: %q", logged)
	}
}
