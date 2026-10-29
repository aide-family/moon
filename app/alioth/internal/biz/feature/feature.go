package feature

import (
	"context"

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

// ListFeatures 返回代码维护的功能目录。功能不是租户数据，不分页。
func (provisioning *Provisioning) ListFeatures(ctx context.Context) ([]bo.Feature, error) {
	if _, err := provisioning.access.RequireSystem(ctx, perm.SystemPackageRead); err != nil {
		return nil, err
	}
	items := perm.Features()
	out := make([]bo.Feature, 0, len(items))
	for _, item := range items {
		out = append(out, bo.Feature{
			Code:        item.Code,
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

func (provisioning *Provisioning) CreatePackage(ctx context.Context, name, remark string, features []string) (*bo.FeaturePackage, error) {
	if _, err := provisioning.access.RequireSystem(ctx, perm.SystemPackageWrite); err != nil {
		return nil, err
	}
	if err := validateFeatures(features); err != nil {
		return nil, err
	}
	item, err := provisioning.features.CreatePackage(ctx, name, remark, features)
	return item, provisioning.access.Wrap(err)
}

// UpdatePackage 更新套餐。基础套餐只允许改功能和备注：改名会破坏它的识别，
// 所以名称必须保持原值。
func (provisioning *Provisioning) UpdatePackage(ctx context.Context, uid, name, remark string, features []string) (*bo.FeaturePackage, error) {
	if _, err := provisioning.access.RequireSystem(ctx, perm.SystemPackageWrite); err != nil {
		return nil, err
	}
	current, err := provisioning.features.GetPackage(ctx, uid)
	if err != nil {
		return nil, provisioning.access.Wrap(err)
	}
	if current.IsBase && name != current.Name {
		return nil, merr.Client("base package cannot be renamed")
	}
	if err := validateFeatures(features); err != nil {
		return nil, err
	}
	item, err := provisioning.features.UpdatePackage(ctx, uid, name, remark, features)
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

// GetNamespaceFeatures 返回租户的套餐、套餐功能和有效功能。
func (provisioning *Provisioning) GetNamespaceFeatures(ctx context.Context, namespaceUID string) (*bo.NamespaceFeature, error) {
	if _, err := provisioning.access.RequireSystem(ctx, perm.SystemPackageRead); err != nil {
		return nil, err
	}
	item, err := provisioning.features.NamespaceFeatureState(ctx, namespaceUID)
	return item, provisioning.access.Wrap(err)
}

// UpdateNamespaceFeatures 原子设置租户的套餐和目标有效功能。
// 服务端按套餐功能换算差异，调用方只表达最终结果。
func (provisioning *Provisioning) UpdateNamespaceFeatures(ctx context.Context, namespaceUID, packageUID string, features []string) (*bo.NamespaceFeature, error) {
	if _, err := provisioning.access.RequireSystem(ctx, perm.SystemPackageWrite); err != nil {
		return nil, err
	}
	if _, err := provisioning.features.GetPackage(ctx, packageUID); err != nil {
		return nil, provisioning.access.Wrap(err)
	}
	if err := validateFeatures(features); err != nil {
		return nil, err
	}
	if err := provisioning.features.SetNamespaceFeatureState(ctx, namespaceUID, packageUID, features); err != nil {
		return nil, provisioning.access.Wrap(err)
	}
	item, err := provisioning.features.NamespaceFeatureState(ctx, namespaceUID)
	return item, provisioning.access.Wrap(err)
}

// validateFeatures 拒绝目录外的功能编码和重复项。
func validateFeatures(features []string) error {
	seen := make(map[string]struct{}, len(features))
	for _, code := range features {
		if !perm.IsFeature(code) {
			return merr.Client("invalid feature")
		}
		if _, ok := seen[code]; ok {
			return merr.Client("invalid feature")
		}
		seen[code] = struct{}{}
	}
	return nil
}
