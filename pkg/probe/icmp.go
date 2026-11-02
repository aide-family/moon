package probe

import (
	"bytes"
	"context"
	"fmt"
	"math"
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

// ICMP 请求的取值范围与默认值。
const (
	maxICMPCount          = 10
	maxICMPPacketSize     = 1400
	defaultICMPPacketSize = 56
	// minICMPPacketWindow 是单包最短读等待：总超时很小时也要给回包留出时间。
	minICMPPacketWindow = 50 * time.Millisecond
)

// runICMP 连续发送 icmp_count 个 echo 请求并统计回包：至少收到 1 个即成功，
// 全部丢失按 timeout 失败。优先用非特权 ping 套接字，失败再回退 raw 套接字；
// 两者都没有权限时返回 permission 分类，不 panic。ctx 取消或超时时关闭连接，
// 避免阻塞在 ReadFrom 上。
func runICMP(ctx context.Context, spec string, req Request) Result {
	req = req.NormalizeForDriver(DriverICMP)
	timeout := time.Duration(req.TimeoutMS) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	started := time.Now()
	addresses, err := net.DefaultResolver.LookupIPAddr(ctx, spec)
	if err != nil {
		return transportFailure(err, time.Since(started), "")
	}
	target, found := preferIPAddr(addresses)
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

	totalDeadline, ok := ctx.Deadline()
	if !ok {
		totalDeadline = time.Now().Add(timeout)
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

	// 每包读超时取 timeout/count（最少 50ms），总时长仍受 ctx 约束。
	perPacket := timeout / time.Duration(req.ICMPCount)
	if perPacket < minICMPPacketWindow {
		perPacket = minICMPPacketWindow
	}
	data := bytes.Repeat([]byte("a"), req.ICMPPacketSize)
	buffer := make([]byte, 1500)
	sentTimes := make(map[int]time.Time, req.ICMPCount)
	received := make(map[int]bool, req.ICMPCount)
	rtts := make([]time.Duration, 0, req.ICMPCount)

	for sequence := 1; sequence <= req.ICMPCount; sequence++ {
		if ctx.Err() != nil {
			break
		}
		message := &icmp.Message{
			Type: messageType,
			Code: 0,
			Body: &icmp.Echo{
				ID:   os.Getpid() & 0xffff,
				Seq:  sequence,
				Data: data,
			},
		}
		payload, err := message.Marshal(nil)
		if err != nil {
			if len(rtts) > 0 {
				break
			}
			return transportFailure(err, time.Since(started), target.IP.String())
		}
		sentTimes[sequence] = time.Now()
		if _, err := conn.WriteTo(payload, destination); err != nil {
			// 已经收到回包说明权限与链路都正常，后面的包按丢包处理。
			if len(rtts) > 0 {
				break
			}
			return transportFailure(err, time.Since(started), target.IP.String())
		}

		readDeadline := time.Now().Add(perPacket)
		if readDeadline.After(totalDeadline) {
			readDeadline = totalDeadline
		}
		if err := conn.SetReadDeadline(readDeadline); err != nil {
			break
		}
		for {
			count, _, err := conn.ReadFrom(buffer)
			if err != nil {
				// 该包读取超时：记为丢失，继续发下一包；ctx 结束由外层退出。
				break
			}
			reply, err := icmp.ParseMessage(protocol, buffer[:count])
			if err != nil {
				continue
			}
			if !isEchoReply(reply.Type, ipv4Target) {
				continue
			}
			echo, ok := reply.Body.(*icmp.Echo)
			if !ok {
				continue
			}
			// 非特权套接字由内核分配 ident，只按 seq 认自己的回包。
			seq := echo.Seq
			if seq < 1 || seq > req.ICMPCount || received[seq] {
				continue
			}
			sentAt, ok := sentTimes[seq]
			if !ok {
				continue
			}
			received[seq] = true
			rtts = append(rtts, time.Since(sentAt))
			// 乱序回包先记下，继续等当前包的应答。
			if seq == sequence {
				break
			}
		}
	}

	avgMS, jitterMS, minMS, maxMS, loss := icmpStats(req.ICMPCount, rtts)
	out := Result{
		DurationMS: avgMS,
		PacketLoss: loss,
		JitterMS:   jitterMS,
		MinRTTMS:   minMS,
		MaxRTTMS:   maxMS,
		ResolvedIP: target.IP.String(),
	}
	if len(rtts) == 0 {
		out.ErrorKind = ErrorTimeout
		out.ErrorMessage = fmt.Sprintf("all %d icmp packets lost", req.ICMPCount)
		return out
	}
	out.OK = true
	return out
}

// icmpStats 根据发送次数与收到的 RTT 算出平均 RTT、抖动、极值与丢包率。
// 抖动取相邻回包 RTT 的平均绝对差；少于两个回包时为 0。丢包率四舍五入。
func icmpStats(sent int, rtts []time.Duration) (avgMS, jitterMS, minMS, maxMS int64, loss int) {
	if sent <= 0 {
		return 0, 0, 0, 0, 0
	}
	loss = int(math.Round(float64(sent-len(rtts)) * 100 / float64(sent)))
	if len(rtts) == 0 {
		return 0, 0, 0, 0, loss
	}
	var total, minRTT, maxRTT time.Duration
	for index, rtt := range rtts {
		total += rtt
		if index == 0 || rtt < minRTT {
			minRTT = rtt
		}
		if index == 0 || rtt > maxRTT {
			maxRTT = rtt
		}
	}
	avgMS = millis(total / time.Duration(len(rtts)))
	minMS = millis(minRTT)
	maxMS = millis(maxRTT)
	if len(rtts) > 1 {
		var diff time.Duration
		for index := 1; index < len(rtts); index++ {
			d := rtts[index] - rtts[index-1]
			if d < 0 {
				d = -d
			}
			diff += d
		}
		jitterMS = millis(diff / time.Duration(len(rtts)-1))
	}
	return avgMS, jitterMS, minMS, maxMS, loss
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
