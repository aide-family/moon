package bo

import (
	"strings"
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/strutil"
	"github.com/bwmarrin/snowflake"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

type CreateNamespaceBo struct {
	Name     string
	Metadata map[string]string
	Status   enum.GlobalStatus
	Logo     string
	Secret   string
	Banners  []string
	Remark   string
	Leader   snowflake.ID
}

func NewCreateNamespaceBo(req *goddessv1.CreateNamespaceRequest, leader snowflake.ID) *CreateNamespaceBo {
	return &CreateNamespaceBo{
		Name:     req.Name,
		Metadata: req.Metadata,
		Status:   enum.GlobalStatus_ENABLED,
		Logo:     req.Logo,
		Secret:   strings.ToUpper(strutil.RandomID(16)),
		Banners:  req.Banners,
		Remark:   req.Remark,
		Leader:   leader,
	}
}

type UpdateNamespaceBo struct {
	UID      snowflake.ID
	Name     string
	Metadata map[string]string
	Logo     string
	Banners  []string
	Remark   string
}

func NewUpdateNamespaceBo(req *goddessv1.UpdateNamespaceRequest) *UpdateNamespaceBo {
	return &UpdateNamespaceBo{
		UID:      snowflake.ParseInt64(req.Uid),
		Name:     req.Name,
		Metadata: req.Metadata,
		Logo:     req.Logo,
		Banners:  req.Banners,
		Remark:   req.Remark,
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
	Logo      string
	Secret    string
	Banners   []string
	Remark    string
	Leader    snowflake.ID
}

func (b *NamespaceItemBo) ToAPIV1NamespaceItem() *goddessv1.NamespaceItem {
	return &goddessv1.NamespaceItem{
		Uid:       b.UID.Int64(),
		Name:      b.Name,
		Metadata:  b.Metadata,
		Status:    b.Status,
		CreatedAt: b.CreatedAt.Format(time.DateTime),
		UpdatedAt: b.UpdatedAt.Format(time.DateTime),
		Logo:      b.Logo,
		Secret:    b.Secret,
		Banners:   b.Banners,
		Remark:    b.Remark,
		Leader:    b.Leader.Int64(),
	}
}

func NewUpdateNamespaceStatusBo(req *goddessv1.UpdateNamespaceStatusRequest) *UpdateNamespaceStatusBo {
	return &UpdateNamespaceStatusBo{
		UID:    snowflake.ParseInt64(req.Uid),
		Status: req.Status,
	}
}

func NewListNamespaceBo(req *goddessv1.ListNamespaceRequest) *ListNamespaceBo {
	return &ListNamespaceBo{
		PageRequestBo: NewPageRequestBo(req.Page, req.PageSize),
		Keyword:       req.Keyword,
		Status:        req.Status,
	}
}

func ToAPIV1ListNamespaceReply(pageResponseBo *PageResponseBo[*NamespaceItemBo]) *goddessv1.ListNamespaceReply {
	items := make([]*goddessv1.NamespaceItem, 0, len(pageResponseBo.GetItems()))
	for _, item := range pageResponseBo.GetItems() {
		items = append(items, item.ToAPIV1NamespaceItem())
	}
	return &goddessv1.ListNamespaceReply{
		Items:    items,
		Total:    pageResponseBo.GetTotal(),
		Page:     pageResponseBo.GetPage(),
		PageSize: pageResponseBo.GetPageSize(),
	}
}

// SelectNamespaceBo is the BO for select-namespace request.
type SelectNamespaceBo struct {
	Keyword string
	Limit   int32
	LastUID snowflake.ID
	Status  enum.GlobalStatus
}

// NewSelectNamespaceBo builds SelectNamespaceBo from API request.
func NewSelectNamespaceBo(req *goddessv1.SelectNamespaceRequest) *SelectNamespaceBo {
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

// NamespaceItemSelectBo is one namespace item in select result.
type NamespaceItemSelectBo struct {
	UID      snowflake.ID
	Name     string
	Status   enum.GlobalStatus
	Disabled bool
	Tooltip  string
}

// ToAPIV1NamespaceItemSelect converts BO to API response.
func (b *NamespaceItemSelectBo) ToAPIV1NamespaceItemSelect() *goddessv1.NamespaceItemSelect {
	return &goddessv1.NamespaceItemSelect{
		Value:    b.UID.Int64(),
		Label:    b.Name,
		Disabled: b.Disabled,
		Tooltip:  b.Tooltip,
	}
}

// SelectNamespaceBoResult is the biz result for select namespace.
type SelectNamespaceBoResult struct {
	Items   []*NamespaceItemSelectBo
	Total   int64
	LastUID snowflake.ID
	HasMore bool
}

// ToAPIV1SelectNamespaceReply converts BO result to API response.
func ToAPIV1SelectNamespaceReply(result *SelectNamespaceBoResult) *goddessv1.SelectNamespaceReply {
	selectItems := make([]*goddessv1.NamespaceItemSelect, 0, len(result.Items))
	for _, item := range result.Items {
		selectItems = append(selectItems, item.ToAPIV1NamespaceItemSelect())
	}

	return &goddessv1.SelectNamespaceReply{
		Items:   selectItems,
		Total:   result.Total,
		LastUID: result.LastUID.Int64(),
		HasMore: result.HasMore,
	}
}
