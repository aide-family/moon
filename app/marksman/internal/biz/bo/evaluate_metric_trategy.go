package bo

import (
	"encoding/json"
	"fmt"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
)

func NewEvaluateMetricStrategyBo(
	namespaceUID snowflake.ID,
	strategyGroup *StrategyGroupItemBo,
	strategy *StrategyItemBo,
	strategyMetric *StrategyMetricItemBo,
	strategyLevel *StrategyMetricLevelItemBo,
	datasource *DatasourceItemBo,
) (*EvaluateMetricStrategyBo, error) {
	if namespaceUID == 0 {
		return nil, merr.ErrorParams("namespace UID is required")
	}
	if strategyGroup == nil {
		return nil, merr.ErrorParams("strategy group is required")
	}
	if strategy == nil {
		return nil, merr.ErrorParams("strategy is required")
	}
	if strategyMetric == nil {
		return nil, merr.ErrorParams("strategy metric is required")
	}
	if strategyLevel == nil {
		return nil, merr.ErrorParams("strategy level is required")
	}
	if datasource == nil {
		return nil, merr.ErrorParams("datasource is required")
	}
	return &EvaluateMetricStrategyBo{
		strategyGroup: strategyGroup,
		strategy:      strategy,
		namespaceUID:  namespaceUID,
		datasource:    datasource,
		expr:          strategyMetric.Expr,
		labels:        strategyMetric.Labels,
		summary:       strategyMetric.Summary,
		description:   strategyMetric.Description,
		level:         strategyLevel.Level,
		mode:          strategyLevel.Mode,
		condition:     strategyLevel.Condition,
		values:        strategyLevel.Values,
		durationSec:   strategyLevel.DurationSec,
	}, nil
}

type EvaluateMetricStrategyBo struct {
	strategyGroup *StrategyGroupItemBo
	strategy      *StrategyItemBo
	namespaceUID  snowflake.ID
	datasource    *DatasourceItemBo
	expr          string
	labels        map[string]string
	summary       string
	description   string
	level         *LevelItemBo
	mode          enum.SampleMode
	condition     enum.ConditionMetric
	values        []float64
	durationSec   int64
}

func (e *EvaluateMetricStrategyBo) BuildMetricEvaluatorIndex() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf(
		"metric-%d-%d-%d-%d-%d",
		e.GetNamespaceUID().Int64(),
		e.GetDatasourceUID().Int64(),
		e.GetStrategyGroupUID().Int64(),
		e.GetStrategyUID().Int64(),
		e.GetLevelUID().Int64(),
	)
}

func (e *EvaluateMetricStrategyBo) GetStrategyGroupUID() snowflake.ID {
	if e.strategyGroup == nil {
		return 0
	}
	return e.strategyGroup.UID
}

func (e *EvaluateMetricStrategyBo) GetStrategyGroupName() string {
	if e.strategyGroup == nil {
		return ""
	}
	return e.strategyGroup.Name
}

func (e *EvaluateMetricStrategyBo) GetStrategyUID() snowflake.ID {
	if e.strategy == nil {
		return 0
	}
	return e.strategy.UID
}

func (e *EvaluateMetricStrategyBo) GetStrategyName() string {
	if e.strategy == nil {
		return ""
	}
	return e.strategy.Name
}

func (e *EvaluateMetricStrategyBo) GetNamespaceUID() snowflake.ID {
	return e.namespaceUID
}

func (e *EvaluateMetricStrategyBo) GetDatasource() *DatasourceItemBo {
	if e.datasource == nil {
		return &DatasourceItemBo{}
	}
	return e.datasource
}

func (e *EvaluateMetricStrategyBo) GetDatasourceUID() snowflake.ID {
	if e.datasource == nil {
		return 0
	}
	return e.datasource.UID
}

func (e *EvaluateMetricStrategyBo) GetDatasourceName() string {
	if e.datasource == nil {
		return ""
	}
	return e.datasource.Name
}

func (e *EvaluateMetricStrategyBo) GetDatasourceLevelName() string {
	if e.datasource == nil || e.datasource.Level == nil {
		return ""
	}
	return e.datasource.Level.Name
}

func (e *EvaluateMetricStrategyBo) GetExpr() string {
	return e.expr
}

func (e *EvaluateMetricStrategyBo) GetLabels() map[string]string {
	return e.labels
}

func (e *EvaluateMetricStrategyBo) GetSummary() string {
	return e.summary
}

func (e *EvaluateMetricStrategyBo) GetDescription() string {
	return e.description
}

func (e *EvaluateMetricStrategyBo) GetLevelUID() snowflake.ID {
	if e.level == nil {
		return 0
	}
	return e.level.UID
}

func (e *EvaluateMetricStrategyBo) GetLevelName() string {
	if e.level == nil {
		return ""
	}
	return e.level.Name
}

func (e *EvaluateMetricStrategyBo) GetLevelBgColor() string {
	if e.level == nil {
		return ""
	}
	return e.level.BgColor
}

func (e *EvaluateMetricStrategyBo) GetMode() enum.SampleMode {
	return e.mode
}

func (e *EvaluateMetricStrategyBo) GetCondition() enum.ConditionMetric {
	return e.condition
}

func (e *EvaluateMetricStrategyBo) GetValues() []float64 {
	return e.values
}

func (e *EvaluateMetricStrategyBo) GetDurationSec() int64 {
	return e.durationSec
}

// ToSnapshot builds a JSON-serializable snapshot from EvaluateMetricStrategyBo (metric evaluator payload).
func (e *EvaluateMetricStrategyBo) ToSnapshot() *MetricEvaluatorSnapshot {
	if e == nil {
		return nil
	}
	return &MetricEvaluatorSnapshot{
		StrategyUID:   e.strategy.UID.Int64(),
		NamespaceUID:  e.namespaceUID.Int64(),
		Expr:          e.expr,
		Labels:        e.labels,
		Summary:       e.summary,
		Description:   e.description,
		Mode:          int32(e.mode),
		Condition:     int32(e.condition),
		Values:        e.values,
		DurationSec:   e.durationSec,
		StrategyGroup: NewStrategyGroupSnapshot(e.strategyGroup),
		Datasource:    NewDatasourceSnapshot(e.datasource),
		Level:         NewLevelSnapshot(e.level),
	}
}

// MarshalEvaluatorSnapshotJSON returns the JSON string of the evaluator snapshot (for storage); returns empty string on error.
func (e *EvaluateMetricStrategyBo) MarshalEvaluatorSnapshotJSON() string {
	snap := e.ToSnapshot()
	if snap == nil {
		return ""
	}
	b, err := json.Marshal(snap)
	if err != nil {
		return ""
	}
	return string(b)
}
