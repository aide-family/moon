package bo

import (
	"strconv"
	"time"

	"github.com/bwmarrin/snowflake"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// parseSnowflakeID 将字符串转为 snowflake.ID
func parseSnowflakeID(s string) snowflake.ID {
	id, _ := strconv.ParseInt(s, 10, 64)
	return snowflake.ID(id)
}

// SysUserRoleItemBo 用户角色业务对象
type SysUserRoleItemBo struct {
	RoleCode      string
	RoleName      string
	RoleDataScope int32
	RoleStatus    int32
	CreatedAt     time.Time
	Grantor       int64
}

// ListSysUserRoleBo 用户角色列表查询业务对象
type ListSysUserRoleBo struct {
	*PageRequestBo
	UserUID snowflake.ID
}

// NewListSysUserRoleBo 从 proto 请求构造查询业务对象
func NewListSysUserRoleBo(req *goddessv1.ListSysUserRoleRequest) *ListSysUserRoleBo {
	return &ListSysUserRoleBo{
		PageRequestBo: NewPageRequestBo(req.Page, req.PageSize),
		UserUID:       parseSnowflakeID(req.UserUid),
	}
}

// BindSysUserRolesBo 绑定用户角色业务对象
type BindSysUserRolesBo struct {
	UserUID   snowflake.ID
	RoleCodes []string
	Creator   snowflake.ID
}

// NewBindSysUserRolesBo 从 proto 请求构造绑定业务对象
func NewBindSysUserRolesBo(req *goddessv1.BindSysUserRolesRequest) *BindSysUserRolesBo {
	return &BindSysUserRolesBo{
		UserUID:   parseSnowflakeID(req.UserUid),
		RoleCodes: req.RoleCodes,
	}
}

// UnbindSysUserRolesBo 解绑用户角色业务对象
type UnbindSysUserRolesBo struct {
	UserUID   snowflake.ID
	RoleCodes []string
}

// NewUnbindSysUserRolesBo 从 proto 请求构造解绑业务对象
func NewUnbindSysUserRolesBo(req *goddessv1.UnbindSysUserRolesRequest) *UnbindSysUserRolesBo {
	return &UnbindSysUserRolesBo{
		UserUID:   parseSnowflakeID(req.UserUid),
		RoleCodes: req.RoleCodes,
	}
}
