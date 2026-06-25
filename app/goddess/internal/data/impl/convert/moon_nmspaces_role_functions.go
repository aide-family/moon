package convert

import (
	"github.com/aide-family/goddess/internal/biz/bo"
)

// NmspacesRoleFunctionJoinRow 用于接收 JOIN 查询结果的结构体
type NmspacesRoleFunctionJoinRow struct {
	MenuCode          string
	FunctionCode      string
	FunctionName      string
	FunctionSortOrder int32
	FunctionStatus    int32
	Creator           int64
	CreatedAt         string
}

// NmspacesRoleFunctionJoinToBo 将 JOIN 查询结果转换为业务对象
func NmspacesRoleFunctionJoinToBo(row *NmspacesRoleFunctionJoinRow) *bo.NmspacesRoleFunctionItemBo {
	if row == nil {
		return nil
	}
	return &bo.NmspacesRoleFunctionItemBo{
		MenuCode:          row.MenuCode,
		FunctionCode:      row.FunctionCode,
		FunctionName:      row.FunctionName,
		FunctionSortOrder: row.FunctionSortOrder,
		FunctionStatus:    row.FunctionStatus,
		Creator:           row.Creator,
	}
}
