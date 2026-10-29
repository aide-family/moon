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
	"time"
)

// Result 是一次探测的结果。Value 是给规则判定用的观测值，取决于规则的
// probe_metric：UP 取 1/0，LATENCY_MS 取总耗时，STATUS_CODE 取状态码，
// CERT_DAYS 取证书剩余天数。
type Result struct {
	// OK 表示这次探测是否算成功（状态码、关键字与连接都满足）。
	OK bool
	// StatusCode 是 HTTP 状态码，连接失败时为 0。
	StatusCode int
	// DurationMS 是整个请求的耗时。
	DurationMS int64
	// ErrorKind 与 ErrorMessage 在失败时有值。
	ErrorKind    string
	ErrorMessage string
	// 耗时分段，只用于展示。
	DNSMS  int64
	TCPMS  int64
	TLSMS  int64
	TTFBMS int64
	// CertDays 是证书剩余天数，非 HTTPS 或跳过校验时为 0。
	CertDays int
	// ResolvedIP 是本次解析到的地址。
	ResolvedIP string
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
	default:
		if r.OK {
			return 1
		}
		return 0
	}
}

// maxBodyBytes 是关键字判定最多读取的响应体大小。
const maxBodyBytes = 64 << 10

// maxRedirects 是最多跟随的重定向跳数。
const maxRedirects = 5

// Run 对归一化后的目标执行一次探测。调用方负责传入 ctx 与超时。
func Run(ctx context.Context, spec string, req Request) Result {
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
	if req.BodyKeyword != "" && !strings.Contains(string(body), req.BodyKeyword) {
		out.ErrorKind = ErrorKeyword
		out.ErrorMessage = "response body does not contain the keyword"
		return out
	}
	out.OK = true
	return out
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
