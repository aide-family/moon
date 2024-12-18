package types

import "time"

var _ Matcher = (*HourRange)(nil)
var _ Matcher = (*DaysOfWeek)(nil)
var _ Matcher = (*DaysOfMonth)(nil)
var _ Matcher = (*Months)(nil)

type (
	// Matcher 匹配器接口
	Matcher interface {
		// Match 匹配时间
		Match(time.Time) bool
	}

	// HourRange 小时范围 24小时制
	HourRange struct {
		Start int `json:"start"`
		End   int `json:"end"`
	}

	// DaysOfWeek 星期 1-7
	DaysOfWeek []int

	// DaysOfMonth 日期范围 1-31
	DaysOfMonth struct {
		Start int `json:"start"`
		End   int `json:"end"`
	}

	// Months 月份范围 1-12
	Months struct {
		Start int `json:"start"`
		End   int `json:"end"`
	}

	// TimeEngineer 时间引擎接口
	TimeEngineer interface {
		IsAllowed(time.Time) bool
	}

	// TimeEngine 时间引擎结构体
	TimeEngine struct {
		Configurations []Matcher
	}

	// Option 时间引擎选项
	Option func(*TimeEngine)
)

// NewTimeEngine 创建时间引擎
func NewTimeEngine(opts ...Option) TimeEngineer {
	engine := &TimeEngine{}
	for _, opt := range opts {
		opt(engine)
	}
	return engine
}

// WithConfigurations 设置配置/匹配器
func WithConfigurations(configurations []Matcher) Option {
	return func(engine *TimeEngine) {
		engine.Configurations = append(engine.Configurations, configurations...)
	}
}

// WithConfiguration 设置配置/匹配器
func WithConfiguration(configuration Matcher) Option {
	return func(engine *TimeEngine) {
		engine.Configurations = append(engine.Configurations, configuration)
	}
}

// Match 匹配月份
func (m *Months) Match(t time.Time) bool {
	if m == nil {
		return true
	}
	month := t.Month()
	return month >= time.Month(m.Start) && month <= time.Month(m.End)
}

// Match 匹配每月的日期
func (d *DaysOfMonth) Match(t time.Time) bool {
	if d == nil {
		return true
	}
	day := t.Day()
	return day >= d.Start && day <= d.End
}

// Match 匹配星期
func (d *DaysOfWeek) Match(t time.Time) bool {
	if d == nil {
		return true
	}
	if len(*d) > 7 {
		*d = (*d)[:7]
	}
	weekDay := t.Weekday()
	for _, item := range *d {
		if weekDay == time.Weekday(item) {
			return true
		}
	}
	return false
}

// Match 匹配小时
func (h *HourRange) Match(t time.Time) bool {
	if h == nil {
		return true
	}
	hour := t.Hour()
	return hour >= h.Start && hour <= h.End
}

// IsAllowed 判断当前时间是否符合时间范围
func (tr *TimeEngine) IsAllowed(t time.Time) bool {
	return tr.matches(t)
}

// Matches 检查时间是否符合当前配置
func (tr *TimeEngine) matches(t time.Time) bool {
	for _, config := range tr.Configurations {
		if config == nil {
			continue
		}
		if !config.Match(t) {
			return false
		}
	}
	return true
}
