package vobj

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
)

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

	// MQConditionGT 大于
	MQConditionGT // 大于

	// MQConditionGTE 大于等于
	MQConditionGTE // 大于

	// MQConditionLT 小于
	MQConditionLT // 小于

	// MQConditionLTE 小于等于
	MQConditionLTE // 小于等于

	// MQConditionContain 包含
	MQConditionContain // 包含

	// MQConditionPrefix 前缀
	MQConditionPrefix // 前缀

	// MQConditionSuffix 后缀
	MQConditionSuffix // 后缀

	// MQConditionRegular 正则
	MQConditionRegular // 正则
)

// Judge 判断是否符合条件
func (c MQCondition) Judge(data []byte, dataType MQDataType, threshold string) bool {
	switch dataType {
	case MQDataTypeNumber:
		return c.numberJudge(data, threshold)
	default:
		return c.stringJudge(string(data), threshold)
	}
}

// stringJudge 字符串判断
func (c MQCondition) stringJudge(data string, threshold string) bool {
	switch c {
	case MQConditionEQ:
		return data == threshold
	case MQConditionNE:
		return data != threshold
	case MQConditionGT:
		return data > threshold
	case MQConditionGTE:
		return data >= threshold
	case MQConditionLT:
		return data < threshold
	case MQConditionLTE:
		return data <= threshold
	case MQConditionContain:
		return strings.Contains(data, threshold)
	case MQConditionPrefix:
		return strings.HasPrefix(data, threshold)
	case MQConditionSuffix:
		return strings.HasSuffix(data, threshold)
	case MQConditionRegular:
		matchReg := threshold
		compile, err := regexp.Compile(matchReg)
		if err != nil {
			log.Warnw("method", "stringJudge", "acton", "regexp compile error", "err", err)
			return false
		}
		return compile.MatchString(data)
	default:
		return false
	}
}

// numberJudge 数字判断
func (c MQCondition) numberJudge(data []byte, threshold string) bool {
	num, err := strconv.ParseFloat(string(data), 64)
	if err != nil {
		log.Warnw("method", "numberJudge", "acton", "parse value float error", "err", err)
		return false
	}
	thresholdVal, err := strconv.ParseFloat(threshold, 64)
	if err != nil {
		log.Warnw("method", "numberJudge", "acton", "parse threshold float error", "err", err)
		return false
	}
	switch c {
	case MQConditionEQ:
		return num == thresholdVal
	case MQConditionNE:
		return num != thresholdVal
	case MQConditionGT:
		return num > thresholdVal
	case MQConditionGTE:
		return num >= thresholdVal
	case MQConditionLT:
		return num < thresholdVal
	case MQConditionLTE:
		return num <= thresholdVal
	default:
		return false
	}
}

// objectJudge 对象数据判断
func (c MQCondition) objectJudge(data []byte, threshold string) bool {
	// TODO 待实现
	return false
}
