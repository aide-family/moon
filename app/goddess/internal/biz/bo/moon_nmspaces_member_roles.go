package bo

import (
	"time"

	"github.com/aide-family/magicbox/timex"
	"github.com/bwmarrin/snowflake"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// NmspacesMemberRoleItemBo 租户成员角色业务对象
type NmspacesMemberRoleItemBo struct {
	RoleCode      string
	RoleName      string
	RoleDataScope int32
	RoleStatus    int32
	GrantedAt     time.Time
	Grantor       int64
}

// ToAPIV1NmspacesMemberRoleItem 转换为 proto 对象
func (b *NmspacesMemberRoleItemBo) ToAPIV1NmspacesMemberRoleItem() *goddessv1.NmspacesMemberRoleItem {
	return &goddessv1.NmspacesMemberRoleItem{
		RoleCode:      b.RoleCode,
		RoleName:      b.RoleName,
		RoleDataScope: b.RoleDataScope,
		RoleStatus:    b.RoleStatus,
		GrantedAt:     timex.FormatTime(&b.GrantedAt),
		Grantor:       b.Grantor,
	}
}

// ListNmspacesMemberRoleBo 查询租户成员角色列表业务对象
type ListNmspacesMemberRoleBo struct {
	*PageRequestBo
	NamespaceUID snowflake.ID
	MemberUID    snowflake.ID
}

// NewListNmspacesMemberRoleBo 从 proto 请求构造
func NewListNmspacesMemberRoleBo(namespaceUID snowflake.ID, req *goddessv1.ListNmspacesMemberRoleRequest) *ListNmspacesMemberRoleBo {
	return &ListNmspacesMemberRoleBo{
		PageRequestBo: NewPageRequestBo(req.Page, req.PageSize),
		NamespaceUID:  namespaceUID,
		MemberUID:     parseSnowflakeID(req.MemberUID),
	}
}

// BindNmspacesMemberRolesBo 绑定租户成员角色业务对象
type BindNmspacesMemberRolesBo struct {
	NamespaceUID snowflake.ID
	MemberUID    snowflake.ID
	RoleCodes    []string
	Grantor      snowflake.ID
}

// NewBindNmspacesMemberRolesBo 从 proto 请求构造
func NewBindNmspacesMemberRolesBo(namespaceUID snowflake.ID, grantor snowflake.ID, req *goddessv1.BindNmspacesMemberRolesRequest) *BindNmspacesMemberRolesBo {
	return &BindNmspacesMemberRolesBo{
		NamespaceUID: namespaceUID,
		MemberUID:    parseSnowflakeID(req.MemberUID),
		RoleCodes:    req.RoleCodes,
		Grantor:      grantor,
	}
}

// UnbindNmspacesMemberRolesBo 解绑租户成员角色业务对象
type UnbindNmspacesMemberRolesBo struct {
	NamespaceUID snowflake.ID
	MemberUID    snowflake.ID
	RoleCodes    []string
}

// NewUnbindNmspacesMemberRolesBo 从 proto 请求构造
func NewUnbindNmspacesMemberRolesBo(namespaceUID snowflake.ID, req *goddessv1.UnbindNmspacesMemberRolesRequest) *UnbindNmspacesMemberRolesBo {
	return &UnbindNmspacesMemberRolesBo{
		NamespaceUID: namespaceUID,
		MemberUID:    parseSnowflakeID(req.MemberUID),
		RoleCodes:    req.RoleCodes,
	}
}
