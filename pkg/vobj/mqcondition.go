package vobj

// MQCondition MQ条件判断
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=MQCondition -linecomment
type MQCondition int

const (
	// MQConditionUnknown 未知
	MQConditionUnknown MQCondition = iota // 未知

	// MQConditionEQ 等于
	MQConditionEQ // 等于

	// MQConditionNE 不等于
	MQConditionNE // 不等于

	// MQConditionGTE 大于等于
	MQConditionGTE // 大于

	// MQConditionLT 小于
	MQConditionLT // 小于

	// MQConditionContain 包含
	MQConditionContain // 包含

	// MQConditionPrefix 前缀
	MQConditionPrefix // 前缀

	// MQConditionSuffix 后缀
	MQConditionSuffix // 后缀

	// MQConditionRegular 正则
	MQConditionRegular // 正则
)
