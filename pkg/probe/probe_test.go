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

	ok := Run(context.Background(), server.URL, DefaultRequest())
	if !ok.OK || ok.StatusCode != http.StatusNoContent {
		t.Fatalf("expected success, got %+v", ok)
	}
	if ok.Value("UP") != 1 || ok.Value("STATUS_CODE") != 204 {
		t.Fatalf("unexpected values: %+v", ok)
	}

	bad := Run(context.Background(), server.URL+"/500", DefaultRequest())
	if bad.OK || bad.ErrorKind != ErrorStatus {
		t.Fatalf("expected status failure, got %+v", bad)
	}

	keyword := DefaultRequest()
	keyword.BodyKeyword = "healthy"
	if got := Run(context.Background(), server.URL+"/keyword", keyword); !got.OK {
		t.Fatalf("keyword should match, got %+v", got)
	}
	keyword.BodyKeyword = "unhealthy"
	if got := Run(context.Background(), server.URL+"/keyword", keyword); got.OK || got.ErrorKind != ErrorKeyword {
		t.Fatalf("keyword should mismatch, got %+v", got)
	}
}
