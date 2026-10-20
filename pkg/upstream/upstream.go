package upstream

import (
	"net/url"
	"strings"

	"github.com/aide-family/Alioth/pkg/merr"
)

// apiVersionSuffix 是数据源地址里常见的 API 版本尾巴。
// Prometheus 兼容客户端会自己拼接 /api/v1/...，地址里再带一遍会拼出重复路径。
const apiVersionSuffix = "/api/v1"

// Resolve 把数据源根地址和 API 路径拼成上游请求地址。
//
// 数据源地址允许带路径前缀，例如 VictoriaMetrics 集群的
// http://vmselect:8481/select/0/prometheus。API 路径必须以 / 开头，
// 拼接后保留前缀，不能像绝对路径引用那样把前缀丢掉。
func Resolve(endpoint, apiPath string) (*url.URL, error) {
	base, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil || base.Host == "" || (base.Scheme != "http" && base.Scheme != "https") {
		return nil, merr.Invalid("invalid request")
	}
	if strings.Contains(base.Path, "..") {
		return nil, merr.Invalid("invalid request")
	}
	ref, err := url.Parse(apiPath)
	if err != nil || ref.Scheme != "" || ref.Host != "" {
		return nil, merr.Invalid("invalid request")
	}
	if !strings.HasPrefix(ref.Path, "/") || strings.Contains(ref.Path, "..") || ref.Fragment != "" {
		return nil, merr.Invalid("invalid request")
	}

	prefix := strings.TrimRight(base.Path, "/")
	if hasSuffixFold(prefix, apiVersionSuffix) {
		prefix = strings.TrimRight(prefix[:len(prefix)-len(apiVersionSuffix)], "/")
	}

	target := *base
	target.User = nil
	target.Fragment = ""
	target.RawFragment = ""
	target.RawPath = ""
	target.RawQuery = ref.RawQuery
	if prefix != "" && (ref.Path == prefix || strings.HasPrefix(ref.Path, prefix+"/")) {
		target.Path = ref.Path
	} else {
		target.Path = prefix + ref.Path
	}
	return &target, nil
}

func hasSuffixFold(value, suffix string) bool {
	return len(value) >= len(suffix) && strings.EqualFold(value[len(value)-len(suffix):], suffix)
}
