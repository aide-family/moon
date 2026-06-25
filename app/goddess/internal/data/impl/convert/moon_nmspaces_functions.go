package convert

import (
	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/data/impl/do"
)

// NmspacesFunctionToBo 将数据库模型转换为租户功能业务对象
func NmspacesFunctionToBo(fn *do.NmspacesFunction) *bo.NmspacesFunctionBo {
	if fn == nil {
		return nil
	}
	return &bo.NmspacesFunctionBo{
		MenuCode:          fn.MenuCode,
		FunctionCode:      fn.FunctionCode,
		FunctionName:      fn.FunctionName,
		FunctionSortOrder: fn.FunctionSortOrder,
		FunctionStatus:    fn.FunctionStatus,
		FunctionRemark:    fn.FunctionRemark,
		Creator:           fn.Creator.Int64(),
		CreatedAt:         fn.CreatedAt,
		UpdatedAt:         fn.UpdatedAt,
	}
}
