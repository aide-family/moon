package probe

import (
	"context"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"

	"github.com/aide-family/moon/pkg/merr"
)

// normalizeICMPTarget 归一化 ICMP 目标：只接受主机名或 IP 字面量，不带端口、
// scheme、路径。IPv6 字面量去掉方括号，主机统一小写。
func normalizeICMPTarget(raw string) (string, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return "", merr.Invalid("probe target is required")
	}
	if strings.Contains(text, "://") {
		return "", merr.Invalid("probe target must be a host without scheme")
	}
	// [::1] 这种带方括号的写法照收，输出统一去掉方括号。
	if len(text) >= 2 && text[0] == '[' && text[len(text)-1] == ']' {
		text = text[1 : len(text)-1]
	}
	if strings.ContainsAny(text, "/?#") {
		return "", merr.Invalid("probe target must be a host without path")
	}
	if ip := net.ParseIP(text); ip != nil {
		return normalizeProbeHost(ip.String())
	}
	// 带冒号却解析不出 IP，通常是 host:port 或非法 IPv6。
	if strings.Contains(text, ":") {
		return "", merr.Invalid("probe target must be a host without port")
	}
	return normalizeProbeHost(text)
}

// runICMP 发一次 ICMP echo 请求并等待回包：收到 echo reply 即成功，耗时取 RTT。
// 优先用非特权 ping 套接字，失败再回退 raw 套接字；两者都没有权限时返回
// permission 分类，不 panic。ctx 取消或超时时关闭连接，避免阻塞在 ReadFrom 上。
func runICMP(ctx context.Context, spec string, req Request) Result {
	req = req.Normalize()
	timeout := time.Duration(req.TimeoutMS) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	started := time.Now()
	addresses, err := net.DefaultResolver.LookupIPAddr(ctx, spec)
	if err != nil {
		return transportFailure(err, time.Since(started), "")
	}
	// 优先 IPv4：多数内网目标只开 IPv4，双栈主机先试 v4 更稳。
	var target net.IPAddr
	found := false
	for _, address := range addresses {
		if !found {
			target = address
			found = true
		}
		if address.IP.To4() != nil {
			target = address
			break
		}
	}
	if !found {
		return Result{DurationMS: millis(time.Since(started)), ErrorKind: ErrorDNS, ErrorMessage: "probe target has no address"}
	}
	ipv4Target := target.IP.To4() != nil

	conn, datagram, err := listenICMP(ipv4Target)
	if err != nil {
		return transportFailure(err, time.Since(started), target.IP.String())
	}
	defer func() { _ = conn.Close() }()

	// ReadFrom 会一直阻塞：ctx 结束时关掉连接让读立即返回。
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			_ = conn.Close()
		case <-done:
		}
	}()

	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(timeout)
	}
	if err := conn.SetReadDeadline(deadline); err != nil {
		return transportFailure(err, time.Since(started), target.IP.String())
	}

	var (
		messageType icmp.Type
		destination net.Addr
		protocol    int
	)
	if ipv4Target {
		messageType = ipv4.ICMPTypeEcho
		protocol = 1
	} else {
		messageType = ipv6.ICMPTypeEchoRequest
		protocol = 58
	}
	if datagram {
		destination = &net.UDPAddr{IP: target.IP, Zone: target.Zone}
	} else {
		destination = &net.IPAddr{IP: target.IP, Zone: target.Zone}
	}

	const sequence = 1
	message := &icmp.Message{
		Type: messageType,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  sequence,
			Data: []byte("alioth-probe"),
		},
	}
	payload, err := message.Marshal(nil)
	if err != nil {
		return transportFailure(err, time.Since(started), target.IP.String())
	}
	if _, err := conn.WriteTo(payload, destination); err != nil {
		return transportFailure(err, time.Since(started), target.IP.String())
	}

	buffer := make([]byte, 1500)
	for {
		count, _, err := conn.ReadFrom(buffer)
		if err != nil {
			// 连接是被 ctx 关闭的：优先按 ctx 的原因分类，避免误报成 connect。
			if ctxErr := ctx.Err(); ctxErr != nil {
				return transportFailure(ctxErr, time.Since(started), target.IP.String())
			}
			return transportFailure(err, time.Since(started), target.IP.String())
		}
		reply, err := icmp.ParseMessage(protocol, buffer[:count])
		if err != nil {
			continue
		}
		if !isEchoReply(reply.Type, ipv4Target) {
			continue
		}
		echo, ok := reply.Body.(*icmp.Echo)
		// 非特权套接字由内核分配 ident，只按 seq 认自己的回包。
		if !ok || echo.Seq != sequence {
			continue
		}
		return Result{
			OK:         true,
			DurationMS: millis(time.Since(started)),
			ResolvedIP: target.IP.String(),
		}
	}
}

// icmpEndpoint 是一组套接字候选：先非特权 datagram，再 raw。
type icmpEndpoint struct {
	network  string
	address  string
	datagram bool
}

// listenICMP 先尝试非特权 ping 套接字，再回退 raw 套接字。返回的 datagram
// 表示是否用的 udp4/udp6 套接字，写目标地址时类型不同。
func listenICMP(ipv4Target bool) (*icmp.PacketConn, bool, error) {
	endpoints := []icmpEndpoint{
		{network: "udp4", address: "0.0.0.0", datagram: true},
		{network: "ip4:icmp", address: "0.0.0.0", datagram: false},
	}
	if !ipv4Target {
		endpoints = []icmpEndpoint{
			{network: "udp6", address: "::", datagram: true},
			{network: "ip6:ipv6-icmp", address: "::", datagram: false},
		}
	}
	var (
		permissionErr error
		lastErr       error
	)
	for _, endpoint := range endpoints {
		conn, err := icmp.ListenPacket(endpoint.network, endpoint.address)
		if err == nil {
			return conn, endpoint.datagram, nil
		}
		if permissionErr == nil && isPermissionError(err) {
			permissionErr = err
		}
		lastErr = err
	}
	// 有权限错误时优先返回它：调用方据此把 error_kind 记成 permission。
	if permissionErr != nil {
		return nil, false, permissionErr
	}
	return nil, false, lastErr
}

// isEchoReply 判断收到的 ICMP 消息是不是 echo reply。
func isEchoReply(messageType icmp.Type, ipv4Target bool) bool {
	if ipv4Target {
		return messageType == ipv4.ICMPTypeEchoReply
	}
	return messageType == ipv6.ICMPTypeEchoReply
}
