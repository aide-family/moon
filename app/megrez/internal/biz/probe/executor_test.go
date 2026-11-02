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

// TestExecutePassesRequestFields 用 runProbe 注入，断言执行器把各驱动新增的
// 请求字段原样交给 pkg/probe：默认值补齐与按驱动的字段校验由 pkg/probe 负责，
// 执行器只做搬运。
func TestExecutePassesRequestFields(t *testing.T) {
	original := runProbe
	t.Cleanup(func() { runProbe = original })

	cases := []struct {
		name    string
		driver  string
		request bo.ProbeRequest
		assert  func(t *testing.T, request probesdk.Request)
	}{
		{
			name:   "http asserts",
			driver: probesdk.DriverHTTP,
			request: bo.ProbeRequest{
				Method: "HEAD", ExpectStatus: "200-299", TimeoutMS: 2000,
				FollowRedirects: true, SkipTLSVerify: true, VerifyCert: true,
				Headers:       map[string]string{"X-Token": "token"},
				BodyKeyword:   "ok",
				ExpectHeaders: map[string]string{"Content-Type": `application/json`},
				BodyRegex:     `"status":"ok"`,
			},
			assert: func(t *testing.T, request probesdk.Request) {
				t.Helper()
				if request.Method != "HEAD" || request.ExpectStatus != "200-299" || request.TimeoutMS != 2000 {
					t.Fatalf("http method/status/timeout = %+v", request)
				}
				if !request.FollowRedirects || !request.SkipTLSVerify || !request.VerifyCert {
					t.Fatalf("http switches = %+v", request)
				}
				if request.Headers["X-Token"] != "token" || request.BodyKeyword != "ok" {
					t.Fatalf("http headers/keyword = %+v", request)
				}
				if request.ExpectHeaders["Content-Type"] != `application/json` || request.BodyRegex != `"status":"ok"` {
					t.Fatalf("http asserts = %+v", request)
				}
			},
		},
		{
			name:   "icmp stats",
			driver: probesdk.DriverICMP,
			request: bo.ProbeRequest{
				TimeoutMS: 1000, ICMPCount: 3, ICMPPacketSize: 64,
			},
			assert: func(t *testing.T, request probesdk.Request) {
				t.Helper()
				if request.TimeoutMS != 1000 || request.ICMPCount != 3 || request.ICMPPacketSize != 64 {
					t.Fatalf("icmp request = %+v", request)
				}
			},
		},
		{
			name:   "dns options",
			driver: probesdk.DriverDNS,
			request: bo.ProbeRequest{
				TimeoutMS: 1500, DNSRecordType: "AAAA",
				DNSExpectValues: `127\.0\.0\.1`, DNSServer: "127.0.0.1:5353",
			},
			assert: func(t *testing.T, request probesdk.Request) {
				t.Helper()
				if request.TimeoutMS != 1500 || request.DNSRecordType != "AAAA" {
					t.Fatalf("dns record/timeout = %+v", request)
				}
				if request.DNSExpectValues != `127\.0\.0\.1` || request.DNSServer != "127.0.0.1:5353" {
					t.Fatalf("dns expect/server = %+v", request)
				}
			},
		},
		{
			name:   "tls options",
			driver: probesdk.DriverTLS,
			request: bo.ProbeRequest{
				TimeoutMS: 1200, TLSServerName: "api.example.com",
				SkipTLSVerify: true, VerifyCert: true,
			},
			assert: func(t *testing.T, request probesdk.Request) {
				t.Helper()
				if request.TimeoutMS != 1200 || request.TLSServerName != "api.example.com" {
					t.Fatalf("tls timeout/sni = %+v", request)
				}
				if !request.SkipTLSVerify || !request.VerifyCert {
					t.Fatalf("tls switches = %+v", request)
				}
			},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var got probesdk.Request
			runProbe = func(_ context.Context, driver, spec string, request probesdk.Request) probesdk.Result {
				if driver != testCase.driver || spec != "probe-target" {
					t.Fatalf("run called with driver/spec = %q/%q", driver, spec)
				}
				got = request
				return probesdk.Result{OK: true}
			}
			logger, _ := testLogger()
			group := bo.ProbeGroup{Request: testCase.request}
			target := bo.ProbeTarget{UID: "target-uid", Spec: "probe-target"}
			if _, reportable := execute(context.Background(), logger, testCase.driver, group, target, "UP", 0); !reportable {
				t.Fatal("result should be reportable")
			}
			testCase.assert(t, got)
		})
	}
}

// TestExecuteMapsExtendedResult 用 runProbe 注入带统计字段的结果，断言执行器把
// ICMP / DNS / TLS 的新结果字段原样映射，并按新指标取观测值。
func TestExecuteMapsExtendedResult(t *testing.T) {
	original := runProbe
	t.Cleanup(func() { runProbe = original })
	runProbe = func(context.Context, string, string, probesdk.Request) probesdk.Result {
		return probesdk.Result{
			OK: true, StatusCode: 200, DurationMS: 42,
			DNSMS: 3, TCPMS: 4, TLSMS: 5, TTFBMS: 6, CertDays: 30, ResolvedIP: "10.0.0.1",
			PacketLoss: 25, JitterMS: 4, MinRTTMS: 10, MaxRTTMS: 30,
			AnswerCount: 2, DNSAnswers: "10.0.0.1,10.0.0.2",
			TLSVersion: "TLSv1.3", TLSCipher: "TLS_AES_128_GCM_SHA256",
			CertSubject: "CN=api.example.com", CertIssuer: "CN=test-ca",
			CertSANs: "api.example.com,www.example.com",
		}
	}

	logger, _ := testLogger()
	group := bo.ProbeGroup{Request: bo.ProbeRequest{TimeoutMS: 1000}}
	target := bo.ProbeTarget{UID: "tls-target", Spec: "api.example.com:443"}
	result, reportable := execute(context.Background(), logger, probesdk.DriverTLS, group, target, "ANSWER_COUNT", 0)
	if !reportable {
		t.Fatalf("result should be reportable, got %+v", result)
	}
	want := bo.ProbeResult{
		TargetUID: "tls-target", OK: true, Value: "2", DurationMS: 42, StatusCode: 200,
		DNSMS: 3, TCPMS: 4, TLSMS: 5, TTFBMS: 6, CertDays: 30, ResolvedIP: "10.0.0.1",
		PacketLoss: 25, JitterMS: 4, MinRTTMS: 10, MaxRTTMS: 30,
		AnswerCount: 2, DNSAnswers: "10.0.0.1,10.0.0.2",
		TLSVersion: "TLSv1.3", TLSCipher: "TLS_AES_128_GCM_SHA256",
		CertSubject: "CN=api.example.com", CertIssuer: "CN=test-ca",
		CertSANs: "api.example.com,www.example.com",
	}
	if result != want {
		t.Fatalf("mapped result = %+v, want %+v", result, want)
	}

	// 新指标都走 Result.Value，执行器只按规则指标取值。
	metrics := map[string]string{
		"LATENCY_MS": "42", "DNS_MS": "3", "TCP_CONNECT_MS": "4", "TLS_MS": "5",
		"TTFB_MS": "6", "PACKET_LOSS": "25", "JITTER_MS": "4", "MIN_RTT_MS": "10",
		"MAX_RTT_MS": "30", "ANSWER_COUNT": "2",
	}
	for metric, wantValue := range metrics {
		result, reportable := execute(context.Background(), logger, probesdk.DriverTLS, group, target, metric, 0)
		if !reportable || result.Value != wantValue {
			t.Fatalf("metric %s value = %q reportable=%v, want %q", metric, result.Value, reportable, wantValue)
		}
	}
}
