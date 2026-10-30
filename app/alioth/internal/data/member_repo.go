package data

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/member"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/memberrole"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/predicate"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/role"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/systemmessage"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/user"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

// CountJoined 统计租户当前 JOINED 成员数，用于校验套餐人数上限。
// 邀请中、已移除和已过期的成员都不占名额。
func (r *memberRepo) CountJoined(ctx context.Context, namespaceUID string) (int, error) {
	total, err := r.client.Member.Query().Where(
		member.NamespaceUIDEQ(namespaceUID),
		member.DeletedAtIsNil(),
		member.StatusEQ(member.StatusJOINED),
	).Count(ctx)
	if err != nil {
		return 0, r.fail(err)
	}
	return total, nil
}

func (r *memberRepo) ListMembers(ctx context.Context, namespaceUID string, page, pageSize int, name string) ([]bo.Member, int, error) {
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

// ListMembersByUIDs 只返回当前 Namespace 内、未删除的成员。
// 通知组选择成员时使用，避免跨租户读取成员信息。
func (r *memberRepo) ListMembersByUIDs(ctx context.Context, namespaceUID string, memberUIDs []string) ([]bo.Member, error) {
	if len(memberUIDs) == 0 {
		return nil, nil
	}
	rows, err := r.client.Member.Query().Where(
		member.NamespaceUIDEQ(namespaceUID),
		member.UIDIn(memberUIDs...),
		member.DeletedAtIsNil(),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return r.memberViews(ctx, namespaceUID, rows)
}

func (r *memberRepo) GetMember(ctx context.Context, namespaceUID, uid string) (*bo.Member, error) {
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

func (r *memberRepo) MemberByUser(ctx context.Context, namespaceUID, userUID string) (*bo.Member, error) {
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

func (r *memberRepo) SelectMembers(ctx context.Context, namespaceUID string, limit int, lastUID string) ([]bo.SelectItem, error) {
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
	out := make([]bo.SelectItem, 0, len(items))
	for _, item := range items {
		label := item.Name
		if label == "" {
			label = item.Email
		}
		out = append(out, bo.SelectItem{UID: item.UID, Name: label})
	}
	return out, nil
}

func (r *memberRepo) DeleteMember(ctx context.Context, namespaceUID, uid, actorUID string) error {
	ns, err := r.GetNamespace(ctx, namespaceUID)
	if err != nil {
		return err
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	row, err := tx.Member.Query().Where(member.NamespaceUIDEQ(namespaceUID), member.UIDEQ(uid), member.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return r.fail(err)
	}
	if row.UserUID == ns.OwnerUID {
		return merr.Invalid("invalid request")
	}
	if _, err := tx.MemberRole.Delete().Where(memberrole.NamespaceUIDEQ(namespaceUID), memberrole.MemberUIDEQ(uid)).Exec(ctx); err != nil {
		return r.fail(err)
	}
	n, err := tx.Member.Update().Where(member.UIDEQ(uid), member.NamespaceUIDEQ(namespaceUID), member.DeletedAtIsNil()).SetDeletedAt(time.Now()).Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("not found")
	}
	recipientEmail := ""
	if recipient, err := tx.User.Query().Where(user.UIDEQ(row.UserUID)).Only(ctx); err == nil {
		recipientEmail = recipient.Email
	}
	actorName := "Alioth"
	if actorUID != "" {
		if actor, err := tx.User.Query().Where(user.UIDEQ(actorUID)).Only(ctx); err == nil {
			actorName = actor.Name
			if actorName == "" {
				actorName = actor.Email
			}
		}
	}
	if _, err := tx.SystemMessage.Create().
		SetUID(id.New()).
		SetUserUID(row.UserUID).
		SetEmail(recipientEmail).
		SetKind(systemmessage.KindMEMBER_REMOVED).
		SetActorName(actorName).
		SetNamespaceUID(namespaceUID).
		SetNamespaceName(ns.Name).
		Save(ctx); err != nil {
		return r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return r.fail(err)
	}
	return nil
}

func (r *memberRepo) UpdateMemberStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.Member, error) {
	value, err := parseMemberStatus(status)
	if err != nil {
		return nil, err
	}
	ns, err := r.GetNamespace(ctx, namespaceUID)
	if err != nil {
		return nil, err
	}
	row, err := r.client.Member.Query().Where(member.NamespaceUIDEQ(namespaceUID), member.UIDEQ(uid), member.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if row.UserUID == ns.OwnerUID {
		return nil, merr.Invalid("cannot change owner status")
	}
	n, err := r.client.Member.Update().Where(member.NamespaceUIDEQ(namespaceUID), member.UIDEQ(uid), member.DeletedAtIsNil()).SetStatus(value).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	return r.GetMember(ctx, namespaceUID, uid)
}

func (r *memberRepo) UpdateMemberRoles(ctx context.Context, namespaceUID, memberUID string, roleUIDs []string) (*bo.Member, error) {
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
		return nil, merr.Invalid("invalid request")
	}
	hasOwner := false
	for _, item := range roles {
		if item.Name == perm.RoleOwner {
			hasOwner = true
		}
	}
	if hasOwner && row.UserUID != ns.OwnerUID {
		return nil, merr.Invalid("owner role is assigned by transfer owner")
	}
	if row.UserUID == ns.OwnerUID && !hasOwner {
		return nil, merr.Invalid("invalid request")
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
			SetUID(id.New()).
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

func (r *memberRepo) memberViews(ctx context.Context, namespaceUID string, rows []*ent.Member) ([]bo.Member, error) {
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
	items := make([]bo.Member, 0, len(rows))
	for _, row := range rows {
		item := bo.Member{
			UID:          row.UID,
			NamespaceUID: row.NamespaceUID,
			UserUID:      row.UserUID,
			Status:       string(row.Status),
		}
		if account := byUser[row.UserUID]; account != nil {
			item.Email = account.Email
			item.Phone = account.Phone
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

func parseMemberStatus(status string) (member.Status, error) {
	switch status {
	case "JOINED":
		return member.StatusJOINED, nil
	case "INVITED":
		return member.StatusINVITED, nil
	case "EXPIRED":
		return member.StatusEXPIRED, nil
	default:
		return "", merr.Invalid("invalid request")
	}
}
