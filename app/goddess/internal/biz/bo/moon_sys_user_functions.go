package bo

import (
	"github.com/bwmarrin/snowflake"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// UserFunctionItemBo 用户直授功能业务对象
type UserFunctionItemBo struct {
	MenuCode          string
	FunctionCode      string
	FunctionName      string
	FunctionSortOrder int32
	FunctionStatus    int32
	Creator           int64
	CreatedAt         string
}

// ListUserFunctionBo 用户直授功能查询业务对象
type ListUserFunctionBo struct {
	UserUID       snowflake.ID
	MenuCode      string
	PageRequestBo *PageRequestBo
}

// NewListUserFunctionBo 从 proto 请求构造
func NewListUserFunctionBo(req *goddessv1.ListUserFunctionRequest) *ListUserFunctionBo {
	return &ListUserFunctionBo{
		UserUID:       snowflake.ID(req.UserUID),
		MenuCode:      req.MenuCode,
		PageRequestBo: NewPageRequestBo(req.Page, req.PageSize),
	}
}

// UserFunctionKey 用户功能键
type UserFunctionKey struct {
	MenuCode     string
	FunctionCode string
}

// GrantUserFunctionsBo 直授用户功能业务对象
type GrantUserFunctionsBo struct {
	UserUID       snowflake.ID
	FunctionKeys  []UserFunctionKey
	AutoGrantMenu bool
	Creator       snowflake.ID
}

// NewGrantUserFunctionsBo 从 proto 请求构造
func NewGrantUserFunctionsBo(req *goddessv1.GrantUserFunctionsRequest) *GrantUserFunctionsBo {
	keys := make([]UserFunctionKey, 0, len(req.FunctionKeys))
	for _, fk := range req.FunctionKeys {
		keys = append(keys, UserFunctionKey{
			MenuCode:     fk.MenuCode,
			FunctionCode: fk.FunctionCode,
		})
	}
	return &GrantUserFunctionsBo{
		UserUID:       snowflake.ID(req.UserUID),
		FunctionKeys:  keys,
		AutoGrantMenu: req.AutoGrantMenu,
	}
}

// RevokeUserFunctionsBo 撤销用户直授功能业务对象
type RevokeUserFunctionsBo struct {
	UserUID        snowflake.ID
	FunctionKeys   []UserFunctionKey
	AutoRevokeMenu bool
}

// NewRevokeUserFunctionsBo 从 proto 请求构造
func NewRevokeUserFunctionsBo(req *goddessv1.RevokeUserFunctionsRequest) *RevokeUserFunctionsBo {
	keys := make([]UserFunctionKey, 0, len(req.FunctionKeys))
	for _, fk := range req.FunctionKeys {
		keys = append(keys, UserFunctionKey{
			MenuCode:     fk.MenuCode,
			FunctionCode: fk.FunctionCode,
		})
	}
	return &RevokeUserFunctionsBo{
		UserUID:        snowflake.ID(req.UserUID),
		FunctionKeys:   keys,
		AutoRevokeMenu: req.AutoRevokeMenu,
	}
}

// GrantUserFunctionResultBo 授予用户功能结果
type GrantUserFunctionResultBo struct {
	GrantedCount     int32
	AutoGrantedMenus []string
}

// RevokeUserFunctionResultBo 撤销用户功能结果
type RevokeUserFunctionResultBo struct {
	RevokedCount     int32
	AutoRevokedMenus []string
}
