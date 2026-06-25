package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/goddess/internal/biz/bo"
)

// NmspacesFunction 租户功能仓储接口
type NmspacesFunction interface {
	// ListNmspacesFunction 分页查询租户功能列表
	ListNmspacesFunction(ctx context.Context, req *bo.ListNmspacesFunctionBo) ([]*bo.NmspacesFunctionBo, int64, error)
	// GetNmspacesFunction 根据联合主键查询功能详情
	GetNmspacesFunction(ctx context.Context, namespaceUID snowflake.ID, menuCode, functionCode string) (*bo.NmspacesFunctionBo, error)
	// CreateNmspacesFunction 新增租户功能
	CreateNmspacesFunction(ctx context.Context, req *bo.CreateNmspacesFunctionBo) (*bo.NmspacesFunctionBo, error)
	// UpdateNmspacesFunction 编辑租户功能
	UpdateNmspacesFunction(ctx context.Context, req *bo.UpdateNmspacesFunctionBo) (*bo.NmspacesFunctionBo, error)
	// DeleteNmspacesFunction 删除租户功能（软删除）
	DeleteNmspacesFunction(ctx context.Context, namespaceUID snowflake.ID, menuCode, functionCode string) error
	// ExistsByNamespaceAndMenuCodeAndFunctionCode 检查联合主键是否已存在（含软删除行）
	ExistsByNamespaceAndMenuCodeAndFunctionCode(ctx context.Context, namespaceUID snowflake.ID, menuCode, functionCode string) (bool, error)
	// ExistsByNamespaceAndMenuCodeAndFunctionCodeNotDeleted 检查联合主键在未删除行中是否存在
	ExistsByNamespaceAndMenuCodeAndFunctionCodeNotDeleted(ctx context.Context, namespaceUID snowflake.ID, menuCode, functionCode string) (bool, error)
	// ExistsByNamespaceAndMenuCodeAndFunctionName 检查同一 menuCode 下 functionName 是否已存在（未删除行）
	ExistsByNamespaceAndMenuCodeAndFunctionName(ctx context.Context, namespaceUID snowflake.ID, menuCode, functionName string) (bool, error)
}
