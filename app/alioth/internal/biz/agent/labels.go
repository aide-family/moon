package agent

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

const (
	// maxAgentLabelPairs 限制单台机器的标签对数，避免一行 JSON 无上限膨胀。
	maxAgentLabelPairs = 32
	// maxAgentLabelKeyLength 与 maxAgentLabelValueLength 限制键值长度。
	maxAgentLabelKeyLength   = 63
	maxAgentLabelValueLength = 63
)

// agentLabelPattern 是标签键值允许的字符集：字母、数字、点、下划线和短横线。
var agentLabelPattern = regexp.MustCompile(`^[A-Za-z0-9._-]*$`)

// NormalizeAgentLabels 校验机器自定义标签并重新编码为紧凑 JSON 对象。
// 空串与空白视为空对象 {}；非 JSON 对象、非字符串值、键值超长或含非法字符都返回客户端错误。
func NormalizeAgentLabels(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "{}", nil
	}
	var values map[string]string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return "", merr.Invalid("invalid labels")
	}
	if values == nil {
		values = map[string]string{}
	}
	if len(values) > maxAgentLabelPairs {
		return "", merr.Invalid("too many labels")
	}
	for key, value := range values {
		if !validAgentLabelKey(key) || !validAgentLabelValue(value) {
			return "", merr.Invalid("invalid labels")
		}
	}
	out, err := json.Marshal(values)
	if err != nil {
		return "", merr.Invalid("invalid labels")
	}
	return string(out), nil
}

// validAgentLabelKey 判断标签键是否满足 1-63 字符且字符集合法。
func validAgentLabelKey(key string) bool {
	return len(key) > 0 && len(key) <= maxAgentLabelKeyLength && agentLabelPattern.MatchString(key)
}

// validAgentLabelValue 判断标签值是否满足 0-63 字符且字符集合法。
func validAgentLabelValue(value string) bool {
	return len(value) <= maxAgentLabelValueLength && agentLabelPattern.MatchString(value)
}

// Match 按标签解析候选主机。只返回当前租户 ACTIVE 的机器，
// 在线状态以 Hub 为准；标签为空属于客户端错误，不返回全部机器。
func (registry *Registry) Match(ctx context.Context, labels map[string]string) ([]bo.Agent, error) {
	ns, err := registry.access.Allow(ctx, perm.SSHExec)
	if err != nil {
		return nil, err
	}
	if len(labels) == 0 {
		return nil, merr.Invalid("labels are required")
	}
	items, err := registry.agents.MatchAgents(ctx, ns, labels)
	if err != nil {
		return nil, registry.access.Wrap(err)
	}
	for i := range items {
		items[i].Online = registry.hub.Online(items[i].UID)
	}
	return items, nil
}
