package bo

import (
	"time"

	"github.com/aide-family/magicbox/timex"
	"github.com/bwmarrin/snowflake"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// NmspacesRoleFunctionItemBo 租户角色功能项业务对象
type NmspacesRoleFunctionItemBo struct {
	MenuCode          string
	FunctionCode      string
	FunctionName      string
	FunctionSortOrder int32
	FunctionStatus    int32
	Creator           int64
	CreatedAt         time.Time
}

// ToAPIV1NmspacesRoleFunctionItem 转换为 proto 对象
func (b *NmspacesRoleFunctionItemBo) ToAPIV1NmspacesRoleFunctionItem() *goddessv1.NmspacesRoleFunctionItem {
	return &goddessv1.NmspacesRoleFunctionItem{
		MenuCode:          b.MenuCode,
		FunctionCode:      b.FunctionCode,
		FunctionName:      b.FunctionName,
		FunctionSortOrder: b.FunctionSortOrder,
		FunctionStatus:    b.FunctionStatus,
		Creator:           b.Creator,
		CreatedAt:         timex.FormatTime(&b.CreatedAt),
	}
}

// NmspacesFunctionKey 租户功能键
type NmspacesFunctionKey struct {
	MenuCode     string
	FunctionCode string
}

// ListNmspacesRoleFunctionBo 查询租户角色功能列表业务对象
type ListNmspacesRoleFunctionBo struct {
	*PageRequestBo
	NamespaceUID snowflake.ID
	RoleCode     string
	MenuCode     string
}

// NewListNmspacesRoleFunctionBo 从 proto 请求构造
func NewListNmspacesRoleFunctionBo(namespaceUID snowflake.ID, req *goddessv1.ListNmspacesRoleFunctionRequest) *ListNmspacesRoleFunctionBo {
	return &ListNmspacesRoleFunctionBo{
		PageRequestBo: NewPageRequestBo(req.Page, req.PageSize),
		NamespaceUID:  namespaceUID,
		RoleCode:      req.RoleCode,
		MenuCode:      req.MenuCode,
	}
}

// GrantNmspacesRoleFunctionsBo 授予租户角色功能业务对象
type GrantNmspacesRoleFunctionsBo struct {
	NamespaceUID  snowflake.ID
	RoleCode      string
	FunctionKeys  []NmspacesFunctionKey
	AutoGrantMenu bool
	Creator       snowflake.ID
}

// NewGrantNmspacesRoleFunctionsBo 从 proto 请求构造
func NewGrantNmspacesRoleFunctionsBo(namespaceUID snowflake.ID, creator snowflake.ID, req *goddessv1.GrantNmspacesRoleFunctionsRequest) *GrantNmspacesRoleFunctionsBo {
	keys := make([]NmspacesFunctionKey, 0, len(req.FunctionKeys))
	for _, k := range req.FunctionKeys {
		keys = append(keys, NmspacesFunctionKey{MenuCode: k.MenuCode, FunctionCode: k.FunctionCode})
	}
	return &GrantNmspacesRoleFunctionsBo{
		NamespaceUID:  namespaceUID,
		RoleCode:      req.RoleCode,
		FunctionKeys:  keys,
		AutoGrantMenu: req.AutoGrantMenu,
		Creator:       creator,
	}
}

// RevokeNmspacesRoleFunctionsBo 撤销租户角色功能业务对象
type RevokeNmspacesRoleFunctionsBo struct {
	NamespaceUID   snowflake.ID
	RoleCode       string
	FunctionKeys   []NmspacesFunctionKey
	AutoRevokeMenu bool
}

// NewRevokeNmspacesRoleFunctionsBo 从 proto 请求构造
func NewRevokeNmspacesRoleFunctionsBo(namespaceUID snowflake.ID, req *goddessv1.RevokeNmspacesRoleFunctionsRequest) *RevokeNmspacesRoleFunctionsBo {
	keys := make([]NmspacesFunctionKey, 0, len(req.FunctionKeys))
	for _, k := range req.FunctionKeys {
		keys = append(keys, NmspacesFunctionKey{MenuCode: k.MenuCode, FunctionCode: k.FunctionCode})
	}
	return &RevokeNmspacesRoleFunctionsBo{
		NamespaceUID:   namespaceUID,
		RoleCode:       req.RoleCode,
		FunctionKeys:   keys,
		AutoRevokeMenu: req.AutoRevokeMenu,
	}
}
