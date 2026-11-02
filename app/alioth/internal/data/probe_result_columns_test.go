package data

import (
	"testing"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
)

// TestToProbeResultMapsM27Columns M27 追加的 ICMP / DNS / TLS 结果列必须原样
// 读出，否则记录列表与事件图表看不到丢包、解析答案与证书信息。
func TestToProbeResultMapsM27Columns(t *testing.T) {
	row := &ent.ProbeResult{
		PacketLoss:  40,
		JitterMs:    7,
		MinRttMs:    12,
		MaxRttMs:    33,
		AnswerCount: 3,
		DNSAnswers:  "1.1.1.1, 2.2.2.2",
		TLSVersion:  "TLSv1.3",
		TLSCipher:   "TLS_AES_128_GCM_SHA256",
		CertSubject: "CN=example.com",
		CertIssuer:  "CN=Test CA",
		CertSans:    "example.com, www.example.com",
	}
	item := toProbeResult(row)
	if item.PacketLoss != 40 || item.JitterMS != 7 || item.MinRTTMS != 12 || item.MaxRTTMS != 33 {
		t.Fatalf("icmp columns = loss:%d jitter:%d min:%d max:%d", item.PacketLoss, item.JitterMS, item.MinRTTMS, item.MaxRTTMS)
	}
	if item.AnswerCount != 3 || item.DNSAnswers != "1.1.1.1, 2.2.2.2" {
		t.Fatalf("dns columns = count:%d answers:%q", item.AnswerCount, item.DNSAnswers)
	}
	if item.TLSVersion != "TLSv1.3" || item.TLSCipher != "TLS_AES_128_GCM_SHA256" {
		t.Fatalf("tls columns = version:%q cipher:%q", item.TLSVersion, item.TLSCipher)
	}
	if item.CertSubject != "CN=example.com" || item.CertIssuer != "CN=Test CA" || item.CertSANS != "example.com, www.example.com" {
		t.Fatalf("cert columns = subject:%q issuer:%q sans:%q", item.CertSubject, item.CertIssuer, item.CertSANS)
	}
}
