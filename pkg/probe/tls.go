package probe

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/aide-family/moon/pkg/merr"
)

// maxTLSServerNameLength 是 SNI 最大长度。
const maxTLSServerNameLength = 255

// normalizeTLSName 归一化 SNI：转小写、去掉末尾点，空串保持空串（执行时取
// 目标 host）。
func normalizeTLSName(raw string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(raw)), ".")
}

// validateTLSName 校验显式 SNI：只能是主机名，不能带端口、路径、scheme，
// 也不能是 IP 字面量。空串合法，表示使用目标 host。
func validateTLSName(name string) error {
	if name == "" {
		return nil
	}
	if strings.Contains(name, "://") || strings.ContainsAny(name, "/?#@:[] \t\r\n") {
		return merr.Invalid("probe tls server name must be a hostname")
	}
	if net.ParseIP(name) != nil {
		return merr.Invalid("probe tls server name must be a hostname")
	}
	return validateDomain(name)
}

// runTLS 执行一次 TLS 探测：解析 host、建连、握手，再读取协议版本、套件与
// 对端证书信息。verify_cert=true 时要求证书在有效期内，即使跳过了链校验。
func runTLS(ctx context.Context, spec string, req Request) Result {
	req = req.NormalizeForDriver(DriverTLS)
	timeout := time.Duration(req.TimeoutMS) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	started := time.Now()
	host, port, err := net.SplitHostPort(spec)
	if err != nil {
		return Result{ErrorKind: ErrorProtocol, ErrorMessage: "invalid probe target"}
	}

	dnsStart := time.Now()
	addresses, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	dnsMS := millis(time.Since(dnsStart))
	if err != nil {
		out := transportFailure(err, time.Since(started), "")
		out.DNSMS = dnsMS
		return out
	}
	target, found := preferIPAddr(addresses)
	if !found {
		return Result{DurationMS: dnsMS, DNSMS: dnsMS, ErrorKind: ErrorDNS, ErrorMessage: "probe target has no address"}
	}
	resolvedIP := target.IP.String()

	connStart := time.Now()
	dialer := &net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(resolvedIP, port))
	tcpMS := millis(time.Since(connStart))
	if err != nil {
		out := transportFailure(err, time.Since(started), resolvedIP)
		out.DNSMS = dnsMS
		out.TCPMS = tcpMS
		return out
	}

	serverName := req.TLSServerName
	if serverName == "" {
		serverName = host
	}
	tlsConn := tls.Client(conn, &tls.Config{
		ServerName:         serverName,
		InsecureSkipVerify: req.SkipTLSVerify,
	})
	tlsStart := time.Now()
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		tlsMS := millis(time.Since(tlsStart))
		_ = tlsConn.Close()
		out := tlsFailure(err, time.Since(started), resolvedIP)
		out.DNSMS = dnsMS
		out.TCPMS = tcpMS
		out.TLSMS = tlsMS
		return out
	}
	tlsMS := millis(time.Since(tlsStart))
	state := tlsConn.ConnectionState()
	_ = tlsConn.Close()

	out := Result{
		OK:         true,
		DurationMS: millis(time.Since(started)),
		DNSMS:      dnsMS,
		TCPMS:      tcpMS,
		TLSMS:      tlsMS,
		ResolvedIP: resolvedIP,
		TLSVersion: tlsVersionName(state.Version),
		TLSCipher:  tls.CipherSuiteName(state.CipherSuite),
	}
	if len(state.PeerCertificates) == 0 {
		if req.VerifyCert {
			out.OK = false
			out.ErrorKind = ErrorTLS
			out.ErrorMessage = "tls peer certificate is missing"
		}
		return out
	}
	cert := state.PeerCertificates[0]
	out.CertDays = int(time.Until(cert.NotAfter).Hours() / 24)
	out.CertSubject = cert.Subject.String()
	out.CertIssuer = cert.Issuer.String()
	out.CertSANs = strings.Join(cert.DNSNames, ",")
	if req.VerifyCert {
		now := time.Now()
		switch {
		case now.After(cert.NotAfter):
			out.OK = false
			out.ErrorKind = ErrorTLS
			out.ErrorMessage = "tls certificate has expired"
		case now.Before(cert.NotBefore):
			out.OK = false
			out.ErrorKind = ErrorTLS
			out.ErrorMessage = "tls certificate is not valid yet"
		}
	}
	return out
}

// tlsFailure 给握手失败分类：权限、超时单独记，其余都算 tls。
func tlsFailure(err error, elapsed time.Duration, resolvedIP string) Result {
	out := Result{
		DurationMS:   millis(elapsed),
		ResolvedIP:   resolvedIP,
		ErrorMessage: truncate(err.Error(), 200),
	}
	switch {
	case isPermissionError(err):
		out.ErrorKind = ErrorPermission
	case isTimeoutError(err):
		out.ErrorKind = ErrorTimeout
	default:
		out.ErrorKind = ErrorTLS
	}
	return out
}

// tlsVersionName 把 crypto/tls 的版本号转成展示用文本。
func tlsVersionName(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLSv1.0"
	case tls.VersionTLS11:
		return "TLSv1.1"
	case tls.VersionTLS12:
		return "TLSv1.2"
	case tls.VersionTLS13:
		return "TLSv1.3"
	default:
		return "0x" + strconv.FormatUint(uint64(version), 16)
	}
}
