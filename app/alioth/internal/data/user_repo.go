package data

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alerteventcolumn"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertpage"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/member"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/memberinvite"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/memberrole"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/namespace"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/namespaceorder"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/notificationgroup"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/systemmessage"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/user"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/useralertpage"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/useroauth"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/pkg/merr"
)

func (r *userRepo) CreateUser(ctx context.Context, in bo.Account) (*bo.Account, error) {
	status := user.StatusACTIVE
	if in.Status == "BANNED" {
		status = user.StatusBANNED
	}
	count, err := r.client.User.Query().Where(user.DeletedAtIsNil()).Count(ctx)
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
	row, err := r.client.User.Query().Where(user.EmailEQ(email), user.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	account := toAccount(row)
	return &account, nil
}

func (r *userRepo) UserByUID(ctx context.Context, uid string) (*bo.Account, error) {
	row, err := r.client.User.Query().Where(user.UIDEQ(uid), user.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	account := toAccount(row)
	return &account, nil
}

func (r *userRepo) UpdateUserProfile(ctx context.Context, uid string, email, phone, avatar, remark *string) (*bo.Account, error) {
	update := r.client.User.Update().Where(user.UIDEQ(uid), user.DeletedAtIsNil())
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
	n, err := r.client.User.Update().Where(user.UIDEQ(uid), user.DeletedAtIsNil()).SetStatus(value).Save(ctx)
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
		Where(user.UIDEQ(uid), user.DeletedAtIsNil()).
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
	query := r.client.User.Query().Where(user.DeletedAtIsNil())
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
	query := r.client.User.Query().Where(user.DeletedAtIsNil())
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

// MergeAccounts 把源账号的全部数据并入目标账号（母本），并在同一事务里软删源账号。
// 第三方历史占位账号（@users.alioth.local）补真实邮箱时，如果邮箱已经属于另一个账号，
// 会调用它把占位账号合并过去。整个迁移放在一个事务里：任何一步失败都回滚，
// 避免出现成员搬了一半、租户没有负责人、超级管理员空缺或者个人数据丢一半的中间态。
//
// 目标是母本：成员与角色取并集，租户负责人改到目标，平台角色走 mergeSystemRole；
// 个人数据（告警页面归属、告警页选择、告警列配置、系统消息、租户订单、邀请记录、
// 第三方绑定）整体迁到目标。目标已有冲突行时保留目标、丢弃源行，按各自的唯一键去重。
// 源账号已不存在或已软删时按"已经合并过"处理，返回成功，保证重复提交幂等。
func (r *userRepo) MergeAccounts(ctx context.Context, fromUID, toUID string) error {
	if fromUID == "" || toUID == "" || fromUID == toUID {
		return merr.Invalid("invalid request")
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	// 目标是并入的母本，必须存在且未删除；源账号允许已被前一次合并软删。
	to, err := tx.User.Query().Where(user.UIDEQ(toUID), user.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return r.fail(err)
	}
	from, err := tx.User.Query().Where(user.UIDEQ(fromUID), user.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return r.fail(err)
	}
	if err := r.mergeSystemRole(ctx, tx, from, to); err != nil {
		return err
	}
	if err := r.fillUserProfile(ctx, tx, from, to); err != nil {
		return err
	}
	if err := r.mergeUserData(ctx, tx, fromUID, toUID); err != nil {
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
	// 也可能刚由源成员改写而来，统一在这里保证目标成员可用。
	namespaces, err := tx.Namespace.Query().Where(namespace.OwnerUIDEQ(fromUID), namespace.DeletedAtIsNil()).All(ctx)
	if err != nil {
		return r.fail(err)
	}
	for _, space := range namespaces {
		if err := r.moveNamespaceOwner(ctx, tx, space, toUID); err != nil {
			return err
		}
	}
	// 全部数据迁完再软删源账号，源账号不再出现在用户列表里，邮箱也能被重新使用。
	if _, err := tx.User.Update().
		Where(user.UIDEQ(fromUID), user.DeletedAtIsNil()).
		SetDeletedAt(time.Now()).
		Save(ctx); err != nil {
		return r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return r.fail(err)
	}
	return nil
}

// fillUserProfile 只用源账号的非空字段填补目标为空的资料，目标已有值一律保留。
// 邮箱、名称、昵称、状态和平台角色属于账号身份，不在这里迁移：
// 目标账号是这次绑定的母本，身份以目标为准；平台角色由 mergeSystemRole 处理。
func (r *userRepo) fillUserProfile(ctx context.Context, tx *ent.Tx, from, to *ent.User) error {
	update := tx.User.Update().Where(user.UIDEQ(to.UID), user.DeletedAtIsNil())
	changed := false
	if to.Phone == "" && from.Phone != "" {
		update.SetPhone(from.Phone)
		changed = true
	}
	if to.Avatar == "" && from.Avatar != "" {
		update.SetAvatar(from.Avatar)
		changed = true
	}
	if to.Remark == "" && from.Remark != "" {
		update.SetRemark(from.Remark)
		changed = true
	}
	if !changed {
		return nil
	}
	if _, err := update.Save(ctx); err != nil {
		return r.fail(err)
	}
	return nil
}

// mergeUserData 把源账号的个人数据整体迁到目标账号。各表按自己的唯一键去重：
// 目标已有同键行时保留目标、删除源行，其余行直接改归属。
func (r *userRepo) mergeUserData(ctx context.Context, tx *ent.Tx, fromUID, toUID string) error {
	// 源账号创建过的告警页面归属到目标账号；告警页面本身是租户配置，不删除。
	if _, err := tx.AlertPage.Update().
		Where(alertpage.OwnerUIDEQ(fromUID)).
		SetOwnerUID(toUID).
		Save(ctx); err != nil {
		return r.fail(err)
	}
	if err := r.mergeUserAlertPages(ctx, tx, fromUID, toUID); err != nil {
		return err
	}
	if err := r.mergeAlertEventColumns(ctx, tx, fromUID, toUID); err != nil {
		return err
	}
	// 系统消息按接收人迁移，目标账号登录后能看到源账号收到的邀请和成员变动消息。
	if _, err := tx.SystemMessage.Update().
		Where(systemmessage.UserUIDEQ(fromUID)).
		SetUserUID(toUID).
		Save(ctx); err != nil {
		return r.fail(err)
	}
	// 租户订单属于下单人：迁移后由目标账号继续查看和支付，源账号删除时才不会留下悬空订单。
	if _, err := tx.NamespaceOrder.Update().
		Where(namespaceorder.UserUIDEQ(fromUID)).
		SetUserUID(toUID).
		Save(ctx); err != nil {
		return r.fail(err)
	}
	// 邀请记录既可能是"发给源账号的邀请"，也可能是"源账号发出的邀请"，两个字段都要迁。
	if _, err := tx.MemberInvite.Update().
		Where(memberinvite.UserUIDEQ(fromUID)).
		SetUserUID(toUID).
		Save(ctx); err != nil {
		return r.fail(err)
	}
	if _, err := tx.MemberInvite.Update().
		Where(memberinvite.InviterUIDEQ(fromUID)).
		SetInviterUID(toUID).
		Save(ctx); err != nil {
		return r.fail(err)
	}
	// 第三方绑定随账号一起搬走，源账号软删后不会留下无法登录的绑定。
	// (provider, open_id) 全局唯一，源账号和目标账号不可能持有同一条绑定。
	if _, err := tx.UserOAuth.Update().
		Where(useroauth.UserUIDEQ(fromUID)).
		SetUserUID(toUID).
		Save(ctx); err != nil {
		return r.fail(err)
	}
	return nil
}

// mergeUserAlertPages 合并实时告警页选择。唯一索引是 (namespace_uid, user_uid, alert_page_uid)，
// 目标已经选择了同一个页面时保留目标行的排序和创建时间，删掉源行。
func (r *userRepo) mergeUserAlertPages(ctx context.Context, tx *ent.Tx, fromUID, toUID string) error {
	rows, err := tx.UserAlertPage.Query().Where(useralertpage.UserUIDEQ(fromUID)).All(ctx)
	if err != nil {
		return r.fail(err)
	}
	for _, row := range rows {
		exists, err := tx.UserAlertPage.Query().Where(
			useralertpage.NamespaceUIDEQ(row.NamespaceUID),
			useralertpage.UserUIDEQ(toUID),
			useralertpage.AlertPageUIDEQ(row.AlertPageUID),
		).Exist(ctx)
		if err != nil {
			return r.fail(err)
		}
		if exists {
			if err := tx.UserAlertPage.DeleteOneID(row.ID).Exec(ctx); err != nil {
				return r.fail(err)
			}
			continue
		}
		if _, err := tx.UserAlertPage.UpdateOneID(row.ID).SetUserUID(toUID).Save(ctx); err != nil {
			return r.fail(err)
		}
	}
	return nil
}

// mergeAlertEventColumns 合并告警列表列配置。唯一索引是 (namespace_uid, user_uid, scope)，
// 同一视图上的配置保留目标、丢弃源。
func (r *userRepo) mergeAlertEventColumns(ctx context.Context, tx *ent.Tx, fromUID, toUID string) error {
	rows, err := tx.AlertEventColumn.Query().Where(alerteventcolumn.UserUIDEQ(fromUID)).All(ctx)
	if err != nil {
		return r.fail(err)
	}
	for _, row := range rows {
		exists, err := tx.AlertEventColumn.Query().Where(
			alerteventcolumn.NamespaceUIDEQ(row.NamespaceUID),
			alerteventcolumn.UserUIDEQ(toUID),
			alerteventcolumn.ScopeEQ(row.Scope),
		).Exist(ctx)
		if err != nil {
			return r.fail(err)
		}
		if exists {
			if err := tx.AlertEventColumn.DeleteOneID(row.ID).Exec(ctx); err != nil {
				return r.fail(err)
			}
			continue
		}
		if _, err := tx.AlertEventColumn.UpdateOneID(row.ID).SetUserUID(toUID).Save(ctx); err != nil {
			return r.fail(err)
		}
	}
	return nil
}

// DeleteUser 软删用户并清理挂在它身上的个人数据，用于平台清理错误账号。
// 用户仍是某个未删除租户的负责人，或者还有租户订单时拒绝：这两种数据牵涉租户和其他用户，
// 不能让账号被删后失去负责人或让订单找不到下单人。
// 清理和软删在同一个事务里完成：第三方绑定、成员与成员角色、告警页个人选择、
// 告警列配置、系统消息、邀请记录（被邀请人和邀请人）一并处理；
// 用户创建过的 alert_page 是租户配置，保留不动，只把归属随账号删除留在原处。
func (r *userRepo) DeleteUser(ctx context.Context, uid string) error {
	if uid == "" {
		return merr.Invalid("invalid request")
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.User.Query().Where(user.UIDEQ(uid), user.DeletedAtIsNil()).Only(ctx); err != nil {
		return r.fail(err)
	}
	owned, err := tx.Namespace.Query().Where(namespace.OwnerUIDEQ(uid), namespace.DeletedAtIsNil()).Exist(ctx)
	if err != nil {
		return r.fail(err)
	}
	if owned {
		return merr.Client("cannot delete a namespace owner")
	}
	hasOrders, err := tx.NamespaceOrder.Query().Where(namespaceorder.UserUIDEQ(uid)).Exist(ctx)
	if err != nil {
		return r.fail(err)
	}
	if hasOrders {
		return merr.Client("cannot delete a user with orders")
	}
	// 成员行软删；成员角色行全部删除，包括已经软删成员的残留授权，避免留下悬空角色授权。
	members, err := tx.Member.Query().Where(member.UserUIDEQ(uid)).All(ctx)
	if err != nil {
		return r.fail(err)
	}
	if len(members) > 0 {
		memberUIDs := make([]string, 0, len(members))
		for _, row := range members {
			memberUIDs = append(memberUIDs, row.UID)
		}
		if _, err := tx.MemberRole.Delete().Where(memberrole.MemberUIDIn(memberUIDs...)).Exec(ctx); err != nil {
			return r.fail(err)
		}
		if _, err := tx.Member.Update().
			Where(member.UserUIDEQ(uid), member.DeletedAtIsNil()).
			SetDeletedAt(time.Now()).
			Save(ctx); err != nil {
			return r.fail(err)
		}
	}
	if _, err := tx.UserOAuth.Delete().Where(useroauth.UserUIDEQ(uid)).Exec(ctx); err != nil {
		return r.fail(err)
	}
	if _, err := tx.UserAlertPage.Delete().Where(useralertpage.UserUIDEQ(uid)).Exec(ctx); err != nil {
		return r.fail(err)
	}
	if _, err := tx.AlertEventColumn.Delete().Where(alerteventcolumn.UserUIDEQ(uid)).Exec(ctx); err != nil {
		return r.fail(err)
	}
	if _, err := tx.SystemMessage.Delete().Where(systemmessage.UserUIDEQ(uid)).Exec(ctx); err != nil {
		return r.fail(err)
	}
	if _, err := tx.MemberInvite.Delete().
		Where(memberinvite.Or(memberinvite.UserUIDEQ(uid), memberinvite.InviterUIDEQ(uid))).
		Exec(ctx); err != nil {
		return r.fail(err)
	}
	if _, err := tx.User.Update().
		Where(user.UIDEQ(uid), user.DeletedAtIsNil()).
		SetDeletedAt(time.Now()).
		Save(ctx); err != nil {
		return r.fail(err)
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

// moveNamespaceOwner 把租户负责人改到目标账号。负责人权限现在由 namespace.owner_uid 隐式判定，
// 不再依赖预置 owner 角色，所以这里只保证目标在该租户有未删除的 JOINED 成员，然后切换 owner_uid。
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
		if _, err := tx.Member.Create().
			SetUID(id.New()).
			SetNamespaceUID(space.UID).
			SetUserUID(toUID).
			SetStatus(member.StatusJOINED).
			Save(ctx); err != nil {
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
	// 只改 owner_uid 仍是源账号的记录；并发请求已经改过时按成功处理。
	if _, err := tx.Namespace.Update().
		Where(namespace.UIDEQ(space.UID), namespace.OwnerUIDEQ(space.OwnerUID)).
		SetOwnerUID(toUID).
		Save(ctx); err != nil {
		return r.fail(err)
	}
	return nil
}
