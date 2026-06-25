package convert

import (
	"time"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/data/impl/do"
)

// SysFunctionToBo 将数据库模型转换为功能业务对象
func SysFunctionToBo(fn *do.SysFunction) *bo.SysFunctionBo {
	if fn == nil {
		return nil
	}
	// 解析时间字符串
	createdAt, _ := time.Parse(time.DateTime, fn.CreatedAt)
	updatedAt, _ := time.Parse(time.DateTime, fn.UpdatedAt)
	return &bo.SysFunctionBo{
		MenuCode:          fn.MenuCode,
		FunctionCode:      fn.FunctionCode,
		FunctionName:      fn.FunctionName,
		FunctionSortOrder: fn.FunctionSortOrder,
		FunctionStatus:    fn.FunctionStatus,
		FunctionRemark:    fn.FunctionRemark,
		Creator:           fn.Creator,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
	}
}
