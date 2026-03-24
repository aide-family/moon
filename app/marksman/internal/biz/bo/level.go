package bo

import (
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/timex"
	"github.com/bwmarrin/snowflake"

	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
)

type CreateLevelBo struct {
	Name     string
	Remark   string
	Metadata map[string]string
	BgColor  string
}

func NewCreateLevelBo(req *apiv1.CreateLevelRequest) *CreateLevelBo {
	return &CreateLevelBo{
		Name:     req.Name,
		Remark:   req.Remark,
		Metadata: req.Metadata,
		BgColor:  req.BgColor,
	}
}

type UpdateLevelBo struct {
	UID      snowflake.ID
	Name     string
	Remark   string
	Metadata map[string]string
	BgColor  string
}

func NewUpdateLevelBo(req *apiv1.UpdateLevelRequest) *UpdateLevelBo {
	return &UpdateLevelBo{
		UID:      snowflake.ParseInt64(req.GetUid()),
		Name:     req.Name,
		Remark:   req.Remark,
		Metadata: req.Metadata,
		BgColor:  req.BgColor,
	}
}

type UpdateLevelStatusBo struct {
	UID    snowflake.ID
	Status enum.GlobalStatus
}

func NewUpdateLevelStatusBo(req *apiv1.UpdateLevelStatusRequest) *UpdateLevelStatusBo {
	return &UpdateLevelStatusBo{
		UID:    snowflake.ID(req.GetUid()),
		Status: req.Status,
	}
}

type LevelItemBo struct {
	UID       snowflake.ID
	Name      string
	Remark    string
	BgColor   string
	Status    enum.GlobalStatus
	Metadata  map[string]string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func ToAPIV1LevelItem(b *LevelItemBo) *apiv1.LevelItem {
	if b == nil {
		return nil
	}
	return &apiv1.LevelItem{
		Uid:       b.UID.Int64(),
		Name:      b.Name,
		Remark:    b.Remark,
		BgColor:   b.BgColor,
		Status:    b.Status,
		Metadata:  b.Metadata,
		CreatedAt: timex.FormatTime(&b.CreatedAt),
		UpdatedAt: timex.FormatTime(&b.UpdatedAt),
	}
}

type ListLevelBo struct {
	*PageRequestBo
	Keyword string
	Status  enum.GlobalStatus
}

func NewListLevelBo(req *apiv1.ListLevelRequest) *ListLevelBo {
	return &ListLevelBo{
		PageRequestBo: NewPageRequestBo(req.Page, req.PageSize),
		Keyword:       req.Keyword,
		Status:        req.Status,
	}
}

func ToAPIV1ListLevelReply(pageResponseBo *PageResponseBo[*LevelItemBo]) *apiv1.ListLevelReply {
	items := make([]*apiv1.LevelItem, 0, len(pageResponseBo.GetItems()))
	for _, item := range pageResponseBo.GetItems() {
		items = append(items, ToAPIV1LevelItem(item))
	}
	return &apiv1.ListLevelReply{
		Total:    pageResponseBo.GetTotal(),
		Page:     pageResponseBo.GetPage(),
		PageSize: pageResponseBo.GetPageSize(),
		Items:    items,
	}
}

type SelectLevelBo struct {
	Keyword string
	Limit   int32
	LastUID snowflake.ID
	Status  enum.GlobalStatus
}

func NewSelectLevelBo(req *apiv1.SelectLevelRequest) *SelectLevelBo {
	return &SelectLevelBo{
		Keyword: req.Keyword,
		Limit:   req.Limit,
		LastUID: snowflake.ParseInt64(req.LastUID),
		Status:  req.Status,
	}
}

type LevelItemSelectBo struct {
	Value    int64
	Label    string
	Disabled bool
	Tooltip  string
}

func ToAPIV1LevelItemSelect(b *LevelItemSelectBo) *apiv1.LevelItemSelect {
	if b == nil {
		return nil
	}
	return &apiv1.LevelItemSelect{
		Value:    b.Value,
		Label:    b.Label,
		Disabled: b.Disabled,
		Tooltip:  b.Tooltip,
	}
}

type SelectLevelBoResult struct {
	Items   []*LevelItemSelectBo
	Total   int64
	LastUID snowflake.ID
	HasMore bool
}

func ToAPIV1SelectLevelReply(result *SelectLevelBoResult) *apiv1.SelectLevelReply {
	selectItems := make([]*apiv1.LevelItemSelect, 0, len(result.Items))
	for _, item := range result.Items {
		selectItems = append(selectItems, ToAPIV1LevelItemSelect(item))
	}
	return &apiv1.SelectLevelReply{
		Items:   selectItems,
		Total:   result.Total,
		LastUID: result.LastUID.Int64(),
		HasMore: result.HasMore,
	}
}
