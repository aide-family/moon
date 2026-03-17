package bo

import (
	"time"

	"github.com/bwmarrin/snowflake"

	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
)

// AlertEventBo is the business object for an alert event produced by metric strategy evaluation (in-memory).
type AlertEventBo struct {
	StrategyUID   snowflake.ID
	NamespaceUID  snowflake.ID
	Level         *LevelItemBo
	Summary       string
	Description   string
	Expr          string
	FiredAt       time.Time
	Value         float64
	Labels        map[string]string
	DatasourceUID snowflake.ID
}

// AlertEventItemBo is the business object for a persisted alert event (real-time alert).
type AlertEventItemBo struct {
	UID             snowflake.ID
	StrategyUID     snowflake.ID
	NamespaceUID    snowflake.ID
	LevelUID        snowflake.ID
	LevelName       string
	Summary         string
	Description     string
	Expr            string
	FiredAt         time.Time
	Value           float64
	Labels          map[string]string
	DatasourceUID   snowflake.ID
	Status          apiv1.AlertEventStatus
	IntervenedAt    *time.Time
	IntervenedBy    snowflake.ID
	SuppressedUntil *time.Time
	RecoveredAt     *time.Time
	RecoveredBy     snowflake.ID
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func ToAPIV1AlertEventItem(b *AlertEventItemBo) *apiv1.AlertEventItem {
	if b == nil {
		return nil
	}
	item := &apiv1.AlertEventItem{
		Uid:           b.UID.Int64(),
		StrategyUid:    b.StrategyUID.Int64(),
		NamespaceUid:  b.NamespaceUID.Int64(),
		LevelUid:      b.LevelUID.Int64(),
		LevelName:     b.LevelName,
		Summary:       b.Summary,
		Description:   b.Description,
		Expr:          b.Expr,
		FiredAt:       b.FiredAt.Format(time.RFC3339),
		Value:         b.Value,
		Labels:        b.Labels,
		DatasourceUid: b.DatasourceUID.Int64(),
		Status:        b.Status,
		CreatedAt:     b.CreatedAt.Format(time.DateTime),
		UpdatedAt:     b.UpdatedAt.Format(time.DateTime),
	}
	if b.IntervenedAt != nil {
		item.IntervenedAt = b.IntervenedAt.Format(time.RFC3339)
		item.IntervenedBy = b.IntervenedBy.Int64()
	}
	if b.SuppressedUntil != nil {
		item.SuppressedUntil = b.SuppressedUntil.Format(time.RFC3339)
	}
	if b.RecoveredAt != nil {
		item.RecoveredAt = b.RecoveredAt.Format(time.RFC3339)
		item.RecoveredBy = b.RecoveredBy.Int64()
	}
	return item
}

type ListRealtimeAlertBo struct {
	*PageRequestBo
	AlertPageUID snowflake.ID
	Status       apiv1.AlertEventStatus
}

func NewListRealtimeAlertBo(req *apiv1.ListRealtimeAlertRequest) *ListRealtimeAlertBo {
	return &ListRealtimeAlertBo{
		PageRequestBo: NewPageRequestBo(req.GetPage(), req.GetPageSize()),
		AlertPageUID: snowflake.ParseInt64(req.GetAlertPageUid()),
		Status:       req.GetStatus(),
	}
}

func ToAPIV1ListRealtimeAlertReply(pageResponseBo *PageResponseBo[*AlertEventItemBo]) *apiv1.ListRealtimeAlertReply {
	items := make([]*apiv1.AlertEventItem, 0, len(pageResponseBo.GetItems()))
	for _, item := range pageResponseBo.GetItems() {
		items = append(items, ToAPIV1AlertEventItem(item))
	}
	return &apiv1.ListRealtimeAlertReply{
		Total:    pageResponseBo.GetTotal(),
		Page:     pageResponseBo.GetPage(),
		PageSize: pageResponseBo.GetPageSize(),
		Items:    items,
	}
}
