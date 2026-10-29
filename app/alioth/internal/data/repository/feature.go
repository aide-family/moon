package repository

import (
	"context"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// FeatureRepo 读写功能套餐和租户级功能差异。
// 有效功能由套餐功能叠加租户差异计算，不落表。
type FeatureRepo interface {
	ListPackages(ctx context.Context, page, pageSize int, name string) ([]bo.FeaturePackage, int, error)
	GetPackage(ctx context.Context, uid string) (*bo.FeaturePackage, error)
	CreatePackage(ctx context.Context, name, remark string, features []string) (*bo.FeaturePackage, error)
	UpdatePackage(ctx context.Context, uid, name, remark string, features []string) (*bo.FeaturePackage, error)
	DeletePackage(ctx context.Context, uid string) error
	// NamespaceFeatureState 返回租户的套餐、套餐功能和有效功能。
	NamespaceFeatureState(ctx context.Context, namespaceUID string) (*bo.NamespaceFeature, error)
	// EffectiveFeatures 只返回租户的有效功能，按 perm.Features() 的顺序。
	// 每次租户权限检查都会调用它，查询量固定为套餐明细和差异两张小表。
	EffectiveFeatures(ctx context.Context, namespaceUID string) ([]string, error)
	// SetNamespaceFeatureState 原子设置租户的套餐和功能差异。
	// features 是目标有效功能，仓储负责和套餐功能对比换算出差异记录。
	SetNamespaceFeatureState(ctx context.Context, namespaceUID, packageUID string, features []string) error
	// CountNamespacesUsingPackage 统计使用该套餐的未删除租户数，被使用的套餐不能删除。
	CountNamespacesUsingPackage(ctx context.Context, packageUID string) (int, error)
}
