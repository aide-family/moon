package do

import (
	"encoding/json"

	"github.com/moon-monitor/moon/cmd/houyi/internal/biz/bo"
	"github.com/moon-monitor/moon/pkg/util/slices"
)

type MetricQueryValue struct {
	Value     float64
	Timestamp int64
}

func (m *MetricQueryValue) GetValue() float64 {
	if m == nil {
		return 0
	}
	return m.Value
}

func (m *MetricQueryValue) GetTimestamp() int64 {
	if m == nil {
		return 0
	}
	return m.Timestamp
}

type MetricQueryRangeReply struct {
	Labels     map[string]string
	Values     []*MetricQueryValue
	ResultType string `json:"resultType"`
}

func (m *MetricQueryRangeReply) GetLabels() map[string]string {
	if m.Labels == nil {
		return nil
	}
	return m.Labels
}

func (m *MetricQueryRangeReply) GetValues() []bo.MetricJudgeDataValue {
	if m.Values == nil {
		return nil
	}
	return slices.Map(m.Values, func(v *MetricQueryValue) bo.MetricJudgeDataValue {
		return v
	})
}

type MetricQueryReply struct {
	Labels     map[string]string
	Value      *MetricQueryValue
	ResultType string `json:"resultType"`
}

type MetricItem struct {
	// Name metric name
	Name string `json:"name"`
	// Help metric help
	Help string `json:"help"`
	// Type metric type
	Type string `json:"type"`
	// Labels metric labels
	Labels map[string][]string `json:"labels"`
	// Unit metric unit
	Unit string `json:"unit"`
}

func (m *MetricItem) String() string {
	bs, _ := json.Marshal(m)
	return string(bs)
}
