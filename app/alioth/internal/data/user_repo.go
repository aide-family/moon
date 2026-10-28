package data

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/member"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/memberrole"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/namespace"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationgroup"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/role"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/user"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

func (r *userRepo) CreateUser(ctx context.Context, in bo.Account) (*bo.Account, error) {
	status := user.StatusACTIVE
	if in.Status == "BANNED" {
		status = user.StatusBANNED
	}
	count, err := r.client.User.Query().Count(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	systemRole := user.SystemRoleUSER
	if count == 0 {
		systemRole = user.SystemRoleSUPER_ADMIN
	}
	create := func(role user.SystemRole) (*ent.User, error) {
		return r.client.User.Create().
			SetUID(in.UID).
			SetName(in.Name).
			SetNickname(in.Nickname).
			SetEmail(in.Email).
			SetAvatar(in.Avatar).
			SetPhone(in.Phone).
			SetRemark(in.Remark).
			SetStatus(status).
			SetSystemRole(role).
			Save(ctx)
	}
	row, err := create(systemRole)
	if err != nil && systemRole == user.SystemRoleSUPER_ADMIN {
		// 并发注册时另一个请求已经拿到唯一超级管理员，当前用户按普通用户创建。
		if ent.IsConstraintError(err) {
			row, err = create(user.SystemRoleUSER)
		}
	}
	if err != nil {
		return nil, r.fail(err)
	}
	account := toAccount(row)
	return &account, nil
}

func (r *userRepo) UserByEmail(ctx context.Context, email string) (*bo.Account, error) {
	row, err := r.client.User.Query().Where(user.EmailEQ(email)).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	account := toAccount(row)
	return &account, nil
}

func (r *userRepo) UserByUID(ctx context.Context, uid string) (*bo.Account, error) {
	row, err := r.client.User.Query().Where(user.UIDEQ(uid)).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	account := toAccount(row)
	return &account, nil
}

func (r *userRepo) UpdateUserProfile(ctx context.Context, uid string, email, phone, avatar, remark *string) (*bo.Account, error) {
	update := r.client.User.Update().Where(user.UIDEQ(uid))
	if email != nil {
		update.SetEmail(*email)
	}
	if phone != nil {
		update.SetPhone(*phone)
	}
	if avatar != nil {
		update.SetAvatar(*avatar)
	}
	if remark != nil {
		update.SetRemark(*remark)
	}
	n, err := update.Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	return r.UserByUID(ctx, uid)
}

func (r *userRepo) UpdateUserStatus(ctx context.Context, uid, status string) (*bo.Account, error) {
	value := user.StatusACTIVE
	if status == "BANNED" {
		value = user.StatusBANNED
	} else if status != "ACTIVE" {
		return nil, merr.Invalid("invalid request")
	}
	n, err := r.client.User.Update().Where(user.UIDEQ(uid)).SetStatus(value).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	return r.UserByUID(ctx, uid)
}

func (r *userRepo) UpdateUserSystemRole(ctx context.Context, uid, role, authorizedBy string) (*bo.Account, error) {
	value := user.SystemRole(role)
	switch value {
	case user.SystemRoleSYSTEM_ADMIN, user.SystemRoleUSER:
	default:
		return nil, merr.Invalid("invalid request")
	}
	n, err := r.client.User.Update().
		Where(user.UIDEQ(uid)).
		SetSystemRole(value).
		SetAuthorizedBy(authorizedBy).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("not found")
	}
	return r.UserByUID(ctx, uid)
}

func (r *userRepo) ListUsers(ctx context.Context, page, pageSize int, name string) ([]bo.Account, int, error) {
	query := r.client.User.Query()
	if name != "" {
		query = query.Where(user.Or(user.NameContainsFold(name), user.EmailContainsFold(name)))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Asc(user.FieldUID)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.Account, 0, len(rows))
	for _, row := range rows {
		items = append(items, toAccount(row))
	}
	return items, total, nil
}

func (r *userRepo) SelectUsers(ctx context.Context, limit int, lastUID string) ([]bo.SelectItem, error) {
	query := r.client.User.Query()
	if lastUID != "" {
		query = query.Where(user.UIDGT(lastUID))
	}
	rows, err := query.Order(ent.Asc(user.FieldUID)).Limit(limit).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]bo.SelectItem, 0, len(rows))
	for _, row := range rows {
		label := row.Name
		if label == "" {
			label = row.Email
		}
		items = append(items, bo.SelectItem{UID: row.UID, Name: label})
	}
	return items, nil
}

func toAccount(row *ent.User) bo.Account {
	return bo.Account{
		UID:          row.UID,
		Name:         row.Name,
		Nickname:     row.Nickname,
		Email:        row.Email,
		Phone:        row.Phone,
		Avatar:       row.Avatar,
		Remark:       row.Remark,
		Status:       string(row.Status),
		SystemRole:   string(row.SystemRole),
		AuthorizedBy: row.AuthorizedBy,
	}
}

// MergeAccounts 把源账号的租户成员、租户角色、租户负责人和平台系统角色迁移到目标账号。
// 第三方历史占位账号（@users.alioth.local）补真实邮箱时，如果邮箱已经属于另一个账号，
// 会调用它把占位账号上的成员关系并过去。整个迁移放在一个事务里：任何一步失败都回滚，
// 避免出现成员搬了一半、租户没有负责人或者超级管理员空缺的中间态。
// 个人数据不迁移：user_alert_page（实时告警页选择）、alert_event_column（列表列）、
// system_message（系统消息）等仍挂在源账号上；源账号保留为空账号供平台管理员查看。
func (r *userRepo) MergeAccounts(ctx context.Context, fromUID, toUID string) error {
	if fromUID == "" || toUID == "" || fromUID == toUID {
		return merr.Invalid("invalid request")
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	from, err := tx.User.Query().Where(user.UIDEQ(fromUID)).Only(ctx)
	if err != nil {
		return r.fail(err)
	}
	to, err := tx.User.Query().Where(user.UIDEQ(toUID)).Only(ctx)
	if err != nil {
		return r.fail(err)
	}
	if err := r.mergeSystemRole(ctx, tx, from, to); err != nil {
		return err
	}
	members, err := tx.Member.Query().Where(member.UserUIDEQ(fromUID), member.DeletedAtIsNil()).All(ctx)
	if err != nil {
		return r.fail(err)
	}
	for _, source := range members {
		if err := r.mergeMember(ctx, tx, source, toUID); err != nil {
			return err
		}
	}
	// 成员迁完之后再处理租户负责人：目标此时可能已经在该租户里有成员身份，
	// 也可能刚由源成员改写而来，统一在这里补齐 owner 角色。
	namespaces, err := tx.Namespace.Query().Where(namespace.OwnerUIDEQ(fromUID), namespace.DeletedAtIsNil()).All(ctx)
	if err != nil {
		return r.fail(err)
	}
	for _, space := range namespaces {
		if err := r.moveNamespaceOwner(ctx, tx, space, toUID); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return r.fail(err)
	}
	return nil
}

// mergeSystemRole 把源账号的平台角色让给目标账号。
// 超级管理员有唯一部分索引，from 更高时必须先把源降成 USER 再提升目标，顺序反了会撞索引。
func (r *userRepo) mergeSystemRole(ctx context.Context, tx *ent.Tx, from, to *ent.User) error {
	if from.SystemRole == user.SystemRoleUSER {
		return nil
	}
	// 目标已经同等或更高时只需要把源降下来，避免空账号留着平台管理权限。
	if _, err := tx.User.Update().Where(user.UIDEQ(from.UID)).SetSystemRole(user.SystemRoleUSER).Save(ctx); err != nil {
		return r.fail(err)
	}
	if systemRoleRank(from.SystemRole) > systemRoleRank(to.SystemRole) {
		if _, err := tx.User.Update().Where(user.UIDEQ(to.UID)).SetSystemRole(from.SystemRole).Save(ctx); err != nil {
			return r.fail(err)
		}
	}
	return nil
}

// systemRoleRank 把平台角色换成可比较的档位，数值越大权限越高。
func systemRoleRank(value user.SystemRole) int {
	switch value {
	case user.SystemRoleSUPER_ADMIN:
		return 3
	case user.SystemRoleSYSTEM_ADMIN:
		return 2
	default:
		return 1
	}
}

// mergeMember 合并源账号在一个租户里的成员身份。
// 目标已经是该租户成员时，把源成员的角色并到目标成员上，清理源成员的角色行并软删源成员；
// 目标还不是成员时，直接把成员行的归属改成目标账号，角色授权仍挂在同一个成员行上。
// member 的唯一索引是 (namespace_uid, user_uid) WHERE deleted_at IS NULL，两个分支互斥：
// 目标已有未删除成员时不能再改源成员的 user_uid，否则会撞索引。
func (r *userRepo) mergeMember(ctx context.Context, tx *ent.Tx, source *ent.Member, toUID string) error {
	target, err := tx.Member.Query().Where(
		member.NamespaceUIDEQ(source.NamespaceUID),
		member.UserUIDEQ(toUID),
		member.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return r.fail(err)
	}
	if ent.IsNotFound(err) {
		n, err := tx.Member.Update().
			Where(member.UIDEQ(source.UID), member.UserUIDEQ(source.UserUID), member.DeletedAtIsNil()).
			SetUserUID(toUID).
			Save(ctx)
		if err != nil {
			return r.fail(err)
		}
		if n == 0 {
			// 同一个事务里刚查到的成员行不应该消失；真出现说明有并发删除，直接失败让调用方回滚。
			return merr.Conflict("member is already merged")
		}
		return nil
	}
	// 目标成员可能还停在 INVITED/EXPIRED。源成员已经加入租户时要把目标成员补成 JOINED：
	// 权限校验（role_repo、access）只认 JOINED，否则占位账号已经加入的租户在账号并过去后会失去访问权。
	if source.Status == member.StatusJOINED && target.Status != member.StatusJOINED {
		if _, err := tx.Member.Update().
			Where(member.UIDEQ(target.UID), member.DeletedAtIsNil()).
			SetStatus(member.StatusJOINED).
			Save(ctx); err != nil {
			return r.fail(err)
		}
	}
	links, err := tx.MemberRole.Query().Where(memberrole.MemberUIDEQ(source.UID)).All(ctx)
	if err != nil {
		return r.fail(err)
	}
	for _, link := range links {
		exists, err := tx.MemberRole.Query().Where(
			memberrole.MemberUIDEQ(target.UID),
			memberrole.RoleUIDEQ(link.RoleUID),
		).Exist(ctx)
		if err != nil {
			return r.fail(err)
		}
		if exists {
			continue
		}
		if _, err := tx.MemberRole.Create().
			SetUID(id.New()).
			SetNamespaceUID(source.NamespaceUID).
			SetMemberUID(target.UID).
			SetRoleUID(link.RoleUID).
			Save(ctx); err != nil {
			return r.fail(err)
		}
	}
	// 角色并完之后先清源成员的角色行再软删成员，避免留下指向已删除成员的角色授权。
	if _, err := tx.MemberRole.Delete().Where(memberrole.MemberUIDEQ(source.UID)).Exec(ctx); err != nil {
		return r.fail(err)
	}
	// 通知组按成员 UID 保存收件人，源成员马上被软删；把引用改到目标成员，
	// 否则 resolveGroupReceivers 会把它们过滤掉，收件人静默消失。
	if err := remapGroupMemberUIDs(ctx, tx, source.NamespaceUID, source.UID, target.UID); err != nil {
		return r.fail(err)
	}
	if _, err := tx.Member.Update().
		Where(member.UIDEQ(source.UID), member.DeletedAtIsNil()).
		SetDeletedAt(time.Now()).
		Save(ctx); err != nil {
		return r.fail(err)
	}
	return nil
}

// remapGroupMemberUIDs 把通知组里引用的源成员 UID 换成目标成员 UID。
// 三个数组都要处理：email_member_uids、sms_member_uids，以及旧版 member_uids。
func remapGroupMemberUIDs(ctx context.Context, tx *ent.Tx, namespaceUID, fromUID, toUID string) error {
	groups, err := tx.NotificationGroup.Query().
		Where(
			notificationgroup.NamespaceUIDEQ(namespaceUID),
			notificationgroup.DeletedAtIsNil(),
		).
		All(ctx)
	if err != nil {
		return err
	}
	for _, group := range groups {
		emailUIDs, emailChanged := replaceMemberUID(group.EmailMemberUids, fromUID, toUID)
		smsUIDs, smsChanged := replaceMemberUID(group.SmsMemberUids, fromUID, toUID)
		legacyUIDs, legacyChanged := replaceMemberUID(group.MemberUids, fromUID, toUID)
		if !emailChanged && !smsChanged && !legacyChanged {
			continue
		}
		update := tx.NotificationGroup.UpdateOne(group)
		if emailChanged {
			update.SetEmailMemberUids(emailUIDs)
		}
		if smsChanged {
			update.SetSmsMemberUids(smsUIDs)
		}
		if legacyChanged {
			update.SetMemberUids(legacyUIDs)
		}
		if _, err := update.Save(ctx); err != nil {
			return err
		}
	}
	return nil
}

// replaceMemberUID 返回替换后的数组以及是否发生变化；发生变化时顺带去重。
func replaceMemberUID(values []string, from, to string) ([]string, bool) {
	changed := false
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == from {
			value = to
			changed = true
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	if !changed {
		return values, false
	}
	return out, true
}

// moveNamespaceOwner 把租户负责人改到目标账号。改之前先保证目标在该租户有未删除成员并持有 owner 角色，
// 否则租户会出现没有负责人的状态。
func (r *userRepo) moveNamespaceOwner(ctx context.Context, tx *ent.Tx, space *ent.Namespace, toUID string) error {
	target, err := tx.Member.Query().Where(
		member.NamespaceUIDEQ(space.UID),
		member.UserUIDEQ(toUID),
		member.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return r.fail(err)
	}
	if ent.IsNotFound(err) {
		target, err = tx.Member.Create().
			SetUID(id.New()).
			SetNamespaceUID(space.UID).
			SetUserUID(toUID).
			SetStatus(member.StatusJOINED).
			Save(ctx)
		if err != nil {
			return r.fail(err)
		}
	} else if target.Status != member.StatusJOINED {
		// 负责人必须是 JOINED 成员，权限校验也只认 JOINED；源账号是占位账号时源成员状态可能不是 JOINED，
		// 这里把目标成员直接补成 JOINED，避免租户负责人看不到自己的租户。
		if _, err := tx.Member.Update().
			Where(member.UIDEQ(target.UID), member.DeletedAtIsNil()).
			SetStatus(member.StatusJOINED).
			Save(ctx); err != nil {
			return r.fail(err)
		}
	}
	ownerRole, err := tx.Role.Query().Where(
		role.NamespaceUIDEQ(space.UID),
		role.NameEQ(perm.RoleOwner),
		role.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return r.fail(err)
	}
	hasOwner, err := tx.MemberRole.Query().Where(
		memberrole.MemberUIDEQ(target.UID),
		memberrole.RoleUIDEQ(ownerRole.UID),
	).Exist(ctx)
	if err != nil {
		return r.fail(err)
	}
	if !hasOwner {
		if _, err := tx.MemberRole.Create().
			SetUID(id.New()).
			SetNamespaceUID(space.UID).
			SetMemberUID(target.UID).
			SetRoleUID(ownerRole.UID).
			Save(ctx); err != nil {
			return r.fail(err)
		}
	}
	// 只改 owner_uid 仍是源账号的记录；并发请求已经改过时按成功处理，前面补的成员和角色不冲突。
	if _, err := tx.Namespace.Update().
		Where(namespace.UIDEQ(space.UID), namespace.OwnerUIDEQ(space.OwnerUID)).
		SetOwnerUID(toUID).
		Save(ctx); err != nil {
		return r.fail(err)
	}
	return nil
}
