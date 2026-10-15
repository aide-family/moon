package data

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/biz"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/member"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/memberrole"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/namespace"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/permission"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/predicate"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/role"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/rolepermission"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/user"
	"github.com/aide-family/Alioth/app/alioth/internal/perm"
)

func (r *identityRepo) CreateNamespace(ctx context.Context, name, remark, ownerUID string) (*biz.Namespace, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()

	secret := make([]byte, 24)
	_, _ = rand.Read(secret)
	ns, err := tx.Namespace.Create().
		SetUID(newID()).
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
	for _, roleName := range perm.SystemRoles() {
		created, err := tx.Role.Create().
			SetUID(newID()).
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
				return nil, biz.ErrInvalid
			}
			if _, err := tx.RolePermission.Create().
				SetUID(newID()).
				SetNamespaceUID(ns.UID).
				SetRoleUID(created.UID).
				SetPermissionUID(permissionUID).
				Save(ctx); err != nil {
				return nil, r.fail(err)
			}
		}
	}
	memberRow, err := tx.Member.Create().
		SetUID(newID()).
		SetNamespaceUID(ns.UID).
		SetUserUID(ownerUID).
		SetStatus(member.StatusJOINED).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if _, err := tx.MemberRole.Create().
		SetUID(newID()).
		SetNamespaceUID(ns.UID).
		SetMemberUID(memberRow.UID).
		SetRoleUID(ownerRoleUID).
		Save(ctx); err != nil {
		return nil, r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	item := toNamespace(ns)
	return &item, nil
}

func (r *identityRepo) joinedNamespaceUIDs(ctx context.Context, userUID string) ([]string, error) {
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

func (r *identityRepo) ListUserNamespaces(ctx context.Context, userUID string) ([]biz.Namespace, error) {
	uids, err := r.joinedNamespaceUIDs(ctx, userUID)
	if err != nil || len(uids) == 0 {
		return nil, err
	}
	rows, err := r.client.Namespace.Query().Where(namespace.UIDIn(uids...), namespace.DeletedAtIsNil()).Order(ent.Asc(namespace.FieldName)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return toNamespaces(rows), nil
}

func (r *identityRepo) ListNamespaces(ctx context.Context, userUID string, page, pageSize int, name string) ([]biz.Namespace, int, error) {
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
	return toNamespaces(rows), total, nil
}

func (r *identityRepo) SelectNamespaces(ctx context.Context, userUID string, limit int, lastUID string) ([]biz.SelectItem, error) {
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
	items := make([]biz.SelectItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, biz.SelectItem{UID: row.UID, Name: row.Name})
	}
	return items, nil
}

func (r *identityRepo) GetNamespace(ctx context.Context, uid string) (*biz.Namespace, error) {
	row, err := r.client.Namespace.Query().Where(namespace.UIDEQ(uid), namespace.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toNamespace(row)
	return &item, nil
}

func (r *identityRepo) UpdateNamespace(ctx context.Context, uid, name, remark, logo string) (*biz.Namespace, error) {
	n, err := r.client.Namespace.Update().Where(namespace.UIDEQ(uid), namespace.DeletedAtIsNil()).
		SetName(name).SetRemark(remark).SetLogo(logo).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, biz.ErrNotFound
	}
	return r.GetNamespace(ctx, uid)
}

func (r *identityRepo) UpdateNamespaceStatus(ctx context.Context, uid, status string) (*biz.Namespace, error) {
	value := namespace.StatusACTIVE
	if status == "DISABLED" {
		value = namespace.StatusDISABLED
	} else if status != "ACTIVE" {
		return nil, biz.ErrInvalid
	}
	n, err := r.client.Namespace.Update().Where(namespace.UIDEQ(uid), namespace.DeletedAtIsNil()).SetStatus(value).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, biz.ErrNotFound
	}
	return r.GetNamespace(ctx, uid)
}

func (r *identityRepo) DeleteNamespace(ctx context.Context, uid string) error {
	n, err := r.client.Namespace.Update().Where(namespace.UIDEQ(uid), namespace.DeletedAtIsNil()).SetDeletedAt(time.Now()).Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return biz.ErrNotFound
	}
	return nil
}

func (r *identityRepo) TransferOwner(ctx context.Context, namespaceUID, targetUserUID string) (*biz.Namespace, error) {
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
		return nil, biz.ErrInvalid
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
			SetUID(newID()).
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
			SetUID(newID()).
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

func (r *identityRepo) ListMembers(ctx context.Context, namespaceUID string, page, pageSize int, name string) ([]biz.Member, int, error) {
	preds := []predicate.Member{member.NamespaceUIDEQ(namespaceUID), member.DeletedAtIsNil()}
	if name != "" {
		users, err := r.client.User.Query().Where(user.Or(user.NameContainsFold(name), user.EmailContainsFold(name))).All(ctx)
		if err != nil {
			return nil, 0, r.fail(err)
		}
		if len(users) == 0 {
			return nil, 0, nil
		}
		uids := make([]string, len(users))
		for i, row := range users {
			uids[i] = row.UID
		}
		preds = append(preds, member.UserUIDIn(uids...))
	}
	query := r.client.Member.Query().Where(preds...)
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Asc(member.FieldUID)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items, err := r.memberViews(ctx, namespaceUID, rows)
	return items, total, err
}

func (r *identityRepo) GetMember(ctx context.Context, namespaceUID, uid string) (*biz.Member, error) {
	row, err := r.client.Member.Query().Where(member.NamespaceUIDEQ(namespaceUID), member.UIDEQ(uid), member.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items, err := r.memberViews(ctx, namespaceUID, []*ent.Member{row})
	if err != nil || len(items) == 0 {
		return nil, err
	}
	return &items[0], nil
}

func (r *identityRepo) MemberByUser(ctx context.Context, namespaceUID, userUID string) (*biz.Member, error) {
	row, err := r.client.Member.Query().Where(
		member.NamespaceUIDEQ(namespaceUID),
		member.UserUIDEQ(userUID),
		member.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items, err := r.memberViews(ctx, namespaceUID, []*ent.Member{row})
	if err != nil || len(items) == 0 {
		return nil, err
	}
	return &items[0], nil
}

func (r *identityRepo) SelectMembers(ctx context.Context, namespaceUID string, limit int, lastUID string) ([]biz.SelectItem, error) {
	query := r.client.Member.Query().Where(member.NamespaceUIDEQ(namespaceUID), member.DeletedAtIsNil())
	if lastUID != "" {
		query = query.Where(member.UIDGT(lastUID))
	}
	rows, err := query.Order(ent.Asc(member.FieldUID)).Limit(limit).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items, err := r.memberViews(ctx, namespaceUID, rows)
	if err != nil {
		return nil, err
	}
	out := make([]biz.SelectItem, 0, len(items))
	for _, item := range items {
		label := item.Name
		if label == "" {
			label = item.Email
		}
		out = append(out, biz.SelectItem{UID: item.UID, Name: label})
	}
	return out, nil
}

func (r *identityRepo) InviteMember(ctx context.Context, namespaceUID, email, roleUID string) (*biz.Member, error) {
	account, err := r.client.User.Query().Where(user.EmailEQ(email)).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	exists, err := r.client.Member.Query().Where(
		member.NamespaceUIDEQ(namespaceUID),
		member.UserUIDEQ(account.UID),
		member.DeletedAtIsNil(),
	).Exist(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if exists {
		return nil, biz.ErrConflict
	}
	roleRow, err := r.client.Role.Query().Where(
		role.NamespaceUIDEQ(namespaceUID),
		role.UIDEQ(roleUID),
		role.DeletedAtIsNil(),
		role.StatusEQ(role.StatusACTIVE),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if roleRow.Name == perm.RoleOwner {
		return nil, biz.ErrInvalid
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	created, err := tx.Member.Create().
		SetUID(newID()).
		SetNamespaceUID(namespaceUID).
		SetUserUID(account.UID).
		SetStatus(member.StatusJOINED).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if _, err := tx.MemberRole.Create().
		SetUID(newID()).
		SetNamespaceUID(namespaceUID).
		SetMemberUID(created.UID).
		SetRoleUID(roleRow.UID).
		Save(ctx); err != nil {
		return nil, r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetMember(ctx, namespaceUID, created.UID)
}

func (r *identityRepo) DeleteMember(ctx context.Context, namespaceUID, uid string) error {
	ns, err := r.GetNamespace(ctx, namespaceUID)
	if err != nil {
		return err
	}
	row, err := r.client.Member.Query().Where(member.NamespaceUIDEQ(namespaceUID), member.UIDEQ(uid), member.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return r.fail(err)
	}
	if row.UserUID == ns.OwnerUID {
		return biz.ErrInvalid
	}
	if _, err := r.client.MemberRole.Delete().Where(memberrole.NamespaceUIDEQ(namespaceUID), memberrole.MemberUIDEQ(uid)).Exec(ctx); err != nil {
		return r.fail(err)
	}
	n, err := r.client.Member.Update().Where(member.UIDEQ(uid), member.NamespaceUIDEQ(namespaceUID), member.DeletedAtIsNil()).SetDeletedAt(time.Now()).Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return biz.ErrNotFound
	}
	return nil
}

func (r *identityRepo) UpdateMemberStatus(ctx context.Context, namespaceUID, uid, status string) (*biz.Member, error) {
	value, err := parseMemberStatus(status)
	if err != nil {
		return nil, err
	}
	n, err := r.client.Member.Update().Where(member.NamespaceUIDEQ(namespaceUID), member.UIDEQ(uid), member.DeletedAtIsNil()).SetStatus(value).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, biz.ErrNotFound
	}
	return r.GetMember(ctx, namespaceUID, uid)
}

func (r *identityRepo) UpdateMemberRoles(ctx context.Context, namespaceUID, memberUID string, roleUIDs []string) (*biz.Member, error) {
	row, err := r.client.Member.Query().Where(member.NamespaceUIDEQ(namespaceUID), member.UIDEQ(memberUID), member.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	ns, err := r.GetNamespace(ctx, namespaceUID)
	if err != nil {
		return nil, err
	}
	roles, err := r.client.Role.Query().Where(role.NamespaceUIDEQ(namespaceUID), role.UIDIn(roleUIDs...), role.DeletedAtIsNil()).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if len(roles) != len(uniqueStrings(roleUIDs)) {
		return nil, biz.ErrInvalid
	}
	hasOwner := false
	for _, item := range roles {
		if item.Name == perm.RoleOwner {
			hasOwner = true
		}
	}
	if row.UserUID == ns.OwnerUID && !hasOwner {
		return nil, biz.ErrInvalid
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.MemberRole.Delete().Where(memberrole.NamespaceUIDEQ(namespaceUID), memberrole.MemberUIDEQ(memberUID)).Exec(ctx); err != nil {
		return nil, r.fail(err)
	}
	for _, item := range roles {
		if _, err := tx.MemberRole.Create().
			SetUID(newID()).
			SetNamespaceUID(namespaceUID).
			SetMemberUID(memberUID).
			SetRoleUID(item.UID).
			Save(ctx); err != nil {
			return nil, r.fail(err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetMember(ctx, namespaceUID, memberUID)
}

func (r *identityRepo) ListRoles(ctx context.Context, namespaceUID string, page, pageSize int, name string) ([]biz.Role, int, error) {
	query := r.client.Role.Query().Where(role.NamespaceUIDEQ(namespaceUID), role.DeletedAtIsNil())
	if name != "" {
		query = query.Where(role.NameContainsFold(name))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Asc(role.FieldUID)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items, err := r.roleViews(ctx, rows)
	return items, total, err
}

func (r *identityRepo) GetRole(ctx context.Context, namespaceUID, uid string) (*biz.Role, error) {
	row, err := r.client.Role.Query().Where(role.NamespaceUIDEQ(namespaceUID), role.UIDEQ(uid), role.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items, err := r.roleViews(ctx, []*ent.Role{row})
	if err != nil || len(items) == 0 {
		return nil, err
	}
	return &items[0], nil
}

func (r *identityRepo) SelectRoles(ctx context.Context, namespaceUID string, limit int, lastUID string) ([]biz.SelectItem, error) {
	query := r.client.Role.Query().Where(role.NamespaceUIDEQ(namespaceUID), role.DeletedAtIsNil())
	if lastUID != "" {
		query = query.Where(role.UIDGT(lastUID))
	}
	rows, err := query.Order(ent.Asc(role.FieldUID)).Limit(limit).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]biz.SelectItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, biz.SelectItem{UID: row.UID, Name: row.Name})
	}
	return items, nil
}

func (r *identityRepo) CreateRole(ctx context.Context, namespaceUID, name string) (*biz.Role, error) {
	row, err := r.client.Role.Create().
		SetUID(newID()).
		SetNamespaceUID(namespaceUID).
		SetName(name).
		SetIsSystem(false).
		SetStatus(role.StatusACTIVE).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return r.GetRole(ctx, namespaceUID, row.UID)
}

func (r *identityRepo) UpdateRole(ctx context.Context, namespaceUID, uid, name string) (*biz.Role, error) {
	row, err := r.GetRole(ctx, namespaceUID, uid)
	if err != nil {
		return nil, err
	}
	if row.IsSystem {
		return nil, biz.ErrInvalid
	}
	if _, err := r.client.Role.Update().Where(role.UIDEQ(uid), role.NamespaceUIDEQ(namespaceUID), role.DeletedAtIsNil()).SetName(name).Save(ctx); err != nil {
		return nil, r.fail(err)
	}
	return r.GetRole(ctx, namespaceUID, uid)
}

func (r *identityRepo) UpdateRoleStatus(ctx context.Context, namespaceUID, uid, status string) (*biz.Role, error) {
	value := role.StatusACTIVE
	if status == "DISABLED" {
		value = role.StatusDISABLED
	} else if status != "ACTIVE" {
		return nil, biz.ErrInvalid
	}
	n, err := r.client.Role.Update().Where(role.UIDEQ(uid), role.NamespaceUIDEQ(namespaceUID), role.DeletedAtIsNil()).SetStatus(value).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, biz.ErrNotFound
	}
	return r.GetRole(ctx, namespaceUID, uid)
}

func (r *identityRepo) DeleteRole(ctx context.Context, namespaceUID, uid string) error {
	row, err := r.GetRole(ctx, namespaceUID, uid)
	if err != nil {
		return err
	}
	if row.IsSystem {
		return biz.ErrInvalid
	}
	if _, err := r.client.RolePermission.Delete().Where(rolepermission.NamespaceUIDEQ(namespaceUID), rolepermission.RoleUIDEQ(uid)).Exec(ctx); err != nil {
		return r.fail(err)
	}
	if _, err := r.client.MemberRole.Delete().Where(memberrole.NamespaceUIDEQ(namespaceUID), memberrole.RoleUIDEQ(uid)).Exec(ctx); err != nil {
		return r.fail(err)
	}
	n, err := r.client.Role.Update().Where(role.UIDEQ(uid), role.NamespaceUIDEQ(namespaceUID), role.DeletedAtIsNil()).SetDeletedAt(time.Now()).Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return biz.ErrNotFound
	}
	return nil
}

func (r *identityRepo) UpdateRolePermissions(ctx context.Context, namespaceUID, uid string, names []string) (*biz.Role, error) {
	row, err := r.GetRole(ctx, namespaceUID, uid)
	if err != nil {
		return nil, err
	}
	if row.IsSystem {
		return nil, biz.ErrInvalid
	}
	for _, name := range names {
		if name == perm.NamespaceDelete || name == perm.NamespaceTransfer {
			return nil, biz.ErrInvalid
		}
	}
	rows, err := r.client.Permission.Query().Where(permission.NameIn(names...)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if len(rows) != len(uniqueStrings(names)) {
		return nil, biz.ErrInvalid
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.RolePermission.Delete().Where(rolepermission.NamespaceUIDEQ(namespaceUID), rolepermission.RoleUIDEQ(uid)).Exec(ctx); err != nil {
		return nil, r.fail(err)
	}
	for _, item := range rows {
		if _, err := tx.RolePermission.Create().
			SetUID(newID()).
			SetNamespaceUID(namespaceUID).
			SetRoleUID(uid).
			SetPermissionUID(item.UID).
			Save(ctx); err != nil {
			return nil, r.fail(err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetRole(ctx, namespaceUID, uid)
}

func (r *identityRepo) ListPermissions(ctx context.Context) ([]biz.Permission, error) {
	rows, err := r.client.Permission.Query().Order(ent.Asc(permission.FieldName)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]biz.Permission, 0, len(rows))
	for _, row := range rows {
		items = append(items, biz.Permission{UID: row.UID, Name: row.Name, Resource: row.Resource, Action: row.Action})
	}
	return items, nil
}

func (r *identityRepo) Permissions(ctx context.Context, namespaceUID, userUID string) (map[string]struct{}, error) {
	mem, err := r.client.Member.Query().Where(
		member.NamespaceUIDEQ(namespaceUID),
		member.UserUIDEQ(userUID),
		member.DeletedAtIsNil(),
		member.StatusEQ(member.StatusJOINED),
	).Only(ctx)
	if ent.IsNotFound(err) {
		return map[string]struct{}{}, nil
	}
	if err != nil {
		return nil, r.fail(err)
	}
	links, err := r.client.MemberRole.Query().Where(memberrole.NamespaceUIDEQ(namespaceUID), memberrole.MemberUIDEQ(mem.UID)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if len(links) == 0 {
		return map[string]struct{}{}, nil
	}
	roleUIDs := make([]string, 0, len(links))
	for _, link := range links {
		roleUIDs = append(roleUIDs, link.RoleUID)
	}
	roles, err := r.client.Role.Query().Where(
		role.NamespaceUIDEQ(namespaceUID),
		role.UIDIn(roleUIDs...),
		role.DeletedAtIsNil(),
		role.StatusEQ(role.StatusACTIVE),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if len(roles) == 0 {
		return map[string]struct{}{}, nil
	}
	active := make([]string, 0, len(roles))
	for _, item := range roles {
		active = append(active, item.UID)
	}
	grants, err := r.client.RolePermission.Query().Where(
		rolepermission.NamespaceUIDEQ(namespaceUID),
		rolepermission.RoleUIDIn(active...),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if len(grants) == 0 {
		return map[string]struct{}{}, nil
	}
	permUIDs := make([]string, 0, len(grants))
	for _, grant := range grants {
		permUIDs = append(permUIDs, grant.PermissionUID)
	}
	rows, err := r.client.Permission.Query().Where(permission.UIDIn(uniqueStrings(permUIDs)...)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	out := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		out[row.Name] = struct{}{}
	}
	return out, nil
}

func (r *identityRepo) memberViews(ctx context.Context, namespaceUID string, rows []*ent.Member) ([]biz.Member, error) {
	if len(rows) == 0 {
		return nil, nil
	}
	userUIDs := make([]string, len(rows))
	memberUIDs := make([]string, len(rows))
	for i, row := range rows {
		userUIDs[i] = row.UserUID
		memberUIDs[i] = row.UID
	}
	users, err := r.client.User.Query().Where(user.UIDIn(userUIDs...)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	byUser := make(map[string]*ent.User, len(users))
	for _, row := range users {
		byUser[row.UID] = row
	}
	links, err := r.client.MemberRole.Query().Where(memberrole.NamespaceUIDEQ(namespaceUID), memberrole.MemberUIDIn(memberUIDs...)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	roleUIDs := make([]string, 0, len(links))
	for _, link := range links {
		roleUIDs = append(roleUIDs, link.RoleUID)
	}
	byRole := map[string]*ent.Role{}
	if len(roleUIDs) > 0 {
		roles, err := r.client.Role.Query().Where(role.NamespaceUIDEQ(namespaceUID), role.UIDIn(uniqueStrings(roleUIDs)...), role.DeletedAtIsNil()).All(ctx)
		if err != nil {
			return nil, r.fail(err)
		}
		for _, item := range roles {
			byRole[item.UID] = item
		}
	}
	rolesByMember := map[string][]*ent.Role{}
	for _, link := range links {
		if item, ok := byRole[link.RoleUID]; ok {
			rolesByMember[link.MemberUID] = append(rolesByMember[link.MemberUID], item)
		}
	}
	items := make([]biz.Member, 0, len(rows))
	for _, row := range rows {
		item := biz.Member{
			UID:          row.UID,
			NamespaceUID: row.NamespaceUID,
			UserUID:      row.UserUID,
			Status:       string(row.Status),
		}
		if account := byUser[row.UserUID]; account != nil {
			item.Email = account.Email
			item.Name = account.Name
		}
		for _, linked := range rolesByMember[row.UID] {
			item.RoleUIDs = append(item.RoleUIDs, linked.UID)
			item.RoleNames = append(item.RoleNames, linked.Name)
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *identityRepo) roleViews(ctx context.Context, rows []*ent.Role) ([]biz.Role, error) {
	if len(rows) == 0 {
		return nil, nil
	}
	roleUIDs := make([]string, len(rows))
	for i, row := range rows {
		roleUIDs[i] = row.UID
	}
	links, err := r.client.RolePermission.Query().Where(
		rolepermission.NamespaceUIDEQ(rows[0].NamespaceUID),
		rolepermission.RoleUIDIn(roleUIDs...),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	permUIDs := make([]string, 0, len(links))
	for _, link := range links {
		permUIDs = append(permUIDs, link.PermissionUID)
	}
	byPerm := map[string]string{}
	if len(permUIDs) > 0 {
		perms, err := r.client.Permission.Query().Where(permission.UIDIn(uniqueStrings(permUIDs)...)).All(ctx)
		if err != nil {
			return nil, r.fail(err)
		}
		for _, item := range perms {
			byPerm[item.UID] = item.Name
		}
	}
	namesByRole := map[string][]string{}
	for _, link := range links {
		if name, ok := byPerm[link.PermissionUID]; ok {
			namesByRole[link.RoleUID] = append(namesByRole[link.RoleUID], name)
		}
	}
	items := make([]biz.Role, 0, len(rows))
	for _, row := range rows {
		items = append(items, biz.Role{
			UID:          row.UID,
			NamespaceUID: row.NamespaceUID,
			Name:         row.Name,
			Status:       string(row.Status),
			IsSystem:     row.IsSystem,
			Permissions:  namesByRole[row.UID],
		})
	}
	return items, nil
}

func toNamespace(row *ent.Namespace) biz.Namespace {
	return biz.Namespace{
		UID:      row.UID,
		Name:     row.Name,
		OwnerUID: row.OwnerUID,
		Logo:     row.Logo,
		Remark:   row.Remark,
		Status:   string(row.Status),
	}
}

func toNamespaces(rows []*ent.Namespace) []biz.Namespace {
	items := make([]biz.Namespace, 0, len(rows))
	for _, row := range rows {
		items = append(items, toNamespace(row))
	}
	return items
}

func parseMemberStatus(status string) (member.Status, error) {
	switch status {
	case "JOINED":
		return member.StatusJOINED, nil
	case "INVITED":
		return member.StatusINVITED, nil
	case "EXPIRED":
		return member.StatusEXPIRED, nil
	default:
		return "", biz.ErrInvalid
	}
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
