package convert

import (
	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/data/impl/do"
)

// SysRoleToBo 将数据库模型转换为业务对象
func SysRoleToBo(role *do.SysRole) *bo.SysRoleItemBo {
	if role == nil {
		return nil
	}
	return &bo.SysRoleItemBo{
		RoleCode:      role.RoleCode,
		RoleName:      role.RoleName,
		RoleDataScope: role.RoleDataScope,
		RoleStatus:    role.RoleStatus,
		RoleSortOrder: role.RoleSortOrder,
		RoleRemark:    role.RoleRemark,
		Creator:       role.Creator,
		CreatedAt:     role.CreatedAt,
		UpdatedAt:     role.UpdatedAt,
	}
}
