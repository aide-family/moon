package convert

import (
	"github.com/aide-family/goddess/internal/biz/bo"
)

// NmspacesMemberMenuJoinRow 用于接收 JOIN 查询结果的结构体
type NmspacesMemberMenuJoinRow struct {
	MenuCode       string
	ParentMenuCode string
	MenuName       string
	MenuStatus     int32
}

// NmspacesMemberMenuJoinToBo 将 JOIN 查询结果转换为业务对象
func NmspacesMemberMenuJoinToBo(row *NmspacesMemberMenuJoinRow) *bo.NmspacesMemberMenuItemBo {
	if row == nil {
		return nil
	}
	return &bo.NmspacesMemberMenuItemBo{
		MenuCode:       row.MenuCode,
		ParentMenuCode: row.ParentMenuCode,
		MenuName:       row.MenuName,
		MenuStatus:     row.MenuStatus,
	}
}
