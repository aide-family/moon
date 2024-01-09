package bo

import (
	"strconv"

	"prometheus-manager/api"
)

// BuildApiDuration 字符串转为api时间
func BuildApiDuration(duration string) *api.Duration {
	durationLen := len(duration)
	if duration == "" || durationLen < 2 {
		return nil
	}
	value, _ := strconv.Atoi(duration[:durationLen-1])
	// 获取字符串最后一个字符
	unit := string(duration[durationLen-1])
	return &api.Duration{
		Value: int64(value),
		Unit:  unit,
	}
}

// BuildApiDurationString 时间转换为字符串
func BuildApiDurationString(duration *api.Duration) string {
	if duration == nil {
		return ""
	}
	return strconv.FormatInt(duration.Value, 10) + duration.Unit
}
