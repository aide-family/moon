package probe

import (
	"context"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/aide-family/moon/pkg/merr"
)

// normalizeTCPTarget 归一化 TCP 目标：只接受 host:port，主机小写，IPv6 字面量
// 统一带方括号。带 scheme、路径、查询串或片段的写法一律拒绝，避免把 HTTP
// 地址误配成 TCP 目标。端口必须显式给出，不做默认端口推断。
func normalizeTCPTarget(raw string) (string, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return "", merr.Invalid("probe target is required")
	}
	if strings.Contains(text, "://") {
		return "", merr.Invalid("probe target must be host:port")
	}
	if strings.ContainsAny(text, "/?#") {
		return "", merr.Invalid("probe target must be host:port")
	}
	host, port, err := net.SplitHostPort(text)
	if err != nil {
		return "", merr.Invalid("probe target must be host:port")
	}
	host, err = normalizeProbeHost(host)
	if err != nil {
		return "", err
	}
	if !validPort(port) {
		return "", merr.Invalid("probe target port must be 1..65535")
	}
	return net.JoinHostPort(host, port), nil
}

// validPort 只接受十进制 1..65535，拒绝前导零与正负号。
func validPort(port string) bool {
	if port == "" || len(port) > 5 {
		return false
	}
	for _, char := range port {
		if char < '0' || char > '9' {
			return false
		}
	}
	value, err := strconv.Atoi(port)
	if err != nil || value < 1 || value > 65535 {
		return false
	}
	return true
}

// runTCP 对 TCP 目标做一次连接探测：建连成功即成功，耗时取建连时间。
// 先单独解析域名并记录 dns_ms，再对解析出的地址建连，这样 DNS_MS 与
// TCP_CONNECT_MS 两个指标都有值，解析失败也能单独按 dns 分类。
func runTCP(ctx context.Context, spec string, req Request) Result {
	req = req.Normalize()
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
	target, ok := preferIPAddr(addresses)
	if !ok {
		return Result{DurationMS: dnsMS, DNSMS: dnsMS, ErrorKind: ErrorDNS, ErrorMessage: "probe target has no address"}
	}

	connStart := time.Now()
	dialer := &net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(target.IP.String(), port))
	tcpMS := millis(time.Since(connStart))
	if err != nil {
		out := transportFailure(err, time.Since(started), target.IP.String())
		out.DNSMS = dnsMS
		out.TCPMS = tcpMS
		return out
	}
	defer func() { _ = conn.Close() }()
	// LATENCY_MS 对 TCP 取建连耗时，与 M26 语义一致。
	return Result{
		OK:         true,
		DurationMS: tcpMS,
		DNSMS:      dnsMS,
		TCPMS:      tcpMS,
		ResolvedIP: remoteIP(conn),
	}
}

// remoteIP 取已连接对端的 IP 字面量。
func remoteIP(conn net.Conn) string {
	address, ok := conn.RemoteAddr().(*net.TCPAddr)
	if !ok {
		return ""
	}
	return address.IP.String()
}
