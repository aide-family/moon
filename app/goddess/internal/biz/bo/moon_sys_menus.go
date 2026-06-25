package bo

import (
	"sort"
	"time"

	"github.com/aide-family/magicbox/timex"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// SysMenuItemBo 系统菜单业务对象
type SysMenuItemBo struct {
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

// ToAPIV1SysMenuItem 转换为 proto SysMenuItem
func (b *SysMenuItemBo) ToAPIV1SysMenuItem() *goddessv1.SysMenuItem {
	return &goddessv1.SysMenuItem{
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

// ToAPIV1SysMenuListItem 转换为列表项（含父菜单名称）
func (b *SysMenuItemBo) ToAPIV1SysMenuListItem(parentMenuName string) *goddessv1.SysMenuListItem {
	return &goddessv1.SysMenuListItem{
		MenuCode:        b.MenuCode,
		ParentMenuCode:  b.ParentMenuCode,
		ParentMenuName:  parentMenuName,
		MenuName:        b.MenuName,
		MenuPath:        b.MenuPath,
		MenuIcon:        b.MenuIcon,
		MenuComponent:   b.MenuComponent,
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

// ToAPIV1SysMenuTreeItem 转换为树节点（无子节点）
func (b *SysMenuItemBo) ToAPIV1SysMenuTreeItem() *goddessv1.SysMenuTreeItem {
	return &goddessv1.SysMenuTreeItem{
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
	}
}

// ListSysMenuBo 菜单列表查询业务对象
type ListSysMenuBo struct {
	*PageRequestBo
	Keyword    string
	MenuStatus *int32
	SortBy     string
	SortOrder  string
}

// NewListSysMenuBo 从 proto 请求构造查询业务对象
func NewListSysMenuBo(req *goddessv1.ListSysMenuRequest) *ListSysMenuBo {
	return &ListSysMenuBo{
		PageRequestBo: NewPageRequestBo(req.Page, req.PageSize),
		Keyword:       req.Keyword,
		MenuStatus:    req.MenuStatus,
		SortBy:        req.SortBy,
		SortOrder:     req.SortOrder,
	}
}

// TreeSysMenuBo 菜单树查询业务对象
type TreeSysMenuBo struct {
	MenuStatus *int32
	Keyword    string
}

// NewTreeSysMenuBo 从 proto 请求构造树查询业务对象
func NewTreeSysMenuBo(req *goddessv1.TreeSysMenuRequest) *TreeSysMenuBo {
	return &TreeSysMenuBo{
		MenuStatus: req.MenuStatus,
		Keyword:    req.Keyword,
	}
}

// CreateSysMenuBo 新增菜单业务对象
type CreateSysMenuBo struct {
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
}

// NewCreateSysMenuBo 从 proto 请求构造新增菜单业务对象
func NewCreateSysMenuBo(req *goddessv1.CreateSysMenuRequest) *CreateSysMenuBo {
	return &CreateSysMenuBo{
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

// UpdateSysMenuBo 编辑菜单业务对象（指针字段区分未传和传零值）
type UpdateSysMenuBo struct {
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

// NewUpdateSysMenuBo 从 proto 请求构造编辑菜单业务对象
func NewUpdateSysMenuBo(req *goddessv1.UpdateSysMenuRequest) *UpdateSysMenuBo {
	return &UpdateSysMenuBo{
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

// BuildMenuTree 将平铺的菜单列表构建为树形结构
func BuildMenuTree(items []*SysMenuItemBo) []*goddessv1.SysMenuTreeItem {
	// 按 menuSortOrder 升序排序
	sort.Slice(items, func(i, j int) bool {
		return items[i].MenuSortOrder < items[j].MenuSortOrder
	})

	nodeMap := make(map[string]*goddessv1.SysMenuTreeItem, len(items))
	roots := make([]*goddessv1.SysMenuTreeItem, 0)

	// 第一遍：创建所有节点
	for _, item := range items {
		node := item.ToAPIV1SysMenuTreeItem()
		node.Children = make([]*goddessv1.SysMenuTreeItem, 0)
		nodeMap[item.MenuCode] = node
	}

	// 第二遍：挂载到父节点或根列表
	for _, item := range items {
		node := nodeMap[item.MenuCode]
		if item.ParentMenuCode == "" {
			roots = append(roots, node)
		} else if parent, ok := nodeMap[item.ParentMenuCode]; ok {
			parent.Children = append(parent.Children, node)
		} else {
			// 父菜单不在结果集中，作为根节点
			roots = append(roots, node)
		}
	}

	return roots
}
