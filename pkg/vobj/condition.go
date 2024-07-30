package vobj

// Condition 条件判断
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=Condition -linecomment
type Condition int

const (
	// ConditionUnknown 未知
	ConditionUnknown Condition = iota // 未知

	// ConditionEQ 等于
	ConditionEQ // 等于

	// ConditionNE 不等于
	ConditionNE // 不等于

	// ConditionGT 大于
	ConditionGT // 大于

	// ConditionGTE 大于等于
	ConditionGTE // 大于等于

	// ConditionLT 小于
	ConditionLT // 小于

	// ConditionLTE 小于等于
	ConditionLTE // 小于等于
)
