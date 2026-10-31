package data

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/featurepackage"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/featurepackageitem"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/namespace"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/namespacefeature"
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

// ListPublicPackages 返回用户创建租户时可选的功能套餐：公开、启用且未删除。
// 按价格升序，同价按名称升序，让免费和低价套餐排在前面。
func (r *featureRepo) ListPublicPackages(ctx context.Context) ([]bo.FeaturePackage, error) {
	rows, err := r.client.FeaturePackage.Query().
		Where(
			featurepackage.DeletedAtIsNil(),
			featurepackage.StatusEQ(featurepackage.StatusACTIVE),
			featurepackage.VisibilityEQ(featurepackage.VisibilityPUBLIC),
		).
		Order(ent.Asc(featurepackage.FieldPriceCents), ent.Asc(featurepackage.FieldName)).
		All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items, err := r.packageViews(ctx, rows)
	if err != nil {
		return nil, err
	}
	// 列表接口对空结果返回空数组而不是 null，前端不需要额外判空。
	if items == nil {
		items = []bo.FeaturePackage{}
	}
	return items, nil
}

func (r *featureRepo) CreatePackage(ctx context.Context, pkg bo.FeaturePackage) (*bo.FeaturePackage, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	currency := featurepackage.Currency(pkg.Currency)
	if pkg.Currency == "" {
		currency = featurepackage.CurrencyCNY
	}
	row, err := tx.FeaturePackage.Create().
		SetUID(id.New()).
		SetName(pkg.Name).
		SetRemark(pkg.Remark).
		SetIsBase(false).
		SetStatus(featurepackage.StatusACTIVE).
		SetDurationDays(pkg.DurationDays).
		SetVisibility(featurepackage.Visibility(pkg.Visibility)).
		SetMaxMembers(pkg.MaxMembers).
		SetPriceCents(pkg.PriceCents).
		SetCurrency(currency).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if err := r.createPackageItems(ctx, tx, row.UID, pkg.Features); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetPackage(ctx, row.UID)
}

// UpdatePackage 同时替换名称、备注、默认时长、版本、人数上限、价格和能力明细。
// 基础套餐的改名由用例层拒绝，这里只负责在一个事务里写入，避免能力明细出现半更新。
func (r *featureRepo) UpdatePackage(ctx context.Context, pkg bo.FeaturePackage) (*bo.FeaturePackage, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	currency := featurepackage.Currency(pkg.Currency)
	if pkg.Currency == "" {
		currency = featurepackage.CurrencyCNY
	}
	n, err := tx.FeaturePackage.Update().
		Where(featurepackage.UIDEQ(pkg.UID), featurepackage.DeletedAtIsNil()).
		SetName(pkg.Name).
		SetRemark(pkg.Remark).
		SetDurationDays(pkg.DurationDays).
		SetVisibility(featurepackage.Visibility(pkg.Visibility)).
		SetMaxMembers(pkg.MaxMembers).
		SetPriceCents(pkg.PriceCents).
		SetCurrency(currency).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	if _, err := tx.FeaturePackageItem.Delete().Where(featurepackageitem.PackageUIDEQ(pkg.UID)).Exec(ctx); err != nil {
		return nil, r.fail(err)
	}
	if err := r.createPackageItems(ctx, tx, pkg.UID, pkg.Features); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetPackage(ctx, pkg.UID)
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

// NamespaceFeatureState 返回租户的套餐、套餐能力、额外能力和到期状态。
// 这里返回的是「已配置」状态：租户停用或套餐到期时 Features 仍然保留，方便平台
// 管理端展示和重新启用；是否真正生效由 EffectiveCapabilities 判断。
func (r *featureRepo) NamespaceFeatureState(ctx context.Context, namespaceUID string) (*bo.NamespaceFeature, error) {
	ns, err := r.client.Namespace.Query().Where(namespace.UIDEQ(namespaceUID), namespace.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	state := &bo.NamespaceFeature{NamespaceUID: namespaceUID, PackageUID: ns.PackageUID}
	// Unix <= 0 的 1970 占位时间按永久处理，和 namespaces 读取保持同一口径。
	if ns.PackageExpiresAt != nil && ns.PackageExpiresAt.Unix() > 0 {
		state.ExpiresAt = *ns.PackageExpiresAt
		state.Expired = !ns.PackageExpiresAt.After(time.Now())
	}
	if ns.PackageUID != "" {
		pkg, err := r.client.FeaturePackage.Query().Where(featurepackage.UIDEQ(ns.PackageUID)).Only(ctx)
		switch {
		case err == nil:
			state.PackageName = pkg.Name
			state.DurationDays = pkg.DurationDays
			// 授权内容读租户快照：管理员之后编辑套餐不会改变已生效租户的权益。
			features, err := namespacePackageCapabilities(ctx, r.client, namespaceUID)
			if err != nil {
				return nil, r.fail(err)
			}
			state.PackageFeatures = features
		case ent.IsNotFound(err):
			// 套餐被物理删除时按空套餐处理；正常删除走软删除，租户仍保留套餐能力。
		default:
			return nil, r.fail(err)
		}
	}
	extra, err := namespaceExtraCapabilityCodes(ctx, r.client, namespaceUID)
	if err != nil {
		return nil, r.fail(err)
	}
	state.ExtraFeatures = extra
	state.Features = perm.EffectiveCapabilities(state.PackageFeatures, state.ExtraFeatures)
	return state, nil
}

// EffectiveCapabilities 返回租户当前生效的能力：套餐能力并上额外能力。
// 租户停用、已归档（deleted_at 非空）或套餐已到期时返回空，配置保留在库里。
func (r *featureRepo) EffectiveCapabilities(ctx context.Context, namespaceUID string) ([]string, error) {
	ns, err := r.client.Namespace.Query().Where(namespace.UIDEQ(namespaceUID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// 租户被物理删除后按空能力处理；鉴权只会因此更严格。
			return nil, nil
		}
		return nil, r.fail(err)
	}
	if ns.DeletedAt != nil || ns.Status == namespace.StatusDISABLED {
		return nil, nil
	}
	if ns.PackageExpiresAt != nil &&
		ns.PackageExpiresAt.Unix() > 0 &&
		!ns.PackageExpiresAt.After(time.Now()) {
		return nil, nil
	}
	packageFeatures, err := namespacePackageCapabilities(ctx, r.client, namespaceUID)
	if err != nil {
		return nil, r.fail(err)
	}
	extra, err := namespaceExtraCapabilityCodes(ctx, r.client, namespaceUID)
	if err != nil {
		return nil, r.fail(err)
	}
	return perm.EffectiveCapabilities(packageFeatures, extra), nil
}

// EffectiveModules 把有效能力反推成模块编码。
func (r *featureRepo) EffectiveModules(ctx context.Context, namespaceUID string) ([]string, error) {
	capabilities, err := r.EffectiveCapabilities(ctx, namespaceUID)
	if err != nil {
		return nil, err
	}
	return perm.ModulesOfCapabilities(capabilities), nil
}

// EffectiveFeatures 是 EffectiveModules 的旧名字，保留给尚未改用能力模型的调用方。
// Deprecated: 新代码使用 EffectiveModules 或 EffectiveCapabilities。
func (r *featureRepo) EffectiveFeatures(ctx context.Context, namespaceUID string) ([]string, error) {
	return r.EffectiveModules(ctx, namespaceUID)
}

// SetNamespaceFeatureState 把目标能力换算成相对套餐的额外能力。
// 能力只增不减：目标必须包含套餐的全部能力，差异只写额外能力（enabled=true）；
// 旧的 enabled=false 记录在这里被全部清掉。同一事务里更新套餐、到期时间和预置角色权限。
func (r *featureRepo) SetNamespaceFeatureState(ctx context.Context, namespaceUID, packageUID string, capabilities []string, expiresAt time.Time) error {
	// 调用方可能还在提交 M24 拆分前的历史能力码；先严格校验每一项可识别，
	// 再归一化成当前能力，避免同义编码被当成缺失的套餐能力。
	for _, code := range capabilities {
		if !perm.IsCapability(code) && !perm.IsLegacyCapability(code) {
			return merr.Client("invalid capability")
		}
	}
	capabilities = perm.NormalizeCapabilities(capabilities)
	target := make(map[string]struct{}, len(capabilities))
	for _, code := range capabilities {
		target[code] = struct{}{}
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	// 先确认租户存在，避免给已删除或不存在的租户留下孤立的能力差异。
	ns, err := tx.Namespace.Query().Where(namespace.UIDEQ(namespaceUID), namespace.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return r.fail(err)
	}
	if _, err := tx.FeaturePackage.Query().Where(featurepackage.UIDEQ(packageUID), featurepackage.DeletedAtIsNil()).Only(ctx); err != nil {
		return r.fail(err)
	}
	// 换套餐时按新套餐当前明细授权并重新快照；同一个套餐只调整额外能力时，
	// 必须沿用租户已有快照，不能让管理员后来修改套餐内容反向影响租户权益。
	var packageFeatures []string
	if ns.PackageUID == packageUID {
		packageFeatures, err = namespacePackageCapabilities(ctx, tx.Client(), namespaceUID)
	} else {
		packageFeatures, err = packageSnapshotCapabilities(ctx, tx.Client(), packageUID)
	}
	if err != nil {
		return r.fail(err)
	}
	included := make(map[string]struct{}, len(packageFeatures))
	for _, code := range packageFeatures {
		included[code] = struct{}{}
		if _, ok := target[code]; !ok {
			return merr.Client("features must include the package features")
		}
	}
	if _, err := tx.NamespaceFeature.Delete().Where(namespacefeature.NamespaceUIDEQ(namespaceUID)).Exec(ctx); err != nil {
		return r.fail(err)
	}
	order := perm.Capabilities()
	for _, item := range order {
		if _, wanted := target[item.Code]; !wanted {
			continue
		}
		if _, inPackage := included[item.Code]; inPackage {
			continue
		}
		if _, err := tx.NamespaceFeature.Create().
			SetUID(id.New()).
			SetNamespaceUID(namespaceUID).
			SetFeature(item.Code).
			SetEnabled(true).
			Save(ctx); err != nil {
			return r.fail(err)
		}
	}
	update := tx.Namespace.Update().
		Where(namespace.UIDEQ(namespaceUID), namespace.DeletedAtIsNil()).
		SetPackageUID(packageUID).
		SetPackageFeatures(packageFeatures)
	if expiresAt.IsZero() {
		update = update.ClearPackageExpiresAt()
	} else {
		update = update.SetPackageExpiresAt(expiresAt)
	}
	if _, err := update.Save(ctx); err != nil {
		return r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return r.fail(err)
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

// ListPackageNamespaces 返回使用该套餐的未删除租户，按名称排序。
func (r *featureRepo) ListPackageNamespaces(ctx context.Context, packageUID string) ([]bo.Namespace, error) {
	rows, err := r.client.Namespace.Query().
		Where(namespace.PackageUIDEQ(packageUID), namespace.DeletedAtIsNil()).
		Order(ent.Asc(namespace.FieldName)).
		All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return toNamespaces(rows), nil
}

// SyncPackageToNamespaces 把套餐当前能力写入使用它的租户快照，并补齐预置角色权限。
// 只更新快照：额外能力、到期时间、价格和成员上限都不动。逐个租户独立事务，
// 避免一个租户失败拖垮整批；返回成功数和跳过数。
func (r *featureRepo) SyncPackageToNamespaces(ctx context.Context, packageUID string, namespaceUIDs []string) (int, int, error) {
	if _, err := r.client.FeaturePackage.Query().
		Where(featurepackage.UIDEQ(packageUID), featurepackage.DeletedAtIsNil()).
		Only(ctx); err != nil {
		return 0, 0, r.fail(err)
	}
	snapshot, err := packageSnapshotCapabilities(ctx, r.client, packageUID)
	if err != nil {
		return 0, 0, r.fail(err)
	}
	requested := uniqueStrings(namespaceUIDs)
	query := r.client.Namespace.Query().
		Where(namespace.PackageUIDEQ(packageUID), namespace.DeletedAtIsNil())
	if len(requested) > 0 {
		query = query.Where(namespace.UIDIn(requested...))
	}
	rows, err := query.Order(ent.Asc(namespace.FieldName)).All(ctx)
	if err != nil {
		return 0, 0, r.fail(err)
	}
	skipped := 0
	if len(requested) > 0 {
		skipped = len(requested) - len(rows)
	}
	synced := 0
	for _, row := range rows {
		if err := r.syncNamespacePackage(ctx, row.UID, snapshot); err != nil {
			return synced, skipped, err
		}
		synced++
	}
	return synced, skipped, nil
}

// syncNamespacePackage 在一个事务里更新租户套餐快照并清理已经并入套餐的额外能力。
func (r *featureRepo) syncNamespacePackage(ctx context.Context, namespaceUID string, snapshot []string) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Namespace.Update().
		Where(namespace.UIDEQ(namespaceUID), namespace.DeletedAtIsNil()).
		SetPackageFeatures(snapshot).
		Save(ctx); err != nil {
		return r.fail(err)
	}
	// 新快照里已经包含的能力不再作为「额外能力」保留，避免租户差异页重复展示。
	if len(snapshot) > 0 {
		if _, err := tx.NamespaceFeature.Delete().
			Where(namespacefeature.NamespaceUIDEQ(namespaceUID), namespacefeature.FeatureIn(snapshot...)).
			Exec(ctx); err != nil {
			return r.fail(err)
		}
	}
	if err := tx.Commit(); err != nil {
		return r.fail(err)
	}
	return nil
}

// NamespaceMemberLimit 返回租户当前套餐的成员数上限。套餐缺失或为空时返回 0（不限制），
// 与「0 表示不限制」的字段语义一致。
func (r *featureRepo) NamespaceMemberLimit(ctx context.Context, namespaceUID string) (int32, error) {
	limit, err := namespaceMemberLimit(ctx, r.client, namespaceUID)
	if err != nil {
		return 0, r.fail(err)
	}
	return limit, nil
}

// namespaceMemberLimit 按租户当前套餐读成员数上限。传 tx.Client() 时在事务内读取，
// 邀请确认用它和成员计数保持同一个快照。套餐已删除或未绑定按不限制处理。
func namespaceMemberLimit(ctx context.Context, client *ent.Client, namespaceUID string) (int32, error) {
	ns, err := client.Namespace.Query().Where(namespace.UIDEQ(namespaceUID), namespace.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return 0, err
	}
	if ns.PackageUID == "" {
		return 0, nil
	}
	pkg, err := client.FeaturePackage.Query().Where(featurepackage.UIDEQ(ns.PackageUID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return 0, nil
		}
		return 0, err
	}
	return pkg.MaxMembers, nil
}

// packageViews 批量补齐套餐的能力明细，并把能力按目录顺序排列。
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
			UID:          row.UID,
			Name:         row.Name,
			Remark:       row.Remark,
			IsBase:       row.IsBase,
			Status:       string(row.Status),
			Visibility:   string(row.Visibility),
			DurationDays: row.DurationDays,
			MaxMembers:   row.MaxMembers,
			PriceCents:   row.PriceCents,
			Currency:     string(row.Currency),
			Features:     capabilityCodesOf(codesByPackage[row.UID]),
			CreatedAt:    row.CreatedAt,
		})
	}
	return out, nil
}

func (r *featureRepo) createPackageItems(ctx context.Context, tx *ent.Tx, packageUID string, features []string) error {
	for _, code := range uniqueStrings(features) {
		if !perm.IsCapability(code) {
			return merr.Invalid("invalid capability")
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

// capabilityCodesOf 把一组编码归一化成当前能力目录里的编码：
// 过滤目录外的编码，把模块级编码展开成该模块的全部能力，并把历史类型级编码
// （例如 M24 拆分前的 notify_channel_webhook / notify_channel_sms）展开成新能力，
// 按能力目录顺序去重返回。存量租户的套餐快照可能是旧编码，读取时必须过这里。
func capabilityCodesOf(codes []string) []string {
	return perm.NormalizeCapabilities(codes)
}

// packageSnapshotCapabilities 返回「现在把该套餐分配给租户」时应写入的能力快照。
// 基础套餐缺少明细时回落到代码里的基础能力，与新建租户的预置角色保持一致；
// 其他套餐只按当前明细快照，空套餐就是空授权。
func packageSnapshotCapabilities(ctx context.Context, client *ent.Client, packageUID string) ([]string, error) {
	if packageUID == "" {
		return nil, nil
	}
	pkg, err := client.FeaturePackage.Query().Where(featurepackage.UIDEQ(packageUID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	codes, err := packageCapabilityCodes(ctx, client, pkg.UID)
	if err != nil {
		return nil, err
	}
	if len(codes) == 0 && pkg.IsBase {
		return perm.BaseCapabilities(), nil
	}
	return codes, nil
}

// namespacePackageCapabilities 返回租户购买套餐时的能力快照。
// 快照为空只出现在升级前创建的租户：回落到套餐当前明细，并在下一次写入时补快照。
func namespacePackageCapabilities(ctx context.Context, client *ent.Client, namespaceUID string) ([]string, error) {
	ns, err := client.Namespace.Query().Where(namespace.UIDEQ(namespaceUID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if len(ns.PackageFeatures) > 0 {
		return capabilityCodesOf(ns.PackageFeatures), nil
	}
	return packageCapabilityCodes(ctx, client, ns.PackageUID)
}

// packageCapabilityCodes 返回套餐携带的能力。packageUID 为空时返回空。
func packageCapabilityCodes(ctx context.Context, client *ent.Client, packageUID string) ([]string, error) {
	if packageUID == "" {
		return nil, nil
	}
	rows, err := client.FeaturePackageItem.Query().Where(featurepackageitem.PackageUIDEQ(packageUID)).All(ctx)
	if err != nil {
		return nil, err
	}
	codes := make([]string, 0, len(rows))
	for _, row := range rows {
		codes = append(codes, row.Feature)
	}
	return capabilityCodesOf(codes), nil
}

// namespaceExtraCapabilityCodes 返回租户额外开通的能力。
// 只认 enabled=true：升级前 enabled=false 的差异记录不再参与计算，语义上能力只增不减。
func namespaceExtraCapabilityCodes(ctx context.Context, client *ent.Client, namespaceUID string) ([]string, error) {
	rows, err := client.NamespaceFeature.Query().
		Where(namespacefeature.NamespaceUIDEQ(namespaceUID), namespacefeature.Enabled(true)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	codes := make([]string, 0, len(rows))
	for _, row := range rows {
		codes = append(codes, row.Feature)
	}
	return capabilityCodesOf(codes), nil
}

// namespaceDataTablePattern 限定归档时动态拼进 SQL 的表名，只允许小写字母、数字和下划线。
// 表名来自 information_schema，仍然做白名单校验，避免任何拼接风险。
var namespaceDataTablePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

// ArchiveNamespace 软删除租户及其全部业务数据。
// 先在同一个事务里标记 namespaces.deleted_at，再按 information_schema 找出所有同时有
// namespace_uid 与 deleted_at 列的表逐表标记；表名经白名单校验后引用，只处理仍未删除的行。
// 归档不可恢复。
func (r *featureRepo) ArchiveNamespace(ctx context.Context, namespaceUID string) error {
	if strings.TrimSpace(namespaceUID) == "" {
		return merr.Invalid("invalid request")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	affected, err := tx.ExecContext(ctx,
		`UPDATE namespaces SET deleted_at = now(), updated_at = now() WHERE uid = $1 AND deleted_at IS NULL`,
		namespaceUID)
	if err != nil {
		return r.fail(err)
	}
	updated, err := affected.RowsAffected()
	if err != nil {
		return r.fail(err)
	}
	if updated == 0 {
		return merr.NotFound("not found")
	}
	tables, err := namespaceDataTables(ctx, tx)
	if err != nil {
		return r.fail(err)
	}
	for _, table := range tables {
		if table == namespace.Table {
			// namespaces 表没有 namespace_uid 列，正常不会出现在结果里；显式跳过，避免重复更新。
			continue
		}
		if !namespaceDataTablePattern.MatchString(table) {
			return merr.Server(fmt.Sprintf("unsupported table name %q", table))
		}
		query := fmt.Sprintf(
			`UPDATE %s SET deleted_at = now() WHERE namespace_uid = $1 AND deleted_at IS NULL`,
			quoteIdentifier(table))
		if _, err := tx.ExecContext(ctx, query, namespaceUID); err != nil {
			return r.fail(err)
		}
	}
	if err := tx.Commit(); err != nil {
		return r.fail(err)
	}
	return nil
}

// namespaceDataTables 返回同时带 namespace_uid 与 deleted_at 列的业务表。
// 分区表在 information_schema 里同样可见，父表 UPDATE 会覆盖子分区，重复处理无害。
func namespaceDataTables(ctx context.Context, tx *sql.Tx) ([]string, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT c.table_name
FROM information_schema.columns c
JOIN information_schema.tables t
  ON t.table_schema = c.table_schema AND t.table_name = c.table_name
WHERE c.table_schema = current_schema()
  AND t.table_type = 'BASE TABLE'
  AND c.column_name IN ('namespace_uid', 'deleted_at')
GROUP BY c.table_name
HAVING COUNT(DISTINCT c.column_name) = 2
ORDER BY c.table_name`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	tables := make([]string, 0, 32)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, rows.Err()
}

// quoteIdentifier 给 PostgreSQL 标识符加双引号；表名已经过白名单校验。
func quoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
