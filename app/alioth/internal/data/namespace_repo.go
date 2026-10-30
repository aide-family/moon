package data

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/featurepackage"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/featurepackageitem"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/member"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/memberrole"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/namespace"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/predicate"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/role"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/user"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

func (r *namespaceRepo) CreateNamespace(ctx context.Context, name, remark, ownerUID, packageUID string) (*bo.Namespace, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	pkg, err := tx.FeaturePackage.Query().
		Where(featurepackage.UIDEQ(packageUID), featurepackage.DeletedAtIsNil(), featurepackage.StatusEQ(featurepackage.StatusACTIVE)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, merr.NotFound("package not found")
		}
		return nil, r.fail(err)
	}
	// 直接创建按套餐当前的授权时长计算到期时间；订单支付传的是下单时的时长快照。
	namespaceUID, err := r.createNamespaceTx(ctx, tx, name, remark, ownerUID, pkg, pkg.DurationDays)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetNamespace(ctx, namespaceUID)
}

// createNamespaceTx 在调用方的事务里创建一个完整租户：绑定套餐、按套餐能力创建预置角色、
// 写入负责人成员和 owner 角色。直接把租户创建和订单支付放在同一个事务里，避免出现
// 「租户建好但订单没置为已支付」的中间状态。调用方负责提交事务。
// durationDays 是本次授权的时长（0 表示永久），由调用方决定用套餐当前值还是订单快照。
func (base *repoBase) createNamespaceTx(ctx context.Context, tx *ent.Tx, name, remark, ownerUID string, pkg *ent.FeaturePackage, durationDays int32) (string, error) {
	// 预置角色跟随套餐当前的能力，而不是代码里的常量：超级管理员调整套餐后，
	// 新建租户的角色权限要立即一致。只有基础套餐在明细缺失时才回落到代码默认能力。
	items, err := tx.FeaturePackageItem.Query().
		Where(featurepackageitem.PackageUIDEQ(pkg.UID)).
		All(ctx)
	if err != nil {
		return "", base.fail(err)
	}
	codes := make([]string, 0, len(items))
	for _, item := range items {
		codes = append(codes, item.Feature)
	}
	capabilities := capabilityCodesOf(codes)
	if len(capabilities) == 0 && pkg.IsBase {
		capabilities = perm.BaseCapabilities()
	}
	secret := make([]byte, 24)
	_, _ = rand.Read(secret)
	create := tx.Namespace.Create().
		SetUID(id.New()).
		SetName(name).
		SetRemark(remark).
		SetOwnerUID(ownerUID).
		SetPackageUID(pkg.UID).
		SetSecret(hex.EncodeToString(secret)).
		SetStatus(namespace.StatusACTIVE)
	if durationDays > 0 {
		create = create.SetPackageExpiresAt(time.Now().Add(time.Duration(durationDays) * 24 * time.Hour))
	}
	ns, err := create.Save(ctx)
	if err != nil {
		return "", base.fail(err)
	}
	catalog, err := tx.Permission.Query().All(ctx)
	if err != nil {
		return "", base.fail(err)
	}
	byName := make(map[string]string, len(catalog))
	for _, item := range catalog {
		byName[item.Name] = item.UID
	}
	var ownerRoleUID string
	for _, roleName := range perm.TenantRoles() {
		created, err := tx.Role.Create().
			SetUID(id.New()).
			SetNamespaceUID(ns.UID).
			SetName(roleName).
			SetRemark(perm.DefaultRoleRemark(roleName)).
			SetIsSystem(true).
			SetStatus(role.StatusACTIVE).
			Save(ctx)
		if err != nil {
			return "", base.fail(err)
		}
		if roleName == perm.RoleOwner {
			ownerRoleUID = created.UID
		}
		// 预置角色只授套餐能力范围内的权限；其余能力由租户开通后再授予角色。
		for _, code := range perm.RolePermissionsForCapabilities(roleName, capabilities) {
			permissionUID, ok := byName[code]
			if !ok {
				return "", merr.Invalid("invalid request")
			}
			if _, err := tx.RolePermission.Create().
				SetUID(id.New()).
				SetNamespaceUID(ns.UID).
				SetRoleUID(created.UID).
				SetPermissionUID(permissionUID).
				Save(ctx); err != nil {
				return "", base.fail(err)
			}
		}
	}
	memberRow, err := tx.Member.Create().
		SetUID(id.New()).
		SetNamespaceUID(ns.UID).
		SetUserUID(ownerUID).
		SetStatus(member.StatusJOINED).
		Save(ctx)
	if err != nil {
		return "", base.fail(err)
	}
	if _, err := tx.MemberRole.Create().
		SetUID(id.New()).
		SetNamespaceUID(ns.UID).
		SetMemberUID(memberRow.UID).
		SetRoleUID(ownerRoleUID).
		Save(ctx); err != nil {
		return "", base.fail(err)
	}
	return ns.UID, nil
}

// GetNamespace 返回租户，并补齐负责人的名称和邮箱。
func (r *namespaceRepo) GetNamespace(ctx context.Context, uid string) (*bo.Namespace, error) {
	item, err := r.repoBase.GetNamespace(ctx, uid)
	if err != nil {
		return nil, err
	}
	items := []bo.Namespace{*item}
	if err := r.fillOwners(ctx, items); err != nil {
		return nil, err
	}
	if err := r.fillPackageInfo(ctx, items); err != nil {
		return nil, err
	}
	return &items[0], nil
}

// fillOwners 按负责人 UID 批量补齐用户展示信息，避免逐行查询。
func (r *namespaceRepo) fillOwners(ctx context.Context, items []bo.Namespace) error {
	ownerUIDs := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if item.OwnerUID == "" {
			continue
		}
		if _, ok := seen[item.OwnerUID]; ok {
			continue
		}
		seen[item.OwnerUID] = struct{}{}
		ownerUIDs = append(ownerUIDs, item.OwnerUID)
	}
	if len(ownerUIDs) == 0 {
		return nil
	}
	rows, err := r.client.User.Query().Where(user.UIDIn(ownerUIDs...)).All(ctx)
	if err != nil {
		return r.fail(err)
	}
	byUID := make(map[string]*ent.User, len(rows))
	for _, row := range rows {
		byUID[row.UID] = row
	}
	for i := range items {
		row, ok := byUID[items[i].OwnerUID]
		if !ok {
			continue
		}
		items[i].OwnerName = row.Name
		items[i].OwnerEmail = row.Email
	}
	return nil
}

// fillPackageInfo 按套餐 UID 批量补齐套餐名称和成员数上限。套餐可能已被软删除，
// 这里不过滤删除标记，保证历史租户仍能显示当初选择的套餐名；已删除套餐的容量按不限制处理。
func (r *namespaceRepo) fillPackageInfo(ctx context.Context, items []bo.Namespace) error {
	packageUIDs := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if item.PackageUID == "" {
			continue
		}
		if _, ok := seen[item.PackageUID]; ok {
			continue
		}
		seen[item.PackageUID] = struct{}{}
		packageUIDs = append(packageUIDs, item.PackageUID)
	}
	if len(packageUIDs) == 0 {
		return nil
	}
	rows, err := r.client.FeaturePackage.Query().Where(featurepackage.UIDIn(packageUIDs...)).All(ctx)
	if err != nil {
		return r.fail(err)
	}
	byUID := make(map[string]*ent.FeaturePackage, len(rows))
	for _, row := range rows {
		byUID[row.UID] = row
	}
	for i := range items {
		row, ok := byUID[items[i].PackageUID]
		if !ok {
			continue
		}
		items[i].PackageName = row.Name
		items[i].MaxMembers = row.MaxMembers
	}
	return nil
}

func (r *namespaceRepo) joinedNamespaceUIDs(ctx context.Context, userUID string) ([]string, error) {
	rows, err := r.client.Member.Query().Where(
		member.UserUIDEQ(userUID),
		member.DeletedAtIsNil(),
		member.StatusEQ(member.StatusJOINED),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	uids := make([]string, 0, len(rows))
	for _, row := range rows {
		uids = append(uids, row.NamespaceUID)
	}
	return uids, nil
}

func (r *namespaceRepo) ListUserNamespaces(ctx context.Context, userUID string) ([]bo.Namespace, error) {
	// 停用租户不可切换进入：只列出 ACTIVE 的租户，停用后前端会自动回落到其他租户。
	return r.userNamespaces(ctx, userUID, false)
}

// ListUserNamespacesIncludingDisabled 保留停用租户，仅用于解释切换失败的原因。
func (r *namespaceRepo) ListUserNamespacesIncludingDisabled(ctx context.Context, userUID string) ([]bo.Namespace, error) {
	return r.userNamespaces(ctx, userUID, true)
}

func (r *namespaceRepo) userNamespaces(ctx context.Context, userUID string, includeDisabled bool) ([]bo.Namespace, error) {
	uids, err := r.joinedNamespaceUIDs(ctx, userUID)
	if err != nil || len(uids) == 0 {
		return nil, err
	}
	preds := []predicate.Namespace{namespace.UIDIn(uids...), namespace.DeletedAtIsNil()}
	if !includeDisabled {
		preds = append(preds, namespace.StatusNEQ(namespace.StatusDISABLED))
	}
	rows, err := r.client.Namespace.Query().
		Where(preds...).
		Order(ent.Asc(namespace.FieldName)).
		All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := toNamespaces(rows)
	if err := r.fillOwners(ctx, items); err != nil {
		return nil, err
	}
	// 切换器和成员页要显示「已用 / 上限」，所以用户自己的租户列表也补齐容量和成员数。
	if err := r.fillMemberCounts(ctx, items); err != nil {
		return nil, err
	}
	if err := r.fillPackageInfo(ctx, items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *namespaceRepo) ListNamespaces(ctx context.Context, userUID string, page, pageSize int, name string) ([]bo.Namespace, int, error) {
	uids, err := r.joinedNamespaceUIDs(ctx, userUID)
	if err != nil || len(uids) == 0 {
		return nil, 0, err
	}
	preds := []predicate.Namespace{namespace.UIDIn(uids...), namespace.DeletedAtIsNil()}
	if name != "" {
		preds = append(preds, namespace.NameContainsFold(name))
	}
	query := r.client.Namespace.Query().Where(preds...)
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Asc(namespace.FieldUID)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := toNamespaces(rows)
	if err := r.fillOwners(ctx, items); err != nil {
		return nil, 0, err
	}
	if err := r.fillMemberCounts(ctx, items); err != nil {
		return nil, 0, err
	}
	if err := r.fillPackageInfo(ctx, items); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// ListAllNamespaces 列出平台全部租户，供系统管理员查看。不按成员身份过滤。
// 成员数与套餐容量在租户列表里补齐；GetNamespace 在每次业务请求的租户校验里都会调用，
// 不做额外统计，只有平台列表和成员自己的租户列表需要这两项展示信息。
func (r *namespaceRepo) ListAllNamespaces(ctx context.Context, page, pageSize int, name, status string) ([]bo.Namespace, int, error) {
	preds := []predicate.Namespace{namespace.DeletedAtIsNil()}
	if name != "" {
		preds = append(preds, namespace.NameContainsFold(name))
	}
	if status != "" {
		preds = append(preds, namespace.StatusEQ(namespace.Status(status)))
	}
	query := r.client.Namespace.Query().Where(preds...)
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Asc(namespace.FieldUID)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := toNamespaces(rows)
	if err := r.fillOwners(ctx, items); err != nil {
		return nil, 0, err
	}
	if err := r.fillMemberCounts(ctx, items); err != nil {
		return nil, 0, err
	}
	if err := r.fillPackageInfo(ctx, items); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// fillMemberCounts 按租户批量统计 JOINED 成员数。未确认的邀请和已移除的成员不计入。
func (r *namespaceRepo) fillMemberCounts(ctx context.Context, items []bo.Namespace) error {
	if len(items) == 0 {
		return nil
	}
	uids := make([]string, 0, len(items))
	for _, item := range items {
		uids = append(uids, item.UID)
	}
	rows, err := r.client.Member.Query().Where(
		member.NamespaceUIDIn(uids...),
		member.DeletedAtIsNil(),
		member.StatusEQ(member.StatusJOINED),
	).All(ctx)
	if err != nil {
		return r.fail(err)
	}
	counts := make(map[string]int, len(rows))
	for _, row := range rows {
		counts[row.NamespaceUID]++
	}
	for i := range items {
		items[i].MemberCount = counts[items[i].UID]
	}
	return nil
}

func (r *namespaceRepo) SelectNamespaces(ctx context.Context, userUID string, limit int, lastUID string) ([]bo.SelectItem, error) {
	uids, err := r.joinedNamespaceUIDs(ctx, userUID)
	if err != nil || len(uids) == 0 {
		return nil, err
	}
	query := r.client.Namespace.Query().
		Where(namespace.UIDIn(uids...), namespace.DeletedAtIsNil(), namespace.StatusNEQ(namespace.StatusDISABLED))
	if lastUID != "" {
		query = query.Where(namespace.UIDGT(lastUID))
	}
	rows, err := query.Order(ent.Asc(namespace.FieldUID)).Limit(limit).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]bo.SelectItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, bo.SelectItem{UID: row.UID, Name: row.Name})
	}
	return items, nil
}

func (r *namespaceRepo) UpdateNamespace(ctx context.Context, uid, name, remark, logo string) (*bo.Namespace, error) {
	n, err := r.client.Namespace.Update().Where(namespace.UIDEQ(uid), namespace.DeletedAtIsNil()).
		SetName(name).SetRemark(remark).SetLogo(logo).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	return r.GetNamespace(ctx, uid)
}

func (r *namespaceRepo) UpdateNamespaceStatus(ctx context.Context, uid, status string) (*bo.Namespace, error) {
	value := namespace.StatusACTIVE
	if status == "DISABLED" {
		value = namespace.StatusDISABLED
	} else if status != "ACTIVE" {
		return nil, merr.Invalid("invalid request")
	}
	n, err := r.client.Namespace.Update().Where(namespace.UIDEQ(uid), namespace.DeletedAtIsNil()).SetStatus(value).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	return r.GetNamespace(ctx, uid)
}

// SetNamespaceStatus 是平台停用/启用租户的入口。停用必须带原因，启用清空原因；
// 状态与原因在同一个更新里写入，避免出现停用但没有原因的中间状态。
func (r *namespaceRepo) SetNamespaceStatus(ctx context.Context, uid, status, reason string) (*bo.Namespace, error) {
	value := namespace.StatusACTIVE
	reason = strings.TrimSpace(reason)
	switch status {
	case "ACTIVE":
		reason = ""
	case "DISABLED":
		if reason == "" {
			return nil, merr.Invalid("invalid request")
		}
		value = namespace.StatusDISABLED
	default:
		return nil, merr.Invalid("invalid request")
	}
	n, err := r.client.Namespace.Update().
		Where(namespace.UIDEQ(uid), namespace.DeletedAtIsNil()).
		SetStatus(value).
		SetDisableReason(reason).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	return r.GetNamespace(ctx, uid)
}

func (r *namespaceRepo) DeleteNamespace(ctx context.Context, uid string) error {
	n, err := r.client.Namespace.Update().Where(namespace.UIDEQ(uid), namespace.DeletedAtIsNil()).SetDeletedAt(time.Now()).Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("not found")
	}
	return nil
}

func (r *namespaceRepo) TransferOwner(ctx context.Context, namespaceUID, targetUserUID string) (*bo.Namespace, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	ns, err := tx.Namespace.Query().Where(namespace.UIDEQ(namespaceUID), namespace.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if ns.OwnerUID == targetUserUID {
		return nil, merr.Invalid("invalid request")
	}
	current, err := tx.Member.Query().Where(
		member.NamespaceUIDEQ(namespaceUID),
		member.UserUIDEQ(ns.OwnerUID),
		member.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	target, err := tx.Member.Query().Where(
		member.NamespaceUIDEQ(namespaceUID),
		member.UserUIDEQ(targetUserUID),
		member.DeletedAtIsNil(),
		member.StatusEQ(member.StatusJOINED),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	ownerRole, err := tx.Role.Query().Where(role.NamespaceUIDEQ(namespaceUID), role.NameEQ(perm.RoleOwner), role.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	adminRole, err := tx.Role.Query().Where(role.NamespaceUIDEQ(namespaceUID), role.NameEQ(perm.RoleAdmin), role.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if _, err := tx.MemberRole.Delete().Where(
		memberrole.NamespaceUIDEQ(namespaceUID),
		memberrole.MemberUIDEQ(current.UID),
		memberrole.RoleUIDEQ(ownerRole.UID),
	).Exec(ctx); err != nil {
		return nil, r.fail(err)
	}
	hasAdmin, err := tx.MemberRole.Query().Where(
		memberrole.NamespaceUIDEQ(namespaceUID),
		memberrole.MemberUIDEQ(current.UID),
		memberrole.RoleUIDEQ(adminRole.UID),
	).Exist(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if !hasAdmin {
		if _, err := tx.MemberRole.Create().
			SetUID(id.New()).
			SetNamespaceUID(namespaceUID).
			SetMemberUID(current.UID).
			SetRoleUID(adminRole.UID).
			Save(ctx); err != nil {
			return nil, r.fail(err)
		}
	}
	hasOwner, err := tx.MemberRole.Query().Where(
		memberrole.NamespaceUIDEQ(namespaceUID),
		memberrole.MemberUIDEQ(target.UID),
		memberrole.RoleUIDEQ(ownerRole.UID),
	).Exist(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if !hasOwner {
		if _, err := tx.MemberRole.Create().
			SetUID(id.New()).
			SetNamespaceUID(namespaceUID).
			SetMemberUID(target.UID).
			SetRoleUID(ownerRole.UID).
			Save(ctx); err != nil {
			return nil, r.fail(err)
		}
	}
	if _, err := tx.Namespace.UpdateOne(ns).SetOwnerUID(targetUserUID).Save(ctx); err != nil {
		return nil, r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetNamespace(ctx, namespaceUID)
}

func toNamespace(row *ent.Namespace) bo.Namespace {
	item := bo.Namespace{
		UID:           row.UID,
		Name:          row.Name,
		OwnerUID:      row.OwnerUID,
		Logo:          row.Logo,
		Remark:        row.Remark,
		Status:        string(row.Status),
		CreatedAt:     row.CreatedAt,
		PackageUID:    row.PackageUID,
		DisableReason: row.DisableReason,
	}
	if row.PackageExpiresAt != nil {
		item.PackageExpiresAt = *row.PackageExpiresAt
	}
	return item
}

func toNamespaces(rows []*ent.Namespace) []bo.Namespace {
	items := make([]bo.Namespace, 0, len(rows))
	for _, row := range rows {
		items = append(items, toNamespace(row))
	}
	return items
}
