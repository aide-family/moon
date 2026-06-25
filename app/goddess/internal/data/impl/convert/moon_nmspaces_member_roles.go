package convert

import (
	"github.com/aide-family/goddess/internal/biz/bo"
)

// NmspacesMemberRoleJoinRow 用于接收 JOIN 查询结果的结构体
type NmspacesMemberRoleJoinRow struct {
	RoleCode      string
	RoleName      string
	RoleDataScope int32
	RoleStatus    int32
	GrantedAt     string
	Grantor       int64
}

// NmspacesMemberRoleJoinToBo 将 JOIN 查询结果转换为业务对象
func NmspacesMemberRoleJoinToBo(row *NmspacesMemberRoleJoinRow) *bo.NmspacesMemberRoleItemBo {
	if row == nil {
		return nil
	}
	return &bo.NmspacesMemberRoleItemBo{
		RoleCode:      row.RoleCode,
		RoleName:      row.RoleName,
		RoleDataScope: row.RoleDataScope,
		RoleStatus:    row.RoleStatus,
		Grantor:       row.Grantor,
	}
}
