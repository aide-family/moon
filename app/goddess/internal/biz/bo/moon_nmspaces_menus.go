package bo

import (
	"time"

	"github.com/aide-family/magicbox/timex"
	"github.com/bwmarrin/snowflake"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// NmspacesMenuItemBo 租户菜单业务对象
type NmspacesMenuItemBo struct {
	MenuCode        string
	ParentMenuCode  string
	MenuName        string
	MenuPath        string
	MenuIcon        string
	MenuComponent   string
	MenuRedirect    string
	MenuIsHidden    int32
	MenuIsKeepAlive int32
	MenuSortOrder   int32
	MenuStatus      int32
	MenuRemark      string
	Creator         int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// ToAPIV1NmspacesMenuItem 转换为 proto 响应对象
func (b *NmspacesMenuItemBo) ToAPIV1NmspacesMenuItem() *goddessv1.NmspacesMenuItem {
	return &goddessv1.NmspacesMenuItem{
		MenuCode:        b.MenuCode,
		ParentMenuCode:  b.ParentMenuCode,
		MenuName:        b.MenuName,
		MenuPath:        b.MenuPath,
		MenuIcon:        b.MenuIcon,
		MenuComponent:   b.MenuComponent,
		MenuRedirect:    b.MenuRedirect,
		MenuIsHidden:    b.MenuIsHidden,
		MenuIsKeepAlive: b.MenuIsKeepAlive,
		MenuSortOrder:   b.MenuSortOrder,
		MenuStatus:      b.MenuStatus,
		MenuRemark:      b.MenuRemark,
		Creator:         b.Creator,
		CreatedAt:       timex.FormatTime(&b.CreatedAt),
		UpdatedAt:       timex.FormatTime(&b.UpdatedAt),
	}
}

// ListNmspacesMenuBo 租户菜单列表查询业务对象
type ListNmspacesMenuBo struct {
	*PageRequestBo
	NamespaceUID snowflake.ID
	Keyword      string
	MenuStatus   *int32
	SortBy       string
	SortOrder    string
}

// NewListNmspacesMenuBo 从 proto 请求构造查询业务对象
func NewListNmspacesMenuBo(namespaceUID snowflake.ID, req *goddessv1.ListNmspacesMenuRequest) *ListNmspacesMenuBo {
	return &ListNmspacesMenuBo{
		PageRequestBo: NewPageRequestBo(req.Page, req.PageSize),
		NamespaceUID:  namespaceUID,
		Keyword:       req.Keyword,
		MenuStatus:    req.MenuStatus,
		SortBy:        req.SortBy,
		SortOrder:     req.SortOrder,
	}
}

// ToAPIV1ListNmspacesMenuReply 将分页结果转换为 proto 响应
func ToAPIV1ListNmspacesMenuReply(items []*NmspacesMenuItemBo, total int64, page int32, pageSize int32) *goddessv1.ListNmspacesMenuReply {
	apiItems := make([]*goddessv1.NmspacesMenuItem, 0, len(items))
	for _, item := range items {
		apiItems = append(apiItems, item.ToAPIV1NmspacesMenuItem())
	}
	return &goddessv1.ListNmspacesMenuReply{
		Items:    apiItems,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}

// CreateNmspacesMenuBo 新增租户菜单业务对象
type CreateNmspacesMenuBo struct {
	NamespaceUID    snowflake.ID
	MenuCode        string
	ParentMenuCode  string
	MenuName        string
	MenuPath        string
	MenuIcon        string
	MenuComponent   string
	MenuRedirect    string
	MenuIsHidden    int32
	MenuIsKeepAlive int32
	MenuSortOrder   int32
	MenuStatus      int32
	MenuRemark      string
	Creator         snowflake.ID
}

// NewCreateNmspacesMenuBoFromSysMenu 从系统菜单业务对象构造新增租户菜单业务对象
func NewCreateNmspacesMenuBoFromSysMenu(namespaceUID snowflake.ID, creator snowflake.ID, sysMenu *SysMenuItemBo) *CreateNmspacesMenuBo {
	return &CreateNmspacesMenuBo{
		NamespaceUID:    namespaceUID,
		MenuCode:        sysMenu.MenuCode,
		ParentMenuCode:  sysMenu.ParentMenuCode,
		MenuName:        sysMenu.MenuName,
		MenuPath:        sysMenu.MenuPath,
		MenuIcon:        sysMenu.MenuIcon,
		MenuComponent:   sysMenu.MenuComponent,
		MenuRedirect:    sysMenu.MenuRedirect,
		MenuIsHidden:    sysMenu.MenuIsHidden,
		MenuIsKeepAlive: sysMenu.MenuIsKeepAlive,
		MenuSortOrder:   sysMenu.MenuSortOrder,
		MenuStatus:      sysMenu.MenuStatus,
		MenuRemark:      sysMenu.MenuRemark,
		Creator:         creator,
	}
}

// UpdateNmspacesMenuBo 编辑租户菜单业务对象
type UpdateNmspacesMenuBo struct {
	NamespaceUID    snowflake.ID
	MenuCode        string
	ParentMenuCode  *string
	MenuName        *string
	MenuPath        *string
	MenuIcon        *string
	MenuComponent   *string
	MenuRedirect    *string
	MenuIsHidden    *int32
	MenuIsKeepAlive *int32
	MenuSortOrder   *int32
	MenuStatus      *int32
	MenuRemark      *string
}

// NewUpdateNmspacesMenuBo 从 proto 请求构造编辑菜单业务对象
func NewUpdateNmspacesMenuBo(namespaceUID snowflake.ID, req *goddessv1.UpdateNmspacesMenuRequest) *UpdateNmspacesMenuBo {
	return &UpdateNmspacesMenuBo{
		NamespaceUID:    namespaceUID,
		MenuCode:        req.MenuCode,
		ParentMenuCode:  req.ParentMenuCode,
		MenuName:        req.MenuName,
		MenuPath:        req.MenuPath,
		MenuIcon:        req.MenuIcon,
		MenuComponent:   req.MenuComponent,
		MenuRedirect:    req.MenuRedirect,
		MenuIsHidden:    req.MenuIsHidden,
		MenuIsKeepAlive: req.MenuIsKeepAlive,
		MenuSortOrder:   req.MenuSortOrder,
		MenuStatus:      req.MenuStatus,
		MenuRemark:      req.MenuRemark,
	}
}
