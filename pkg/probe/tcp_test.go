package probe

import (
	"context"
	"errors"
	"net"
	"os"
	"testing"
)

func TestRunTCP(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = listener.Close() }()

	request := DefaultRequest()
	request.TimeoutMS = 2000
	outcome := Run(context.Background(), DriverTCP, listener.Addr().String(), request)
	if !outcome.OK {
		t.Fatalf("expected TCP connect to succeed, got %+v", outcome)
	}
	if outcome.ResolvedIP != "127.0.0.1" {
		t.Fatalf("ResolvedIP=%q, want 127.0.0.1", outcome.ResolvedIP)
	}
	if outcome.Value("UP") != 1 {
		t.Fatalf("UP should be 1, got %v", outcome.Value("UP"))
	}
	if outcome.Value("LATENCY_MS") != float64(outcome.DurationMS) {
		t.Fatalf("LATENCY_MS=%v, want %d", outcome.Value("LATENCY_MS"), outcome.DurationMS)
	}
	if outcome.TCPMS != outcome.DurationMS {
		t.Fatalf("TCPMS=%d, DurationMS=%d", outcome.TCPMS, outcome.DurationMS)
	}
}

func TestRunTCPConnectionRefused(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	address := listener.Addr().String()
	_ = listener.Close()

	request := DefaultRequest()
	request.TimeoutMS = 2000
	outcome := Run(context.Background(), DriverTCP, address, request)
	if outcome.OK {
		t.Fatalf("closed port should fail, got %+v", outcome)
	}
	if outcome.ErrorKind == "" || outcome.ErrorMessage == "" {
		t.Fatalf("failure should carry classification, got %+v", outcome)
	}
	if outcome.Value("UP") != 0 {
		t.Fatalf("UP should be 0, got %v", outcome.Value("UP"))
	}
}

func TestTransportErrorKind(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{err: context.DeadlineExceeded, want: ErrorTimeout},
		{err: os.ErrPermission, want: ErrorPermission},
		{err: &net.DNSError{Err: "no such host", Name: "probe.invalid"}, want: ErrorDNS},
		{err: errors.New("connect: connection refused"), want: ErrorConnect},
	}
	for _, tc := range cases {
		if got := transportErrorKind(tc.err); got != tc.want {
			t.Fatalf("transportErrorKind(%v)=%q, want %q", tc.err, got, tc.want)
		}
	}
}
