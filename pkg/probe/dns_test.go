package probe

import (
	"context"
	"net"
	"strings"
	"testing"

	"golang.org/x/net/dns/dnsmessage"
)

// startTestDNSServer 起一个只回 A 记录的本机 DNS 服务器，records 的键是
// 不带末尾点的小写域名；未登记的域名回 NXDOMAIN。
func startTestDNSServer(t *testing.T, records map[string][]net.IP) string {
	t.Helper()
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen dns server: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	go func() {
		buffer := make([]byte, 1500)
		for {
			count, peer, err := conn.ReadFrom(buffer)
			if err != nil {
				return
			}
			var query dnsmessage.Message
			if err := query.Unpack(buffer[:count]); err != nil {
				continue
			}
			response := dnsmessage.Message{
				Header: dnsmessage.Header{
					ID:                 query.Header.ID,
					Response:           true,
					Authoritative:      true,
					RecursionDesired:   query.Header.RecursionDesired,
					RecursionAvailable: true,
				},
				Questions: query.Questions,
			}
			answered := false
			for _, question := range query.Questions {
				if question.Type != dnsmessage.TypeA {
					continue
				}
				domain := strings.TrimSuffix(strings.ToLower(question.Name.String()), ".")
				for _, address := range records[domain] {
					ipv4 := address.To4()
					if ipv4 == nil {
						continue
					}
					var packed [4]byte
					copy(packed[:], ipv4)
					response.Answers = append(response.Answers, dnsmessage.Resource{
						Header: dnsmessage.ResourceHeader{
							Name:  question.Name,
							Type:  dnsmessage.TypeA,
							Class: dnsmessage.ClassINET,
							TTL:   60,
						},
						Body: &dnsmessage.AResource{A: packed},
					})
					answered = true
				}
			}
			if !answered {
				response.Header.RCode = dnsmessage.RCodeNameError
			}
			packed, err := response.Pack()
			if err != nil {
				continue
			}
			_, _ = conn.WriteTo(packed, peer)
		}
	}()
	return conn.LocalAddr().String()
}

func TestRunDNSCustomServer(t *testing.T) {
	server := startTestDNSServer(t, map[string][]net.IP{
		"probe.test": {net.ParseIP("127.0.0.1")},
	})
	request := DefaultRequest()
	request.TimeoutMS = 2000
	request.DNSServer = server
	request.DNSRecordType = DNSRecordA

	outcome := Run(context.Background(), DriverDNS, "probe.test", request)
	if !outcome.OK {
		t.Fatalf("dns probe should succeed, got %+v", outcome)
	}
	if outcome.AnswerCount != 1 || outcome.DNSAnswers != "127.0.0.1" || outcome.ResolvedIP != "127.0.0.1" {
		t.Fatalf("unexpected dns result: %+v", outcome)
	}
	if outcome.Value("ANSWER_COUNT") != 1 || outcome.Value("UP") != 1 {
		t.Fatalf("unexpected dns metrics: %+v", outcome)
	}
}

func TestRunDNSExpectValues(t *testing.T) {
	server := startTestDNSServer(t, map[string][]net.IP{
		"probe.test": {net.ParseIP("127.0.0.1")},
	})
	request := DefaultRequest()
	request.TimeoutMS = 2000
	request.DNSServer = server

	request.DNSExpectValues = `127\.0\.0\.1`
	if outcome := Run(context.Background(), DriverDNS, "probe.test", request); !outcome.OK {
		t.Fatalf("expect value should match, got %+v", outcome)
	}

	request.DNSExpectValues = `10\.0\.0\.\d+`
	outcome := Run(context.Background(), DriverDNS, "probe.test", request)
	if outcome.OK || outcome.ErrorKind != ErrorAssert {
		t.Fatalf("expect value mismatch should be assert, got %+v", outcome)
	}

	// 多个期望值按逗号分隔，每个都要至少命中一条记录。
	request.DNSExpectValues = `127\.0\.0\.1,127\.0\.0\.2`
	if outcome := Run(context.Background(), DriverDNS, "probe.test", request); outcome.OK || outcome.ErrorKind != ErrorAssert {
		t.Fatalf("one of two expect values missing should fail, got %+v", outcome)
	}
}

func TestRunDNSNXDomain(t *testing.T) {
	server := startTestDNSServer(t, map[string][]net.IP{})
	request := DefaultRequest()
	request.TimeoutMS = 2000
	request.DNSServer = server
	outcome := Run(context.Background(), DriverDNS, "missing.test", request)
	if outcome.OK || outcome.ErrorKind != ErrorDNS {
		t.Fatalf("NXDOMAIN should be dns failure, got %+v", outcome)
	}
}

// TestRunDNSLocalhost 用系统 resolver 解析 localhost，不依赖外网。
func TestRunDNSLocalhost(t *testing.T) {
	request := DefaultRequest()
	request.TimeoutMS = 2000
	outcome := Run(context.Background(), DriverDNS, "localhost", request)
	if !outcome.OK {
		t.Fatalf("localhost should resolve, got %+v", outcome)
	}
	if outcome.AnswerCount < 1 || outcome.ResolvedIP == "" || outcome.DNSAnswers == "" {
		t.Fatalf("unexpected localhost result: %+v", outcome)
	}
	if outcome.DNSMS < 0 || outcome.DurationMS != outcome.DNSMS {
		t.Fatalf("dns timing: dns=%d duration=%d", outcome.DNSMS, outcome.DurationMS)
	}
}

func TestNormalizeDNSTarget(t *testing.T) {
	cases := []struct {
		raw  string
		want string
	}{
		{raw: "  API.Example.COM. ", want: "api.example.com"},
		{raw: "probe.test", want: "probe.test"},
		{raw: "a-b.example.com", want: "a-b.example.com"},
	}
	for _, tc := range cases {
		got, err := NormalizeTarget(DriverDNS, tc.raw)
		if err != nil {
			t.Fatalf("NormalizeTarget(DNS, %q): %v", tc.raw, err)
		}
		if got != tc.want {
			t.Fatalf("NormalizeTarget(DNS, %q)=%q, want %q", tc.raw, got, tc.want)
		}
	}
	bad := []string{
		"", "127.0.0.1", "::1", "[::1]",
		"example.com:53", "http://example.com", "example.com/path",
		"example.com?x=1", "example.com#frag", "@example.com",
		"-example.com", "example.-com", "exa_mple.com", "exa mple.com",
		strings.Repeat("a", 64) + ".example.com",
	}
	for _, raw := range bad {
		if _, err := NormalizeTarget(DriverDNS, raw); err == nil {
			t.Fatalf("NormalizeTarget(DNS, %q) should fail", raw)
		}
	}
}
