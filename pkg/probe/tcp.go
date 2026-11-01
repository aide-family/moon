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
// 建连包含 DNS 解析，解析失败按 dns 分类。
func runTCP(ctx context.Context, spec string, req Request) Result {
	req = req.Normalize()
	timeout := time.Duration(req.TimeoutMS) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	started := time.Now()
	dialer := &net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", spec)
	if err != nil {
		return transportFailure(err, time.Since(started), "")
	}
	defer func() { _ = conn.Close() }()
	elapsed := time.Since(started)
	return Result{
		OK:         true,
		DurationMS: millis(elapsed),
		TCPMS:      millis(elapsed),
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
