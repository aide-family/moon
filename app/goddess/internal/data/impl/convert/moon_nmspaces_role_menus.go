package convert

import (
	"github.com/aide-family/goddess/internal/biz/bo"
)

// NmspacesRoleMenuJoinRow 用于接收 JOIN 查询结果的结构体
type NmspacesRoleMenuJoinRow struct {
	MenuCode       string
	ParentMenuCode string
	MenuName       string
	MenuStatus     int32
}

// NmspacesRoleMenuJoinToBo 将 JOIN 查询结果转换为业务对象
func NmspacesRoleMenuJoinToBo(row *NmspacesRoleMenuJoinRow) *bo.NmspacesRoleMenuItemBo {
	if row == nil {
		return nil
	}
	return &bo.NmspacesRoleMenuItemBo{
		MenuCode:       row.MenuCode,
		ParentMenuCode: row.ParentMenuCode,
		MenuName:       row.MenuName,
		MenuStatus:     row.MenuStatus,
	}
}
