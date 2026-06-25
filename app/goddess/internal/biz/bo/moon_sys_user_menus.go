package bo

import (
	"github.com/bwmarrin/snowflake"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// UserMenuItemBo 用户直授菜单业务对象
type UserMenuItemBo struct {
	MenuCode       string
	ParentMenuCode string
	MenuName       string
	MenuStatus     int32
}

// ListUserMenuBo 用户直授菜单查询业务对象
type ListUserMenuBo struct {
	UserUID    snowflake.ID
	AsTree     bool
	MenuStatus *int32
}

// NewListUserMenuBo 从 proto 请求构造
func NewListUserMenuBo(req *goddessv1.ListUserMenuRequest) *ListUserMenuBo {
	return &ListUserMenuBo{
		UserUID:    snowflake.ID(req.UserUID),
		AsTree:     req.AsTree,
		MenuStatus: req.MenuStatus,
	}
}

// GrantUserMenusBo 直授用户菜单业务对象
type GrantUserMenusBo struct {
	UserUID          snowflake.ID
	MenuCodes        []string
	AutoAddAncestors bool
	Creator          snowflake.ID
}

// NewGrantUserMenusBo 从 proto 请求构造
func NewGrantUserMenusBo(req *goddessv1.GrantUserMenusRequest) *GrantUserMenusBo {
	return &GrantUserMenusBo{
		UserUID:          snowflake.ID(req.UserUID),
		MenuCodes:        req.MenuCodes,
		AutoAddAncestors: req.AutoAddAncestors,
	}
}

// RevokeUserMenusBo 撤销用户直授菜单业务对象
type RevokeUserMenusBo struct {
	UserUID            snowflake.ID
	MenuCodes          []string
	CascadeDescendants bool
}

// NewRevokeUserMenusBo 从 proto 请求构造
func NewRevokeUserMenusBo(req *goddessv1.RevokeUserMenusRequest) *RevokeUserMenusBo {
	return &RevokeUserMenusBo{
		UserUID:            snowflake.ID(req.UserUID),
		MenuCodes:          req.MenuCodes,
		CascadeDescendants: req.CascadeDescendants,
	}
}
