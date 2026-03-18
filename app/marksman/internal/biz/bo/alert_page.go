package bo

import (
	"time"

	"github.com/aide-family/magicbox/timex"
	"github.com/bwmarrin/snowflake"

	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
)

type CreateAlertPageBo struct {
	Name      string
	Color     string
	SortOrder int32
	Filter    *AlertPageFilterBo
}

func NewCreateAlertPageBo(req *apiv1.CreateAlertPageRequest) *CreateAlertPageBo {
	return &CreateAlertPageBo{
		Name:      req.GetName(),
		Color:     req.GetColor(),
		SortOrder: req.GetSortOrder(),
		Filter:    ToAlertPageFilterBo(req.GetFilter()),
	}
}

type AlertPageFilterBo struct {
	StrategyGroupUIDs []int64
	LevelUIDs         []int64
	StrategyUIDs      []int64
}

func ToAlertPageFilterBo(p *apiv1.AlertPageFilter) *AlertPageFilterBo {
	if p == nil {
		return nil
	}
	return &AlertPageFilterBo{
		StrategyGroupUIDs: p.GetStrategyGroupUids(),
		LevelUIDs:         p.GetLevelUids(),
		StrategyUIDs:      p.GetStrategyUids(),
	}
}

type UpdateAlertPageBo struct {
	UID       snowflake.ID
	Name      string
	Color     string
	SortOrder int32
	Filter    *AlertPageFilterBo
}

func NewUpdateAlertPageBo(req *apiv1.UpdateAlertPageRequest) *UpdateAlertPageBo {
	return &UpdateAlertPageBo{
		UID:       snowflake.ParseInt64(req.GetUid()),
		Name:      req.GetName(),
		Color:     req.GetColor(),
		SortOrder: req.GetSortOrder(),
		Filter:    ToAlertPageFilterBo(req.GetFilter()),
	}
}

type AlertPageItemBo struct {
	UID       snowflake.ID
	Name      string
	Color     string
	SortOrder int32
	Filter    *AlertPageFilterBo
	CreatedAt time.Time
	UpdatedAt time.Time
}

func ToAPIV1AlertPageItem(b *AlertPageItemBo) *apiv1.AlertPageItem {
	if b == nil {
		return nil
	}
	var filter *apiv1.AlertPageFilter
	if b.Filter != nil {
		filter = &apiv1.AlertPageFilter{
			StrategyGroupUids: b.Filter.StrategyGroupUIDs,
			LevelUids:         b.Filter.LevelUIDs,
			StrategyUids:      b.Filter.StrategyUIDs,
		}
	}
	return &apiv1.AlertPageItem{
		Uid:       b.UID.Int64(),
		Name:      b.Name,
		Color:     b.Color,
		SortOrder: b.SortOrder,
		Filter:    filter,
		CreatedAt: timex.FormatTime(&b.CreatedAt),
		UpdatedAt: timex.FormatTime(&b.UpdatedAt),
	}
}

type ListAlertPageBo struct {
	*PageRequestBo
	Keyword string
}

func NewListAlertPageBo(req *apiv1.ListAlertPageRequest) *ListAlertPageBo {
	return &ListAlertPageBo{
		PageRequestBo: NewPageRequestBo(req.GetPage(), req.GetPageSize()),
		Keyword:       req.GetKeyword(),
	}
}

func ToAPIV1ListAlertPageReply(pageResponseBo *PageResponseBo[*AlertPageItemBo]) *apiv1.ListAlertPageReply {
	items := make([]*apiv1.AlertPageItem, 0, len(pageResponseBo.GetItems()))
	for _, item := range pageResponseBo.GetItems() {
		items = append(items, ToAPIV1AlertPageItem(item))
	}
	return &apiv1.ListAlertPageReply{
		Total:    pageResponseBo.GetTotal(),
		Page:     pageResponseBo.GetPage(),
		PageSize: pageResponseBo.GetPageSize(),
		Items:    items,
	}
}

// ToAPIV1ListUserAlertPagesReply converts biz items to ListUserAlertPagesReply.
func ToAPIV1ListUserAlertPagesReply(items []*AlertPageItemBo) *apiv1.ListUserAlertPagesReply {
	out := make([]*apiv1.AlertPageItem, 0, len(items))
	for _, item := range items {
		out = append(out, ToAPIV1AlertPageItem(item))
	}
	return &apiv1.ListUserAlertPagesReply{Items: out}
}
