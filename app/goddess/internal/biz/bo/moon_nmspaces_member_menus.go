package bo

import (
	"github.com/bwmarrin/snowflake"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// NmspacesMemberMenuItemBo 租户成员直授菜单项业务对象
type NmspacesMemberMenuItemBo struct {
	MenuCode       string
	ParentMenuCode string
	MenuName       string
	MenuStatus     int32
}

// ListNmspacesMemberMenuBo 查询租户成员直授菜单业务对象
type ListNmspacesMemberMenuBo struct {
	NamespaceUID snowflake.ID
	MemberUID    snowflake.ID
	AsTree       bool
	MenuStatus   *int32
}

// NewListNmspacesMemberMenuBo 从 proto 请求构造
func NewListNmspacesMemberMenuBo(namespaceUID snowflake.ID, req *goddessv1.ListNmspacesMemberMenuRequest) *ListNmspacesMemberMenuBo {
	return &ListNmspacesMemberMenuBo{
		NamespaceUID: namespaceUID,
		MemberUID:    parseSnowflakeID(req.MemberUID),
		AsTree:       req.AsTree,
		MenuStatus:   req.MenuStatus,
	}
}

// GrantNmspacesMemberMenusBo 直授租户菜单业务对象
type GrantNmspacesMemberMenusBo struct {
	NamespaceUID     snowflake.ID
	MemberUID        snowflake.ID
	MenuCodes        []string
	AutoAddAncestors bool
	Creator          snowflake.ID
}

// NewGrantNmspacesMemberMenusBo 从 proto 请求构造
func NewGrantNmspacesMemberMenusBo(namespaceUID snowflake.ID, creator snowflake.ID, req *goddessv1.GrantNmspacesMemberMenusRequest) *GrantNmspacesMemberMenusBo {
	return &GrantNmspacesMemberMenusBo{
		NamespaceUID:     namespaceUID,
		MemberUID:        parseSnowflakeID(req.MemberUID),
		MenuCodes:        req.MenuCodes,
		AutoAddAncestors: req.AutoAddAncestors,
		Creator:          creator,
	}
}

// RevokeNmspacesMemberMenusBo 撤销直授租户菜单业务对象
type RevokeNmspacesMemberMenusBo struct {
	NamespaceUID       snowflake.ID
	MemberUID          snowflake.ID
	MenuCodes          []string
	CascadeDescendants bool
}

// NewRevokeNmspacesMemberMenusBo 从 proto 请求构造
func NewRevokeNmspacesMemberMenusBo(namespaceUID snowflake.ID, req *goddessv1.RevokeNmspacesMemberMenusRequest) *RevokeNmspacesMemberMenusBo {
	return &RevokeNmspacesMemberMenusBo{
		NamespaceUID:       namespaceUID,
		MemberUID:          parseSnowflakeID(req.MemberUID),
		MenuCodes:          req.MenuCodes,
		CascadeDescendants: req.CascadeDescendants,
	}
}
