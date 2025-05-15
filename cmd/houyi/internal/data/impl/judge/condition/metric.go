package condition

import (
	"github.com/moon-monitor/moon/pkg/api/houyi/common"
)

var metricConditions = make(map[common.MetricStrategyItem_Condition]MetricCondition, 8)

func NewMetricCondition(conditionType common.MetricStrategyItem_Condition) MetricCondition {
	condition, ok := metricConditions[conditionType]
	if ok {
		return condition
	}
	return &metricConditionEQ{}
}

func init() {
	conditions := []MetricCondition{
		&metricConditionEQ{},
		&metricConditionNE{},
		&metricConditionGT{},
		&metricConditionGTE{},
		&metricConditionLT{},
		&metricConditionLTE{},
		&metricConditionInRange{},
		&metricConditionNotInRange{},
	}
	RegisterMetricCondition(conditions...)
}

func RegisterMetricCondition(conditions ...MetricCondition) {
	for _, condition := range conditions {
		metricConditions[condition.Type()] = condition
	}
}

type MetricCondition interface {
	Comparable(conditionValues []float64, originValue float64) bool
	Type() common.MetricStrategyItem_Condition
}

type metricConditionEQ struct{}

func (m *metricConditionEQ) Type() common.MetricStrategyItem_Condition {
	return common.MetricStrategyItem_EQ
}

func (m *metricConditionEQ) Comparable(conditionValues []float64, originValue float64) bool {
	if len(conditionValues) != 1 {
		return false
	}
	return conditionValues[0] == originValue
}

type metricConditionNE struct{}

func (m *metricConditionNE) Type() common.MetricStrategyItem_Condition {
	return common.MetricStrategyItem_NE
}

func (m *metricConditionNE) Comparable(conditionValues []float64, originValue float64) bool {
	if len(conditionValues) != 1 {
		return false
	}
	return conditionValues[0] != originValue
}

type metricConditionGT struct{}

func (m *metricConditionGT) Type() common.MetricStrategyItem_Condition {
	return common.MetricStrategyItem_GT
}

func (m *metricConditionGT) Comparable(conditionValues []float64, originValue float64) bool {
	if len(conditionValues) != 1 {
		return false
	}
	return originValue > conditionValues[0]
}

type metricConditionGTE struct{}

func (m *metricConditionGTE) Type() common.MetricStrategyItem_Condition {
	return common.MetricStrategyItem_GTE
}

func (m *metricConditionGTE) Comparable(conditionValues []float64, originValue float64) bool {
	if len(conditionValues) != 1 {
		return false
	}
	return originValue >= conditionValues[0]
}

type metricConditionLT struct{}

func (m *metricConditionLT) Type() common.MetricStrategyItem_Condition {
	return common.MetricStrategyItem_LT
}

func (m *metricConditionLT) Comparable(conditionValues []float64, originValue float64) bool {
	if len(conditionValues) != 1 {
		return false
	}
	return originValue < conditionValues[0]
}

type metricConditionLTE struct{}

func (m *metricConditionLTE) Type() common.MetricStrategyItem_Condition {
	return common.MetricStrategyItem_LTE
}

func (m *metricConditionLTE) Comparable(conditionValues []float64, originValue float64) bool {
	if len(conditionValues) != 1 {
		return false
	}
	return originValue <= conditionValues[0]
}

type metricConditionInRange struct{}

func (m *metricConditionInRange) Type() common.MetricStrategyItem_Condition {
	return common.MetricStrategyItem_IN
}

func (m *metricConditionInRange) Comparable(conditionValues []float64, originValue float64) bool {
	if len(conditionValues) != 2 {
		return false
	}
	return originValue >= conditionValues[0] && originValue <= conditionValues[1]
}

type metricConditionNotInRange struct{}

func (m *metricConditionNotInRange) Type() common.MetricStrategyItem_Condition {
	return common.MetricStrategyItem_NOT_IN
}

func (m *metricConditionNotInRange) Comparable(conditionValues []float64, originValue float64) bool {
	if len(conditionValues) != 2 {
		return false
	}
	return originValue < conditionValues[0] || originValue > conditionValues[1]
}
