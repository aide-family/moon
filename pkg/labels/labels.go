// Package labels 解析标签 JSON，并判断匹配器是否被包含。
package labels

import (
	"encoding/json"
	"strings"

	"github.com/aide-family/Alioth/pkg/merr"
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

// MatchersContained 判断 matchers 的每个键值是否都出现在 target 中。
func MatchersContained(matchers, target map[string]string) bool {
	for key, value := range matchers {
		if target[key] != value {
			return false
		}
	}
	return true
}

// MatchersJSONContained 解析两个 JSON 对象后做包含判断。
func MatchersJSONContained(matchersJSON, targetJSON string) (bool, error) {
	matchers, err := ParseObject(matchersJSON)
	if err != nil {
		return false, err
	}
	target, err := ParseObject(targetJSON)
	if err != nil {
		return false, err
	}
	return MatchersContained(matchers, target), nil
}
