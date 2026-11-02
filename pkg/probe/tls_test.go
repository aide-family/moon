package probe

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestCertificate 生成一张自签服务端证书，SAN 覆盖 127.0.0.1 与 probe.example。
func newTestCertificate(t *testing.T, notBefore, notAfter time.Time) tls.Certificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "probe.test"},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"probe.example"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
}

// startTestTLSServer 起一个本机 TLS 握手服务器，返回目标地址与收到的 SNI 通道。
func startTestTLSServer(t *testing.T, cert tls.Certificate) (string, <-chan string) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen tls: %v", err)
	}
	captured := make(chan string, 1)
	tlsListener := tls.NewListener(listener, &tls.Config{
		Certificates: []tls.Certificate{cert},
		GetConfigForClient: func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
			select {
			case captured <- hello.ServerName:
			default:
			}
			return nil, nil
		},
	})
	t.Cleanup(func() { _ = tlsListener.Close() })
	go func() {
		for {
			conn, err := tlsListener.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				defer func() { _ = conn.Close() }()
				if tlsConn, ok := conn.(*tls.Conn); ok {
					_ = tlsConn.Handshake()
				}
			}(conn)
		}
	}()
	return listener.Addr().String(), captured
}

func TestRunTLSSelfSigned(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	request := DefaultRequest()
	request.TimeoutMS = 3000
	request.SkipTLSVerify = true
	request.VerifyCert = true
	outcome := Run(context.Background(), DriverTLS, server.Listener.Addr().String(), request)
	if !outcome.OK {
		t.Fatalf("tls handshake should succeed, got %+v", outcome)
	}
	if !strings.HasPrefix(outcome.TLSVersion, "TLSv1.") {
		t.Fatalf("unexpected tls version: %q", outcome.TLSVersion)
	}
	if outcome.TLSCipher == "" || outcome.CertSubject == "" || outcome.CertIssuer == "" {
		t.Fatalf("missing tls/cert details: %+v", outcome)
	}
	if outcome.CertDays <= 0 {
		t.Fatalf("cert days should be positive: %d", outcome.CertDays)
	}
	if outcome.ResolvedIP != "127.0.0.1" {
		t.Fatalf("ResolvedIP=%q", outcome.ResolvedIP)
	}
	if outcome.Value("CERT_DAYS") != float64(outcome.CertDays) ||
		outcome.Value("TLS_MS") != float64(outcome.TLSMS) ||
		outcome.Value("TCP_CONNECT_MS") != float64(outcome.TCPMS) ||
		outcome.Value("LATENCY_MS") != float64(outcome.DurationMS) {
		t.Fatalf("unexpected tls metrics: %+v", outcome)
	}
}

func TestRunTLSVerificationFailure(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	request := DefaultRequest()
	request.TimeoutMS = 3000
	request.SkipTLSVerify = false
	request.VerifyCert = true
	outcome := Run(context.Background(), DriverTLS, server.Listener.Addr().String(), request)
	if outcome.OK || outcome.ErrorKind != ErrorTLS {
		t.Fatalf("untrusted certificate should be tls failure, got %+v", outcome)
	}
	if outcome.TLSMS < 0 || outcome.TCPMS < 0 {
		t.Fatalf("timings should not be negative: %+v", outcome)
	}
}

func TestRunTLSExpiredCertificate(t *testing.T) {
	cert := newTestCertificate(t, time.Now().Add(-48*time.Hour), time.Now().Add(-24*time.Hour))
	address, _ := startTestTLSServer(t, cert)

	request := DefaultRequest()
	request.TimeoutMS = 3000
	request.SkipTLSVerify = true
	request.VerifyCert = true
	outcome := Run(context.Background(), DriverTLS, address, request)
	if outcome.OK || outcome.ErrorKind != ErrorTLS {
		t.Fatalf("expired certificate should be tls failure, got %+v", outcome)
	}
	if !strings.Contains(outcome.ErrorMessage, "expired") {
		t.Fatalf("error message should mention expiry: %q", outcome.ErrorMessage)
	}
}

func TestRunTLSExpiredCertificateNotVerified(t *testing.T) {
	cert := newTestCertificate(t, time.Now().Add(-48*time.Hour), time.Now().Add(-24*time.Hour))
	address, _ := startTestTLSServer(t, cert)

	// 跳过链校验且不校验证书有效期时，握手成功即算成功。
	request := DefaultRequest()
	request.TimeoutMS = 3000
	request.SkipTLSVerify = true
	request.VerifyCert = false
	outcome := Run(context.Background(), DriverTLS, address, request)
	if !outcome.OK {
		t.Fatalf("handshake should succeed without cert verification, got %+v", outcome)
	}
	if outcome.CertDays >= 0 {
		t.Fatalf("expired certificate days should be negative: %d", outcome.CertDays)
	}
}

func TestRunTLSServerName(t *testing.T) {
	cert := newTestCertificate(t, time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	address, captured := startTestTLSServer(t, cert)

	request := DefaultRequest()
	request.TimeoutMS = 3000
	request.SkipTLSVerify = true
	request.TLSServerName = "probe.example"
	if outcome := Run(context.Background(), DriverTLS, address, request); !outcome.OK {
		t.Fatalf("tls probe should succeed, got %+v", outcome)
	}
	select {
	case sni := <-captured:
		if sni != "probe.example" {
			t.Fatalf("SNI=%q, want probe.example", sni)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("tls server did not receive the client hello")
	}
}

func TestTLSVersionName(t *testing.T) {
	cases := map[uint16]string{
		tls.VersionTLS10: "TLSv1.0",
		tls.VersionTLS11: "TLSv1.1",
		tls.VersionTLS12: "TLSv1.2",
		tls.VersionTLS13: "TLSv1.3",
	}
	for version, want := range cases {
		if got := tlsVersionName(version); got != want {
			t.Fatalf("tlsVersionName(%d)=%q, want %q", version, got, want)
		}
	}
	if got := tlsVersionName(0x1234); !strings.HasPrefix(got, "0x") {
		t.Fatalf("unknown version should render as hex, got %q", got)
	}
}
