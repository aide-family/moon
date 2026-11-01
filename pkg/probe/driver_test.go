package probe

import "testing"

func TestValidDriver(t *testing.T) {
	for _, driver := range []string{DriverHTTP, DriverTCP, DriverICMP, " probe_tcp "} {
		if !ValidDriver(driver) {
			t.Fatalf("ValidDriver(%q) should be true", driver)
		}
	}
	for _, driver := range []string{"", "PROBE_UDP", "HTTP"} {
		if ValidDriver(driver) {
			t.Fatalf("ValidDriver(%q) should be false", driver)
		}
	}
}

func TestNormalizeTarget(t *testing.T) {
	cases := []struct {
		driver string
		raw    string
		want   string
	}{
		{driver: DriverHTTP, raw: "  HTTPS://API.Example.com:443/healthz#frag ", want: "https://api.example.com/healthz"},
		{driver: DriverTCP, raw: " API.Example.com:8443 ", want: "api.example.com:8443"},
		{driver: DriverTCP, raw: "[::1]:80", want: "[::1]:80"},
		{driver: DriverICMP, raw: "API.Example.com", want: "api.example.com"},
		{driver: DriverICMP, raw: "::1", want: "::1"},
		{driver: DriverICMP, raw: "[::1]", want: "::1"},
	}
	for _, tc := range cases {
		got, err := NormalizeTarget(tc.driver, tc.raw)
		if err != nil {
			t.Fatalf("NormalizeTarget(%q, %q): %v", tc.driver, tc.raw, err)
		}
		if got != tc.want {
			t.Fatalf("NormalizeTarget(%q, %q)=%q, want %q", tc.driver, tc.raw, got, tc.want)
		}
	}

	bad := []struct {
		driver string
		raw    string
	}{
		{driver: DriverTCP, raw: ""},
		{driver: DriverTCP, raw: "api.example.com"},
		{driver: DriverTCP, raw: "https://api.example.com:443"},
		{driver: DriverTCP, raw: "api.example.com/path:80"},
		{driver: DriverTCP, raw: "api.example.com:0"},
		{driver: DriverTCP, raw: "api.example.com:65536"},
		{driver: DriverTCP, raw: "api.example.com:8o"},
		{driver: DriverTCP, raw: "user@api.example.com:80"},
		{driver: DriverTCP, raw: "169.254.169.254:80"},
		{driver: DriverICMP, raw: ""},
		{driver: DriverICMP, raw: "api.example.com:80"},
		{driver: DriverICMP, raw: "http://api.example.com"},
		{driver: DriverICMP, raw: "api.example.com/healthz"},
		{driver: DriverICMP, raw: "169.254.169.254"},
		{driver: DriverICMP, raw: "0.0.0.0"},
		{driver: "PROBE_UDP", raw: "api.example.com"},
	}
	for _, tc := range bad {
		if _, err := NormalizeTarget(tc.driver, tc.raw); err == nil {
			t.Fatalf("NormalizeTarget(%q, %q) should fail", tc.driver, tc.raw)
		}
	}
}

func TestValidateRequestForDriver(t *testing.T) {
	for _, driver := range []string{DriverTCP, DriverICMP} {
		if err := ValidateRequestForDriver(driver, DefaultRequest()); err != nil {
			t.Fatalf("ValidateRequestForDriver(%q) default: %v", driver, err)
		}
		// 零值超时补成默认 3000，仍通过。
		if err := ValidateRequestForDriver(driver, Request{}); err != nil {
			t.Fatalf("ValidateRequestForDriver(%q) zero timeout: %v", driver, err)
		}
		if err := ValidateRequestForDriver(driver, Request{TimeoutMS: 50}); err == nil {
			t.Fatalf("ValidateRequestForDriver(%q) should reject short timeout", driver)
		}
		if err := ValidateRequestForDriver(driver, Request{TimeoutMS: 60001}); err == nil {
			t.Fatalf("ValidateRequestForDriver(%q) should reject long timeout", driver)
		}
	}

	// TCP/ICMP 不使用 method、期望状态码与请求头，这些字段不参与校验。
	if err := ValidateRequestForDriver(DriverTCP, Request{Method: "POST", ExpectStatus: "abc", Headers: map[string]string{"Host": "x"}}); err != nil {
		t.Fatalf("TCP only checks timeout, got %v", err)
	}
	if err := ValidateRequestForDriver(DriverHTTP, Request{Method: "POST"}); err == nil {
		t.Fatal("HTTP should reject unsupported method")
	}
	if err := ValidateRequestForDriver("PROBE_UDP", DefaultRequest()); err == nil {
		t.Fatal("unknown driver should fail")
	}
}

func TestRunUnknownDriver(t *testing.T) {
	out := Run(t.Context(), "PROBE_UDP", "example.com", DefaultRequest())
	if out.OK || out.ErrorKind != ErrorProtocol {
		t.Fatalf("unexpected result for unknown driver: %+v", out)
	}
}
