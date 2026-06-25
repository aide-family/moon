package bo

import (
	"time"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// FunctionKey 功能键
type FunctionKey struct {
	MenuCode     string
	FunctionCode string
}

// RoleFunctionItemBo 角色功能业务对象
type RoleFunctionItemBo struct {
	MenuCode          string
	FunctionCode      string
	FunctionName      string
	FunctionSortOrder int32
	FunctionStatus    int32
	Creator           int64
	CreatedAt         time.Time
}

// ListRoleFunctionBo 角色功能列表查询业务对象
type ListRoleFunctionBo struct {
	RoleCode      string
	MenuCode      string
	PageRequestBo *PageRequestBo
}

// NewListRoleFunctionBo 从 proto 请求构造
func NewListRoleFunctionBo(req *goddessv1.ListRoleFunctionRequest) *ListRoleFunctionBo {
	return &ListRoleFunctionBo{
		RoleCode:      req.RoleCode,
		MenuCode:      req.MenuCode,
		PageRequestBo: NewPageRequestBo(req.Page, req.PageSize),
	}
}

// GrantRoleFunctionsBo 授予角色功能业务对象
type GrantRoleFunctionsBo struct {
	RoleCode      string
	FunctionKeys  []FunctionKey
	AutoGrantMenu bool
	Creator       int64
}

// NewGrantRoleFunctionsBo 从 proto 请求构造
func NewGrantRoleFunctionsBo(req *goddessv1.GrantRoleFunctionsRequest) *GrantRoleFunctionsBo {
	keys := make([]FunctionKey, 0, len(req.FunctionKeys))
	for _, fk := range req.FunctionKeys {
		keys = append(keys, FunctionKey{
			MenuCode:     fk.MenuCode,
			FunctionCode: fk.FunctionCode,
		})
	}
	return &GrantRoleFunctionsBo{
		RoleCode:      req.RoleCode,
		FunctionKeys:  keys,
		AutoGrantMenu: req.AutoGrantMenu,
	}
}

// RevokeRoleFunctionsBo 撤销角色功能业务对象
type RevokeRoleFunctionsBo struct {
	RoleCode       string
	FunctionKeys   []FunctionKey
	AutoRevokeMenu bool
}

// NewRevokeRoleFunctionsBo 从 proto 请求构造
func NewRevokeRoleFunctionsBo(req *goddessv1.RevokeRoleFunctionsRequest) *RevokeRoleFunctionsBo {
	keys := make([]FunctionKey, 0, len(req.FunctionKeys))
	for _, fk := range req.FunctionKeys {
		keys = append(keys, FunctionKey{
			MenuCode:     fk.MenuCode,
			FunctionCode: fk.FunctionCode,
		})
	}
	return &RevokeRoleFunctionsBo{
		RoleCode:       req.RoleCode,
		FunctionKeys:   keys,
		AutoRevokeMenu: req.AutoRevokeMenu,
	}
}

// GrantFunctionResultBo 授予功能结果业务对象
type GrantFunctionResultBo struct {
	GrantedCount     int32
	AutoGrantedMenus []string
}

// RevokeFunctionResultBo 撤销功能结果业务对象
type RevokeFunctionResultBo struct {
	RevokedCount     int32
	AutoRevokedMenus []string
}
