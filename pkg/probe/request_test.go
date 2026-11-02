package probe

import (
	"strings"
	"testing"
)

func TestNormalizeForDriver(t *testing.T) {
	icmp := Request{}.NormalizeForDriver(DriverICMP)
	if icmp.ICMPCount != 1 || icmp.ICMPPacketSize != defaultICMPPacketSize {
		t.Fatalf("ICMP defaults: %+v", icmp)
	}
	kept := Request{ICMPCount: 3, ICMPPacketSize: 128}.NormalizeForDriver(DriverICMP)
	if kept.ICMPCount != 3 || kept.ICMPPacketSize != 128 {
		t.Fatalf("ICMP explicit values changed: %+v", kept)
	}

	dns := Request{}.NormalizeForDriver(DriverDNS)
	if dns.DNSRecordType != DNSRecordA {
		t.Fatalf("DNS default record type: %q", dns.DNSRecordType)
	}
	dns = Request{DNSRecordType: " aaaa ", DNSExpectValues: "  127\\.0\\.0\\.1 ", DNSServer: " 8.8.8.8:53 "}.NormalizeForDriver(DriverDNS)
	if dns.DNSRecordType != DNSRecordAAAA || dns.DNSExpectValues != `127\.0\.0\.1` || dns.DNSServer != "8.8.8.8:53" {
		t.Fatalf("DNS normalization: %+v", dns)
	}

	tls := Request{TLSServerName: " API.Example.COM. "}.NormalizeForDriver(DriverTLS)
	if tls.TLSServerName != "api.example.com" {
		t.Fatalf("TLS SNI normalization: %q", tls.TLSServerName)
	}
	// SNI 留空保持空串，执行时取目标 host。
	if empty := (Request{}).NormalizeForDriver(DriverTLS); empty.TLSServerName != "" {
		t.Fatalf("TLS empty SNI should stay empty: %q", empty.TLSServerName)
	}

	// HTTP / TCP 不受新字段默认值影响。
	if got := (Request{}).NormalizeForDriver(DriverHTTP); got.ICMPCount != 0 || got.DNSRecordType != "" {
		t.Fatalf("HTTP should not apply ICMP/DNS defaults: %+v", got)
	}
	if got := (Request{}).NormalizeForDriver(DriverTCP); got.ICMPCount != 0 || got.DNSRecordType != "" {
		t.Fatalf("TCP should not apply ICMP/DNS defaults: %+v", got)
	}
}

func TestValidateRequestForDriverICMP(t *testing.T) {
	cases := []struct {
		name string
		req  Request
		ok   bool
	}{
		{name: "default", req: Request{}, ok: true},
		{name: "count 0 defaults to 1", req: Request{ICMPCount: 0}, ok: true},
		{name: "count 10", req: Request{ICMPCount: 10}, ok: true},
		{name: "count 11", req: Request{ICMPCount: 11}},
		{name: "size 0 defaults to 56", req: Request{ICMPPacketSize: 0}, ok: true},
		{name: "size 1400", req: Request{ICMPPacketSize: 1400}, ok: true},
		{name: "size 1401", req: Request{ICMPPacketSize: 1401}},
		// ICMP 不使用 HTTP 字段。
		{name: "ignores http fields", req: Request{Method: "POST", ExpectStatus: "abc", Headers: map[string]string{"Host": "x"}}, ok: true},
	}
	for _, tc := range cases {
		err := ValidateRequestForDriver(DriverICMP, tc.req)
		if tc.ok && err != nil {
			t.Fatalf("%s: expected ok, got %v", tc.name, err)
		}
		if !tc.ok && err == nil {
			t.Fatalf("%s: expected error", tc.name)
		}
	}
}

func TestValidateRequestForDriverDNS(t *testing.T) {
	cases := []struct {
		name string
		req  Request
		ok   bool
	}{
		{name: "default", req: Request{}, ok: true},
		{name: "record type aaaa", req: Request{DNSRecordType: "aaaa"}, ok: true},
		{name: "record type mx", req: Request{DNSRecordType: "MX"}, ok: true},
		{name: "record type invalid", req: Request{DNSRecordType: "CAA"}},
		{name: "expect values", req: Request{DNSExpectValues: `127\.0\.0\.1, 10\.0\.0\.\d+`}, ok: true},
		{name: "expect values invalid regex", req: Request{DNSExpectValues: "("}},
		{name: "custom server", req: Request{DNSServer: "8.8.8.8:53"}, ok: true},
		{name: "custom server ipv6", req: Request{DNSServer: "[::1]:53"}, ok: true},
		{name: "custom server missing port", req: Request{DNSServer: "8.8.8.8"}},
		{name: "custom server bad port", req: Request{DNSServer: "8.8.8.8:0"}},
		{name: "custom server scheme", req: Request{DNSServer: "udp://8.8.8.8:53"}},
		{name: "custom server blocked", req: Request{DNSServer: "169.254.169.254:53"}},
		{name: "timeout short", req: Request{TimeoutMS: 50}},
	}
	for _, tc := range cases {
		err := ValidateRequestForDriver(DriverDNS, tc.req)
		if tc.ok && err != nil {
			t.Fatalf("%s: expected ok, got %v", tc.name, err)
		}
		if !tc.ok && err == nil {
			t.Fatalf("%s: expected error", tc.name)
		}
	}
}

func TestValidateRequestForDriverTLS(t *testing.T) {
	cases := []struct {
		name string
		req  Request
		ok   bool
	}{
		{name: "empty sni", req: Request{}, ok: true},
		{name: "hostname sni", req: Request{TLSServerName: "api.example.com"}, ok: true},
		{name: "sni with port", req: Request{TLSServerName: "api.example.com:443"}},
		{name: "sni with scheme", req: Request{TLSServerName: "https://api.example.com"}},
		{name: "sni with path", req: Request{TLSServerName: "api.example.com/x"}},
		{name: "sni ip literal", req: Request{TLSServerName: "127.0.0.1"}},
		{name: "sni with space", req: Request{TLSServerName: "api example.com"}},
		{name: "sni too long", req: Request{TLSServerName: strings.Repeat("a", 256)}},
	}
	for _, tc := range cases {
		err := ValidateRequestForDriver(DriverTLS, tc.req)
		if tc.ok && err != nil {
			t.Fatalf("%s: expected ok, got %v", tc.name, err)
		}
		if !tc.ok && err == nil {
			t.Fatalf("%s: expected error", tc.name)
		}
	}
}

func TestValidateRequestForDriverHTTPAssertions(t *testing.T) {
	cases := []struct {
		name string
		req  Request
		ok   bool
	}{
		{name: "expect header", req: Request{ExpectHeaders: map[string]string{"X-Health": "healthy"}}, ok: true},
		{name: "expect header empty regex", req: Request{ExpectHeaders: map[string]string{"X-Health": ""}}, ok: true},
		{name: "expect header invalid regex", req: Request{ExpectHeaders: map[string]string{"X-Health": "("}}},
		{name: "expect header empty name", req: Request{ExpectHeaders: map[string]string{" ": "healthy"}}},
		{name: "expect header newline", req: Request{ExpectHeaders: map[string]string{"X-Health": "a\nb"}}},
		{name: "body regex", req: Request{BodyRegex: `payload-\d+`}, ok: true},
		{name: "body regex invalid", req: Request{BodyRegex: "("}},
		{name: "body regex too long", req: Request{BodyRegex: strings.Repeat("a", 513)}},
	}
	for _, tc := range cases {
		err := ValidateRequestForDriver(DriverHTTP, tc.req)
		if tc.ok && err != nil {
			t.Fatalf("%s: expected ok, got %v", tc.name, err)
		}
		if !tc.ok && err == nil {
			t.Fatalf("%s: expected error", tc.name)
		}
	}
	// 断言只对 HTTP 生效：TCP 只校验超时。
	if err := ValidateRequestForDriver(DriverTCP, Request{BodyRegex: "(", ExpectHeaders: map[string]string{"X": "("}}); err != nil {
		t.Fatalf("TCP should ignore HTTP assertions: %v", err)
	}
}

func TestRequestJSONRoundTrip(t *testing.T) {
	in := Request{
		Method:          MethodGet,
		ExpectStatus:    "200-299",
		TimeoutMS:       3000,
		ExpectHeaders:   map[string]string{"X-Health": "healthy"},
		BodyRegex:       `payload-\d+`,
		ICMPCount:       3,
		ICMPPacketSize:  64,
		DNSRecordType:   "MX",
		DNSExpectValues: `10 mail\.example\.com`,
		DNSServer:       "8.8.8.8:53",
		TLSServerName:   "api.example.com",
	}
	parsed, err := ParseRequest(in.Marshal())
	if err != nil {
		t.Fatalf("ParseRequest: %v", err)
	}
	if parsed.ExpectHeaders["X-Health"] != "healthy" || parsed.BodyRegex != in.BodyRegex {
		t.Fatalf("HTTP assertions lost: %+v", parsed)
	}
	if parsed.ICMPCount != 3 || parsed.ICMPPacketSize != 64 {
		t.Fatalf("ICMP fields lost: %+v", parsed)
	}
	if parsed.DNSRecordType != "MX" || parsed.DNSExpectValues != in.DNSExpectValues || parsed.DNSServer != "8.8.8.8:53" {
		t.Fatalf("DNS fields lost: %+v", parsed)
	}
	if parsed.TLSServerName != "api.example.com" {
		t.Fatalf("TLS field lost: %+v", parsed)
	}
}
