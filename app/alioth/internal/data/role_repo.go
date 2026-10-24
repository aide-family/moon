package data

import (
	"context"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/member"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/memberrole"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/permission"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/role"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/rolepermission"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/id"
	"github.com/aide-family/Alioth/app/alioth/pkg/perm"
	"github.com/aide-family/Alioth/pkg/merr"
)

func (r *roleRepo) ListRoles(ctx context.Context, namespaceUID string, page, pageSize int, name string) ([]bo.Role, int, error) {
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

func (r *roleRepo) GetRole(ctx context.Context, namespaceUID, uid string) (*bo.Role, error) {
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

func (r *roleRepo) SelectRoles(ctx context.Context, namespaceUID string, limit int, lastUID string) ([]bo.SelectItem, error) {
	query := r.client.Role.Query().Where(role.NamespaceUIDEQ(namespaceUID), role.DeletedAtIsNil())
	if lastUID != "" {
		query = query.Where(role.UIDGT(lastUID))
	}
	rows, err := query.Order(ent.Asc(role.FieldUID)).Limit(limit).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]bo.SelectItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, bo.SelectItem{UID: row.UID, Name: row.Name})
	}
	return items, nil
}

func (r *roleRepo) CreateRole(ctx context.Context, namespaceUID, name, remark string) (*bo.Role, error) {
	row, err := r.client.Role.Create().
		SetUID(id.New()).
		SetNamespaceUID(namespaceUID).
		SetName(name).
		SetRemark(remark).
		SetIsSystem(false).
		SetStatus(role.StatusACTIVE).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return r.GetRole(ctx, namespaceUID, row.UID)
}

func (r *roleRepo) UpdateRole(ctx context.Context, namespaceUID, uid, name string) (*bo.Role, error) {
	row, err := r.GetRole(ctx, namespaceUID, uid)
	if err != nil {
		return nil, err
	}
	if row.IsSystem {
		return nil, merr.Invalid("invalid request")
	}
	if _, err := r.client.Role.Update().Where(role.UIDEQ(uid), role.NamespaceUIDEQ(namespaceUID), role.DeletedAtIsNil()).SetName(name).Save(ctx); err != nil {
		return nil, r.fail(err)
	}
	return r.GetRole(ctx, namespaceUID, uid)
}

// UpdateRoleRemark 只改备注。这里故意不判断 IsSystem：备注是给成员分配角色时的说明，
// 写它不改变预置角色的权限边界，而系统角色恰恰最需要解释自己是干什么的。
func (r *roleRepo) UpdateRoleRemark(ctx context.Context, namespaceUID, uid, remark string) (*bo.Role, error) {
	n, err := r.client.Role.Update().
		Where(role.UIDEQ(uid), role.NamespaceUIDEQ(namespaceUID), role.DeletedAtIsNil()).
		SetRemark(remark).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	return r.GetRole(ctx, namespaceUID, uid)
}

func (r *roleRepo) UpdateRoleStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.Role, error) {
	value := role.StatusACTIVE
	if status == "DISABLED" {
		value = role.StatusDISABLED
	} else if status != "ACTIVE" {
		return nil, merr.Invalid("invalid request")
	}
	n, err := r.client.Role.Update().Where(role.UIDEQ(uid), role.NamespaceUIDEQ(namespaceUID), role.DeletedAtIsNil()).SetStatus(value).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	return r.GetRole(ctx, namespaceUID, uid)
}

func (r *roleRepo) DeleteRole(ctx context.Context, namespaceUID, uid string) error {
	row, err := r.GetRole(ctx, namespaceUID, uid)
	if err != nil {
		return err
	}
	if row.IsSystem {
		return merr.Invalid("invalid request")
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
		return merr.NotFound("not found")
	}
	return nil
}

func (r *roleRepo) UpdateRolePermissions(ctx context.Context, namespaceUID, uid string, names []string) (*bo.Role, error) {
	row, err := r.GetRole(ctx, namespaceUID, uid)
	if err != nil {
		return nil, err
	}
	if row.IsSystem {
		return nil, merr.Invalid("invalid request")
	}
	for _, name := range names {
		if !perm.IsTenantPermission(name) || name == perm.NamespaceDelete || name == perm.NamespaceTransfer {
			return nil, merr.Invalid("invalid request")
		}
	}
	rows, err := r.client.Permission.Query().Where(permission.NameIn(names...)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if len(rows) != len(uniqueStrings(names)) {
		return nil, merr.Invalid("invalid request")
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
			SetUID(id.New()).
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

func (r *roleRepo) ListPermissions(ctx context.Context) ([]bo.Permission, error) {
	// 只返回租户权限目录。系统权限不在这张表里，避免租户角色误授平台能力。
	rows, err := r.client.Permission.Query().
		Where(permission.NameIn(perm.Names()...)).
		Order(ent.Asc(permission.FieldName)).
		All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]bo.Permission, 0, len(rows))
	for _, row := range rows {
		items = append(items, bo.Permission{UID: row.UID, Name: row.Name, Resource: row.Resource, Action: row.Action})
	}
	return items, nil
}

func (r *roleRepo) Permissions(ctx context.Context, namespaceUID, userUID string) (map[string]struct{}, error) {
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

func (r *roleRepo) roleViews(ctx context.Context, rows []*ent.Role) ([]bo.Role, error) {
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
	items := make([]bo.Role, 0, len(rows))
	for _, row := range rows {
		items = append(items, bo.Role{
			UID:          row.UID,
			NamespaceUID: row.NamespaceUID,
			Name:         row.Name,
			Remark:       row.Remark,
			Status:       string(row.Status),
			IsSystem:     row.IsSystem,
			Permissions:  namesByRole[row.UID],
		})
	}
	return items, nil
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
