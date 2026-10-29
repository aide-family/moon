package data

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/featurepackage"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/featurepackageitem"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/namespace"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/namespacefeature"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/role"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/rolepermission"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

func (r *featureRepo) ListPackages(ctx context.Context, page, pageSize int, name string) ([]bo.FeaturePackage, int, error) {
	query := r.client.FeaturePackage.Query().Where(featurepackage.DeletedAtIsNil())
	if name != "" {
		query = query.Where(featurepackage.NameContainsFold(name))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Asc(featurepackage.FieldUID)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items, err := r.packageViews(ctx, rows)
	return items, total, err
}

func (r *featureRepo) GetPackage(ctx context.Context, uid string) (*bo.FeaturePackage, error) {
	row, err := r.client.FeaturePackage.Query().Where(featurepackage.UIDEQ(uid), featurepackage.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items, err := r.packageViews(ctx, []*ent.FeaturePackage{row})
	if err != nil || len(items) == 0 {
		return nil, err
	}
	return &items[0], nil
}

func (r *featureRepo) CreatePackage(ctx context.Context, name, remark string, features []string) (*bo.FeaturePackage, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	row, err := tx.FeaturePackage.Create().
		SetUID(id.New()).
		SetName(name).
		SetRemark(remark).
		SetIsBase(false).
		SetStatus(featurepackage.StatusACTIVE).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if err := r.createPackageItems(ctx, tx, row.UID, features); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetPackage(ctx, row.UID)
}

// UpdatePackage 同时替换名称、备注和功能明细。基础套餐的改名由用例层拒绝，
// 这里只负责在一个事务里写入，避免功能明细出现半更新。
func (r *featureRepo) UpdatePackage(ctx context.Context, uid, name, remark string, features []string) (*bo.FeaturePackage, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	n, err := tx.FeaturePackage.Update().
		Where(featurepackage.UIDEQ(uid), featurepackage.DeletedAtIsNil()).
		SetName(name).
		SetRemark(remark).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	if _, err := tx.FeaturePackageItem.Delete().Where(featurepackageitem.PackageUIDEQ(uid)).Exec(ctx); err != nil {
		return nil, r.fail(err)
	}
	if err := r.createPackageItems(ctx, tx, uid, features); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetPackage(ctx, uid)
}

func (r *featureRepo) DeletePackage(ctx context.Context, uid string) error {
	n, err := r.client.FeaturePackage.Update().
		Where(featurepackage.UIDEQ(uid), featurepackage.DeletedAtIsNil()).
		SetDeletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("not found")
	}
	return nil
}

func (r *featureRepo) NamespaceFeatureState(ctx context.Context, namespaceUID string) (*bo.NamespaceFeature, error) {
	ns, err := r.client.Namespace.Query().Where(namespace.UIDEQ(namespaceUID), namespace.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	state := &bo.NamespaceFeature{NamespaceUID: namespaceUID, PackageUID: ns.PackageUID}
	if ns.PackageUID != "" {
		pkg, err := r.client.FeaturePackage.Query().Where(featurepackage.UIDEQ(ns.PackageUID)).Only(ctx)
		switch {
		case err == nil:
			state.PackageName = pkg.Name
			features, err := r.packageFeatures(ctx, pkg.UID)
			if err != nil {
				return nil, err
			}
			state.PackageFeatures = features
		case ent.IsNotFound(err):
			// 套餐被物理删除时按空套餐处理；正常删除走软删除，租户仍保留套餐功能。
		default:
			return nil, r.fail(err)
		}
	}
	rows, err := r.client.NamespaceFeature.Query().Where(namespacefeature.NamespaceUIDEQ(namespaceUID)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	overrides := make(map[string]bool, len(rows))
	for _, row := range rows {
		overrides[row.Feature] = row.Enabled
	}
	state.Features = perm.EffectiveFeatures(state.PackageFeatures, overrides)
	return state, nil
}

func (r *featureRepo) EffectiveFeatures(ctx context.Context, namespaceUID string) ([]string, error) {
	state, err := r.NamespaceFeatureState(ctx, namespaceUID)
	if err != nil {
		return nil, err
	}
	return state.Features, nil
}

// SetNamespaceFeatureState 把目标有效功能换算成相对套餐的差异。
// 差异只写目录里的功能：不在目录里的旧编码被忽略，避免历史数据被重新写回。
func (r *featureRepo) SetNamespaceFeatureState(ctx context.Context, namespaceUID, packageUID string, features []string) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	// 先确认租户存在，避免给已删除或不存在的租户留下孤立的功能差异。
	if _, err := tx.Namespace.Query().Where(namespace.UIDEQ(namespaceUID), namespace.DeletedAtIsNil()).Only(ctx); err != nil {
		return r.fail(err)
	}
	pkg, err := tx.FeaturePackage.Query().Where(featurepackage.UIDEQ(packageUID), featurepackage.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return r.fail(err)
	}
	packageRows, err := tx.FeaturePackageItem.Query().Where(featurepackageitem.PackageUIDEQ(pkg.UID)).All(ctx)
	if err != nil {
		return r.fail(err)
	}
	inPackage := make(map[string]struct{}, len(packageRows))
	for _, row := range packageRows {
		inPackage[row.Feature] = struct{}{}
	}
	target := make(map[string]struct{}, len(features))
	for _, code := range features {
		if perm.IsFeature(code) {
			target[code] = struct{}{}
		}
	}
	if _, err := tx.NamespaceFeature.Delete().Where(namespacefeature.NamespaceUIDEQ(namespaceUID)).Exec(ctx); err != nil {
		return r.fail(err)
	}
	for _, code := range perm.FeatureCodes() {
		_, wanted := target[code]
		_, included := inPackage[code]
		if wanted == included {
			continue
		}
		if _, err := tx.NamespaceFeature.Create().
			SetUID(id.New()).
			SetNamespaceUID(namespaceUID).
			SetFeature(code).
			SetEnabled(wanted).
			Save(ctx); err != nil {
			return r.fail(err)
		}
	}
	if _, err := tx.Namespace.Update().
		Where(namespace.UIDEQ(namespaceUID), namespace.DeletedAtIsNil()).
		SetPackageUID(packageUID).
		Save(ctx); err != nil {
		return r.fail(err)
	}
	// 功能范围变化后立刻补齐预置系统角色的权限，租户管理员不用等下一次 migrate
	// 就能在新开通的功能里给成员授权；关闭功能时旧授权保留，由权限检查拒绝。
	effective := perm.EffectiveFeatures(nil, enabledOverrides(target))
	if err := r.grantPresetRolePermissions(ctx, tx, namespaceUID, effective); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return r.fail(err)
	}
	return nil
}

// enabledOverrides 把目标功能集合转成 EffectiveFeatures 需要的差异入参。
func enabledOverrides(target map[string]struct{}) map[string]bool {
	overrides := make(map[string]bool, len(target))
	for code := range target {
		overrides[code] = true
	}
	return overrides
}

// grantPresetRolePermissions 给租户的预置系统角色补上当前功能内缺失的权限。
// 只加不删，和 seed 的 syncTenantRoles 保持同一套规则；重复执行不会产生重复授权。
func (r *featureRepo) grantPresetRolePermissions(ctx context.Context, tx *ent.Tx, namespaceUID string, features []string) error {
	permissions, err := tx.Permission.Query().All(ctx)
	if err != nil {
		return r.fail(err)
	}
	byName := make(map[string]string, len(permissions))
	for _, item := range permissions {
		byName[item.Name] = item.UID
	}
	roles, err := tx.Role.Query().
		Where(role.NamespaceUIDEQ(namespaceUID), role.IsSystem(true), role.DeletedAtIsNil()).
		All(ctx)
	if err != nil {
		return r.fail(err)
	}
	if len(roles) == 0 {
		return nil
	}
	roleUIDs := make([]string, len(roles))
	for index, item := range roles {
		roleUIDs[index] = item.UID
	}
	links, err := tx.RolePermission.Query().Where(rolepermission.RoleUIDIn(roleUIDs...)).All(ctx)
	if err != nil {
		return r.fail(err)
	}
	held := make(map[string]map[string]struct{}, len(roles))
	for _, link := range links {
		if held[link.RoleUID] == nil {
			held[link.RoleUID] = make(map[string]struct{})
		}
		held[link.RoleUID][link.PermissionUID] = struct{}{}
	}
	for _, item := range roles {
		for _, code := range perm.RolePermissionsForFeatures(item.Name, features) {
			permissionUID := byName[code]
			if permissionUID == "" {
				continue
			}
			if _, ok := held[item.UID][permissionUID]; ok {
				continue
			}
			if _, err := tx.RolePermission.Create().
				SetUID(id.New()).
				SetNamespaceUID(namespaceUID).
				SetRoleUID(item.UID).
				SetPermissionUID(permissionUID).
				Save(ctx); err != nil {
				return r.fail(err)
			}
		}
	}
	return nil
}

func (r *featureRepo) CountNamespacesUsingPackage(ctx context.Context, packageUID string) (int, error) {
	total, err := r.client.Namespace.Query().
		Where(namespace.PackageUIDEQ(packageUID), namespace.DeletedAtIsNil()).
		Count(ctx)
	if err != nil {
		return 0, r.fail(err)
	}
	return total, nil
}

// packageViews 批量补齐套餐的功能明细，并把功能按目录顺序排列。
func (r *featureRepo) packageViews(ctx context.Context, rows []*ent.FeaturePackage) ([]bo.FeaturePackage, error) {
	if len(rows) == 0 {
		return nil, nil
	}
	uids := make([]string, 0, len(rows))
	for _, row := range rows {
		uids = append(uids, row.UID)
	}
	items, err := r.client.FeaturePackageItem.Query().Where(featurepackageitem.PackageUIDIn(uids...)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	codesByPackage := make(map[string][]string, len(rows))
	for _, item := range items {
		codesByPackage[item.PackageUID] = append(codesByPackage[item.PackageUID], item.Feature)
	}
	out := make([]bo.FeaturePackage, 0, len(rows))
	for _, row := range rows {
		out = append(out, bo.FeaturePackage{
			UID:       row.UID,
			Name:      row.Name,
			Remark:    row.Remark,
			IsBase:    row.IsBase,
			Status:    string(row.Status),
			Features:  orderFeatureCodes(codesByPackage[row.UID]),
			CreatedAt: row.CreatedAt,
		})
	}
	return out, nil
}

func (r *featureRepo) packageFeatures(ctx context.Context, packageUID string) ([]string, error) {
	rows, err := r.client.FeaturePackageItem.Query().Where(featurepackageitem.PackageUIDEQ(packageUID)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	codes := make([]string, 0, len(rows))
	for _, row := range rows {
		codes = append(codes, row.Feature)
	}
	return orderFeatureCodes(codes), nil
}

func (r *featureRepo) createPackageItems(ctx context.Context, tx *ent.Tx, packageUID string, features []string) error {
	for _, code := range uniqueStrings(features) {
		if !perm.IsFeature(code) {
			return merr.Invalid("invalid feature")
		}
		if _, err := tx.FeaturePackageItem.Create().
			SetUID(id.New()).
			SetPackageUID(packageUID).
			SetFeature(code).
			Save(ctx); err != nil {
			return r.fail(err)
		}
	}
	return nil
}

// orderFeatureCodes 去重后按功能目录顺序排列，目录外的编码被丢弃。
func orderFeatureCodes(codes []string) []string {
	seen := make(map[string]struct{}, len(codes))
	for _, code := range codes {
		seen[code] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for _, code := range perm.FeatureCodes() {
		if _, ok := seen[code]; ok {
			out = append(out, code)
		}
	}
	return out
}
