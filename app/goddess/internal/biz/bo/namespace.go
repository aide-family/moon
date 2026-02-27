package bo

import (
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/bwmarrin/snowflake"

	magicboxv1 "github.com/aide-family/magicbox/api/v1"
)

type CreateNamespaceBo struct {
	Name     string
	Metadata map[string]string
	Status   enum.GlobalStatus
}

func NewCreateNamespaceBo(req *magicboxv1.CreateNamespaceRequest) *CreateNamespaceBo {
	return &CreateNamespaceBo{
		Name:     req.Name,
		Metadata: req.Metadata,
		Status:   enum.GlobalStatus_ENABLED,
	}
}

type UpdateNamespaceBo struct {
	UID      snowflake.ID
	Name     string
	Metadata map[string]string
}

func NewUpdateNamespaceBo(req *magicboxv1.UpdateNamespaceRequest) *UpdateNamespaceBo {
	return &UpdateNamespaceBo{
		UID:      snowflake.ParseInt64(req.Uid),
		Name:     req.Name,
		Metadata: req.Metadata,
	}
}

type UpdateNamespaceStatusBo struct {
	UID    snowflake.ID
	Status enum.GlobalStatus
}

type ListNamespaceBo struct {
	*PageRequestBo
	Keyword string
	Status  enum.GlobalStatus
}

type NamespaceItemBo struct {
	UID       snowflake.ID
	Name      string
	Metadata  map[string]string
	Status    enum.GlobalStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (b *NamespaceItemBo) ToAPIV1NamespaceItem() *magicboxv1.NamespaceItem {
	return &magicboxv1.NamespaceItem{
		Uid:       b.UID.Int64(),
		Name:      b.Name,
		Metadata:  b.Metadata,
		Status:    b.Status,
		CreatedAt: b.CreatedAt.Format(time.DateTime),
		UpdatedAt: b.UpdatedAt.Format(time.DateTime),
	}
}

func NewUpdateNamespaceStatusBo(req *magicboxv1.UpdateNamespaceStatusRequest) *UpdateNamespaceStatusBo {
	return &UpdateNamespaceStatusBo{
		UID:    snowflake.ParseInt64(req.Uid),
		Status: req.Status,
	}
}

func NewListNamespaceBo(req *magicboxv1.ListNamespaceRequest) *ListNamespaceBo {
	return &ListNamespaceBo{
		PageRequestBo: NewPageRequestBo(req.Page, req.PageSize),
		Keyword:       req.Keyword,
		Status:        req.Status,
	}
}

func ToAPIV1ListNamespaceReply(pageResponseBo *PageResponseBo[*NamespaceItemBo]) *magicboxv1.ListNamespaceReply {
	items := make([]*magicboxv1.NamespaceItem, 0, len(pageResponseBo.GetItems()))
	for _, item := range pageResponseBo.GetItems() {
		items = append(items, item.ToAPIV1NamespaceItem())
	}
	return &magicboxv1.ListNamespaceReply{
		Items:    items,
		Total:    pageResponseBo.GetTotal(),
		Page:     pageResponseBo.GetPage(),
		PageSize: pageResponseBo.GetPageSize(),
	}
}

// SelectNamespaceBo 选择Namespace的 BO
type SelectNamespaceBo struct {
	Keyword string
	Limit   int32
	LastUID snowflake.ID
	Status  enum.GlobalStatus
}

// NewSelectNamespaceBo 从 API 请求创建 BO
func NewSelectNamespaceBo(req *magicboxv1.SelectNamespaceRequest) *SelectNamespaceBo {
	var lastUID snowflake.ID
	if req.LastUID > 0 {
		lastUID = snowflake.ParseInt64(req.LastUID)
	}
	return &SelectNamespaceBo{
		Keyword: req.Keyword,
		Limit:   req.Limit,
		LastUID: lastUID,
		Status:  req.Status,
	}
}

// NamespaceItemSelectBo Namespace选择项的 BO
type NamespaceItemSelectBo struct {
	UID      snowflake.ID
	Name     string
	Status   enum.GlobalStatus
	Disabled bool
	Tooltip  string
}

// ToAPIV1NamespaceItemSelect 转换为 API 响应
func (b *NamespaceItemSelectBo) ToAPIV1NamespaceItemSelect() *magicboxv1.NamespaceItemSelect {
	return &magicboxv1.NamespaceItemSelect{
		Value:    b.UID.Int64(),
		Label:    b.Name,
		Disabled: b.Disabled,
		Tooltip:  b.Tooltip,
	}
}

// SelectNamespaceBoResult Biz层返回结果
type SelectNamespaceBoResult struct {
	Items   []*NamespaceItemSelectBo
	Total   int64
	LastUID snowflake.ID
	HasMore bool
}

// ToAPIV1SelectNamespaceReply 转换为 API 响应
func ToAPIV1SelectNamespaceReply(result *SelectNamespaceBoResult) *magicboxv1.SelectNamespaceReply {
	selectItems := make([]*magicboxv1.NamespaceItemSelect, 0, len(result.Items))
	for _, item := range result.Items {
		selectItems = append(selectItems, item.ToAPIV1NamespaceItemSelect())
	}

	return &magicboxv1.SelectNamespaceReply{
		Items:   selectItems,
		Total:   result.Total,
		LastUID: result.LastUID.Int64(),
		HasMore: result.HasMore,
	}
}
