package bo

import (
	"github.com/bwmarrin/snowflake"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// NmspacesRoleMenuItemBo 租户角色菜单项业务对象
type NmspacesRoleMenuItemBo struct {
	MenuCode       string
	ParentMenuCode string
	MenuName       string
	MenuStatus     int32
}

// ListNmspacesRoleMenuBo 查询租户角色菜单列表业务对象
type ListNmspacesRoleMenuBo struct {
	NamespaceUID snowflake.ID
	RoleCode     string
	AsTree       bool
	MenuStatus   *int32
}

// NewListNmspacesRoleMenuBo 从 proto 请求构造
func NewListNmspacesRoleMenuBo(namespaceUID snowflake.ID, req *goddessv1.ListNmspacesRoleMenuRequest) *ListNmspacesRoleMenuBo {
	return &ListNmspacesRoleMenuBo{
		NamespaceUID: namespaceUID,
		RoleCode:     req.RoleCode,
		AsTree:       req.AsTree,
		MenuStatus:   req.MenuStatus,
	}
}

// GrantNmspacesRoleMenusBo 授予租户角色菜单业务对象
type GrantNmspacesRoleMenusBo struct {
	NamespaceUID     snowflake.ID
	RoleCode         string
	MenuCodes        []string
	AutoAddAncestors bool
	Creator          snowflake.ID
}

// NewGrantNmspacesRoleMenusBo 从 proto 请求构造
func NewGrantNmspacesRoleMenusBo(namespaceUID snowflake.ID, creator snowflake.ID, req *goddessv1.GrantNmspacesRoleMenusRequest) *GrantNmspacesRoleMenusBo {
	return &GrantNmspacesRoleMenusBo{
		NamespaceUID:     namespaceUID,
		RoleCode:         req.RoleCode,
		MenuCodes:        req.MenuCodes,
		AutoAddAncestors: req.AutoAddAncestors,
		Creator:          creator,
	}
}

// RevokeNmspacesRoleMenusBo 撤销租户角色菜单业务对象
type RevokeNmspacesRoleMenusBo struct {
	NamespaceUID       snowflake.ID
	RoleCode           string
	MenuCodes          []string
	CascadeDescendants bool
}

// NewRevokeNmspacesRoleMenusBo 从 proto 请求构造
func NewRevokeNmspacesRoleMenusBo(namespaceUID snowflake.ID, req *goddessv1.RevokeNmspacesRoleMenusRequest) *RevokeNmspacesRoleMenusBo {
	return &RevokeNmspacesRoleMenusBo{
		NamespaceUID:       namespaceUID,
		RoleCode:           req.RoleCode,
		MenuCodes:          req.MenuCodes,
		CascadeDescendants: req.CascadeDescendants,
	}
}
