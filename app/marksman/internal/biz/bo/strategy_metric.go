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
	Status         enum.GlobalStatus
	DatasourceUIDs []int64
}

func NewSaveStrategyMetricBo(req *apiv1.SaveStrategyMetricRequest) *SaveStrategyMetricBo {
	return &SaveStrategyMetricBo{
		StrategyUID:    snowflake.ParseInt64(req.GetStrategyUID()),
		Expr:           req.GetExpr(),
		Labels:         req.GetLabels(),
		Summary:        req.GetSummary(),
		Description:    req.GetDescription(),
		Status:         req.GetStatus(),
		DatasourceUIDs: req.GetDatasourceUIDs(),
	}
}

type StrategyMetricItemBo struct {
	StrategyUID    snowflake.ID
	Expr           string
	Labels         map[string]string
	Summary        string
	Description    string
	Status         enum.GlobalStatus
	DatasourceUIDs []int64
	Levels         []*StrategyMetricLevelItemBo
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (b *StrategyMetricItemBo) ToAPIV1StrategyMetricItem() *apiv1.StrategyMetricItem {
	levels := make([]*apiv1.StrategyMetricLevelItem, 0, len(b.Levels))
	for _, l := range b.Levels {
		levels = append(levels, l.ToAPIV1StrategyMetricLevelItem())
	}
	return &apiv1.StrategyMetricItem{
		StrategyUID:    b.StrategyUID.Int64(),
		Expr:           b.Expr,
		Labels:         b.Labels,
		Summary:        b.Summary,
		Description:    b.Description,
		Status:         b.Status,
		DatasourceUIDs: b.DatasourceUIDs,
		Levels:         levels,
		CreatedAt:      b.CreatedAt.Format(time.DateTime),
		UpdatedAt:      b.UpdatedAt.Format(time.DateTime),
	}
}

// StrategyMetricLevelItemBo for one strategy_metric_level row; Level is loaded from level table.
type StrategyMetricLevelItemBo struct {
	UID         snowflake.ID
	StrategyUID snowflake.ID
	Level       *LevelItemBo // loaded from levels table
	Mode        enum.SampleMode
	Condition   enum.ConditionMetric
	Values      []float64
	DurationSec int64
	Status      enum.GlobalStatus
}

func (b *StrategyMetricLevelItemBo) ToAPIV1StrategyMetricLevelItem() *apiv1.StrategyMetricLevelItem {
	out := &apiv1.StrategyMetricLevelItem{
		Uid:         b.UID.Int64(),
		StrategyUID: b.StrategyUID.Int64(),
		Mode:        b.Mode,
		Condition:   b.Condition,
		Values:      b.Values,
		Status:      b.Status,
	}
	if b.Level != nil {
		out.Level = b.Level.ToAPIV1LevelItem()
	}
	if b.DurationSec > 0 {
		out.Duration = durationpb.New(time.Duration(b.DurationSec) * time.Second)
	}
	return out
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
	UID         snowflake.ID
	StrategyUID snowflake.ID
	Status      enum.GlobalStatus
}

func NewUpdateStrategyMetricLevelStatusBo(req *apiv1.UpdateStrategyMetricLevelStatusRequest) *UpdateStrategyMetricLevelStatusBo {
	return &UpdateStrategyMetricLevelStatusBo{
		UID:         snowflake.ParseInt64(req.GetUid()),
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
