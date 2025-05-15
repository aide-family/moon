package do

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/moon-monitor/moon/cmd/houyi/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/houyi/internal/biz/vobj"
	"github.com/moon-monitor/moon/pkg/api/houyi/common"
	"github.com/moon-monitor/moon/pkg/util/cnst"
	"github.com/moon-monitor/moon/pkg/util/kv"
	"github.com/moon-monitor/moon/pkg/util/kv/label"
	"github.com/moon-monitor/moon/pkg/util/slices"
)

type MetricRule struct {
	TeamId        uint32                              `json:"teamId,omitempty"`
	DatasourceId  uint32                              `json:"datasourceId,omitempty"`
	Datasource    string                              `json:"datasource,omitempty"`
	StrategyId    uint32                              `json:"strategyId,omitempty"`
	LevelId       uint32                              `json:"levelId,omitempty"`
	Receiver      []string                            `json:"receiver,omitempty"`
	LabelReceiver []*LabelNotices                     `json:"labelReceiver,omitempty"`
	Expr          string                              `json:"expr,omitempty"`
	Labels        *label.Label                        `json:"labels,omitempty"`
	Annotations   *label.Annotation                   `json:"annotations,omitempty"`
	Duration      time.Duration                       `json:"duration,omitempty"`
	Count         int64                               `json:"count,omitempty"`
	Values        []float64                           `json:"values,omitempty"`
	SampleMode    common.SampleMode                   `json:"sampleMode,omitempty"`
	Condition     common.MetricStrategyItem_Condition `json:"condition,omitempty"`
	Enable        bool                                `json:"enable,omitempty"`
}

func (m *MetricRule) Renovate() {
	m.Labels.Appends(map[string]string{
		cnst.LabelKeyTeamID:       strconv.FormatUint(uint64(m.TeamId), 10),
		cnst.LabelKeyStrategyID:   strconv.FormatUint(uint64(m.StrategyId), 10),
		cnst.LabelKeyLevelID:      strconv.FormatUint(uint64(m.LevelId), 10),
		cnst.LabelKeyDatasourceID: strconv.FormatUint(uint64(m.DatasourceId), 10),
	})
}

func (m *MetricRule) GetEnable() bool {
	if m == nil {
		return false
	}
	return m.Enable
}

func (m *MetricRule) MarshalBinary() (data []byte, err error) {
	return json.Marshal(m)
}

func (m *MetricRule) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, m)
}

func (m *MetricRule) UniqueKey() string {
	if m == nil {
		return ""
	}
	return vobj.MetricRuleUniqueKey(m.TeamId, m.StrategyId, m.LevelId, m.Datasource)
}

func (m *MetricRule) GetTeamId() uint32 {
	if m == nil {
		return 0
	}
	return m.TeamId
}

func (m *MetricRule) GetDatasource() string {
	if m == nil {
		return ""
	}
	return m.Datasource
}

func (m *MetricRule) GetStrategyId() uint32 {
	if m == nil {
		return 0
	}
	return m.StrategyId
}

func (m *MetricRule) GetLevelId() uint32 {
	if m == nil {
		return 0
	}
	return m.LevelId
}

func (m *MetricRule) GetReceiverRoutes() []string {
	if m == nil {
		return nil
	}
	return m.Receiver
}

func (m *MetricRule) GetLabelReceiverRoutes() []bo.LabelNotices {
	if m == nil {
		return nil
	}
	return slices.Map(m.LabelReceiver, func(v *LabelNotices) bo.LabelNotices {
		return v
	})
}

func (m *MetricRule) GetExpr() string {
	if m == nil {
		return ""
	}
	return m.Expr
}

func (m *MetricRule) GetLabels() *label.Label {
	if m == nil {
		return nil
	}
	return m.Labels
}

func (m *MetricRule) GetAnnotations() *label.Annotation {
	if m == nil {
		return nil
	}
	return m.Annotations
}

func (m *MetricRule) GetDuration() time.Duration {
	if m == nil {
		return 0
	}
	return m.Duration
}

func (m *MetricRule) GetCount() int64 {
	if m == nil {
		return 0
	}
	return m.Count
}

func (m *MetricRule) GetValues() []float64 {
	if m == nil {
		return nil
	}
	return m.Values
}

func (m *MetricRule) GetSampleMode() common.SampleMode {
	if m == nil {
		return common.SampleMode_SAMPLE_MODE_UNKNOWN
	}
	return m.SampleMode
}

func (m *MetricRule) GetCondition() common.MetricStrategyItem_Condition {
	if m == nil {
		return common.MetricStrategyItem_EQ
	}
	return m.Condition
}

func (m *MetricRule) GetExt() kv.Map[string, any] {
	if m == nil {
		return nil
	}
	return map[string]any{}
}
