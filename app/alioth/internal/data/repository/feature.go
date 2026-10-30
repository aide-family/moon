package repository

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// FeatureRepo 读写功能套餐和租户级额外能力。
// 套餐与租户差异都只存能力编码；有效能力由套餐能力并上租户额外能力计算，不落表。
type FeatureRepo interface {
	ListPackages(ctx context.Context, page, pageSize int, name string) ([]bo.FeaturePackage, int, error)
	// ListPublicPackages 返回公开且启用的套餐，供用户创建租户时选择，按价格和名称排序。
	ListPublicPackages(ctx context.Context) ([]bo.FeaturePackage, error)
	GetPackage(ctx context.Context, uid string) (*bo.FeaturePackage, error)
	// CreatePackage 新建套餐。pkg.Features 是能力编码，pkg.DurationDays 是默认授权时长（0=永久），
	// pkg.Visibility 决定公开/内测，pkg.MaxMembers 是成员数上限（0=不限），pkg.PriceCents 是价格（分）。
	CreatePackage(ctx context.Context, pkg bo.FeaturePackage) (*bo.FeaturePackage, error)
	// UpdatePackage 同时替换名称、备注、能力明细、默认时长、版本、人数上限和价格。
	UpdatePackage(ctx context.Context, pkg bo.FeaturePackage) (*bo.FeaturePackage, error)
	DeletePackage(ctx context.Context, uid string) error
	// NamespaceFeatureState 返回租户的套餐、套餐能力、额外能力、已配置的合并能力和到期状态。
	// 它用于平台管理端展示，即使租户停用或已到期也返回配置；是否生效由 EffectiveCapabilities 判断。
	NamespaceFeatureState(ctx context.Context, namespaceUID string) (*bo.NamespaceFeature, error)
	// EffectiveCapabilities 只返回租户当前生效的能力，按能力目录顺序。
	// 租户停用、已归档、套餐已到期都返回空；每次租户权限检查都会调用它。
	EffectiveCapabilities(ctx context.Context, namespaceUID string) ([]string, error)
	// EffectiveModules 把有效能力反推成模块编码，按模块展示顺序。
	// access 用它做模块级功能检查：模块下还有一个能力开通，模块就还在。
	EffectiveModules(ctx context.Context, namespaceUID string) ([]string, error)
	// EffectiveFeatures 是 EffectiveModules 的旧名字，保留给尚未改用能力模型的调用方。
	// Deprecated: 新代码使用 EffectiveModules 或 EffectiveCapabilities。
	EffectiveFeatures(ctx context.Context, namespaceUID string) ([]string, error)
	// SetNamespaceFeatureState 原子设置租户的套餐、目标能力和到期时间。
	// capabilities 是目标能力集合，必须包含套餐的全部能力（只增不减），仓储负责换算成
	// 租户额外能力记录；expiresAt 零值表示永久。同时写入 namespace.package_uid 和
	// namespace.package_expires_at。
	SetNamespaceFeatureState(ctx context.Context, namespaceUID, packageUID string, capabilities []string, expiresAt time.Time) error
	// ArchiveNamespace 软删除租户及其全部业务数据。归档不可恢复，调用方必须先确认。
	ArchiveNamespace(ctx context.Context, namespaceUID string) error
	// CountNamespacesUsingPackage 统计使用该套餐的未删除租户数，被使用的套餐不能删除。
	CountNamespacesUsingPackage(ctx context.Context, packageUID string) (int, error)
	// NamespaceMemberLimit 返回租户当前套餐的成员数上限，0 表示不限制。
	// 容量跟着套餐实时变化，不在 namespaces 上冗余；套餐缺失时按不限制处理。
	NamespaceMemberLimit(ctx context.Context, namespaceUID string) (int32, error)
}
