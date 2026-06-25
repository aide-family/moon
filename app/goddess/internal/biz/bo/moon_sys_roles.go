package bo

import (
	"time"

	"github.com/aide-family/magicbox/timex"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// SysRoleItemBo 系统角色业务对象
type SysRoleItemBo struct {
	RoleCode      string
	RoleName      string
	RoleDataScope int32
	RoleStatus    int32
	RoleSortOrder int32
	RoleRemark    string
	Creator       int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ToAPIV1SysRoleItem 将业务对象转换为 proto 响应对象
func (b *SysRoleItemBo) ToAPIV1SysRoleItem() *goddessv1.SysRoleItem {
	return &goddessv1.SysRoleItem{
		RoleCode:      b.RoleCode,
		RoleName:      b.RoleName,
		RoleDataScope: b.RoleDataScope,
		RoleStatus:    b.RoleStatus,
		RoleSortOrder: b.RoleSortOrder,
		RoleRemark:    b.RoleRemark,
		Creator:       b.Creator,
		CreatedAt:     timex.FormatTime(&b.CreatedAt),
		UpdatedAt:     timex.FormatTime(&b.UpdatedAt),
	}
}

// ListSysRoleBo 系统角色列表查询业务对象
type ListSysRoleBo struct {
	*PageRequestBo
	Keyword    string
	RoleCode   string
	RoleStatus *int32 // 指针类型，区分未传和传 0（角色状态：0-未知 1-启用 2-禁用）
	SortBy     string
	SortOrder  string
}

// NewListSysRoleBo 从 proto 请求构造查询业务对象
func NewListSysRoleBo(req *goddessv1.ListSysRoleRequest) *ListSysRoleBo {
	return &ListSysRoleBo{
		PageRequestBo: NewPageRequestBo(req.Page, req.PageSize),
		Keyword:       req.Keyword,
		RoleCode:      req.RoleCode,
		RoleStatus:    req.RoleStatus, // optional *int32，nil 表示不过滤
		SortBy:        req.SortBy,
		SortOrder:     req.SortOrder,
	}
}

// ToAPIV1ListSysRoleReply 将分页结果转换为 proto 响应
func ToAPIV1ListSysRoleReply(pageResponseBo *PageResponseBo[*SysRoleItemBo]) *goddessv1.ListSysRoleReply {
	items := make([]*goddessv1.SysRoleItem, 0, len(pageResponseBo.GetItems()))
	for _, item := range pageResponseBo.GetItems() {
		items = append(items, item.ToAPIV1SysRoleItem())
	}
	return &goddessv1.ListSysRoleReply{
		Items:    items,
		Total:    pageResponseBo.GetTotal(),
		Page:     pageResponseBo.GetPage(),
		PageSize: pageResponseBo.GetPageSize(),
	}
}

// SysRoleMenuItemBo 角色关联的菜单业务对象
type SysRoleMenuItemBo struct {
	MenuCode string
	MenuName string
}

// ToAPIV1 转换为 proto 响应对象
func (b *SysRoleMenuItemBo) ToAPIV1() *goddessv1.SysRoleMenuItem {
	return &goddessv1.SysRoleMenuItem{
		MenuCode: b.MenuCode,
		MenuName: b.MenuName,
	}
}

// SysRoleFunctionItemBo 角色关联的功能业务对象
type SysRoleFunctionItemBo struct {
	MenuCode     string
	MenuName     string
	FunctionCode string
	FunctionName string
}

// ToAPIV1 转换为 proto 响应对象
func (b *SysRoleFunctionItemBo) ToAPIV1() *goddessv1.SysRoleFunctionItem {
	return &goddessv1.SysRoleFunctionItem{
		MenuCode:     b.MenuCode,
		MenuName:     b.MenuName,
		FunctionCode: b.FunctionCode,
		FunctionName: b.FunctionName,
	}
}

// GetSysRoleBo 角色详情业务对象（含菜单和功能权限）
type GetSysRoleBo struct {
	*SysRoleItemBo
	MenuCodes    []*SysRoleMenuItemBo
	FunctionKeys []*SysRoleFunctionItemBo
}

// ToAPIV1GetSysRoleReply 转换为 proto 响应
func (b *GetSysRoleBo) ToAPIV1GetSysRoleReply() *goddessv1.GetSysRoleReply {
	menuCodes := make([]*goddessv1.SysRoleMenuItem, 0, len(b.MenuCodes))
	for _, item := range b.MenuCodes {
		menuCodes = append(menuCodes, item.ToAPIV1())
	}
	functionKeys := make([]*goddessv1.SysRoleFunctionItem, 0, len(b.FunctionKeys))
	for _, item := range b.FunctionKeys {
		functionKeys = append(functionKeys, item.ToAPIV1())
	}
	return &goddessv1.GetSysRoleReply{
		Role:         b.SysRoleItemBo.ToAPIV1SysRoleItem(),
		MenuCodes:    menuCodes,
		FunctionKeys: functionKeys,
	}
}

// CreateSysRoleBo 新增系统角色业务对象
type CreateSysRoleBo struct {
	RoleCode      string
	RoleName      string
	RoleDataScope int32
	RoleStatus    int32
	RoleSortOrder int32
	RoleRemark    string
	Creator       int64
}

// NewCreateSysRoleBo 从 proto 请求构造新增角色业务对象
func NewCreateSysRoleBo(req *goddessv1.CreateSysRoleRequest) *CreateSysRoleBo {
	return &CreateSysRoleBo{
		RoleCode:      req.RoleCode,
		RoleName:      req.RoleName,
		RoleDataScope: req.RoleDataScope,
		RoleStatus:    req.RoleStatus,
		RoleSortOrder: req.RoleSortOrder,
		RoleRemark:    req.RoleRemark,
	}
}

// UpdateSysRoleBo 编辑系统角色业务对象（指针字段区分未传和传零值）
type UpdateSysRoleBo struct {
	RoleCode      string
	RoleName      *string
	RoleDataScope *int32
	RoleStatus    *int32
	RoleSortOrder *int32
	RoleRemark    *string
}

// NewUpdateSysRoleBo 从 proto 请求构造编辑角色业务对象
func NewUpdateSysRoleBo(req *goddessv1.UpdateSysRoleRequest) *UpdateSysRoleBo {
	return &UpdateSysRoleBo{
		RoleCode:      req.RoleCode,
		RoleName:      req.RoleName,
		RoleDataScope: req.RoleDataScope,
		RoleStatus:    req.RoleStatus,
		RoleSortOrder: req.RoleSortOrder,
		RoleRemark:    req.RoleRemark,
	}
}
