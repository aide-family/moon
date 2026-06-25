package bo

import (
	"time"

	"github.com/aide-family/magicbox/timex"
	"github.com/bwmarrin/snowflake"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// NmspacesMemberFunctionItemBo 租户成员直授功能项业务对象
type NmspacesMemberFunctionItemBo struct {
	MenuCode          string
	FunctionCode      string
	FunctionName      string
	FunctionSortOrder int32
	FunctionStatus    int32
	Creator           int64
	CreatedAt         time.Time
}

// ToAPIV1NmspacesMemberFunctionItem 转换为 proto 对象
func (b *NmspacesMemberFunctionItemBo) ToAPIV1NmspacesMemberFunctionItem() *goddessv1.NmspacesMemberFunctionItem {
	return &goddessv1.NmspacesMemberFunctionItem{
		MenuCode:          b.MenuCode,
		FunctionCode:      b.FunctionCode,
		FunctionName:      b.FunctionName,
		FunctionSortOrder: b.FunctionSortOrder,
		FunctionStatus:    b.FunctionStatus,
		Creator:           b.Creator,
		CreatedAt:         timex.FormatTime(&b.CreatedAt),
	}
}

// ListNmspacesMemberFunctionBo 查询租户成员直授功能列表业务对象
type ListNmspacesMemberFunctionBo struct {
	*PageRequestBo
	NamespaceUID snowflake.ID
	MemberUID    snowflake.ID
	MenuCode     string
}

// NewListNmspacesMemberFunctionBo 从 proto 请求构造
func NewListNmspacesMemberFunctionBo(namespaceUID snowflake.ID, req *goddessv1.ListNmspacesMemberFunctionRequest) *ListNmspacesMemberFunctionBo {
	return &ListNmspacesMemberFunctionBo{
		PageRequestBo: NewPageRequestBo(req.Page, req.PageSize),
		NamespaceUID:  namespaceUID,
		MemberUID:     parseSnowflakeID(req.MemberUID),
		MenuCode:      req.MenuCode,
	}
}

// GrantNmspacesMemberFunctionsBo 直授租户功能业务对象
type GrantNmspacesMemberFunctionsBo struct {
	NamespaceUID  snowflake.ID
	MemberUID     snowflake.ID
	FunctionKeys  []NmspacesFunctionKey
	AutoGrantMenu bool
	Creator       snowflake.ID
}

// NewGrantNmspacesMemberFunctionsBo 从 proto 请求构造
func NewGrantNmspacesMemberFunctionsBo(namespaceUID snowflake.ID, creator snowflake.ID, req *goddessv1.GrantNmspacesMemberFunctionsRequest) *GrantNmspacesMemberFunctionsBo {
	keys := make([]NmspacesFunctionKey, 0, len(req.FunctionKeys))
	for _, k := range req.FunctionKeys {
		keys = append(keys, NmspacesFunctionKey{MenuCode: k.MenuCode, FunctionCode: k.FunctionCode})
	}
	return &GrantNmspacesMemberFunctionsBo{
		NamespaceUID:  namespaceUID,
		MemberUID:     parseSnowflakeID(req.MemberUID),
		FunctionKeys:  keys,
		AutoGrantMenu: req.AutoGrantMenu,
		Creator:       creator,
	}
}

// RevokeNmspacesMemberFunctionsBo 撤销直授租户功能业务对象
type RevokeNmspacesMemberFunctionsBo struct {
	NamespaceUID   snowflake.ID
	MemberUID      snowflake.ID
	FunctionKeys   []NmspacesFunctionKey
	AutoRevokeMenu bool
}

// NewRevokeNmspacesMemberFunctionsBo 从 proto 请求构造
func NewRevokeNmspacesMemberFunctionsBo(namespaceUID snowflake.ID, req *goddessv1.RevokeNmspacesMemberFunctionsRequest) *RevokeNmspacesMemberFunctionsBo {
	keys := make([]NmspacesFunctionKey, 0, len(req.FunctionKeys))
	for _, k := range req.FunctionKeys {
		keys = append(keys, NmspacesFunctionKey{MenuCode: k.MenuCode, FunctionCode: k.FunctionCode})
	}
	return &RevokeNmspacesMemberFunctionsBo{
		NamespaceUID:   namespaceUID,
		MemberUID:      parseSnowflakeID(req.MemberUID),
		FunctionKeys:   keys,
		AutoRevokeMenu: req.AutoRevokeMenu,
	}
}
