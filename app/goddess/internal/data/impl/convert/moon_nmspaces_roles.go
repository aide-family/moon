package convert

import (
	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/data/impl/do"
)

// NmspacesRoleToBo 将数据库模型转换为业务对象
func NmspacesRoleToBo(role *do.NmspacesRole) *bo.NmspacesRoleItemBo {
	if role == nil {
		return nil
	}
	return &bo.NmspacesRoleItemBo{
		RoleCode:      role.RoleCode,
		RoleName:      role.RoleName,
		RoleDataScope: role.RoleDataScope,
		RoleStatus:    role.RoleStatus,
		RoleSortOrder: role.RoleSortOrder,
		RoleRemark:    role.RoleRemark,
		Creator:       role.Creator.Int64(),
		CreatedAt:     role.CreatedAt,
		UpdatedAt:     role.UpdatedAt,
	}
}
