// Package probe 实现探测：目标归一化、请求校验与一次探测执行。
// 支持 HTTP、TCP、ICMP、DNS、TLS 五种驱动；天权（megrez）的探测 worker 与
// 玉衡（alioth）的试探测共用这份实现，保证两边「什么算成功」的语义一致。
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
	ErrorDNS        = "dns"
	ErrorConnect    = "connect"
	ErrorTLS        = "tls"
	ErrorTimeout    = "timeout"
	ErrorStatus     = "status"
	ErrorKeyword    = "keyword"
	ErrorRedirect   = "redirect"
	ErrorProtocol   = "protocol"
	ErrorPermission = "permission"
	// ErrorAssert 是断言不匹配：响应头、响应体正则或 DNS 期望值不符合预期。
	ErrorAssert = "assert"
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
	// ExpectHeaders 是响应头断言：头名 → 完整匹配正则。
	ExpectHeaders map[string]string `json:"expect_headers"`
	// BodyRegex 是响应体完整匹配正则，≤512；与 BodyKeyword 可同时生效。
	BodyRegex string `json:"body_regex"`
	// ICMPCount 是单次探测连续发送的 echo 数量，0 视为 1。
	ICMPCount int `json:"icmp_count"`
	// ICMPPacketSize 是 ICMP payload 字节数，0 视为 56。
	ICMPPacketSize int `json:"icmp_packet_size"`
	// DNSRecordType 是解析的记录类型，取 A/AAAA/CNAME/MX/TXT/NS，空视为 A。
	DNSRecordType string `json:"dns_record_type"`
	// DNSExpectValues 是逗号分隔的完整匹配正则；每个正则至少命中一条记录。
	DNSExpectValues string `json:"dns_expect_values"`
	// DNSServer 是自定义 DNS 服务器 host:port；留空使用系统 resolver。
	DNSServer string `json:"dns_server"`
	// TLSServerName 是 TLS 握手的 SNI；留空取目标 host。
	TLSServerName string `json:"tls_server_name"`
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
	// 断言对象也补空 map，避免 Marshal 时写成 null。
	if out.ExpectHeaders == nil {
		out.ExpectHeaders = map[string]string{}
	}
	out.BodyKeyword = strings.TrimSpace(out.BodyKeyword)
	return out
}

// NormalizeForDriver 先做通用归一化，再按驱动补默认值：ICMP 补探测次数与
// payload 大小，DNS 补记录类型与修剪字符串，TLS 修剪 SNI（留空由执行器取
// 目标 host）。不适用的字段保持原样，不参与执行。
func (r Request) NormalizeForDriver(driver string) Request {
	out := r.Normalize()
	switch normalizeDriver(driver) {
	case DriverICMP:
		if out.ICMPCount == 0 {
			out.ICMPCount = 1
		}
		if out.ICMPPacketSize == 0 {
			out.ICMPPacketSize = defaultICMPPacketSize
		}
	case DriverDNS:
		out.DNSRecordType = normalizeDNSRecordType(out.DNSRecordType)
		if out.DNSRecordType == "" {
			out.DNSRecordType = DNSRecordA
		}
		out.DNSExpectValues = strings.TrimSpace(out.DNSExpectValues)
		out.DNSServer = strings.TrimSpace(out.DNSServer)
	case DriverTLS:
		out.TLSServerName = normalizeTLSName(out.TLSServerName)
	}
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
	// 响应头断言与响应体正则是 HTTP 专属，校验口径与执行时共用同一份包装。
	if err := req.validateExpectHeaders(); err != nil {
		return err
	}
	if req.BodyRegex != "" {
		if len(req.BodyRegex) > maxBodyRegexLength {
			return merr.Invalid("probe body regex is too long")
		}
		if _, err := compileFullMatch(req.BodyRegex); err != nil {
			return merr.Invalid("invalid probe body regex")
		}
	}
	return nil
}

// validateExpectHeaders 校验响应头断言：头名/值不许有换行，值可编译为正则。
func (r Request) validateExpectHeaders() error {
	for key, value := range r.ExpectHeaders {
		if strings.TrimSpace(key) == "" {
			return merr.Invalid("probe expect header key is required")
		}
		if strings.ContainsAny(key, "\r\n") || strings.ContainsAny(value, "\r\n") {
			return merr.Invalid("probe header contains newline")
		}
		if _, err := compileFullMatch(strings.TrimSpace(value)); err != nil {
			return merr.Invalid("invalid probe expect header regex")
		}
	}
	return nil
}

// ValidateRequestForDriver 按驱动校验请求参数。所有驱动都校验超时，再叠加
// 各自用到的字段：HTTP 走完整校验（含断言），TCP 只用超时，ICMP 加次数与
// payload，DNS 加记录类型/期望值/自定义服务器，TLS 加 SNI。
func ValidateRequestForDriver(driver string, req Request) error {
	driver = normalizeDriver(driver)
	if !ValidDriver(driver) {
		return merr.Invalid("unsupported probe driver")
	}
	req = req.NormalizeForDriver(driver)
	switch driver {
	case DriverHTTP:
		return req.Validate()
	case DriverTCP:
		return validateProbeTimeout(req.TimeoutMS)
	case DriverICMP:
		if err := validateProbeTimeout(req.TimeoutMS); err != nil {
			return err
		}
		if req.ICMPCount < 1 || req.ICMPCount > maxICMPCount {
			return merr.Invalid("probe icmp count out of range")
		}
		if req.ICMPPacketSize < 1 || req.ICMPPacketSize > maxICMPPacketSize {
			return merr.Invalid("probe icmp packet size out of range")
		}
		return nil
	case DriverDNS:
		if err := validateProbeTimeout(req.TimeoutMS); err != nil {
			return err
		}
		if !validDNSRecordType(req.DNSRecordType) {
			return merr.Invalid("unsupported probe dns record type")
		}
		if len(req.DNSExpectValues) > maxDNSExpectValuesLength {
			return merr.Invalid("probe dns expect values is too long")
		}
		if _, err := parseDNSExpectValues(req.DNSExpectValues); err != nil {
			return err
		}
		if req.DNSServer != "" {
			if len(req.DNSServer) > maxDNSServerLength {
				return merr.Invalid("probe dns server is too long")
			}
			if _, err := dnsServerAddress(req.DNSServer); err != nil {
				return err
			}
		}
		return nil
	case DriverTLS:
		if err := validateProbeTimeout(req.TimeoutMS); err != nil {
			return err
		}
		if len(req.TLSServerName) > maxTLSServerNameLength {
			return merr.Invalid("probe tls server name is too long")
		}
		return validateTLSName(req.TLSServerName)
	default:
		return merr.Invalid("unsupported probe driver")
	}
}

// validateProbeTimeout 校验单次超时范围。所有驱动共用同一区间。
func validateProbeTimeout(timeoutMS int) error {
	if timeoutMS < 100 || timeoutMS > 60000 {
		return merr.Invalid("probe timeout out of range")
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
