package probe

import (
	"net"
	"net/url"
	"strings"

	"github.com/aide-family/moon/pkg/merr"
)

// blockedNetworks 是探测默认拒绝的网段：云厂商元数据地址、链路本地与组播。
// 私网地址默认允许——内网探活是探测的主要用途。
var blockedNetworks = []*net.IPNet{
	mustCIDR("169.254.0.0/16"),
	mustCIDR("fe80::/10"),
	mustCIDR("ff00::/8"),
	mustCIDR("224.0.0.0/4"),
}

// BlockedHost 判断一个字面地址是否落在拒绝网段里。DNS 解析后不做二次校验，
// 只挡住直接写死元数据地址的配置。
func BlockedHost(host string) bool {
	ip := net.ParseIP(strings.Trim(host, "[]"))
	if ip == nil {
		return false
	}
	if ip.IsUnspecified() {
		return true
	}
	for _, network := range blockedNetworks {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// NormalizeSpec 归一化探测目标：只接受 http/https，去掉 userinfo、默认端口与
// 片段，其余（路径、查询串）保留，因为不同查询串是不同的观测对象。
func NormalizeSpec(raw string) (string, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return "", merr.Invalid("probe target is required")
	}
	parsed, err := url.Parse(text)
	if err != nil {
		return "", merr.Invalid("invalid probe target")
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", merr.Invalid("probe target must be http or https")
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "" {
		return "", merr.Invalid("probe target host is required")
	}
	if BlockedHost(host) {
		return "", merr.Invalid("probe target host is blocked")
	}
	port := parsed.Port()
	if (parsed.Scheme == "http" && port == "80") || (parsed.Scheme == "https" && port == "443") {
		port = ""
	}
	if port != "" {
		parsed.Host = net.JoinHostPort(host, port)
	} else if strings.Contains(host, ":") {
		// IPv6 字面量需要方括号。
		parsed.Host = "[" + host + "]"
	} else {
		parsed.Host = host
	}
	parsed.User = nil
	parsed.Fragment = ""
	parsed.RawFragment = ""
	return parsed.String(), nil
}

func mustCIDR(text string) *net.IPNet {
	_, network, err := net.ParseCIDR(text)
	if err != nil {
		panic(err)
	}
	return network
}
