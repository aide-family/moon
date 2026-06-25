package repository

import (
	"context"

	"github.com/aide-family/goddess/internal/biz/bo"
)

// SysFunction 系统功能仓储接口
type SysFunction interface {
	// ListSysFunction 分页查询系统功能列表
	ListSysFunction(ctx context.Context, req *bo.ListSysFunctionBo) ([]*bo.SysFunctionBo, int64, error)
	// GetSysFunction 根据联合主键查询功能详情
	GetSysFunction(ctx context.Context, menuCode, functionCode string) (*bo.SysFunctionBo, error)
	// CreateSysFunction 新增系统功能
	CreateSysFunction(ctx context.Context, req *bo.CreateSysFunctionBo) (*bo.SysFunctionBo, error)
	// UpdateSysFunction 编辑系统功能
	UpdateSysFunction(ctx context.Context, req *bo.UpdateSysFunctionBo) (*bo.SysFunctionBo, error)
	// DeleteSysFunction 删除系统功能（软删除）
	DeleteSysFunction(ctx context.Context, menuCode, functionCode string) error
	// ExistsByMenuCodeAndFunctionCode 检查联合主键是否已存在（含软删除行）
	ExistsByMenuCodeAndFunctionCode(ctx context.Context, menuCode, functionCode string) (bool, error)
	// ExistsByMenuCodeAndFunctionCodeNotDeleted 检查联合主键在未删除行中是否存在
	ExistsByMenuCodeAndFunctionCodeNotDeleted(ctx context.Context, menuCode, functionCode string) (bool, error)
	// ExistsByMenuCodeAndFunctionName 检查同一 menuCode 下 functionName 是否已存在（未删除行）
	ExistsByMenuCodeAndFunctionName(ctx context.Context, menuCode, functionName string) (bool, error)
}
