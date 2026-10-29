package probe

import (
	"strconv"
	"strings"

	"github.com/aide-family/moon/pkg/merr"
)

// StatusRange 是一段闭区间状态码。
type StatusRange struct {
	Min int
	Max int
}

// ParseExpectStatus 解析期望状态码表达式：单个值、区间、逗号分隔的列表。
// 结果用于判断一次响应是否算成功。
func ParseExpectStatus(expr string) ([]StatusRange, error) {
	text := strings.TrimSpace(expr)
	if text == "" {
		return []StatusRange{{Min: 200, Max: 299}}, nil
	}
	parts := strings.Split(text, ",")
	out := make([]StatusRange, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if idx := strings.Index(part, "-"); idx > 0 {
			min, errMin := strconv.Atoi(strings.TrimSpace(part[:idx]))
			max, errMax := strconv.Atoi(strings.TrimSpace(part[idx+1:]))
			if errMin != nil || errMax != nil || !validStatus(min) || !validStatus(max) || min > max {
				return nil, merr.Invalid("invalid expected status range")
			}
			out = append(out, StatusRange{Min: min, Max: max})
			continue
		}
		value, err := strconv.Atoi(part)
		if err != nil || !validStatus(value) {
			return nil, merr.Invalid("invalid expected status code")
		}
		out = append(out, StatusRange{Min: value, Max: value})
	}
	if len(out) == 0 {
		return nil, merr.Invalid("invalid expected status code")
	}
	return out, nil
}

// MatchStatus 判断状态码是否落在任一区间内。
func MatchStatus(ranges []StatusRange, code int) bool {
	for _, item := range ranges {
		if code >= item.Min && code <= item.Max {
			return true
		}
	}
	return false
}

// validStatus 只接受 100..599 的 HTTP 状态码。
func validStatus(code int) bool {
	return code >= 100 && code <= 599
}
