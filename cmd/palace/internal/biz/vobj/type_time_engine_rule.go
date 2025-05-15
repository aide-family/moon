package vobj

// TimeEngineRuleType time engine rule type
//
//go:generate stringer -type=TimeEngineRuleType -linecomment -output=type_time_engine_rule.string.go
type TimeEngineRuleType int8

const (
	TimeEngineRuleTypeUnknown         TimeEngineRuleType = iota // unknown
	TimeEngineRuleTypeHourRange                                 // hour-range
	TimeEngineRuleTypeHour                                      // hour
	TimeEngineRuleTypeHourMinuteRange                           // hour-minute-range
	TimeEngineRuleTypeDaysOfWeek                                // days-of-week
	TimeEngineRuleTypeDayOfMonth                                // day-of-month
	TimeEngineRuleTypeMonth                                     // month
)
