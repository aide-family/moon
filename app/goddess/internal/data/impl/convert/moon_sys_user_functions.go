package convert

import (
	"github.com/aide-family/goddess/internal/biz/bo"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// UserFunctionJoinRow 用于接收 JOIN 查询结果的结构体
type UserFunctionJoinRow struct {
	MenuCode          string
	FunctionCode      string
	FunctionName      string
	FunctionSortOrder int32
	FunctionStatus    int32
	Creator           int64
	CreatedAt         string
}

// UserFunctionJoinToBo 将 JOIN 查询结果转换为业务对象
func UserFunctionJoinToBo(row *UserFunctionJoinRow) *bo.UserFunctionItemBo {
	if row == nil {
		return nil
	}
	return &bo.UserFunctionItemBo{
		MenuCode:          row.MenuCode,
		FunctionCode:      row.FunctionCode,
		FunctionName:      row.FunctionName,
		FunctionSortOrder: row.FunctionSortOrder,
		FunctionStatus:    row.FunctionStatus,
		Creator:           row.Creator,
		CreatedAt:         row.CreatedAt,
	}
}

// UserFunctionItemBoToAPIV1 将业务对象转换为 proto 响应
func UserFunctionItemBoToAPIV1(b *bo.UserFunctionItemBo) *goddessv1.UserFunctionItem {
	return &goddessv1.UserFunctionItem{
		MenuCode:          b.MenuCode,
		FunctionCode:      b.FunctionCode,
		FunctionName:      b.FunctionName,
		FunctionSortOrder: b.FunctionSortOrder,
		FunctionStatus:    b.FunctionStatus,
		Creator:           b.Creator,
		CreatedAt:         b.CreatedAt,
	}
}
