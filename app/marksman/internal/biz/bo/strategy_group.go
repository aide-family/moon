package bo

import (
	"strconv"
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/bwmarrin/snowflake"

	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
)

type CreateStrategyGroupBo struct {
	Name     string
	Remark   string
	Metadata map[string]string
}

func NewCreateStrategyGroupBo(req *apiv1.CreateStrategyGroupRequest) *CreateStrategyGroupBo {
	return &CreateStrategyGroupBo{
		Name:     req.GetName(),
		Remark:   req.GetRemark(),
		Metadata: req.GetMetadata(),
	}
}

type UpdateStrategyGroupBo struct {
	UID      snowflake.ID
	Name     string
	Remark   string
	Metadata map[string]string
}

func NewUpdateStrategyGroupBo(req *apiv1.UpdateStrategyGroupRequest) *UpdateStrategyGroupBo {
	return &UpdateStrategyGroupBo{
		UID:      snowflake.ParseInt64(req.GetUid()),
		Name:     req.GetName(),
		Remark:   req.GetRemark(),
		Metadata: req.GetMetadata(),
	}
}

type UpdateStrategyGroupStatusBo struct {
	UID    snowflake.ID
	Status enum.GlobalStatus
}

func NewUpdateStrategyGroupStatusBo(req *apiv1.UpdateStrategyGroupStatusRequest) *UpdateStrategyGroupStatusBo {
	uid := snowflake.ID(0)
	if i, err := strconv.ParseInt(req.GetUid(), 10, 64); err == nil {
		uid = snowflake.ParseInt64(i)
	}
	return &UpdateStrategyGroupStatusBo{
		UID:    uid,
		Status: req.GetStatus(),
	}
}

type StrategyGroupItemBo struct {
	UID       snowflake.ID
	Name      string
	Remark    string
	Status    enum.GlobalStatus
	Metadata  map[string]string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func ToAPIV1StrategyGroupItem(b *StrategyGroupItemBo) *apiv1.StrategyGroupItem {
	if b == nil {
		return nil
	}
	return &apiv1.StrategyGroupItem{
		Uid:       b.UID.Int64(),
		Name:      b.Name,
		Remark:    b.Remark,
		Status:    b.Status,
		Metadata:  b.Metadata,
		CreatedAt: b.CreatedAt.Format(time.DateTime),
		UpdatedAt: b.UpdatedAt.Format(time.DateTime),
	}
}

type ListStrategyGroupBo struct {
	*PageRequestBo
	Keyword string
	Status  enum.GlobalStatus
}

func NewListStrategyGroupBo(req *apiv1.ListStrategyGroupRequest) *ListStrategyGroupBo {
	return &ListStrategyGroupBo{
		PageRequestBo: NewPageRequestBo(req.GetPage(), req.GetPageSize()),
		Keyword:       req.GetKeyword(),
		Status:        req.GetStatus(),
	}
}

func ToAPIV1ListStrategyGroupReply(pageResponseBo *PageResponseBo[*StrategyGroupItemBo]) *apiv1.ListStrategyGroupReply {
	items := make([]*apiv1.StrategyGroupItem, 0, len(pageResponseBo.GetItems()))
	for _, item := range pageResponseBo.GetItems() {
		items = append(items, ToAPIV1StrategyGroupItem(item))
	}
	return &apiv1.ListStrategyGroupReply{
		Total:    pageResponseBo.GetTotal(),
		Page:     pageResponseBo.GetPage(),
		PageSize: pageResponseBo.GetPageSize(),
		Items:    items,
	}
}

type SelectStrategyGroupBo struct {
	Keyword string
	Limit   int32
	LastUID snowflake.ID
	Status  enum.GlobalStatus
}

func NewSelectStrategyGroupBo(req *apiv1.SelectStrategyGroupRequest) *SelectStrategyGroupBo {
	return &SelectStrategyGroupBo{
		Keyword: req.GetKeyword(),
		Limit:   req.GetLimit(),
		LastUID: snowflake.ParseInt64(req.GetLastUID()),
		Status:  req.GetStatus(),
	}
}

type StrategyGroupItemSelectBo struct {
	Value    int64
	Label    string
	Disabled bool
	Tooltip  string
}

func ToAPIV1StrategyGroupItemSelect(b *StrategyGroupItemSelectBo) *apiv1.StrategyGroupItemSelect {
	if b == nil {
		return nil
	}
	return &apiv1.StrategyGroupItemSelect{
		Value:    b.Value,
		Label:    b.Label,
		Disabled: b.Disabled,
		Tooltip:  b.Tooltip,
	}
}

type SelectStrategyGroupBoResult struct {
	Items   []*StrategyGroupItemSelectBo
	Total   int64
	LastUID snowflake.ID
	HasMore bool
}

func ToAPIV1SelectStrategyGroupReply(result *SelectStrategyGroupBoResult) *apiv1.SelectStrategyGroupReply {
	selectItems := make([]*apiv1.StrategyGroupItemSelect, 0, len(result.Items))
	for _, item := range result.Items {
		selectItems = append(selectItems, ToAPIV1StrategyGroupItemSelect(item))
	}
	nextUID := uint32(0)
	if result.LastUID.Int64() > 0 && result.LastUID.Int64() <= 0xFFFFFFFF {
		nextUID = uint32(result.LastUID.Int64())
	}
	return &apiv1.SelectStrategyGroupReply{
		Items:   selectItems,
		Total:   result.Total,
		HasMore: result.HasMore,
		NextUID: nextUID,
	}
}

type StrategyGroupBindReceiversBo struct {
	StrategyGroupUID snowflake.ID
	ReceiverUIDs     []snowflake.ID
}

func NewStrategyGroupBindReceiversBo(req *apiv1.StrategyGroupBindReceiversRequest) *StrategyGroupBindReceiversBo {
	uids := make([]snowflake.ID, 0, len(req.GetReceiverUIDs()))
	for _, u := range req.GetReceiverUIDs() {
		uids = append(uids, snowflake.ParseInt64(u))
	}
	return &StrategyGroupBindReceiversBo{
		StrategyGroupUID: snowflake.ParseInt64(req.GetUid()),
		ReceiverUIDs:     uids,
	}
}
