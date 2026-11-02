package probe

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"os"
	"strings"
	"syscall"
	"time"
)

// Result 是一次探测的结果。Value 是给规则判定用的观测值，取决于规则的
// probe_metric：UP 取 1/0，LATENCY_MS 取本次探测耗时，STATUS_CODE 取状态码，
// CERT_DAYS 取证书剩余天数，其余指标取对应的细分字段。
type Result struct {
	// OK 表示这次探测是否算成功（HTTP 看连接、状态码、关键字与断言，TCP 看
	// 建连，ICMP 看是否收到回包，DNS 看解析与期望值，TLS 看握手与证书）。
	OK bool
	// StatusCode 是 HTTP 状态码，连接失败或非 HTTP 驱动时为 0。
	StatusCode int
	// DurationMS 是这次探测的耗时：HTTP 是整请求，TCP 是建连，ICMP 是平均
	// RTT，DNS 是解析耗时，TLS 是整次握手耗时。
	DurationMS int64
	// ErrorKind 与 ErrorMessage 在失败时有值。
	ErrorKind    string
	ErrorMessage string
	// 耗时分段，只用于展示。
	DNSMS  int64
	TCPMS  int64
	TLSMS  int64
	TTFBMS int64
	// CertDays 是证书剩余天数，非 HTTPS/TLS 或未读到证书时为 0。
	CertDays int
	// ResolvedIP 是本次解析到的地址。
	ResolvedIP string
	// ICMP 统计：丢包百分比 0–100、相邻回包抖动、RTT 极值。
	PacketLoss int
	JitterMS   int64
	MinRTTMS   int64
	MaxRTTMS   int64
	// DNS 结果：记录条数与逗号分隔的答案文本。
	AnswerCount int
	DNSAnswers  string
	// TLS 握手与对端证书信息。
	TLSVersion  string
	TLSCipher   string
	CertSubject string
	CertIssuer  string
	CertSANs    string
}

// Value 按观测指标取出用于判定的数值。
func (r Result) Value(metric string) float64 {
	switch metric {
	case "LATENCY_MS":
		return float64(r.DurationMS)
	case "STATUS_CODE":
		return float64(r.StatusCode)
	case "CERT_DAYS":
		return float64(r.CertDays)
	case "DNS_MS":
		return float64(r.DNSMS)
	case "TCP_CONNECT_MS":
		return float64(r.TCPMS)
	case "TLS_MS":
		return float64(r.TLSMS)
	case "TTFB_MS":
		return float64(r.TTFBMS)
	case "PACKET_LOSS":
		return float64(r.PacketLoss)
	case "JITTER_MS":
		return float64(r.JitterMS)
	case "MIN_RTT_MS":
		return float64(r.MinRTTMS)
	case "MAX_RTT_MS":
		return float64(r.MaxRTTMS)
	case "ANSWER_COUNT":
		return float64(r.AnswerCount)
	default:
		if r.OK {
			return 1
		}
		return 0
	}
}

// maxBodyBytes 是关键字判定最多读取的响应体大小。
const maxBodyBytes = 64 << 10

// maxBodyRegexLength 是响应体正则的最大长度。
const maxBodyRegexLength = 512

// maxRedirects 是最多跟随的重定向跳数。
const maxRedirects = 5

// Run 按驱动对归一化后的目标执行一次探测。调用方负责传入 ctx；单次超时
// 由 req.TimeoutMS 决定，这里会再包一层，避免调用方漏传时无限等待。
func Run(ctx context.Context, driver, spec string, req Request) Result {
	switch normalizeDriver(driver) {
	case DriverHTTP:
		return runHTTP(ctx, spec, req)
	case DriverTCP:
		return runTCP(ctx, spec, req)
	case DriverICMP:
		return runICMP(ctx, spec, req)
	case DriverDNS:
		return runDNS(ctx, spec, req)
	case DriverTLS:
		return runTLS(ctx, spec, req)
	default:
		return Result{ErrorKind: ErrorProtocol, ErrorMessage: "unsupported probe driver"}
	}
}

// runHTTP 执行一次 HTTP 探测：连接、TLS、期望状态码与关键字都满足才算成功。
func runHTTP(ctx context.Context, spec string, req Request) Result {
	req = req.Normalize()
	timeout := time.Duration(req.TimeoutMS) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var (
		dnsStart, dnsDone           time.Time
		connStart, connDone         time.Time
		tlsStart, tlsDone           time.Time
		firstByte                   time.Time
		resolvedIP                  string
		dnsMS, tcpMS, tlsMS, ttfbMS int64
	)
	trace := &httptrace.ClientTrace{
		DNSStart: func(httptrace.DNSStartInfo) { dnsStart = time.Now() },
		DNSDone: func(info httptrace.DNSDoneInfo) {
			dnsDone = time.Now()
			if len(info.Addrs) > 0 {
				resolvedIP = info.Addrs[0].IP.String()
			}
		},
		ConnectStart:         func(_, _ string) { connStart = time.Now() },
		ConnectDone:          func(_, _ string, _ error) { connDone = time.Now() },
		TLSHandshakeStart:    func() { tlsStart = time.Now() },
		TLSHandshakeDone:     func(tls.ConnectionState, error) { tlsDone = time.Now() },
		GotFirstResponseByte: func() { firstByte = time.Now() },
	}
	ctx = httptrace.WithClientTrace(ctx, trace)

	ranges, err := ParseExpectStatus(req.ExpectStatus)
	if err != nil {
		return Result{ErrorKind: ErrorProtocol, ErrorMessage: err.Error()}
	}

	transport := &http.Transport{
		// 探测直连目标，不经过环境变量里的代理，避免把内网探活条件隐藏掉。
		Proxy:                 nil,
		ForceAttemptHTTP2:     false,
		MaxIdleConns:          1,
		IdleConnTimeout:       time.Second,
		TLSHandshakeTimeout:   timeout,
		ResponseHeaderTimeout: timeout,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: req.SkipTLSVerify},
	}
	client := &http.Client{Transport: transport, Timeout: timeout}
	if !req.FollowRedirects {
		client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	} else {
		client.CheckRedirect = func(_ *http.Request, via []*http.Request) error {
			if len(via) > maxRedirects {
				return errTooManyRedirects
			}
			return nil
		}
	}

	request, err := http.NewRequestWithContext(ctx, req.Method, spec, nil)
	if err != nil {
		return Result{ErrorKind: ErrorProtocol, ErrorMessage: err.Error()}
	}
	for key, value := range req.NormalizeHeaders() {
		request.Header.Set(key, value)
	}

	started := time.Now()
	response, err := client.Do(request)
	if err != nil {
		return classifyFailure(err, time.Since(started), resolvedIP, dnsStart, dnsDone, connStart, connDone, tlsStart, tlsDone)
	}
	defer func() { _ = response.Body.Close() }()

	body := make([]byte, 0, 0)
	if req.Method != MethodHead {
		body, _ = io.ReadAll(io.LimitReader(response.Body, maxBodyBytes))
	}
	finished := time.Now()

	if dnsStart.IsZero() {
		dnsStart = started
	}
	if dnsDone.IsZero() {
		dnsDone = started
	}
	dnsMS = millis(dnsDone.Sub(dnsStart))
	if !connStart.IsZero() {
		tcpMS = millis(connDone.Sub(connStart))
	}
	if !tlsStart.IsZero() {
		tlsMS = millis(tlsDone.Sub(tlsStart))
	}
	if !firstByte.IsZero() {
		ttfbMS = millis(firstByte.Sub(started))
	}

	out := Result{
		StatusCode: response.StatusCode,
		DurationMS: millis(finished.Sub(started)),
		DNSMS:      dnsMS,
		TCPMS:      tcpMS,
		TLSMS:      tlsMS,
		TTFBMS:     ttfbMS,
		ResolvedIP: resolvedIP,
	}
	if response.TLS != nil && len(response.TLS.PeerCertificates) > 0 && req.VerifyCert {
		cert := response.TLS.PeerCertificates[0]
		out.CertDays = int(time.Until(cert.NotAfter).Hours() / 24)
	}
	if !MatchStatus(ranges, response.StatusCode) {
		out.ErrorKind = ErrorStatus
		out.ErrorMessage = "status " + http.StatusText(response.StatusCode)
		return out
	}
	if kind, message := checkResponseAssertions(req, response.Header, body); kind != "" {
		out.ErrorKind = kind
		out.ErrorMessage = message
		return out
	}
	out.OK = true
	return out
}

// checkResponseAssertions 检查响应头断言、响应体关键字与响应体正则。
// 全部通过返回空分类；断言不匹配返回 assert，正则本身非法返回 protocol。
func checkResponseAssertions(req Request, header http.Header, body []byte) (string, string) {
	for name, pattern := range req.ExpectHeaders {
		name = strings.TrimSpace(name)
		values := header.Values(name)
		if len(values) == 0 {
			return ErrorAssert, "response header " + name + " is missing"
		}
		matcher, err := compileFullMatch(strings.TrimSpace(pattern))
		if err != nil {
			return ErrorProtocol, "invalid probe expect header regex"
		}
		if !matcher.MatchString(strings.Join(values, ", ")) {
			return ErrorAssert, "response header " + name + " does not match the expected value"
		}
	}
	if req.BodyKeyword != "" && !strings.Contains(string(body), req.BodyKeyword) {
		return ErrorKeyword, "response body does not contain the keyword"
	}
	if req.BodyRegex != "" {
		matcher, err := compileFullMatch(req.BodyRegex)
		if err != nil {
			return ErrorProtocol, "invalid probe body regex"
		}
		if !matcher.MatchString(string(body)) {
			return ErrorAssert, "response body does not match the regex"
		}
	}
	return "", ""
}

var errTooManyRedirects = errors.New("too many redirects")

// classifyFailure 按错误类型填 error_kind，方便事件描述和排查。
func classifyFailure(err error, elapsed time.Duration, resolvedIP string, dnsStart, dnsDone, connStart, connDone, tlsStart, tlsDone time.Time) Result {
	out := Result{
		DurationMS: millis(elapsed),
		ResolvedIP: resolvedIP,
		ErrorKind:  ErrorConnect,
	}
	if !dnsStart.IsZero() && !dnsDone.IsZero() {
		out.DNSMS = millis(dnsDone.Sub(dnsStart))
	}
	if !connStart.IsZero() && !connDone.IsZero() {
		out.TCPMS = millis(connDone.Sub(connStart))
	}
	if !tlsStart.IsZero() && !tlsDone.IsZero() {
		out.TLSMS = millis(tlsDone.Sub(tlsStart))
	}

	var netErr net.Error
	var dnsErr *net.DNSError
	var recordErr tls.RecordHeaderError
	switch {
	case errors.Is(err, errTooManyRedirects):
		out.ErrorKind = ErrorRedirect
	case errors.As(err, &dnsErr):
		out.ErrorKind = ErrorDNS
	case os.IsTimeout(err) || (errors.As(err, &netErr) && netErr.Timeout()) || errors.Is(err, context.DeadlineExceeded):
		out.ErrorKind = ErrorTimeout
	case errors.As(err, &recordErr):
		out.ErrorKind = ErrorTLS
	case strings.Contains(err.Error(), "tls:") || strings.Contains(err.Error(), "certificate"):
		out.ErrorKind = ErrorTLS
	case strings.Contains(err.Error(), "redirect"):
		out.ErrorKind = ErrorRedirect
	}
	out.ErrorMessage = truncate(err.Error(), 200)
	return out
}

func millis(d time.Duration) int64 {
	if d <= 0 {
		return 0
	}
	return d.Milliseconds()
}

func truncate(text string, limit int) string {
	if len(text) <= limit {
		return text
	}
	return text[:limit]
}

// preferIPAddr 从解析结果里挑一个地址：优先 IPv4，双栈目标先试 v4 更稳。
func preferIPAddr(addresses []net.IPAddr) (net.IPAddr, bool) {
	for _, address := range addresses {
		if address.IP.To4() != nil {
			return address, true
		}
	}
	if len(addresses) > 0 {
		return addresses[0], true
	}
	return net.IPAddr{}, false
}

// isTimeoutError 判断错误是不是超时：DNS/TLS 失败分类时用它把超时单列。
func isTimeoutError(err error) bool {
	var netErr net.Error
	return os.IsTimeout(err) || (errors.As(err, &netErr) && netErr.Timeout()) || errors.Is(err, context.DeadlineExceeded)
}

// transportFailure 把 TCP、ICMP 的底层错误转成失败结果，分类口径与 HTTP 一致。
func transportFailure(err error, elapsed time.Duration, resolvedIP string) Result {
	return Result{
		DurationMS:   millis(elapsed),
		ResolvedIP:   resolvedIP,
		ErrorKind:    transportErrorKind(err),
		ErrorMessage: truncate(err.Error(), 200),
	}
}

// transportErrorKind 按错误类型给 TCP、ICMP 的失败分类。没权限单独记
// permission：线上多半是容器没给 CAP_NET_RAW，可与目标故障区分开。
func transportErrorKind(err error) string {
	var netErr net.Error
	var dnsErr *net.DNSError
	switch {
	case isPermissionError(err):
		return ErrorPermission
	case errors.As(err, &dnsErr):
		return ErrorDNS
	case os.IsTimeout(err) || (errors.As(err, &netErr) && netErr.Timeout()) || errors.Is(err, context.DeadlineExceeded):
		return ErrorTimeout
	default:
		return ErrorConnect
	}
}

// isPermissionError 判断错误是不是权限不足：非特权 ICMP 被关、raw 套接字
// 需要特权都会落到这里。
func isPermissionError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, os.ErrPermission) || errors.Is(err, syscall.EACCES) || errors.Is(err, syscall.EPERM) {
		return true
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "permission denied") || strings.Contains(text, "operation not permitted")
}
