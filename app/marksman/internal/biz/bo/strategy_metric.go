package bo

import (
	"encoding/json"
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

type EvaluateMetricStrategyBo struct {
	StrategyUID  snowflake.ID
	NamespaceUID snowflake.ID
	Datasource   *DatasourceItemBo
	Expr         string
	Labels       map[string]string
	Summary      string
	Description  string
	Level        *LevelItemBo
	Mode         enum.SampleMode
	Condition    enum.ConditionMetric
	Values       []float64
	DurationSec  int64
}

// MetricEvaluatorSnapshot is the JSON-serializable snapshot payload for metric evaluator (stored in evaluator_snapshots.snapshot_json).
type MetricEvaluatorSnapshot struct {
	StrategyUID  int64               `json:"strategy_uid"`
	NamespaceUID int64               `json:"namespace_uid"`
	Datasource   *DatasourceSnapshot `json:"datasource,omitempty"`
	Expr         string              `json:"expr"`
	Labels       map[string]string   `json:"labels,omitempty"`
	Summary      string              `json:"summary"`
	Description  string              `json:"description"`
	Level        *LevelSnapshot      `json:"level,omitempty"`
	Mode         int32               `json:"mode"`
	Condition    int32               `json:"condition"`
	Values       []float64           `json:"values,omitempty"`
	DurationSec  int64               `json:"duration_sec"`
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

// ToSnapshot builds a JSON-serializable snapshot from EvaluateMetricStrategyBo (metric evaluator payload).
func (e *EvaluateMetricStrategyBo) ToSnapshot() *MetricEvaluatorSnapshot {
	if e == nil {
		return nil
	}
	s := &MetricEvaluatorSnapshot{
		StrategyUID:  e.StrategyUID.Int64(),
		NamespaceUID: e.NamespaceUID.Int64(),
		Expr:         e.Expr,
		Labels:       e.Labels,
		Summary:      e.Summary,
		Description:  e.Description,
		Mode:         int32(e.Mode),
		Condition:    int32(e.Condition),
		Values:       e.Values,
		DurationSec:  e.DurationSec,
	}
	if e.Datasource != nil {
		s.Datasource = &DatasourceSnapshot{
			UID:       e.Datasource.UID.Int64(),
			Name:      e.Datasource.Name,
			Type:      int32(e.Datasource.Type),
			Driver:    int32(e.Datasource.Driver),
			Metadata:  e.Datasource.Metadata,
			Status:    int32(e.Datasource.Status),
			URL:       e.Datasource.URL,
			Remark:    e.Datasource.Remark,
			CreatedAt: timex.FormatTime(&e.Datasource.CreatedAt),
			UpdatedAt: timex.FormatTime(&e.Datasource.UpdatedAt),
		}
	}
	if e.Level != nil {
		s.Level = &LevelSnapshot{
			UID:       e.Level.UID.Int64(),
			Name:      e.Level.Name,
			Remark:    e.Level.Remark,
			Status:    int32(e.Level.Status),
			Metadata:  e.Level.Metadata,
			CreatedAt: timex.FormatTime(&e.Level.CreatedAt),
			UpdatedAt: timex.FormatTime(&e.Level.UpdatedAt),
		}
	}
	return s
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
