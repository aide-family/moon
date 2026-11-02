package probe

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNormalizeSpec(t *testing.T) {
	cases := map[string]string{
		"  HTTPS://API.Example.com:443/healthz#frag ": "https://api.example.com/healthz",
		"http://user:pass@example.com:80/data?q=1":    "http://example.com/data?q=1",
		"https://example.com:8443/x":                  "https://example.com:8443/x",
	}
	for in, want := range cases {
		got, err := NormalizeSpec(in)
		if err != nil {
			t.Fatalf("NormalizeSpec(%q): %v", in, err)
		}
		if got != want {
			t.Fatalf("NormalizeSpec(%q)=%q, want %q", in, got, want)
		}
	}
	for _, bad := range []string{"", "example.com", "ftp://example.com", "http://169.254.169.254/latest", "http:///x"} {
		if _, err := NormalizeSpec(bad); err == nil {
			t.Fatalf("NormalizeSpec(%q) should fail", bad)
		}
	}
}

func TestParseExpectStatus(t *testing.T) {
	ranges, err := ParseExpectStatus("200,204-206")
	if err != nil {
		t.Fatal(err)
	}
	if !MatchStatus(ranges, 200) || !MatchStatus(ranges, 205) || MatchStatus(ranges, 201) {
		t.Fatalf("unexpected match result: %+v", ranges)
	}
	if _, err := ParseExpectStatus("200-100"); err == nil {
		t.Fatal("reversed range should fail")
	}
	if _, err := ParseExpectStatus("abc"); err == nil {
		t.Fatal("non numeric should fail")
	}
}

func TestRunSuccessAndFailures(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/keyword" {
			_, _ = w.Write([]byte("healthy"))
			return
		}
		if r.URL.Path == "/500" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	ok := Run(context.Background(), DriverHTTP, server.URL, DefaultRequest())
	if !ok.OK || ok.StatusCode != http.StatusNoContent {
		t.Fatalf("expected success, got %+v", ok)
	}
	if ok.Value("UP") != 1 || ok.Value("STATUS_CODE") != 204 {
		t.Fatalf("unexpected values: %+v", ok)
	}

	bad := Run(context.Background(), DriverHTTP, server.URL+"/500", DefaultRequest())
	if bad.OK || bad.ErrorKind != ErrorStatus {
		t.Fatalf("expected status failure, got %+v", bad)
	}

	keyword := DefaultRequest()
	keyword.BodyKeyword = "healthy"
	if got := Run(context.Background(), DriverHTTP, server.URL+"/keyword", keyword); !got.OK {
		t.Fatalf("keyword should match, got %+v", got)
	}
	keyword.BodyKeyword = "unhealthy"
	if got := Run(context.Background(), DriverHTTP, server.URL+"/keyword", keyword); got.OK || got.ErrorKind != ErrorKeyword {
		t.Fatalf("keyword should mismatch, got %+v", got)
	}
}

func TestRunHTTPAssertions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Health", "healthy")
		_, _ = w.Write([]byte("payload-123"))
	}))
	defer server.Close()

	run := func(mutate func(*Request)) Result {
		request := DefaultRequest()
		mutate(&request)
		return Run(context.Background(), DriverHTTP, server.URL, request)
	}

	ok := run(func(request *Request) {
		request.ExpectHeaders = map[string]string{"X-Health": "healthy"}
		request.BodyRegex = `payload-\d+`
	})
	if !ok.OK {
		t.Fatalf("assertions should match, got %+v", ok)
	}

	headerMismatch := run(func(request *Request) {
		request.ExpectHeaders = map[string]string{"X-Health": "unhealthy"}
	})
	if headerMismatch.OK || headerMismatch.ErrorKind != ErrorAssert {
		t.Fatalf("header mismatch should be assert, got %+v", headerMismatch)
	}
	if headerMismatch.Value("UP") != 0 {
		t.Fatalf("UP should be 0, got %v", headerMismatch.Value("UP"))
	}

	missingHeader := run(func(request *Request) {
		request.ExpectHeaders = map[string]string{"X-Missing": "anything"}
	})
	if missingHeader.OK || missingHeader.ErrorKind != ErrorAssert {
		t.Fatalf("missing header should be assert, got %+v", missingHeader)
	}

	bodyMismatch := run(func(request *Request) {
		request.BodyRegex = `payload-\d{5}`
	})
	if bodyMismatch.OK || bodyMismatch.ErrorKind != ErrorAssert {
		t.Fatalf("body regex mismatch should be assert, got %+v", bodyMismatch)
	}

	// 完整匹配而不是包含：正文是 payload-123，单写 payload 不算命中。
	partial := run(func(request *Request) { request.BodyRegex = "payload" })
	if partial.OK || partial.ErrorKind != ErrorAssert {
		t.Fatalf("partial match should fail, got %+v", partial)
	}

	// 关键字与正则同时配置时两者都要满足。
	both := run(func(request *Request) {
		request.BodyKeyword = "payload"
		request.BodyRegex = `payload-\d+`
	})
	if !both.OK {
		t.Fatalf("keyword and regex should both match, got %+v", both)
	}
}

func TestResultValueMetrics(t *testing.T) {
	result := Result{
		OK:          true,
		DurationMS:  120,
		StatusCode:  204,
		CertDays:    30,
		DNSMS:       3,
		TCPMS:       4,
		TLSMS:       5,
		TTFBMS:      6,
		PacketLoss:  25,
		JitterMS:    7,
		MinRTTMS:    8,
		MaxRTTMS:    9,
		AnswerCount: 2,
	}
	cases := map[string]float64{
		"UP":             1,
		"LATENCY_MS":     120,
		"STATUS_CODE":    204,
		"CERT_DAYS":      30,
		"DNS_MS":         3,
		"TCP_CONNECT_MS": 4,
		"TLS_MS":         5,
		"TTFB_MS":        6,
		"PACKET_LOSS":    25,
		"JITTER_MS":      7,
		"MIN_RTT_MS":     8,
		"MAX_RTT_MS":     9,
		"ANSWER_COUNT":   2,
	}
	for metric, want := range cases {
		if got := result.Value(metric); got != want {
			t.Fatalf("Value(%q)=%v, want %v", metric, got, want)
		}
	}
	failed := Result{ErrorKind: ErrorAssert}
	if failed.Value("UP") != 0 {
		t.Fatalf("failed UP should be 0, got %v", failed.Value("UP"))
	}
}
