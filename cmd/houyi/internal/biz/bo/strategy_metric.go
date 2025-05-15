package bo

import (
	"time"

	"github.com/aide-family/moon/pkg/api/houyi/common"
	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/aide-family/moon/pkg/util/kv"
	"github.com/aide-family/moon/pkg/util/kv/label"
)

type MetricJudgeRule interface {
	GetLabels() *label.Label
	GetAnnotations() *label.Annotation
	GetDuration() time.Duration
	GetCount() int64
	GetValues() []float64
	GetSampleMode() common.SampleMode
	GetCondition() common.MetricStrategyItem_Condition
	GetExt() kv.Map[string, any]
}

type MetricJudgeDataValue interface {
	GetValue() float64
	GetTimestamp() int64
}

type MetricJudgeData interface {
	GetLabels() map[string]string
	GetValues() []MetricJudgeDataValue
}

type MetricRule interface {
	cache.Object
	GetTeamId() uint32
	GetDatasource() string
	GetStrategyId() uint32
	GetLevelId() uint32
	GetReceiverRoutes() []string
	GetLabelReceiverRoutes() []LabelNotices
	GetExpr() string
	GetEnable() bool
	Renovate()

	MetricJudgeRule
}

type MetricJudgeRequest struct {
	JudgeData []MetricJudgeData
	Strategy  MetricRule
	Step      time.Duration
}
