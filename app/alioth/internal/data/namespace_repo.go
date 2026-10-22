package data

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/member"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/memberrole"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/namespace"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/predicate"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/role"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/user"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/id"
	"github.com/aide-family/Alioth/app/alioth/pkg/perm"
	"github.com/aide-family/Alioth/pkg/merr"
)

func (r *namespaceRepo) CreateNamespace(ctx context.Context, name, remark, ownerUID string) (*bo.Namespace, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()

	secret := make([]byte, 24)
	_, _ = rand.Read(secret)
	ns, err := tx.Namespace.Create().
		SetUID(id.New()).
		SetName(name).
		SetRemark(remark).
		SetOwnerUID(ownerUID).
		SetSecret(hex.EncodeToString(secret)).
		SetStatus(namespace.StatusACTIVE).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	catalog, err := tx.Permission.Query().All(ctx)
	if err != nil {
		return nil, r.fail(err)
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
			SetIsSystem(true).
			SetStatus(role.StatusACTIVE).
			Save(ctx)
		if err != nil {
			return nil, r.fail(err)
		}
		if roleName == perm.RoleOwner {
			ownerRoleUID = created.UID
		}
		for _, code := range perm.RolePermissions(roleName) {
			permissionUID, ok := byName[code]
			if !ok {
				return nil, merr.Invalid("invalid request")
			}
			if _, err := tx.RolePermission.Create().
				SetUID(id.New()).
				SetNamespaceUID(ns.UID).
				SetRoleUID(created.UID).
				SetPermissionUID(permissionUID).
				Save(ctx); err != nil {
				return nil, r.fail(err)
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
		return nil, r.fail(err)
	}
	if _, err := tx.MemberRole.Create().
		SetUID(id.New()).
		SetNamespaceUID(ns.UID).
		SetMemberUID(memberRow.UID).
		SetRoleUID(ownerRoleUID).
		Save(ctx); err != nil {
		return nil, r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetNamespace(ctx, ns.UID)
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
	uids, err := r.joinedNamespaceUIDs(ctx, userUID)
	if err != nil || len(uids) == 0 {
		return nil, err
	}
	rows, err := r.client.Namespace.Query().Where(namespace.UIDIn(uids...), namespace.DeletedAtIsNil()).Order(ent.Asc(namespace.FieldName)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := toNamespaces(rows)
	if err := r.fillOwners(ctx, items); err != nil {
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
	return items, total, nil
}

func (r *namespaceRepo) SelectNamespaces(ctx context.Context, userUID string, limit int, lastUID string) ([]bo.SelectItem, error) {
	uids, err := r.joinedNamespaceUIDs(ctx, userUID)
	if err != nil || len(uids) == 0 {
		return nil, err
	}
	query := r.client.Namespace.Query().Where(namespace.UIDIn(uids...), namespace.DeletedAtIsNil())
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
	return bo.Namespace{
		UID:      row.UID,
		Name:     row.Name,
		OwnerUID: row.OwnerUID,
		Logo:     row.Logo,
		Remark:   row.Remark,
		Status:   string(row.Status),
	}
}

func toNamespaces(rows []*ent.Namespace) []bo.Namespace {
	items := make([]bo.Namespace, 0, len(rows))
	for _, row := range rows {
		items = append(items, toNamespace(row))
	}
	return items
}
