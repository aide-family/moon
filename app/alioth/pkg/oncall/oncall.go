package oncall

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aide-family/moon/pkg/merr"
)

const (
	// MaxWindowsPerGroup 是单个通知组允许配置的窗口数量，避免一次保存塞入无界数据。
	MaxWindowsPerGroup = 20
	// DefaultTimezone 是未指定时区时使用的 IANA 时区。
	DefaultTimezone = "Asia/Shanghai"
)

// Spec 是时间窗口的接口和 JSON 形态。Weekdays 取 MONDAY 到 SUNDAY，
// Start/End 取 HH:mm；Start 晚于 End 表示跨午夜。
type Spec struct {
	Weekdays []string `json:"weekdays"`
	Start    string   `json:"start"`
	End      string   `json:"end"`
}

// Window 是编译后的时间窗口。weekdays 用位图保存，周一到周日依次占一位。
type Window struct {
	weekdays uint8
	start    int
	end      int
}

// Normalize 校验窗口并把星期、时间写成稳定形态：星期按周一到周日排序去重，
// 时间补成两位。重复窗口会被合并，避免同一时刻被重复判断。
func Normalize(specs []Spec) ([]Spec, error) {
	if len(specs) > MaxWindowsPerGroup {
		return nil, merr.Invalid("too many on-call time windows")
	}
	out := make([]Spec, 0, len(specs))
	seen := make(map[string]struct{}, len(specs))
	for index, spec := range specs {
		weekdays, err := normalizeWeekdays(spec.Weekdays)
		if err != nil {
			return nil, err
		}
		start, err := parseClock(spec.Start)
		if err != nil {
			return nil, merr.Invalid(fmt.Sprintf("window %d: %v", index+1, err))
		}
		end, err := parseClock(spec.End)
		if err != nil {
			return nil, merr.Invalid(fmt.Sprintf("window %d: %v", index+1, err))
		}
		if start == end {
			return nil, merr.Invalid("on-call window start and end must differ")
		}
		normalized := Spec{
			Weekdays: weekdays,
			Start:    formatClock(start),
			End:      formatClock(end),
		}
		key, err := windowKey(normalized)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, normalized)
	}
	return out, nil
}

// Compile 把窗口编译成可直接参与判断的形态。空窗口表示全天候，编译结果为空切片。
func Compile(specs []Spec) ([]Window, error) {
	normalized, err := Normalize(specs)
	if err != nil {
		return nil, err
	}
	windows := make([]Window, 0, len(normalized))
	for _, spec := range normalized {
		weekdays, err := weekdayMask(spec.Weekdays)
		if err != nil {
			return nil, err
		}
		start, err := parseClock(spec.Start)
		if err != nil {
			return nil, err
		}
		end, err := parseClock(spec.End)
		if err != nil {
			return nil, err
		}
		windows = append(windows, Window{weekdays: weekdays, start: start, end: end})
	}
	return windows, nil
}

// Active 判断 at 是否落在任一窗口内。没有窗口时始终接收，这是“全天候值班”的表达方式。
func Active(windows []Window, at time.Time) bool {
	if len(windows) == 0 {
		return true
	}
	for _, window := range windows {
		if window.match(at) {
			return true
		}
	}
	return false
}

// LoadLocation 读取 IANA 时区。空名回落到 DefaultTimezone。
func LoadLocation(name string) (*time.Location, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = DefaultTimezone
	}
	// Local 依赖进程所在主机的时区，多副本会各算各的；它不是 IANA 时区名，必须显式拒绝。
	if strings.EqualFold(name, "Local") || strings.EqualFold(name, "localtime") {
		return nil, merr.Invalid("invalid on-call timezone")
	}
	location, err := time.LoadLocation(name)
	if err != nil {
		return nil, merr.Invalid("invalid on-call timezone")
	}
	return location, nil
}

// NormalizeTimezone 校验时区并返回规范名称。空名回落到 DefaultTimezone。
func NormalizeTimezone(name string) (string, error) {
	location, err := LoadLocation(name)
	if err != nil {
		return "", err
	}
	return location.String(), nil
}

// Marshal 把窗口编码成紧凑 JSON。nil 和空切片都编码成空数组。
func Marshal(specs []Spec) (string, error) {
	normalized, err := Normalize(specs)
	if err != nil {
		return "", err
	}
	out, err := json.Marshal(normalized)
	if err != nil {
		return "", merr.Server("failed to encode on-call windows")
	}
	return string(out), nil
}

// Unmarshal 解析窗口 JSON。空串和 null 视为没有窗口。
func Unmarshal(raw string) ([]Spec, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "null" {
		return []Spec{}, nil
	}
	var specs []Spec
	if err := json.Unmarshal([]byte(raw), &specs); err != nil {
		return nil, merr.Invalid("invalid on-call windows")
	}
	if specs == nil {
		return []Spec{}, nil
	}
	return specs, nil
}

func (window Window) match(at time.Time) bool {
	minute := at.Hour()*60 + at.Minute()
	day := weekdayBit(at.Weekday())
	if window.start < window.end {
		return window.weekdays&day != 0 && minute >= window.start && minute < window.end
	}
	// 跨午夜：开始那天的星期的晚间时段，或前一天星期的凌晨时段。
	if window.weekdays&day != 0 && minute >= window.start {
		return true
	}
	previous := weekdayBit(at.AddDate(0, 0, -1).Weekday())
	return window.weekdays&previous != 0 && minute < window.end
}

func windowKey(spec Spec) (string, error) {
	start, err := parseClock(spec.Start)
	if err != nil {
		return "", err
	}
	end, err := parseClock(spec.End)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s|%d|%d", strings.Join(spec.Weekdays, ","), start, end), nil
}

func normalizeWeekdays(values []string) ([]string, error) {
	if len(values) == 0 {
		return nil, merr.Invalid("on-call window requires at least one weekday")
	}
	seen := make(map[int]struct{}, len(values))
	for _, value := range values {
		index, ok := weekdayIndex(value)
		if !ok {
			return nil, merr.Invalid("invalid on-call weekday")
		}
		seen[index] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for index := 0; index < 7; index++ {
		if _, ok := seen[index]; ok {
			out = append(out, weekdayNames[index])
		}
	}
	return out, nil
}

func weekdayMask(values []string) (uint8, error) {
	normalized, err := normalizeWeekdays(values)
	if err != nil {
		return 0, err
	}
	var mask uint8
	for _, value := range normalized {
		index, _ := weekdayIndex(value)
		mask |= 1 << index
	}
	return mask, nil
}

func weekdayIndex(value string) (int, bool) {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "MONDAY":
		return 0, true
	case "TUESDAY":
		return 1, true
	case "WEDNESDAY":
		return 2, true
	case "THURSDAY":
		return 3, true
	case "FRIDAY":
		return 4, true
	case "SATURDAY":
		return 5, true
	case "SUNDAY":
		return 6, true
	default:
		return 0, false
	}
}

// weekdayBit 把 Go 的 Sunday=0 映射为周一位图的第 0 位到第 6 位。
func weekdayBit(day time.Weekday) uint8 {
	if day == time.Sunday {
		return 1 << 6
	}
	return 1 << (int(day) - 1)
}

func parseClock(value string) (int, error) {
	value = strings.TrimSpace(value)
	// 允许 9:00 这种省略前导零的写法，保存时统一补成 HH:mm。
	colon := strings.IndexByte(value, ':')
	if colon < 1 || colon > 2 || len(value)-colon-1 != 2 {
		return 0, merr.Invalid("on-call window time must be HH:mm")
	}
	hour, err := strconv.Atoi(value[:colon])
	if err != nil || hour < 0 || hour > 23 {
		return 0, merr.Invalid("on-call window time must be HH:mm")
	}
	minute, err := strconv.Atoi(value[colon+1:])
	if err != nil || minute < 0 || minute > 59 {
		return 0, merr.Invalid("on-call window time must be HH:mm")
	}
	return hour*60 + minute, nil
}

func formatClock(minute int) string {
	return fmt.Sprintf("%02d:%02d", minute/60, minute%60)
}

var weekdayNames = [7]string{
	"MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY", "SUNDAY",
}
