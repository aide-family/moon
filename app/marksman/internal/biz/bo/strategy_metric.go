package bo

import (
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/bwmarrin/snowflake"

	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
	"google.golang.org/protobuf/types/known/durationpb"
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
		CreatedAt:      b.CreatedAt.Format(time.DateTime),
		UpdatedAt:      b.UpdatedAt.Format(time.DateTime),
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

type StrategyMetricBindReceiversBo struct {
	StrategyUID  snowflake.ID
	ReceiverUIDs []snowflake.ID
	LevelUID     snowflake.ID // optional
}

func NewStrategyMetricBindReceiversBo(req *apiv1.StrategyMetricBindReceiversRequest) *StrategyMetricBindReceiversBo {
	uids := make([]snowflake.ID, 0, len(req.GetReceiverUIDs()))
	for _, u := range req.GetReceiverUIDs() {
		uids = append(uids, snowflake.ParseInt64(u))
	}
	return &StrategyMetricBindReceiversBo{
		StrategyUID:  snowflake.ParseInt64(req.GetStrategyUID()),
		ReceiverUIDs: uids,
		LevelUID:     snowflake.ParseInt64(req.GetLevelUID()),
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
