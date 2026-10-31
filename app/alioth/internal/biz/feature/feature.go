package feature

import (
	"context"
	"slices"
	"time"

	"github.com/google/wire"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/page"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

var ProviderSet = wire.NewSet(NewProvisioning)

// Provisioning 维护功能目录和套餐，并把功能开通给租户。
// 所有入口都走系统权限，不读 X-Namespace。
type Provisioning struct {
	features repository.FeatureRepo
	access   access.Checker
}

func NewProvisioning(features repository.FeatureRepo, checker access.Checker) *Provisioning {
	return &Provisioning{features: features, access: checker}
}

// ListFeatures 返回代码维护的能力目录。能力不是租户数据，不分页。
// 每个能力带上所属菜单模块，前端按模块分组展示授权项。
func (provisioning *Provisioning) ListFeatures(ctx context.Context) ([]bo.Feature, error) {
	if _, err := provisioning.access.RequireSystem(ctx, perm.SystemPackageRead); err != nil {
		return nil, err
	}
	items := perm.Capabilities()
	out := make([]bo.Feature, 0, len(items))
	for _, item := range items {
		out = append(out, bo.Feature{
			Code:        item.Code,
			Module:      item.Module,
			Name:        item.Name,
			Description: item.Description,
			Permissions: item.Permissions,
		})
	}
	return out, nil
}

func (provisioning *Provisioning) ListPackages(ctx context.Context, pageNo, pageSize int32, name string) ([]bo.FeaturePackage, int, int, int, error) {
	if _, err := provisioning.access.RequireSystem(ctx, perm.SystemPackageRead); err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := provisioning.features.ListPackages(ctx, p, s, name)
	return items, total, p, s, provisioning.access.Wrap(err)
}

func (provisioning *Provisioning) GetPackage(ctx context.Context, uid string) (*bo.FeaturePackage, error) {
	if _, err := provisioning.access.RequireSystem(ctx, perm.SystemPackageRead); err != nil {
		return nil, err
	}
	item, err := provisioning.features.GetPackage(ctx, uid)
	return item, provisioning.access.Wrap(err)
}

func (provisioning *Provisioning) CreatePackage(ctx context.Context, pkg bo.FeaturePackage) (*bo.FeaturePackage, error) {
	if _, err := provisioning.access.RequireSystem(ctx, perm.SystemPackageWrite); err != nil {
		return nil, err
	}
	if pkg.Currency == "" {
		pkg.Currency = bo.CurrencyCNY
	}
	if err := validatePackage(pkg); err != nil {
		return nil, err
	}
	item, err := provisioning.features.CreatePackage(ctx, pkg)
	return item, provisioning.access.Wrap(err)
}

// UpdatePackage 更新套餐。基础套餐只允许改功能和备注：改名会破坏它的识别，
// 所以名称必须保持原值。
func (provisioning *Provisioning) UpdatePackage(ctx context.Context, pkg bo.FeaturePackage) (*bo.FeaturePackage, error) {
	if _, err := provisioning.access.RequireSystem(ctx, perm.SystemPackageWrite); err != nil {
		return nil, err
	}
	if pkg.Currency == "" {
		pkg.Currency = bo.CurrencyCNY
	}
	if err := validatePackage(pkg); err != nil {
		return nil, err
	}
	current, err := provisioning.features.GetPackage(ctx, pkg.UID)
	if err != nil {
		return nil, provisioning.access.Wrap(err)
	}
	if current.IsBase && pkg.Name != current.Name {
		return nil, merr.Client("base package cannot be renamed")
	}
	item, err := provisioning.features.UpdatePackage(ctx, pkg)
	return item, provisioning.access.Wrap(err)
}

// DeletePackage 删除套餐。基础套餐和被租户使用的套餐不能删除，
// 否则租户会在下一次鉴权时突然失去全部功能。
func (provisioning *Provisioning) DeletePackage(ctx context.Context, uid string) error {
	if _, err := provisioning.access.RequireSystem(ctx, perm.SystemPackageWrite); err != nil {
		return err
	}
	current, err := provisioning.features.GetPackage(ctx, uid)
	if err != nil {
		return provisioning.access.Wrap(err)
	}
	if current.IsBase {
		return merr.Client("base package cannot be deleted")
	}
	used, err := provisioning.features.CountNamespacesUsingPackage(ctx, uid)
	if err != nil {
		return provisioning.access.Wrap(err)
	}
	if used > 0 {
		return merr.Client("package is in use")
	}
	return provisioning.access.Wrap(provisioning.features.DeletePackage(ctx, uid))
}

// ListPackageNamespaces 返回正在使用该套餐的租户，供同步授权选择目标。
func (provisioning *Provisioning) ListPackageNamespaces(ctx context.Context, packageUID string) ([]bo.Namespace, error) {
	if _, err := provisioning.access.RequireSystem(ctx, perm.SystemPackageRead); err != nil {
		return nil, err
	}
	if _, err := provisioning.features.GetPackage(ctx, packageUID); err != nil {
		return nil, provisioning.access.Wrap(err)
	}
	items, err := provisioning.features.ListPackageNamespaces(ctx, packageUID)
	return items, provisioning.access.Wrap(err)
}

// SyncPackage 把套餐当前能力快照同步给指定租户；namespaceUIDs 为空表示该套餐下全部租户。
// 同步只覆盖套餐授权，不动额外能力、到期时间和价格。
func (provisioning *Provisioning) SyncPackage(ctx context.Context, packageUID string, namespaceUIDs []string) (int, int, error) {
	if _, err := provisioning.access.RequireSystem(ctx, perm.SystemPackageWrite); err != nil {
		return 0, 0, err
	}
	if _, err := provisioning.features.GetPackage(ctx, packageUID); err != nil {
		return 0, 0, provisioning.access.Wrap(err)
	}
	synced, skipped, err := provisioning.features.SyncPackageToNamespaces(ctx, packageUID, namespaceUIDs)
	return synced, skipped, provisioning.access.Wrap(err)
}

// GetNamespaceFeatures 返回租户的套餐、套餐功能和有效功能。
func (provisioning *Provisioning) GetNamespaceFeatures(ctx context.Context, namespaceUID string) (*bo.NamespaceFeature, error) {
	if _, err := provisioning.access.RequireSystem(ctx, perm.SystemPackageRead); err != nil {
		return nil, err
	}
	item, err := provisioning.features.NamespaceFeatureState(ctx, namespaceUID)
	return item, provisioning.access.Wrap(err)
}

// UpdateNamespaceFeatures 原子设置租户的套餐和目标有效能力。
// 服务端按套餐能力换算差异，调用方只表达最终结果：
// 目标能力必须包含租户当前生效的套餐能力（只增不减），expiresAt 为 0 时按套餐授权天数计算。
func (provisioning *Provisioning) UpdateNamespaceFeatures(ctx context.Context, namespaceUID, packageUID string, capabilities []string, expiresAt int64) (*bo.NamespaceFeature, error) {
	if _, err := provisioning.access.RequireSystem(ctx, perm.SystemPackageWrite); err != nil {
		return nil, err
	}
	item, err := provisioning.features.GetPackage(ctx, packageUID)
	if err != nil {
		return nil, provisioning.access.Wrap(err)
	}
	// 不传功能表示「不额外调整」，直接使用套餐默认能力。
	if len(capabilities) == 0 {
		capabilities = item.Features
	}
	// 旧客户端或旧套餐快照可能还带着 M24 拆分前的历史能力码，
	// 先展开成当前能力再比较，避免同义编码被当成缺能力。
	normalized, err := normalizeCapabilityInput(capabilities)
	if err != nil {
		return nil, err
	}
	capabilities = normalized
	required, err := provisioning.requiredCapabilities(ctx, namespaceUID, packageUID, item)
	if err != nil {
		return nil, err
	}
	for _, code := range required {
		if !slices.Contains(capabilities, code) {
			return nil, merr.Client("features must include the package features")
		}
	}
	expiry := resolveExpiresAt(item.DurationDays, expiresAt)
	if err := provisioning.features.SetNamespaceFeatureState(ctx, namespaceUID, packageUID, capabilities, expiry); err != nil {
		return nil, provisioning.access.Wrap(err)
	}
	state, err := provisioning.features.NamespaceFeatureState(ctx, namespaceUID)
	return state, provisioning.access.Wrap(err)
}

// requiredCapabilities 返回本次保存必须包含的能力。
// 租户仍使用原套餐时必须包含它当前的套餐快照：管理员后来编辑套餐不会反向要求
// 已生效租户补能力（M22 的权益快照语义）；换套餐时才按新套餐当前明细要求。
// 两者都先归一化，历史编码与新编码按同一含义比较。
func (provisioning *Provisioning) requiredCapabilities(ctx context.Context, namespaceUID, packageUID string, item *bo.FeaturePackage) ([]string, error) {
	state, err := provisioning.features.NamespaceFeatureState(ctx, namespaceUID)
	if err != nil {
		return nil, provisioning.access.Wrap(err)
	}
	if state != nil && state.PackageUID == packageUID {
		return perm.NormalizeCapabilities(state.PackageFeatures), nil
	}
	return perm.NormalizeCapabilities(item.Features), nil
}

// normalizeCapabilityInput 校验并归一化调用方提交的能力：
// 接受当前能力码和 M24 拆分前的历史能力码（展开成当前能力），
// 拒绝未知编码和重复项。归一化后同一含义的编码不会重复。
func normalizeCapabilityInput(capabilities []string) ([]string, error) {
	seen := make(map[string]struct{}, len(capabilities))
	for _, code := range capabilities {
		if !perm.IsCapability(code) && !perm.IsLegacyCapability(code) {
			return nil, merr.Client("invalid capability")
		}
		if _, ok := seen[code]; ok {
			return nil, merr.Client("invalid capability")
		}
		seen[code] = struct{}{}
	}
	return perm.NormalizeCapabilities(capabilities), nil
}

// resolveExpiresAt 把调用方的 Unix 秒和套餐授权时长换算成仓储使用的到期时间。
// 0 表示永久；没有显式到期时间时按套餐的授权天数从现在起算。
func resolveExpiresAt(durationDays int32, expiresAt int64) time.Time {
	if expiresAt > 0 {
		return time.Unix(expiresAt, 0)
	}
	if durationDays > 0 {
		return time.Now().AddDate(0, 0, int(durationDays))
	}
	return time.Time{}
}

// validatePackage 校验套餐的可售属性：版本、授权时长、人数上限、价格和能力目录。
// proto 已经做过一次格式校验，这里保证内部调用（seed、测试）也走同一套规则。
func validatePackage(pkg bo.FeaturePackage) error {
	if err := validateDuration(pkg.DurationDays); err != nil {
		return err
	}
	if pkg.Visibility != perm.PackageVisibilityPublic && pkg.Visibility != perm.PackageVisibilityBeta {
		return merr.Invalid("invalid visibility")
	}
	if pkg.MaxMembers < 0 {
		return merr.Invalid("invalid max members")
	}
	if pkg.PriceCents < 0 {
		return merr.Invalid("invalid price")
	}
	if !bo.IsCurrency(pkg.Currency) {
		return merr.Invalid("invalid currency")
	}
	return validateCapabilities(pkg.Features)
}

// validateDuration 拒绝负数授权天数；0 表示永久。
func validateDuration(durationDays int32) error {
	if durationDays < 0 {
		return merr.Invalid("invalid duration")
	}
	return nil
}

// validateCapabilities 拒绝目录外的能力编码和重复项。
func validateCapabilities(capabilities []string) error {
	seen := make(map[string]struct{}, len(capabilities))
	for _, code := range capabilities {
		if !perm.IsCapability(code) {
			return merr.Client("invalid capability")
		}
		if _, ok := seen[code]; ok {
			return merr.Client("invalid capability")
		}
		seen[code] = struct{}{}
	}
	return nil
}
