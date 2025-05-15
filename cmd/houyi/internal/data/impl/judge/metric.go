package judge

import (
	"time"

	"github.com/aide-family/moon/cmd/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/houyi/internal/data/impl/judge/condition"
	"github.com/aide-family/moon/pkg/api/houyi/common"
)

func NewMetricJudge(sampleModeType common.SampleMode, opts ...MetricJudgeOption) MetricJudge {
	var config MetricJudgeConfig
	for _, opt := range opts {
		opt(&config)
	}
	switch sampleModeType {
	case common.SampleMode_FOR:
		return &metricForJudge{config}
	case common.SampleMode_MAX:
		return &metricMaxJudge{config}
	case common.SampleMode_MIN:
		return &metricMinJudge{config}
	default:
		return &metricForJudge{config}
	}
}

type MetricJudgeConfig struct {
	condition       condition.MetricCondition
	conditionValues []float64
	conditionCount  int64
	step            time.Duration
}

type MetricJudgeOption func(*MetricJudgeConfig)

func WithMetricJudgeCondition(condition condition.MetricCondition) MetricJudgeOption {
	return func(c *MetricJudgeConfig) {
		c.condition = condition
	}
}

func WithMetricJudgeConditionValues(values []float64) MetricJudgeOption {
	return func(c *MetricJudgeConfig) {
		c.conditionValues = values
	}
}

func WithMetricJudgeConditionCount(count int64) MetricJudgeOption {
	return func(c *MetricJudgeConfig) {
		c.conditionCount = count
	}
}

func WithMetricJudgeStep(step time.Duration) MetricJudgeOption {
	return func(c *MetricJudgeConfig) {
		c.step = step
	}
}

type MetricJudge interface {
	Judge(originValues []bo.MetricJudgeDataValue) (bo.MetricJudgeDataValue, bool)
	Type() common.SampleMode
}

type metricForJudge struct {
	MetricJudgeConfig
}

func (m *metricForJudge) Type() common.SampleMode {
	return common.SampleMode_FOR
}

func (m *metricForJudge) Judge(originValues []bo.MetricJudgeDataValue) (bo.MetricJudgeDataValue, bool) {
	total := int64(0)
	var currentValue bo.MetricJudgeDataValue
	step := m.step
	for _, value := range originValues {
		if time.Duration(value.GetTimestamp())-step > 0 {
			total = 0
		}
		if !m.condition.Comparable(m.conditionValues, value.GetValue()) {
			total = 0
			currentValue = nil
			continue
		}
		total += 1
		currentValue = value
		if total >= m.conditionCount {
			return currentValue, true
		}
	}
	return currentValue, total >= m.conditionCount && currentValue != nil
}

type metricMaxJudge struct {
	MetricJudgeConfig
}

func (m *metricMaxJudge) Type() common.SampleMode {
	return common.SampleMode_MAX
}

func (m *metricMaxJudge) Judge(originValues []bo.MetricJudgeDataValue) (bo.MetricJudgeDataValue, bool) {
	total := int64(0)
	var currentValue bo.MetricJudgeDataValue
	for _, value := range originValues {
		if !m.condition.Comparable(m.conditionValues, value.GetValue()) {
			currentValue = value
			continue
		}

		total += 1
		if total > m.conditionCount {
			return nil, false
		}
	}
	return currentValue, currentValue != nil
}

type metricMinJudge struct {
	MetricJudgeConfig
}

func (m *metricMinJudge) Type() common.SampleMode {
	return common.SampleMode_MIN
}

func (m *metricMinJudge) Judge(originValues []bo.MetricJudgeDataValue) (bo.MetricJudgeDataValue, bool) {
	total := int64(0)
	var currentValue bo.MetricJudgeDataValue
	for _, value := range originValues {
		if m.condition.Comparable(m.conditionValues, value.GetValue()) {
			total += 1
			currentValue = value
			if total >= m.conditionCount {
				return currentValue, true
			}
		}
	}
	return nil, false
}
