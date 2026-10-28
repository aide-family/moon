// Package labels 解析标签 JSON，并按标签值正则判断匹配器是否命中。
package labels

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/aide-family/moon/pkg/merr"
)

// ParseObject 把 JSON 对象解析成 map。空串视为空对象。
func ParseObject(raw string) (map[string]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]string{}, nil
	}
	var values map[string]string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil, merr.Invalid("invalid labels")
	}
	if values == nil {
		return map[string]string{}, nil
	}
	return values, nil
}

// NormalizeObject 校验并重新编码为紧凑 JSON 对象。
func NormalizeObject(raw string) (string, error) {
	values, err := ParseObject(raw)
	if err != nil {
		return "", err
	}
	out, err := json.Marshal(values)
	if err != nil {
		return "", merr.Invalid("invalid labels")
	}
	return string(out), nil
}

// Pattern 把匹配器值包装成完整匹配的正则。标签值是 RE2 表达式，
// 必须整段命中，和 Prometheus 的 =~ 语义一致；纯字面值仍然只匹配自身。
func Pattern(value string) string {
	return "^(?:" + value + ")$"
}

// ValidatePattern 检查匹配器值能否作为完整匹配正则使用。
func ValidatePattern(value string) error {
	if _, err := regexp.Compile(Pattern(value)); err != nil {
		return merr.Invalid("invalid label matcher")
	}
	return nil
}

// NormalizeMatchers 校验匹配器值并按紧凑 JSON 重新编码。
// 非法正则直接拒绝，避免保存后才在过滤时静默失效。
func NormalizeMatchers(raw string) (string, error) {
	values, err := ParseObject(raw)
	if err != nil {
		return "", err
	}
	for _, value := range values {
		if err := ValidatePattern(value); err != nil {
			return "", err
		}
	}
	out, err := json.Marshal(values)
	if err != nil {
		return "", merr.Invalid("invalid labels")
	}
	return string(out), nil
}

// CompileMatchers 预编译匹配器，便于在同一个请求里对多条目标标签复用。
// 非法表达式保留字面值，和 MatchersContained 的兼容行为一致。
func CompileMatchers(matchers map[string]string) func(target map[string]string) bool {
	type compiledMatcher struct {
		re      *regexp.Regexp
		literal string
	}
	entries := make(map[string]compiledMatcher, len(matchers))
	for key, matcher := range matchers {
		re, err := regexp.Compile(Pattern(matcher))
		if err != nil {
			// 兼容历史数据里保存过的非法表达式：退回字面比较，
			// 避免一条旧记录让整页过滤或通知分发失效。
			entries[key] = compiledMatcher{literal: matcher}
			continue
		}
		entries[key] = compiledMatcher{re: re}
	}
	return func(target map[string]string) bool {
		for key, matcher := range entries {
			value, ok := target[key]
			if !ok {
				return false
			}
			if matcher.re != nil {
				if !matcher.re.MatchString(value) {
					return false
				}
				continue
			}
			if value != matcher.literal {
				return false
			}
		}
		return true
	}
}

// MatchersContained 判断 matchers 的每个键是否都能在 target 中以正则命中。
// 键不存在时不命中；匹配器值为空时同样要求 target 中存在该键。
func MatchersContained(matchers, target map[string]string) bool {
	return CompileMatchers(matchers)(target)
}

// MatchersJSONContained 解析两个 JSON 对象后做包含判断。
// 空匹配器表示不限制，即使目标标签无法解析也返回 true。
func MatchersJSONContained(matchersJSON, targetJSON string) (bool, error) {
	matchers, err := ParseObject(matchersJSON)
	if err != nil {
		return false, err
	}
	if len(matchers) == 0 {
		return true, nil
	}
	target, err := ParseObject(targetJSON)
	if err != nil {
		return false, err
	}
	return MatchersContained(matchers, target), nil
}
