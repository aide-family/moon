package convert

import (
	"time"

	"github.com/aide-family/magicbox/timex"

	"github.com/aide-family/goddess/internal/biz/bo"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// RoleFunctionJoinRow 用于接收 JOIN 查询结果的结构体
type RoleFunctionJoinRow struct {
	MenuCode          string
	FunctionCode      string
	FunctionName      string
	FunctionSortOrder int32
	FunctionStatus    int32
	Creator           int64
	CreatedAt         time.Time
}

// RoleFunctionJoinToBo 将 JOIN 查询结果转换为业务对象
func RoleFunctionJoinToBo(row *RoleFunctionJoinRow) *bo.RoleFunctionItemBo {
	if row == nil {
		return nil
	}
	return &bo.RoleFunctionItemBo{
		MenuCode:          row.MenuCode,
		FunctionCode:      row.FunctionCode,
		FunctionName:      row.FunctionName,
		FunctionSortOrder: row.FunctionSortOrder,
		FunctionStatus:    row.FunctionStatus,
		Creator:           row.Creator,
		CreatedAt:         row.CreatedAt,
	}
}

// RoleFunctionItemBoToAPIV1 将业务对象转换为 proto 响应
func RoleFunctionItemBoToAPIV1(b *bo.RoleFunctionItemBo) *goddessv1.RoleFunctionItem {
	return &goddessv1.RoleFunctionItem{
		MenuCode:          b.MenuCode,
		FunctionCode:      b.FunctionCode,
		FunctionName:      b.FunctionName,
		FunctionSortOrder: b.FunctionSortOrder,
		FunctionStatus:    b.FunctionStatus,
		Creator:           b.Creator,
		CreatedAt:         timex.FormatTime(&b.CreatedAt),
	}
}
