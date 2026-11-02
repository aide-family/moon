package service

import (
	"testing"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// TestProbeRequestConversionRoundTrip 新请求字段双向转换：expect_headers 在 proto
// 里是 JSON 对象文本、在执行器结构里是 map，其余字段 1:1；空断言对象回读为空串。
func TestProbeRequestConversionRoundTrip(t *testing.T) {
	in := &v1.ProbeRequestConfig{
		Method:          "GET",
		ExpectStatus:    "200-299",
		TimeoutMs:       3000,
		FollowRedirects: true,
		VerifyCert:      true,
		BodyKeyword:     "ok",
		ExpectHeaders:   `{"X-Content-Type":"application/(json|yaml)","X-Trace":"^[a-z0-9]+$"}`,
		BodyRegex:       `"status":"ok"`,
		IcmpCount:       3,
		IcmpPacketSize:  128,
		DnsRecordType:   "AAAA",
		DnsExpectValues: `127\.0\.0\.1, 10\.0\.0\.\d+`,
		DnsServer:       "8.8.8.8:53",
		TlsServerName:   "api.example.com",
	}

	request := probeRequest(in)
	if request.ExpectHeaders["X-Content-Type"] != "application/(json|yaml)" {
		t.Fatalf("expect header not parsed: %+v", request.ExpectHeaders)
	}
	if request.ExpectHeaders["X-Trace"] != "^[a-z0-9]+$" {
		t.Fatalf("expect header not parsed: %+v", request.ExpectHeaders)
	}
	if request.BodyRegex != in.GetBodyRegex() ||
		request.ICMPCount != int(in.GetIcmpCount()) ||
		request.ICMPPacketSize != int(in.GetIcmpPacketSize()) ||
		request.DNSRecordType != in.GetDnsRecordType() ||
		request.DNSExpectValues != in.GetDnsExpectValues() ||
		request.DNSServer != in.GetDnsServer() ||
		request.TLSServerName != in.GetTlsServerName() {
		t.Fatalf("request conversion lost M27 fields: %+v", request)
	}

	back := probeRequestItem(request)
	if back.GetExpectHeaders() != in.GetExpectHeaders() {
		t.Fatalf("expect headers roundtrip = %q, want %q", back.GetExpectHeaders(), in.GetExpectHeaders())
	}
	if back.GetBodyRegex() != in.GetBodyRegex() ||
		back.GetIcmpCount() != in.GetIcmpCount() ||
		back.GetIcmpPacketSize() != in.GetIcmpPacketSize() ||
		back.GetDnsRecordType() != in.GetDnsRecordType() ||
		back.GetDnsExpectValues() != in.GetDnsExpectValues() ||
		back.GetDnsServer() != in.GetDnsServer() ||
		back.GetTlsServerName() != in.GetTlsServerName() {
		t.Fatalf("response conversion lost M27 fields: %+v", back)
	}

	// 没有断言时回读为空串，与前端的「未配置」表示一致。
	empty := probeRequestItem(probeRequest(&v1.ProbeRequestConfig{}))
	if empty.GetExpectHeaders() != "" {
		t.Fatalf("empty expect headers = %q, want empty", empty.GetExpectHeaders())
	}
}

// TestProbeResultItemMapsM27Fields 探测结果新增的 ICMP / DNS / TLS 字段要透传到
// 接口响应，否则页面看不到丢包率、解析答案与证书信息。
func TestProbeResultItemMapsM27Fields(t *testing.T) {
	in := bo.ProbeResult{
		PacketLoss: 40, JitterMS: 7, MinRTTMS: 12, MaxRTTMS: 33,
		AnswerCount: 3, DNSAnswers: "1.1.1.1, 2.2.2.2",
		TLSVersion: "TLSv1.3", TLSCipher: "TLS_AES_128_GCM_SHA256",
		CertSubject: "CN=example.com", CertIssuer: "CN=Test CA", CertSANS: "example.com",
	}
	item := probeResultItem(in)
	if item.GetPacketLoss() != 40 || item.GetJitterMs() != 7 || item.GetMinRttMs() != 12 || item.GetMaxRttMs() != 33 {
		t.Fatalf("icmp fields = %d/%d/%d/%d", item.GetPacketLoss(), item.GetJitterMs(), item.GetMinRttMs(), item.GetMaxRttMs())
	}
	if item.GetAnswerCount() != 3 || item.GetDnsAnswers() != "1.1.1.1, 2.2.2.2" {
		t.Fatalf("dns fields = %d/%q", item.GetAnswerCount(), item.GetDnsAnswers())
	}
	if item.GetTlsVersion() != "TLSv1.3" || item.GetTlsCipher() != "TLS_AES_128_GCM_SHA256" {
		t.Fatalf("tls fields = %q/%q", item.GetTlsVersion(), item.GetTlsCipher())
	}
	if item.GetCertSubject() != "CN=example.com" || item.GetCertIssuer() != "CN=Test CA" || item.GetCertSans() != "example.com" {
		t.Fatalf("cert fields = %q/%q/%q", item.GetCertSubject(), item.GetCertIssuer(), item.GetCertSans())
	}
}
