// Package probe 实现 HTTP 探测：目标归一化、期望状态码解析与一次探测执行。
// 天权（megrez）的探测 worker 与玉衡（alioth）的试探测共用这份实现，保证两边
// 「什么算成功」的语义一致。
package probe

import (
	"encoding/json"
	"strings"

	"github.com/aide-family/moon/pkg/merr"
)

// 请求方法只允许幂等动词：一期不发送请求体。
const (
	MethodGet  = "GET"
	MethodHead = "HEAD"
)

// 失败分类。写进 probe_results.error_kind，也用于事件描述。
const (
	ErrorDNS      = "dns"
	ErrorConnect  = "connect"
	ErrorTLS      = "tls"
	ErrorTimeout  = "timeout"
	ErrorStatus   = "status"
	ErrorKeyword  = "keyword"
	ErrorRedirect = "redirect"
	ErrorProtocol = "protocol"
)

// 探测请求参数。JSON 形态与 probe_target_groups.request 一致。
type Request struct {
	// Method 取 GET 或 HEAD。
	Method string `json:"method"`
	// ExpectStatus 是期望状态码表达式：200、200-299、200,204,301。
	ExpectStatus string `json:"expect_status"`
	// TimeoutMS 是单次超时毫秒数。
	TimeoutMS int `json:"timeout_ms"`
	// FollowRedirects 是否跟随重定向，最多 5 跳。
	FollowRedirects bool `json:"follow_redirects"`
	// SkipTLSVerify 跳过证书校验；与 VerifyCert 互斥使用。
	SkipTLSVerify bool `json:"skip_tls_verify"`
	// VerifyCert 校验证书有效期并计算剩余天数。
	VerifyCert bool `json:"verify_cert"`
	// Headers 是自定义请求头。
	Headers map[string]string `json:"headers"`
	// BodyKeyword 命中即成功；留空只按状态码判定。
	BodyKeyword string `json:"body_keyword"`
}

// 默认参数：GET、200-299、3 秒、跟随重定向、校验证书。
func DefaultRequest() Request {
	return Request{
		Method:          MethodGet,
		ExpectStatus:    "200-299",
		TimeoutMS:       3000,
		FollowRedirects: true,
		VerifyCert:      true,
	}
}

// ParseRequest 解析入库的请求参数 JSON；空文本返回默认值。
func ParseRequest(raw string) (Request, error) {
	in := DefaultRequest()
	if strings.TrimSpace(raw) == "" {
		return in, nil
	}
	if err := json.Unmarshal([]byte(raw), &in); err != nil {
		return Request{}, merr.Invalid("invalid probe request")
	}
	return in.Normalize(), nil
}

// Marshal 返回入库用的紧凑 JSON。
func (r Request) Marshal() string {
	out, err := json.Marshal(r.Normalize())
	if err != nil {
		return "{}"
	}
	return string(out)
}

// Normalize 补齐空值并统一大小写。零值字段按默认值处理，便于旧数据平滑。
func (r Request) Normalize() Request {
	out := r
	if strings.TrimSpace(out.Method) == "" {
		out.Method = MethodGet
	}
	out.Method = strings.ToUpper(strings.TrimSpace(out.Method))
	if strings.TrimSpace(out.ExpectStatus) == "" {
		out.ExpectStatus = "200-299"
	}
	out.ExpectStatus = strings.Join(strings.Fields(strings.ReplaceAll(out.ExpectStatus, " ", "")), ",")
	if out.TimeoutMS == 0 {
		out.TimeoutMS = 3000
	}
	if out.Headers == nil {
		out.Headers = map[string]string{}
	}
	out.BodyKeyword = strings.TrimSpace(out.BodyKeyword)
	return out
}

// 受控请求头：由客户端自己计算，探测不允许覆盖。
var controlledHeaders = map[string]struct{}{
	"host":              {},
	"content-length":    {},
	"transfer-encoding": {},
	"connection":        {},
	"upgrade":           {},
}

// Validate 校验参数。与 proto 的 buf.validate 规则保持一致，并补上跨字段约束。
func (r Request) Validate() error {
	req := r.Normalize()
	if req.Method != MethodGet && req.Method != MethodHead {
		return merr.Invalid("unsupported probe method")
	}
	if _, err := ParseExpectStatus(req.ExpectStatus); err != nil {
		return err
	}
	if req.TimeoutMS < 100 || req.TimeoutMS > 60000 {
		return merr.Invalid("probe timeout out of range")
	}
	for key, value := range req.Headers {
		if strings.TrimSpace(key) == "" {
			return merr.Invalid("probe header key is required")
		}
		if _, blocked := controlledHeaders[strings.ToLower(strings.TrimSpace(key))]; blocked {
			return merr.Invalid("probe header is controlled by the client")
		}
		// 换行会拆出额外的头，直接拒绝。
		if strings.ContainsAny(key, "\r\n") || strings.ContainsAny(value, "\r\n") {
			return merr.Invalid("probe header contains newline")
		}
	}
	if len(req.BodyKeyword) > 256 {
		return merr.Invalid("probe body keyword is too long")
	}
	return nil
}

// NormalizeHeaders 去掉空白键值并复制一份，避免调用方修改原始 map。
func (r Request) NormalizeHeaders() map[string]string {
	if len(r.Headers) == 0 {
		return nil
	}
	out := make(map[string]string, len(r.Headers))
	for key, value := range r.Headers {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		out[key] = value
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
