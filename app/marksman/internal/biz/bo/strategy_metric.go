package bo

import (
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/timex"
	"github.com/bwmarrin/snowflake"
	"google.golang.org/protobuf/types/known/durationpb"

	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
)

// SaveStrategyMetricBo for SaveStrategyMetric (upsert by strategy_uid)
type SaveStrategyMetricBo struct {
	StrategyUID    snowflake.ID
	Expr           string
	Labels         map[string]string
	Summary        string
	Description    string
	DatasourceUIDs []int64
}

func NewSaveStrategyMetricBo(req *apiv1.SaveStrategyMetricRequest) *SaveStrategyMetricBo {
	return &SaveStrategyMetricBo{
		StrategyUID:    snowflake.ParseInt64(req.GetStrategyUID()),
		Expr:           req.GetExpr(),
		Labels:         req.GetLabels(),
		Summary:        req.GetSummary(),
		Description:    req.GetDescription(),
		DatasourceUIDs: req.GetDatasourceUIDs(),
	}
}

type StrategyMetricItemBo struct {
	UID            snowflake.ID // strategy_metrics.id
	StrategyUID    snowflake.ID
	Strategy       *StrategyItemBo
	Expr           string
	Labels         map[string]string
	Summary        string
	Description    string
	DatasourceUIDs []int64
	Levels         []*StrategyMetricLevelItemBo
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func ToAPIV1StrategyMetricItem(b *StrategyMetricItemBo) *apiv1.StrategyMetricItem {
	if b == nil {
		return nil
	}
	levels := make([]*apiv1.StrategyMetricLevelItem, 0, len(b.Levels))
	for _, l := range b.Levels {
		levels = append(levels, ToAPIV1StrategyMetricLevelItem(l))
	}
	return &apiv1.StrategyMetricItem{
		StrategyUID:    b.StrategyUID.Int64(),
		Expr:           b.Expr,
		Labels:         b.Labels,
		Summary:        b.Summary,
		Description:    b.Description,
		DatasourceUIDs: b.DatasourceUIDs,
		Levels:         levels,
		CreatedAt:      timex.FormatTime(&b.CreatedAt),
		UpdatedAt:      timex.FormatTime(&b.UpdatedAt),
		Strategy:       ToAPIV1StrategyItem(b.Strategy),
	}
}

// StrategyMetricLevelItemBo for one strategy_metric_level row; Level is loaded from level table.
type StrategyMetricLevelItemBo struct {
	LevelUID    snowflake.ID
	StrategyUID snowflake.ID
	Level       *LevelItemBo // loaded from levels table
	Mode        enum.SampleMode
	Condition   enum.ConditionMetric
	Values      []float64
	DurationSec int64
	Status      enum.GlobalStatus
}

func ToAPIV1StrategyMetricLevelItem(b *StrategyMetricLevelItemBo) *apiv1.StrategyMetricLevelItem {
	if b == nil {
		return nil
	}
	return &apiv1.StrategyMetricLevelItem{
		LevelUID:    b.LevelUID.Int64(),
		StrategyUID: b.StrategyUID.Int64(),
		Level:       ToAPIV1LevelItem(b.Level),
		Mode:        b.Mode,
		Condition:   b.Condition,
		Values:      b.Values,
		Duration:    durationpb.New(time.Duration(b.DurationSec) * time.Second),
		Status:      b.Status,
	}
}

// SaveStrategyMetricLevelBo for SaveStrategyMetricLevel (upsert by strategy_uid + level_uid)
type SaveStrategyMetricLevelBo struct {
	StrategyUID snowflake.ID
	LevelUID    snowflake.ID
	Mode        enum.SampleMode
	Condition   enum.ConditionMetric
	Values      []float64
	DurationSec int64
	Status      enum.GlobalStatus
}

func NewSaveStrategyMetricLevelBo(req *apiv1.SaveStrategyMetricLevelRequest) *SaveStrategyMetricLevelBo {
	durSec := int64(0)
	if req.GetDuration() != nil {
		durSec = int64(req.GetDuration().AsDuration().Seconds())
	}
	return &SaveStrategyMetricLevelBo{
		StrategyUID: snowflake.ParseInt64(req.GetStrategyUID()),
		LevelUID:    snowflake.ParseInt64(req.GetLevelUID()),
		Mode:        req.GetMode(),
		Condition:   req.GetCondition(),
		Values:      req.GetValues(),
		DurationSec: durSec,
		Status:      req.GetStatus(),
	}
}

type UpdateStrategyMetricLevelStatusBo struct {
	LevelUID    snowflake.ID
	StrategyUID snowflake.ID
	Status      enum.GlobalStatus
}

func NewUpdateStrategyMetricLevelStatusBo(req *apiv1.UpdateStrategyMetricLevelStatusRequest) *UpdateStrategyMetricLevelStatusBo {
	return &UpdateStrategyMetricLevelStatusBo{
		LevelUID:    snowflake.ParseInt64(req.GetLevelUID()),
		StrategyUID: snowflake.ParseInt64(req.GetStrategyUID()),
		Status:      req.GetStatus(),
	}
}

// MetricEvaluatorSnapshot is the JSON-serializable snapshot payload for metric evaluator (stored in evaluator_snapshots.snapshot_json).
type MetricEvaluatorSnapshot struct {
	StrategyGroup *StrategyGroupSnapshot `json:"strategy_group"`
	StrategyUID   int64                  `json:"strategy_uid"`
	NamespaceUID  int64                  `json:"namespace_uid"`
	Datasource    *DatasourceSnapshot    `json:"datasource,omitempty"`
	Expr          string                 `json:"expr"`
	Labels        map[string]string      `json:"labels,omitempty"`
	Summary       string                 `json:"summary"`
	Description   string                 `json:"description"`
	Level         *LevelSnapshot         `json:"level,omitempty"`
	Mode          int32                  `json:"mode"`
	Condition     int32                  `json:"condition"`
	Values        []float64              `json:"values,omitempty"`
	DurationSec   int64                  `json:"duration_sec"`
}

type StrategyGroupSnapshot struct {
	UID       int64             `json:"uid"`
	Name      string            `json:"name"`
	Remark    string            `json:"remark"`
	Status    int32             `json:"status"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt string            `json:"created_at,omitempty"`
	UpdatedAt string            `json:"updated_at,omitempty"`
}

// LevelSnapshot is JSON-serializable level info.
type LevelSnapshot struct {
	UID       int64             `json:"uid"`
	Name      string            `json:"name"`
	Remark    string            `json:"remark"`
	Status    int32             `json:"status"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt string            `json:"created_at,omitempty"`
	UpdatedAt string            `json:"updated_at,omitempty"`
}

// DatasourceSnapshot is JSON-serializable datasource info.
type DatasourceSnapshot struct {
	UID       int64             `json:"uid"`
	Name      string            `json:"name"`
	Type      int32             `json:"type"`
	Driver    int32             `json:"driver"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Status    int32             `json:"status"`
	URL       string            `json:"url,omitempty"`
	Remark    string            `json:"remark,omitempty"`
	CreatedAt string            `json:"created_at,omitempty"`
	UpdatedAt string            `json:"updated_at,omitempty"`
}

func NewStrategyGroupSnapshot(strategyGroup *StrategyGroupItemBo) *StrategyGroupSnapshot {
	if strategyGroup == nil {
		return nil
	}
	return &StrategyGroupSnapshot{
		UID:       strategyGroup.UID.Int64(),
		Name:      strategyGroup.Name,
		Remark:    strategyGroup.Remark,
		Status:    int32(strategyGroup.Status),
		Metadata:  strategyGroup.Metadata,
		CreatedAt: timex.FormatTime(&strategyGroup.CreatedAt),
		UpdatedAt: timex.FormatTime(&strategyGroup.UpdatedAt),
	}
}

func NewDatasourceSnapshot(datasource *DatasourceItemBo) *DatasourceSnapshot {
	if datasource == nil {
		return nil
	}
	return &DatasourceSnapshot{
		UID:       datasource.UID.Int64(),
		Name:      datasource.Name,
		Type:      int32(datasource.Type),
		Driver:    int32(datasource.Driver),
		Metadata:  datasource.Metadata,
		Status:    int32(datasource.Status),
		URL:       datasource.URL,
		Remark:    datasource.Remark,
		CreatedAt: timex.FormatTime(&datasource.CreatedAt),
		UpdatedAt: timex.FormatTime(&datasource.UpdatedAt),
	}
}

func NewLevelSnapshot(level *LevelItemBo) *LevelSnapshot {
	if level == nil {
		return nil
	}
	return &LevelSnapshot{
		UID:       level.UID.Int64(),
		Name:      level.Name,
		Remark:    level.Remark,
		Status:    int32(level.Status),
		Metadata:  level.Metadata,
		CreatedAt: timex.FormatTime(&level.CreatedAt),
		UpdatedAt: timex.FormatTime(&level.UpdatedAt),
	}
}
