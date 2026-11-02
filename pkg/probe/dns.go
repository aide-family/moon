package probe

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/aide-family/moon/pkg/merr"
)

// DNS 记录类型。取值与 proto 的枚举一致，空值在 NormalizeForDriver 里补成 A。
const (
	DNSRecordA     = "A"
	DNSRecordAAAA  = "AAAA"
	DNSRecordCNAME = "CNAME"
	DNSRecordMX    = "MX"
	DNSRecordTXT   = "TXT"
	DNSRecordNS    = "NS"
)

const (
	// maxDNSExpectValuesLength 是期望值表达式最大长度。
	maxDNSExpectValuesLength = 1024
	// maxDNSServerLength 是自定义 DNS 服务器地址最大长度。
	maxDNSServerLength = 255
	// maxDomainLength 是域名总长度上限（不含末尾点）。
	maxDomainLength = 253
	// maxDomainLabelLength 是单个域名标签的长度上限。
	maxDomainLabelLength = 63
)

// dnsRecordTypes 是支持的记录类型白名单。
var dnsRecordTypes = map[string]struct{}{
	DNSRecordA:     {},
	DNSRecordAAAA:  {},
	DNSRecordCNAME: {},
	DNSRecordMX:    {},
	DNSRecordTXT:   {},
	DNSRecordNS:    {},
}

// normalizeDNSRecordType 统一记录类型的大小写，空串保持空串由默认值补齐。
func normalizeDNSRecordType(raw string) string {
	return strings.ToUpper(strings.TrimSpace(raw))
}

// validDNSRecordType 判断记录类型是否在支持范围内。
func validDNSRecordType(recordType string) bool {
	_, ok := dnsRecordTypes[recordType]
	return ok
}

// normalizeDNSTarget 归一化 DNS 目标：只接受域名，转小写并去掉末尾点。
// 带 scheme、端口、路径、空白或 IP 字面量的写法一律拒绝。
func normalizeDNSTarget(raw string) (string, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return "", merr.Invalid("probe target is required")
	}
	if strings.Contains(text, "://") {
		return "", merr.Invalid("probe target must be a domain")
	}
	if strings.ContainsAny(text, "/?#@:") {
		return "", merr.Invalid("probe target must be a domain")
	}
	domain := strings.ToLower(strings.TrimSuffix(text, "."))
	if net.ParseIP(domain) != nil {
		return "", merr.Invalid("probe target must be a domain")
	}
	if err := validateDomain(domain); err != nil {
		return "", err
	}
	return domain, nil
}

// validateDomain 校验域名的标签规则：总长 1–253，标签 1–63，只含字母、
// 数字与连字符，且不以连字符开头或结尾。
func validateDomain(domain string) error {
	if domain == "" || len(domain) > maxDomainLength {
		return merr.Invalid("invalid probe target domain")
	}
	for _, label := range strings.Split(domain, ".") {
		if label == "" || len(label) > maxDomainLabelLength {
			return merr.Invalid("invalid probe target domain")
		}
		for index := 0; index < len(label); index++ {
			char := label[index]
			alpha := char >= 'a' && char <= 'z'
			digit := char >= '0' && char <= '9'
			if !alpha && !digit && char != '-' {
				return merr.Invalid("invalid probe target domain")
			}
			if char == '-' && (index == 0 || index == len(label)-1) {
				return merr.Invalid("invalid probe target domain")
			}
		}
	}
	return nil
}

// dnsServerAddress 校验并归一化自定义 DNS 服务器地址：必须是 host:port，
// 端口 1–65535，主机不能是云元数据等保留网段地址。
func dnsServerAddress(raw string) (string, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return "", merr.Invalid("probe dns server is required")
	}
	if strings.Contains(text, "://") || strings.ContainsAny(text, "/?#") {
		return "", merr.Invalid("probe dns server must be host:port")
	}
	host, port, err := net.SplitHostPort(text)
	if err != nil {
		return "", merr.Invalid("probe dns server must be host:port")
	}
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return "", merr.Invalid("probe dns server host is required")
	}
	if strings.ContainsAny(host, " \t\r\n@[]") {
		return "", merr.Invalid("invalid probe dns server host")
	}
	if BlockedHost(host) {
		return "", merr.Invalid("probe dns server host is blocked")
	}
	if !validPort(port) {
		return "", merr.Invalid("probe dns server port must be 1..65535")
	}
	return net.JoinHostPort(host, port), nil
}

// parseDNSExpectValues 把逗号分隔的完整匹配正则解析成列表；空表达式返回空。
func parseDNSExpectValues(raw string) ([]*regexp.Regexp, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return nil, nil
	}
	parts := strings.Split(text, ",")
	patterns := make([]*regexp.Regexp, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		pattern, err := compileFullMatch(part)
		if err != nil {
			return nil, merr.Invalid("invalid dns expect value regex")
		}
		patterns = append(patterns, pattern)
	}
	return patterns, nil
}

// runDNS 按记录类型解析目标：解析成功且每个期望值正则命中一条记录才算成功。
// 自定义 dns_server 时用独立的 Go resolver，只把查询转发到指定服务器。
func runDNS(ctx context.Context, spec string, req Request) Result {
	req = req.NormalizeForDriver(DriverDNS)
	timeout := time.Duration(req.TimeoutMS) * time.Millisecond
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resolver := net.DefaultResolver
	if req.DNSServer != "" {
		server, err := dnsServerAddress(req.DNSServer)
		if err != nil {
			return Result{ErrorKind: ErrorProtocol, ErrorMessage: truncate(err.Error(), 200)}
		}
		resolver = customResolver(server, timeout)
	}

	started := time.Now()
	answers, resolvedIP, err := resolveDNS(ctx, resolver, spec, req.DNSRecordType)
	dnsMS := millis(time.Since(started))
	if err != nil {
		kind := ErrorDNS
		if isTimeoutError(err) {
			kind = ErrorTimeout
		}
		return Result{
			DurationMS:   dnsMS,
			DNSMS:        dnsMS,
			ErrorKind:    kind,
			ErrorMessage: truncate(err.Error(), 200),
		}
	}

	out := Result{
		OK:          true,
		DurationMS:  dnsMS,
		DNSMS:       dnsMS,
		AnswerCount: len(answers),
		DNSAnswers:  strings.Join(answers, ","),
		ResolvedIP:  resolvedIP,
	}
	if len(answers) == 0 {
		out.OK = false
		out.ErrorKind = ErrorAssert
		out.ErrorMessage = "probe target has no dns records"
		return out
	}
	patterns, err := parseDNSExpectValues(req.DNSExpectValues)
	if err != nil {
		return Result{DurationMS: dnsMS, DNSMS: dnsMS, ErrorKind: ErrorProtocol, ErrorMessage: truncate(err.Error(), 200)}
	}
	for _, pattern := range patterns {
		matched := false
		for _, answer := range answers {
			if pattern.MatchString(answer) {
				matched = true
				break
			}
		}
		if !matched {
			out.OK = false
			out.ErrorKind = ErrorAssert
			out.ErrorMessage = "dns answer does not match the expected value"
			return out
		}
	}
	return out
}

// customResolver 构造只向指定 DNS 服务器发查询的 resolver。Go resolver 在
// 响应被截断时会改用 tcp，所以这里按传入 network 建连。
func customResolver(server string, timeout time.Duration) *net.Resolver {
	dialer := &net.Dialer{Timeout: timeout}
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, server)
		},
	}
}

// resolveDNS 按记录类型解析目标，返回答案文本与（A/AAAA 的）解析地址。
// CNAME/MX/NS 的答案是主机名，MX 带优先级前缀，TXT 是各字符串拼接后的文本。
func resolveDNS(ctx context.Context, resolver *net.Resolver, domain, recordType string) ([]string, string, error) {
	switch recordType {
	case DNSRecordA, DNSRecordAAAA:
		network := "ip4"
		if recordType == DNSRecordAAAA {
			network = "ip6"
		}
		addresses, err := resolver.LookupIP(ctx, network, domain)
		if err != nil {
			return nil, "", err
		}
		answers := make([]string, 0, len(addresses))
		for _, address := range addresses {
			answers = append(answers, address.String())
		}
		resolvedIP := ""
		if len(answers) > 0 {
			resolvedIP = answers[0]
		}
		return answers, resolvedIP, nil
	case DNSRecordCNAME:
		name, err := resolver.LookupCNAME(ctx, domain)
		if err != nil {
			return nil, "", err
		}
		return []string{strings.TrimSuffix(name, ".")}, "", nil
	case DNSRecordMX:
		records, err := resolver.LookupMX(ctx, domain)
		if err != nil {
			return nil, "", err
		}
		answers := make([]string, 0, len(records))
		for _, record := range records {
			answers = append(answers, fmt.Sprintf("%d %s", record.Pref, strings.TrimSuffix(record.Host, ".")))
		}
		return answers, "", nil
	case DNSRecordTXT:
		records, err := resolver.LookupTXT(ctx, domain)
		if err != nil {
			return nil, "", err
		}
		return records, "", nil
	case DNSRecordNS:
		records, err := resolver.LookupNS(ctx, domain)
		if err != nil {
			return nil, "", err
		}
		answers := make([]string, 0, len(records))
		for _, record := range records {
			answers = append(answers, strings.TrimSuffix(record.Host, "."))
		}
		return answers, "", nil
	default:
		return nil, "", merr.Invalid("unsupported probe dns record type")
	}
}
