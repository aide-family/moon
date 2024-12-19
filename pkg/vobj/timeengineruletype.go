package vobj

// TimeEngineRuleType 时间引擎规则类型
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=TimeEngineRuleType -linecomment
type TimeEngineRuleType int

const (
	// TimeEngineRuleTypeUnknown 未知
	TimeEngineRuleTypeUnknown TimeEngineRuleType = iota // 未知

	// TimeEngineRuleTypeHourRange 小时范围 24小时制
	TimeEngineRuleTypeHourRange // 小时范围

	// TimeEngineRuleTypeDaysOfWeek 星期 1-7
	TimeEngineRuleTypeDaysOfWeek // 星期

	// TimeEngineRuleTypeDaysOfMonth 日期范围 1-31
	TimeEngineRuleTypeDaysOfMonth // 日期

	// TimeEngineRuleTypeMonths 月份范围 1-12
	TimeEngineRuleTypeMonths // 月份
)
