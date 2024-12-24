package vobj

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/tidwall/gjson"
)

// EventCondition Event条件判断
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=EventCondition -linecomment
type EventCondition int

const (
	// EventConditionUnknown 未知
	EventConditionUnknown EventCondition = iota // 未知

	// EventConditionEQ 等于
	EventConditionEQ // 等于

	// EventConditionNE 不等于
	EventConditionNE // 不等于

	// EventConditionGT 大于
	EventConditionGT // 大于

	// EventConditionGTE 大于等于
	EventConditionGTE // 大于

	// EventConditionLT 小于
	EventConditionLT // 小于

	// EventConditionLTE 小于等于
	EventConditionLTE // 小于等于

	// EventConditionContain 包含
	EventConditionContain // 包含

	// EventConditionPrefix 前缀
	EventConditionPrefix // 前缀

	// EventConditionSuffix 后缀
	EventConditionSuffix // 后缀

	// EventConditionRegular 正则
	EventConditionRegular // 正则
)

// Judge 判断是否符合条件
func (c EventCondition) Judge(data []byte, dataType EventDataType, key, threshold string) bool {
	switch dataType {
	case EventDataTypeNumber:
		return c.numberJudge(data, threshold)
	case EventDataTypeObject:
		return c.objectJudge(data, key, threshold)
	default:
		return c.stringJudge(string(data), threshold)
	}
}

// stringJudge 字符串判断
func (c EventCondition) stringJudge(data string, threshold string) bool {
	switch c {
	case EventConditionEQ:
		return data == threshold
	case EventConditionNE:
		return data != threshold
	case EventConditionGT:
		return data > threshold
	case EventConditionGTE:
		return data >= threshold
	case EventConditionLT:
		return data < threshold
	case EventConditionLTE:
		return data <= threshold
	case EventConditionContain:
		return strings.Contains(data, threshold)
	case EventConditionPrefix:
		return strings.HasPrefix(data, threshold)
	case EventConditionSuffix:
		return strings.HasSuffix(data, threshold)
	case EventConditionRegular:
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
func (c EventCondition) numberJudge(data []byte, threshold string) bool {
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
	case EventConditionEQ:
		return num == thresholdVal
	case EventConditionNE:
		return num != thresholdVal
	case EventConditionGT:
		return num > thresholdVal
	case EventConditionGTE:
		return num >= thresholdVal
	case EventConditionLT:
		return num < thresholdVal
	case EventConditionLTE:
		return num <= thresholdVal
	default:
		return false
	}
}

// objectJudge 对象数据判断
func (c EventCondition) objectJudge(data []byte, key, threshold string) bool {
	result := gjson.GetBytes(data, key)
	if !result.Exists() {
		log.Warnw("method", "objectJudge", "action", "key not found in json data", "key", key)
		return false
	}
	return c.stringJudge(result.String(), threshold)
}
