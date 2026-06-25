package bo

import (
	"github.com/bwmarrin/snowflake"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// ListNmspaceFinalMenusBo 查询租户成员最终菜单集合请求 BO
type ListNmspaceFinalMenusBo struct {
	NamespaceUID snowflake.ID
	MemberUID    snowflake.ID
	AsTree       bool
}

// NewListNmspaceFinalMenusBo 从 proto 请求构造
func NewListNmspaceFinalMenusBo(namespaceUID snowflake.ID, req *goddessv1.ListNmspaceFinalMenusRequest) *ListNmspaceFinalMenusBo {
	return &ListNmspaceFinalMenusBo{
		NamespaceUID: namespaceUID,
		MemberUID:    parseSnowflakeID(req.MemberUID),
		AsTree:       req.AsTree,
	}
}

// NmspaceFinalMenuItemBo 租户成员最终菜单项 BO
type NmspaceFinalMenuItemBo struct {
	MenuCode       string
	ParentMenuCode string
	MenuName       string
	MenuStatus     int32
}

// ListNmspaceFinalMenusResultBo 租户成员最终菜单集合结果
type ListNmspaceFinalMenusResultBo struct {
	Items     []*NmspaceFinalMenuItemBo
	MenuCodes []string
}

// ListNmspaceFinalFunctionsBo 查询租户成员最终功能集合请求 BO
type ListNmspaceFinalFunctionsBo struct {
	NamespaceUID snowflake.ID
	MemberUID    snowflake.ID
	MenuCode     string
}

// NewListNmspaceFinalFunctionsBo 从 proto 请求构造
func NewListNmspaceFinalFunctionsBo(namespaceUID snowflake.ID, req *goddessv1.ListNmspaceFinalFunctionsRequest) *ListNmspaceFinalFunctionsBo {
	return &ListNmspaceFinalFunctionsBo{
		NamespaceUID: namespaceUID,
		MemberUID:    parseSnowflakeID(req.MemberUID),
		MenuCode:     req.MenuCode,
	}
}

// NmspaceFinalFunctionItemBo 租户成员最终功能项 BO
type NmspaceFinalFunctionItemBo struct {
	MenuCode     string
	FunctionCode string
	FunctionName string
}

// ListNmspaceFinalFunctionsResultBo 租户成员最终功能集合结果
type ListNmspaceFinalFunctionsResultBo struct {
	FunctionKeys []*NmspaceFinalFunctionItemBo
	FunctionMap  map[string][]string
}
