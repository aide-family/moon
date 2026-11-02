package data

import (
	"reflect"
	"testing"

	"google.golang.org/protobuf/proto"

	megrezv1 "github.com/aide-family/moon/api/megrez/v1"
	"github.com/aide-family/moon/app/megrez/pkg/bo"
)

// TestProbeRequestFromSnapshot 覆盖领取快照到业务请求的完整字段映射，包括两段
// JSON 文本字段（自定义请求头、响应头断言）。
func TestProbeRequestFromSnapshot(t *testing.T) {
	snapshot := &megrezv1.ProbeRequestSnapshot{
		Method: "HEAD", ExpectStatus: "200-299", TimeoutMs: 2500,
		FollowRedirects: true, SkipTlsVerify: true, VerifyCert: true,
		Headers:       `{"X-Token":"token"}`,
		BodyKeyword:   "ok",
		ExpectHeaders: `{"Content-Type":"application/json"}`,
		BodyRegex:     `"status":"ok"`,
		IcmpCount:     3, IcmpPacketSize: 64,
		DnsRecordType: "AAAA", DnsExpectValues: `127\.0\.0\.1`, DnsServer: "127.0.0.1:5353",
		TlsServerName: "api.example.com",
	}
	got := probeRequestFromSnapshot(snapshot)
	want := bo.ProbeRequest{
		Method: "HEAD", ExpectStatus: "200-299", TimeoutMS: 2500,
		FollowRedirects: true, SkipTLSVerify: true, VerifyCert: true,
		Headers:       map[string]string{"X-Token": "token"},
		BodyKeyword:   "ok",
		ExpectHeaders: map[string]string{"Content-Type": "application/json"},
		BodyRegex:     `"status":"ok"`,
		ICMPCount:     3, ICMPPacketSize: 64,
		DNSRecordType: "AAAA", DNSExpectValues: `127\.0\.0\.1`, DNSServer: "127.0.0.1:5353",
		TLSServerName: "api.example.com",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("probe request = %+v, want %+v", got, want)
	}
}

// TestProbeRequestFromSnapshotMalformed 覆盖空快照与坏 JSON：不 panic，JSON 字段
// 按空处理，执行端把 nil 当作没有额外请求头 / 没有断言。
func TestProbeRequestFromSnapshotMalformed(t *testing.T) {
	if got := probeRequestFromSnapshot(nil); !reflect.DeepEqual(got, bo.ProbeRequest{}) {
		t.Fatalf("nil snapshot request = %+v", got)
	}
	got := probeRequestFromSnapshot(&megrezv1.ProbeRequestSnapshot{
		Headers: `{`, ExpectHeaders: `[]`, TimeoutMs: 1000,
	})
	if got.Headers != nil || got.ExpectHeaders != nil {
		t.Fatalf("malformed json headers = %+v", got)
	}
	if got.TimeoutMS != 1000 {
		t.Fatalf("malformed json lost scalar field: %+v", got)
	}
}

// TestProbeResultToProto 覆盖回写结果的完整字段映射：原有字段与 M27 新增的
// ICMP 统计、DNS 答案、TLS 握手/证书信息都必须原样提交给玉衡。
func TestProbeResultToProto(t *testing.T) {
	item := bo.ProbeResult{
		TargetUID: "target-uid", OK: true, Value: "2", DurationMS: 42, StatusCode: 200,
		ErrorKind: "assert", ErrorMessage: "body regex mismatch",
		DNSMS: 3, TCPMS: 4, TLSMS: 5, TTFBMS: 6, CertDays: 30, ResolvedIP: "10.0.0.1",
		Executor: "megrez-1", PacketLoss: 25, JitterMS: 4, MinRTTMS: 10, MaxRTTMS: 30,
		AnswerCount: 2, DNSAnswers: "10.0.0.1,10.0.0.2",
		TLSVersion: "TLSv1.3", TLSCipher: "TLS_AES_128_GCM_SHA256",
		CertSubject: "CN=api.example.com", CertIssuer: "CN=test-ca",
		CertSANs: "api.example.com,www.example.com",
	}
	want := &megrezv1.ProbeResult{
		TargetUid: "target-uid", Ok: true, Value: "2", DurationMs: 42, StatusCode: 200,
		ErrorKind: "assert", ErrorMessage: "body regex mismatch",
		DnsMs: 3, TcpMs: 4, TlsMs: 5, TtfbMs: 6, CertDays: 30, ResolvedIp: "10.0.0.1",
		Executor: "megrez-1", PacketLoss: 25, JitterMs: 4, MinRttMs: 10, MaxRttMs: 30,
		AnswerCount: 2, DnsAnswers: "10.0.0.1,10.0.0.2",
		TlsVersion: "TLSv1.3", TlsCipher: "TLS_AES_128_GCM_SHA256",
		CertSubject: "CN=api.example.com", CertIssuer: "CN=test-ca",
		CertSans: "api.example.com,www.example.com",
	}
	if got := probeResultToProto(item); !proto.Equal(got, want) {
		t.Fatalf("probe result proto = %+v, want %+v", got, want)
	}
}
