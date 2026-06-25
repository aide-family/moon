package bo

import (
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// RoleMenuItemBo 角色菜单业务对象
type RoleMenuItemBo struct {
	MenuCode       string
	ParentMenuCode string
	MenuName       string
	MenuStatus     int32
}

// ListRoleMenuBo 角色菜单列表查询业务对象
type ListRoleMenuBo struct {
	RoleCode   string
	AsTree     bool
	MenuStatus *int32
}

// NewListRoleMenuBo 从 proto 请求构造
func NewListRoleMenuBo(req *goddessv1.ListRoleMenuRequest) *ListRoleMenuBo {
	return &ListRoleMenuBo{
		RoleCode:   req.RoleCode,
		AsTree:     req.AsTree,
		MenuStatus: req.MenuStatus,
	}
}

// GrantRoleMenusBo 授予角色菜单业务对象
type GrantRoleMenusBo struct {
	RoleCode         string
	MenuCodes        []string
	AutoAddAncestors bool
	Creator          int64
}

// NewGrantRoleMenusBo 从 proto 请求构造
func NewGrantRoleMenusBo(req *goddessv1.GrantRoleMenusRequest) *GrantRoleMenusBo {
	return &GrantRoleMenusBo{
		RoleCode:         req.RoleCode,
		MenuCodes:        req.MenuCodes,
		AutoAddAncestors: req.AutoAddAncestors,
	}
}

// RevokeRoleMenusBo 撤销角色菜单业务对象
type RevokeRoleMenusBo struct {
	RoleCode           string
	MenuCodes          []string
	CascadeDescendants bool
}

// NewRevokeRoleMenusBo 从 proto 请求构造
func NewRevokeRoleMenusBo(req *goddessv1.RevokeRoleMenusRequest) *RevokeRoleMenusBo {
	return &RevokeRoleMenusBo{
		RoleCode:           req.RoleCode,
		MenuCodes:          req.MenuCodes,
		CascadeDescendants: req.CascadeDescendants,
	}
}

// GrantResultBo 授予结果业务对象
type GrantResultBo struct {
	GrantedCount   int32
	AutoAddedCodes []string
}

// RevokeResultBo 撤销结果业务对象
type RevokeResultBo struct {
	RevokedCount        int32
	CascadeRevokedCodes []string
}
