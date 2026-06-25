package bo

import (
	"github.com/bwmarrin/snowflake"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// ListUserFinalMenusBo 查询用户最终菜单集合请求 BO
type ListUserFinalMenusBo struct {
	UserUID snowflake.ID
	AsTree  bool
}

// NewListUserFinalMenusBo 从 proto 请求构造
func NewListUserFinalMenusBo(req *goddessv1.ListUserFinalMenusRequest) *ListUserFinalMenusBo {
	return &ListUserFinalMenusBo{
		UserUID: snowflake.ID(req.UserUID),
		AsTree:  req.AsTree,
	}
}

// UserFinalMenuItemBo 用户最终菜单项 BO
type UserFinalMenuItemBo struct {
	MenuCode       string
	ParentMenuCode string
	MenuName       string
	MenuStatus     int32
}

// ListUserFinalMenusResultBo 用户最终菜单集合结果
type ListUserFinalMenusResultBo struct {
	Items     []*UserFinalMenuItemBo
	MenuCodes []string
}

// ListUserFinalFunctionsBo 查询用户最终功能集合请求 BO
type ListUserFinalFunctionsBo struct {
	UserUID  snowflake.ID
	MenuCode string
}

// NewListUserFinalFunctionsBo 从 proto 请求构造
func NewListUserFinalFunctionsBo(req *goddessv1.ListUserFinalFunctionsRequest) *ListUserFinalFunctionsBo {
	return &ListUserFinalFunctionsBo{
		UserUID:  snowflake.ID(req.UserUID),
		MenuCode: req.MenuCode,
	}
}

// UserFinalFunctionItemBo 用户最终功能项 BO
type UserFinalFunctionItemBo struct {
	MenuCode     string
	FunctionCode string
	FunctionName string
}

// ListUserFinalFunctionsResultBo 用户最终功能集合结果
type ListUserFinalFunctionsResultBo struct {
	FunctionKeys []*UserFinalFunctionItemBo
	FunctionMap  map[string][]string
}
