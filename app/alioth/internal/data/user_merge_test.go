package data

import (
	"context"
	"log/slog"
	"strconv"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alerteventcolumn"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertpage"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/member"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/memberinvite"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/memberrole"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/namespace"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/namespaceorder"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/role"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/rolepermission"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/systemmessage"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/user"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/useralertpage"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/useroauth"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

// MergeAccounts 和 MoveOAuth 会影响成员、角色、租户负责人和平台系统角色，只有真库能验证
// 事务、软删和唯一索引（member 的 (namespace_uid, user_uid) WHERE deleted_at IS NULL、
// user_oauth 的 (provider, open_id)、super admin 的部分唯一索引）。
// 测试读写独立测试库：
//
//	ALIOTH_TEST_DSN='postgres://alioth:alioth@127.0.0.1:5432/alioth_test?sslmode=disable' \
//	  go test -C app/alioth ./internal/data -run 'TestMergeAccounts|TestMoveOAuth' -count=1 -v
//
// 未设置 ALIOTH_TEST_DSN 时整组测试跳过；testDSN 拒绝使用开发库 alioth。
// 所有测试数据都带 merge-<时间戳> 前缀，t.Cleanup 只按本测试记录的 UID 清理，不会碰库里其他数据。
func mergeTestStore(t *testing.T) *Dependencies {
	t.Helper()
	source := testDSN(t)
	store, cleanup, err := NewDependencies(
		&conf.Data{Database: &conf.Data_Database{Source: source}},
		&conf.Auth{EncryptKey: "dev-only-encrypt-key"},
		slog.Default(),
	)
	if err != nil {
		t.Skip(err)
	}
	t.Cleanup(cleanup)
	if err := ApplySchema(context.Background(), store.client, store.db); err != nil {
		t.Fatalf("ApplySchema: %v", err)
	}
	return store
}

// mergeFixture 造数并记录本测试创建的行，清理时只删这些行。
type mergeFixture struct {
	t          *testing.T
	store      *Dependencies
	prefix     string
	users      []string
	namespaces []string
	openIDs    []string
}

func newMergeFixture(t *testing.T) *mergeFixture {
	t.Helper()
	fixture := &mergeFixture{
		t:      t,
		store:  mergeTestStore(t),
		prefix: "merge-" + strconv.FormatInt(time.Now().UnixNano(), 10),
	}
	t.Cleanup(fixture.cleanup)
	return fixture
}

// cleanup 按本测试记录的 UID 清理：先个人数据和成员/角色，再租户，最后用户和第三方绑定。
// 个人数据按用户 UID 清理，成员数据按租户 UID 清理；都只删本测试造出来的行。
func (fixture *mergeFixture) cleanup() {
	ctx := context.Background()
	client := fixture.store.client
	if len(fixture.users) > 0 {
		_, _ = client.UserAlertPage.Delete().Where(useralertpage.UserUIDIn(fixture.users...)).Exec(ctx)
		_, _ = client.AlertEventColumn.Delete().Where(alerteventcolumn.UserUIDIn(fixture.users...)).Exec(ctx)
		_, _ = client.SystemMessage.Delete().Where(systemmessage.UserUIDIn(fixture.users...)).Exec(ctx)
		_, _ = client.MemberInvite.Delete().Where(memberinvite.UserUIDIn(fixture.users...)).Exec(ctx)
		_, _ = client.MemberInvite.Delete().Where(memberinvite.InviterUIDIn(fixture.users...)).Exec(ctx)
		_, _ = client.NamespaceOrder.Delete().Where(namespaceorder.UserUIDIn(fixture.users...)).Exec(ctx)
		_, _ = client.AlertPage.Delete().Where(alertpage.OwnerUIDIn(fixture.users...)).Exec(ctx)
	}
	if len(fixture.namespaces) > 0 {
		_, _ = client.MemberRole.Delete().Where(memberrole.NamespaceUIDIn(fixture.namespaces...)).Exec(ctx)
		_, _ = client.Member.Delete().Where(member.NamespaceUIDIn(fixture.namespaces...)).Exec(ctx)
		_, _ = client.RolePermission.Delete().Where(rolepermission.NamespaceUIDIn(fixture.namespaces...)).Exec(ctx)
		_, _ = client.Role.Delete().Where(role.NamespaceUIDIn(fixture.namespaces...)).Exec(ctx)
		_, _ = client.Namespace.Delete().Where(namespace.UIDIn(fixture.namespaces...)).Exec(ctx)
	}
	if len(fixture.openIDs) > 0 {
		_, _ = client.UserOAuth.Delete().Where(useroauth.OpenIDIn(fixture.openIDs...)).Exec(ctx)
	}
	if len(fixture.users) > 0 {
		_, _ = client.User.Delete().Where(user.UIDIn(fixture.users...)).Exec(ctx)
	}
}

func (fixture *mergeFixture) ctx() context.Context {
	return context.Background()
}

func (fixture *mergeFixture) createUser(label string, systemRole user.SystemRole) *ent.User {
	fixture.t.Helper()
	uid := fixture.prefix + "-user-" + label
	row, err := fixture.store.client.User.Create().
		SetUID(uid).
		SetName(label).
		SetNickname(label).
		SetEmail(uid + "@test.example.com").
		SetSystemRole(systemRole).
		Save(fixture.ctx())
	if err != nil {
		fixture.t.Fatalf("创建用户 %s: %v", label, err)
	}
	fixture.users = append(fixture.users, uid)
	return row
}

// createNamespace 只建租户和四个系统角色，不写权限目录，避免依赖测试库的种子数据。
func (fixture *mergeFixture) createNamespace(label, ownerUID string) (*ent.Namespace, map[string]*ent.Role) {
	fixture.t.Helper()
	uid := fixture.prefix + "-ns-" + label
	space, err := fixture.store.client.Namespace.Create().
		SetUID(uid).
		SetName(fixture.prefix + "-name-" + label).
		SetOwnerUID(ownerUID).
		Save(fixture.ctx())
	if err != nil {
		fixture.t.Fatalf("创建租户 %s: %v", label, err)
	}
	roles := make(map[string]*ent.Role, 4)
	for _, name := range []string{perm.RoleOwner, perm.RoleAdmin, perm.RoleMember, perm.RoleViewer} {
		row, err := fixture.store.client.Role.Create().
			SetUID(id.New()).
			SetNamespaceUID(uid).
			SetName(name).
			SetIsSystem(true).
			Save(fixture.ctx())
		if err != nil {
			fixture.t.Fatalf("创建角色 %s: %v", name, err)
		}
		roles[name] = row
	}
	fixture.namespaces = append(fixture.namespaces, uid)
	return space, roles
}

func (fixture *mergeFixture) addMember(space *ent.Namespace, userUID string, status member.Status, roles ...*ent.Role) *ent.Member {
	fixture.t.Helper()
	row, err := fixture.store.client.Member.Create().
		SetUID(id.New()).
		SetNamespaceUID(space.UID).
		SetUserUID(userUID).
		SetStatus(status).
		Save(fixture.ctx())
	if err != nil {
		fixture.t.Fatalf("创建成员: %v", err)
	}
	for _, item := range roles {
		if _, err := fixture.store.client.MemberRole.Create().
			SetUID(id.New()).
			SetNamespaceUID(space.UID).
			SetMemberUID(row.UID).
			SetRoleUID(item.UID).
			Save(fixture.ctx()); err != nil {
			fixture.t.Fatalf("绑定成员角色: %v", err)
		}
	}
	return row
}

func (fixture *mergeFixture) mergeAccounts(fromUID, toUID string) error {
	fixture.t.Helper()
	users := &userRepo{newBase(fixture.store, slog.Default())}
	return users.MergeAccounts(fixture.ctx(), fromUID, toUID)
}

// activeMember 返回租户里未删除的成员行，不存在时返回 nil。
func (fixture *mergeFixture) activeMember(namespaceUID, userUID string) *ent.Member {
	fixture.t.Helper()
	row, err := fixture.store.client.Member.Query().Where(
		member.NamespaceUIDEQ(namespaceUID),
		member.UserUIDEQ(userUID),
		member.DeletedAtIsNil(),
	).Only(fixture.ctx())
	if ent.IsNotFound(err) {
		return nil
	}
	if err != nil {
		fixture.t.Fatalf("查询成员: %v", err)
	}
	return row
}

func (fixture *mergeFixture) memberRoles(namespaceUID, memberUID string) map[string]bool {
	fixture.t.Helper()
	links, err := fixture.store.client.MemberRole.Query().Where(
		memberrole.NamespaceUIDEQ(namespaceUID),
		memberrole.MemberUIDEQ(memberUID),
	).All(fixture.ctx())
	if err != nil {
		fixture.t.Fatalf("查询成员角色: %v", err)
	}
	names := make(map[string]bool, len(links))
	for _, link := range links {
		row, err := fixture.store.client.Role.Query().Where(role.UIDEQ(link.RoleUID)).Only(fixture.ctx())
		if err != nil {
			fixture.t.Fatalf("查询角色: %v", err)
		}
		names[row.Name] = true
	}
	return names
}

func (fixture *mergeFixture) reloadUser(uid string) *ent.User {
	fixture.t.Helper()
	row, err := fixture.store.client.User.Query().Where(user.UIDEQ(uid)).Only(fixture.ctx())
	if err != nil {
		fixture.t.Fatalf("查询用户 %s: %v", uid, err)
	}
	return row
}

// updateProfile 直接改用户的个人资料字段，用来验证合并时只补目标为空的值。
func (fixture *mergeFixture) updateProfile(uid, phone, avatar, remark string) {
	fixture.t.Helper()
	if _, err := fixture.store.client.User.Update().
		Where(user.UIDEQ(uid)).
		SetPhone(phone).
		SetAvatar(avatar).
		SetRemark(remark).
		Save(fixture.ctx()); err != nil {
		fixture.t.Fatalf("更新用户资料 %s: %v", uid, err)
	}
}

// createAlertPage 造一个租户告警页面，owner 是创建人。
func (fixture *mergeFixture) createAlertPage(namespaceUID, ownerUID, label string) *ent.AlertPage {
	fixture.t.Helper()
	row, err := fixture.store.client.AlertPage.Create().
		SetUID(id.New()).
		SetNamespaceUID(namespaceUID).
		SetName(fixture.prefix + "-page-" + label).
		SetOwnerUID(ownerUID).
		Save(fixture.ctx())
	if err != nil {
		fixture.t.Fatalf("创建告警页面 %s: %v", label, err)
	}
	return row
}

// selectAlertPage 造一条用户对告警页面的选择记录，sort 用来断言去重时保留目标的顺序。
func (fixture *mergeFixture) selectAlertPage(namespaceUID, userUID, alertPageUID string, sort int) *ent.UserAlertPage {
	fixture.t.Helper()
	row, err := fixture.store.client.UserAlertPage.Create().
		SetUID(id.New()).
		SetNamespaceUID(namespaceUID).
		SetUserUID(userUID).
		SetAlertPageUID(alertPageUID).
		SetSort(sort).
		Save(fixture.ctx())
	if err != nil {
		fixture.t.Fatalf("创建告警页面选择: %v", err)
	}
	return row
}

// saveAlertEventColumn 造一条用户在某个告警视图上的列配置。
func (fixture *mergeFixture) saveAlertEventColumn(namespaceUID, userUID, scope, columns string) *ent.AlertEventColumn {
	fixture.t.Helper()
	row, err := fixture.store.client.AlertEventColumn.Create().
		SetUID(id.New()).
		SetNamespaceUID(namespaceUID).
		SetUserUID(userUID).
		SetScope(scope).
		SetColumnsJSON(columns).
		Save(fixture.ctx())
	if err != nil {
		fixture.t.Fatalf("创建告警列配置 %s: %v", scope, err)
	}
	return row
}

// sendSystemMessage 造一条发给指定用户的系统消息。
func (fixture *mergeFixture) sendSystemMessage(userUID, label string) *ent.SystemMessage {
	fixture.t.Helper()
	row, err := fixture.store.client.SystemMessage.Create().
		SetUID(id.New()).
		SetUserUID(userUID).
		SetKind(systemmessage.KindMEMBER_REMOVED).
		SetNamespaceName(fixture.prefix + "-message-" + label).
		Save(fixture.ctx())
	if err != nil {
		fixture.t.Fatalf("创建系统消息 %s: %v", label, err)
	}
	return row
}

// createNamespaceOrder 造一条指定下单人的租户订单。
func (fixture *mergeFixture) createNamespaceOrder(userUID, namespaceUID, label string) *ent.NamespaceOrder {
	fixture.t.Helper()
	row, err := fixture.store.client.NamespaceOrder.Create().
		SetUID(id.New()).
		SetUserUID(userUID).
		SetNamespaceUID(namespaceUID).
		SetNamespaceName(fixture.prefix + "-order-ns-" + label).
		SetPackageUID(fixture.prefix + "-package-" + label).
		SetPackageName(fixture.prefix + "-package-name-" + label).
		Save(fixture.ctx())
	if err != nil {
		fixture.t.Fatalf("创建订单 %s: %v", label, err)
	}
	return row
}

// createMemberInvite 造一条邀请记录。userUID 是受邀人，inviterUID 是邀请人；
// 邮箱由调用方给出，避免撞同一租户同一邮箱只保留一条待确认邀请的部分唯一索引。
func (fixture *mergeFixture) createMemberInvite(namespaceUID, email, userUID, inviterUID, label string) *ent.MemberInvite {
	fixture.t.Helper()
	row, err := fixture.store.client.MemberInvite.Create().
		SetUID(id.New()).
		SetNamespaceUID(namespaceUID).
		SetEmail(email).
		SetUserUID(userUID).
		SetRoleUID(fixture.prefix + "-role-" + label).
		SetInviterUID(inviterUID).
		SetTokenHash(fixture.prefix + "-token-" + label).
		SetExpiresAt(time.Now().Add(time.Hour)).
		Save(fixture.ctx())
	if err != nil {
		fixture.t.Fatalf("创建邀请 %s: %v", label, err)
	}
	return row
}

// linkOAuth 给用户写入一条第三方绑定。
func (fixture *mergeFixture) linkOAuth(userUID, provider, openID string) {
	fixture.t.Helper()
	fixture.openIDs = append(fixture.openIDs, openID)
	if err := (&oauthRepo{newBase(fixture.store, slog.Default())}).LinkOAuth(
		fixture.ctx(), userUID, provider, openID, fixture.prefix+"@test.example.com",
	); err != nil {
		fixture.t.Fatalf("写入第三方绑定: %v", err)
	}
}

// TestMergeAccountsMembersAndSystemRole 覆盖账号合并的成员迁移、状态补齐、角色并集、
// 租户负责人迁移和平台系统角色转移。每个子测试独立造数，互不影响。
func TestMergeAccountsMembersAndSystemRole(t *testing.T) {
	t.Run("moves member to target when target has no member", func(t *testing.T) {
		fixture := newMergeFixture(t)
		from := fixture.createUser("from", user.SystemRoleUSER)
		to := fixture.createUser("to", user.SystemRoleUSER)
		owner := fixture.createUser("owner", user.SystemRoleUSER)
		space, roles := fixture.createNamespace("move", owner.UID)
		source := fixture.addMember(space, from.UID, member.StatusJOINED, roles[perm.RoleMember])

		if err := fixture.mergeAccounts(from.UID, to.UID); err != nil {
			t.Fatalf("MergeAccounts: %v", err)
		}
		moved := fixture.activeMember(space.UID, to.UID)
		if moved == nil {
			t.Fatal("源成员没有迁移到目标账号")
		}
		if moved.Status != member.StatusJOINED {
			t.Fatalf("成员状态 = %s, want JOINED", moved.Status)
		}
		if !fixture.memberRoles(space.UID, moved.UID)[perm.RoleMember] {
			t.Fatalf("迁移后成员角色 = %v，应保留 member", fixture.memberRoles(space.UID, moved.UID))
		}
		if fixture.activeMember(space.UID, from.UID) != nil {
			t.Fatal("源账号仍保留未删除成员")
		}
		// 目标在该租户没有成员时只改写成员行的归属，角色行仍挂在同一个成员行上，必须保留。
		if got, _ := fixture.store.client.MemberRole.Query().Where(memberrole.MemberUIDEQ(source.UID)).Count(fixture.ctx()); got != 1 {
			t.Fatalf("迁移后成员角色行 = %d, want 1", got)
		}
	})

	t.Run("merges roles and soft deletes source member", func(t *testing.T) {
		fixture := newMergeFixture(t)
		from := fixture.createUser("from", user.SystemRoleUSER)
		to := fixture.createUser("to", user.SystemRoleUSER)
		owner := fixture.createUser("owner", user.SystemRoleUSER)
		space, roles := fixture.createNamespace("union", owner.UID)
		source := fixture.addMember(space, from.UID, member.StatusJOINED, roles[perm.RoleAdmin])
		target := fixture.addMember(space, to.UID, member.StatusINVITED, roles[perm.RoleMember])

		if err := fixture.mergeAccounts(from.UID, to.UID); err != nil {
			t.Fatalf("MergeAccounts: %v", err)
		}
		merged := fixture.activeMember(space.UID, to.UID)
		if merged == nil {
			t.Fatal("目标成员丢失")
		}
		if merged.UID != target.UID {
			t.Fatalf("目标成员行 = %s, want %s（目标已有成员时应复用原成员行）", merged.UID, target.UID)
		}
		// 源成员 JOINED、目标 INVITED：权限校验只认 JOINED，合并后必须补成 JOINED。
		if merged.Status != member.StatusJOINED {
			t.Fatalf("目标成员状态 = %s, want JOINED", merged.Status)
		}
		gotRoles := fixture.memberRoles(space.UID, merged.UID)
		if !gotRoles[perm.RoleAdmin] || !gotRoles[perm.RoleMember] {
			t.Fatalf("角色并集 = %v，应同时含 admin 和 member", gotRoles)
		}
		sourceRow, err := fixture.store.client.Member.Query().Where(member.UIDEQ(source.UID)).Only(fixture.ctx())
		if err != nil {
			t.Fatalf("查询源成员: %v", err)
		}
		if sourceRow.DeletedAt == nil {
			t.Fatal("源成员没有软删")
		}
		if got, _ := fixture.store.client.MemberRole.Query().Where(memberrole.MemberUIDEQ(source.UID)).Count(fixture.ctx()); got != 0 {
			t.Fatalf("源成员角色行 = %d, want 0", got)
		}
		if fixture.activeMember(space.UID, from.UID) != nil {
			t.Fatal("源账号仍保留未删除成员")
		}
	})

	t.Run("transfers namespace owner and owner role", func(t *testing.T) {
		fixture := newMergeFixture(t)
		from := fixture.createUser("from", user.SystemRoleUSER)
		to := fixture.createUser("to", user.SystemRoleUSER)
		space, roles := fixture.createNamespace("owner", from.UID)
		fixture.addMember(space, from.UID, member.StatusJOINED, roles[perm.RoleOwner])
		target := fixture.addMember(space, to.UID, member.StatusINVITED, roles[perm.RoleAdmin])

		if err := fixture.mergeAccounts(from.UID, to.UID); err != nil {
			t.Fatalf("MergeAccounts: %v", err)
		}
		spaceRow, err := fixture.store.client.Namespace.Query().Where(namespace.UIDEQ(space.UID)).Only(fixture.ctx())
		if err != nil {
			t.Fatalf("查询租户: %v", err)
		}
		if spaceRow.OwnerUID != to.UID {
			t.Fatalf("租户负责人 = %s, want %s", spaceRow.OwnerUID, to.UID)
		}
		merged := fixture.activeMember(space.UID, to.UID)
		if merged == nil {
			t.Fatal("目标成员丢失")
		}
		if merged.UID != target.UID {
			t.Fatalf("目标成员行 = %s, want %s", merged.UID, target.UID)
		}
		if merged.Status != member.StatusJOINED {
			t.Fatalf("目标成员状态 = %s, want JOINED", merged.Status)
		}
		gotRoles := fixture.memberRoles(space.UID, merged.UID)
		if !gotRoles[perm.RoleOwner] || !gotRoles[perm.RoleAdmin] {
			t.Fatalf("目标成员角色 = %v，应含 owner 和 admin", gotRoles)
		}
	})

	t.Run("promotes invited owner to joined when source is not joined", func(t *testing.T) {
		fixture := newMergeFixture(t)
		from := fixture.createUser("from", user.SystemRoleUSER)
		to := fixture.createUser("to", user.SystemRoleUSER)
		space, roles := fixture.createNamespace("owner-status", from.UID)
		// 源账号状态不是 JOINED（历史脏数据），负责人迁移仍要保证目标成员是 JOINED。
		fixture.addMember(space, from.UID, member.StatusEXPIRED, roles[perm.RoleOwner])
		fixture.addMember(space, to.UID, member.StatusINVITED, roles[perm.RoleAdmin])

		if err := fixture.mergeAccounts(from.UID, to.UID); err != nil {
			t.Fatalf("MergeAccounts: %v", err)
		}
		spaceRow, err := fixture.store.client.Namespace.Query().Where(namespace.UIDEQ(space.UID)).Only(fixture.ctx())
		if err != nil {
			t.Fatalf("查询租户: %v", err)
		}
		if spaceRow.OwnerUID != to.UID {
			t.Fatalf("租户负责人 = %s, want %s", spaceRow.OwnerUID, to.UID)
		}
		merged := fixture.activeMember(space.UID, to.UID)
		if merged == nil {
			t.Fatal("目标成员丢失")
		}
		if merged.Status != member.StatusJOINED {
			t.Fatalf("目标成员状态 = %s, want JOINED", merged.Status)
		}
		if !fixture.memberRoles(space.UID, merged.UID)[perm.RoleOwner] {
			t.Fatalf("目标成员角色 = %v，应含 owner", fixture.memberRoles(space.UID, merged.UID))
		}
	})

	t.Run("moves system admin role to target", func(t *testing.T) {
		fixture := newMergeFixture(t)
		from := fixture.createUser("from", user.SystemRoleSYSTEM_ADMIN)
		to := fixture.createUser("to", user.SystemRoleUSER)

		if err := fixture.mergeAccounts(from.UID, to.UID); err != nil {
			t.Fatalf("MergeAccounts: %v", err)
		}
		if got := fixture.reloadUser(from.UID).SystemRole; got != user.SystemRoleUSER {
			t.Fatalf("源账号平台角色 = %s, want USER", got)
		}
		if got := fixture.reloadUser(to.UID).SystemRole; got != user.SystemRoleSYSTEM_ADMIN {
			t.Fatalf("目标账号平台角色 = %s, want SYSTEM_ADMIN", got)
		}
	})

	t.Run("moves super admin role to target without hitting unique index", func(t *testing.T) {
		fixture := newMergeFixture(t)
		// 测试库可能残留上一个异常中断的用例留下的超级管理员，遇到时跳过，避免撞唯一索引。
		existing, err := fixture.store.client.User.Query().Where(user.SystemRoleEQ(user.SystemRoleSUPER_ADMIN)).Count(fixture.ctx())
		if err != nil {
			t.Fatalf("统计超级管理员: %v", err)
		}
		if existing > 0 {
			t.Skipf("测试库已有 %d 个超级管理员，跳过一次转移断言", existing)
		}
		from := fixture.createUser("super-from", user.SystemRoleSUPER_ADMIN)
		to := fixture.createUser("super-to", user.SystemRoleUSER)

		if err := fixture.mergeAccounts(from.UID, to.UID); err != nil {
			t.Fatalf("MergeAccounts: %v", err)
		}
		if got := fixture.reloadUser(from.UID).SystemRole; got != user.SystemRoleUSER {
			t.Fatalf("源账号平台角色 = %s, want USER", got)
		}
		if got := fixture.reloadUser(to.UID).SystemRole; got != user.SystemRoleSUPER_ADMIN {
			t.Fatalf("目标账号平台角色 = %s, want SUPER_ADMIN", got)
		}
	})

	t.Run("repeated merge is idempotent", func(t *testing.T) {
		fixture := newMergeFixture(t)
		from := fixture.createUser("from", user.SystemRoleUSER)
		to := fixture.createUser("to", user.SystemRoleUSER)
		owner := fixture.createUser("owner", user.SystemRoleUSER)
		space, roles := fixture.createNamespace("repeat", owner.UID)
		fixture.addMember(space, from.UID, member.StatusJOINED, roles[perm.RoleMember])

		if err := fixture.mergeAccounts(from.UID, to.UID); err != nil {
			t.Fatalf("首次 MergeAccounts: %v", err)
		}
		if err := fixture.mergeAccounts(from.UID, to.UID); err != nil {
			t.Fatalf("重复 MergeAccounts 应幂等成功: %v", err)
		}
		if got, _ := fixture.store.client.Member.Query().Where(
			member.NamespaceUIDEQ(space.UID),
			member.UserUIDEQ(to.UID),
			member.DeletedAtIsNil(),
		).Count(fixture.ctx()); got != 1 {
			t.Fatalf("目标账号在租户里的成员数 = %d, want 1", got)
		}
		merged := fixture.activeMember(space.UID, to.UID)
		if merged == nil || !fixture.memberRoles(space.UID, merged.UID)[perm.RoleMember] {
			t.Fatalf("重复合并后成员/角色 = %+v", merged)
		}
		if fixture.activeMember(space.UID, from.UID) != nil {
			t.Fatal("重复合并后源账号出现成员")
		}
	})
}

// TestMoveOAuth 覆盖第三方绑定改绑：user_uid 和 email 一起更新，原归属第二次改绑返回 NotFound。
func TestMoveOAuth(t *testing.T) {
	fixture := newMergeFixture(t)
	from := fixture.createUser("from", user.SystemRoleUSER)
	to := fixture.createUser("to", user.SystemRoleUSER)
	openID := fixture.prefix + "-open"
	fixture.openIDs = append(fixture.openIDs, openID)
	oauths := &oauthRepo{newBase(fixture.store, slog.Default())}
	ctx := fixture.ctx()

	if err := oauths.LinkOAuth(ctx, from.UID, "DINGTALK", openID, "old@test.example.com"); err != nil {
		t.Fatalf("LinkOAuth: %v", err)
	}
	if err := oauths.MoveOAuth(ctx, "DINGTALK", openID, from.UID, to.UID, "new@test.example.com"); err != nil {
		t.Fatalf("MoveOAuth: %v", err)
	}
	row, err := fixture.store.client.UserOAuth.Query().Where(useroauth.OpenIDEQ(openID)).Only(ctx)
	if err != nil {
		t.Fatalf("查询绑定: %v", err)
	}
	if row.UserUID != to.UID {
		t.Fatalf("绑定 user_uid = %s, want %s", row.UserUID, to.UID)
	}
	if row.Email != "new@test.example.com" {
		t.Fatalf("绑定 email = %s, want new@test.example.com", row.Email)
	}
	// 原归属已经改走，再按 from 改绑应返回 NotFound，由用例层按幂等处理。
	if err := oauths.MoveOAuth(ctx, "DINGTALK", openID, from.UID, to.UID, "again@test.example.com"); !merr.IsNotFound(err) {
		t.Fatalf("0 行改绑应返回 NotFound，实际 %v", err)
	}
}

// TestMergeAccountsPersonalData 覆盖个人数据并入母本账号：告警页面归属、告警页选择、
// 告警列配置、系统消息、租户订单、邀请记录和第三方绑定整体迁移；
// 目标已有同键行时保留目标、丢弃源行。目标是母本，源账号的非空资料只填补目标为空的字段；
// 最后源账号被软删，重复合并和源账号已不存在都按幂等成功处理。
func TestMergeAccountsPersonalData(t *testing.T) {
	fixture := newMergeFixture(t)
	from := fixture.createUser("from", user.SystemRoleUSER)
	to := fixture.createUser("to", user.SystemRoleUSER)
	owner := fixture.createUser("owner", user.SystemRoleUSER)
	space, _ := fixture.createNamespace("personal", owner.UID)

	// 目标已有头像，源账号的手机号、头像和备注只能补空，不能覆盖已有值。
	fixture.updateProfile(from.UID, "13800000000", "https://avatar.example.com/from.png", "源账号备注")
	fixture.updateProfile(to.UID, "", "https://avatar.example.com/to.png", "")
	pageA := fixture.createAlertPage(space.UID, from.UID, "a")
	pageB := fixture.createAlertPage(space.UID, from.UID, "b")
	// 两个账号都选择了页面 A：保留邮箱账号那条和它的排序，源账号的丢弃。
	fixture.selectAlertPage(space.UID, from.UID, pageA.UID, 3)
	fixture.selectAlertPage(space.UID, from.UID, pageB.UID, 4)
	fixture.selectAlertPage(space.UID, to.UID, pageA.UID, 7)
	// 两个账号都在 mine 视图上有列配置：保留邮箱账号的；history 只有源账号有，整行迁过去。
	fixture.saveAlertEventColumn(space.UID, from.UID, "mine", `[{"key":"source"}]`)
	fixture.saveAlertEventColumn(space.UID, from.UID, "history", `[{"key":"history"}]`)
	fixture.saveAlertEventColumn(space.UID, to.UID, "mine", `[{"key":"target"}]`)
	fixture.sendSystemMessage(from.UID, "received")
	order := fixture.createNamespaceOrder(from.UID, space.UID, "order")
	received := fixture.createMemberInvite(space.UID, fixture.prefix+"-received@test.example.com", from.UID, owner.UID, "received")
	sent := fixture.createMemberInvite(space.UID, fixture.prefix+"-sent@test.example.com", "", from.UID, "sent")
	fixture.linkOAuth(from.UID, "GITHUB", fixture.prefix+"-github")

	if err := fixture.mergeAccounts(from.UID, to.UID); err != nil {
		t.Fatalf("MergeAccounts: %v", err)
	}

	// 目标资料只补空：手机号和备注来自源账号，头像保留目标自己的。
	merged := fixture.reloadUser(to.UID)
	if merged.Phone != "13800000000" || merged.Remark != "源账号备注" {
		t.Fatalf("目标资料没有补齐: phone=%q remark=%q", merged.Phone, merged.Remark)
	}
	if merged.Avatar != "https://avatar.example.com/to.png" {
		t.Fatalf("目标已有头像被覆盖: %q", merged.Avatar)
	}
	// 邮箱、名称、状态和平台角色以目标为准，不随源账号变化。
	if merged.Email != to.Email || merged.Name != to.Name || merged.SystemRole != user.SystemRoleUSER {
		t.Fatalf("目标身份字段被改动: %+v", merged)
	}
	// 源账号最后软删，不能按 UID 再查到。
	if sourceRow := fixture.reloadUser(from.UID); sourceRow.DeletedAt == nil {
		t.Fatal("源账号没有软删")
	}
	users := &userRepo{newBase(fixture.store, slog.Default())}
	if account, err := users.UserByUID(fixture.ctx(), from.UID); !merr.IsNotFound(err) {
		t.Fatalf("软删用户仍能按 UID 查到: %+v err=%v", account, err)
	}
	// 源账号创建的告警页面归属到邮箱账号，页面本身保留。
	for _, page := range []*ent.AlertPage{pageA, pageB} {
		row, err := fixture.store.client.AlertPage.Query().Where(alertpage.UIDEQ(page.UID)).Only(fixture.ctx())
		if err != nil {
			t.Fatalf("查询告警页面: %v", err)
		}
		if row.OwnerUID != to.UID {
			t.Fatalf("告警页面 %s 归属 = %s, want %s", page.UID, row.OwnerUID, to.UID)
		}
	}
	// 告警页选择：目标自己的页面 A 记录保留原排序，源账号的页面 B 迁入；源账号一条不剩。
	selections, err := fixture.store.client.UserAlertPage.Query().Where(useralertpage.UserUIDEQ(to.UID)).All(fixture.ctx())
	if err != nil {
		t.Fatalf("查询告警页选择: %v", err)
	}
	if len(selections) != 2 {
		t.Fatalf("目标告警页选择 = %d, want 2: %+v", len(selections), selections)
	}
	selected := map[string]int{}
	for _, row := range selections {
		selected[row.AlertPageUID] = row.Sort
	}
	if selected[pageA.UID] != 7 || selected[pageB.UID] != 4 {
		t.Fatalf("告警页选择和排序 = %v，应保留目标页面 A 的排序并迁入页面 B", selected)
	}
	if got, _ := fixture.store.client.UserAlertPage.Query().Where(useralertpage.UserUIDEQ(from.UID)).Count(fixture.ctx()); got != 0 {
		t.Fatalf("源账号仍保留 %d 条告警页选择", got)
	}
	// 告警列配置：mine 保留目标行，history 迁移源行；源账号一条不剩。
	columns, err := fixture.store.client.AlertEventColumn.Query().Where(alerteventcolumn.UserUIDEQ(to.UID)).All(fixture.ctx())
	if err != nil {
		t.Fatalf("查询告警列配置: %v", err)
	}
	byScope := map[string]string{}
	for _, row := range columns {
		byScope[row.Scope] = row.ColumnsJSON
	}
	if byScope["mine"] != `[{"key":"target"}]` || byScope["history"] != `[{"key":"history"}]` {
		t.Fatalf("告警列配置 = %v，应保留目标 mine、迁入源 history", byScope)
	}
	if got, _ := fixture.store.client.AlertEventColumn.Query().Where(alerteventcolumn.UserUIDEQ(from.UID)).Count(fixture.ctx()); got != 0 {
		t.Fatalf("源账号仍保留 %d 条告警列配置", got)
	}
	// 系统消息按接收人迁移。
	if got, _ := fixture.store.client.SystemMessage.Query().Where(systemmessage.UserUIDEQ(to.UID)).Count(fixture.ctx()); got != 1 {
		t.Fatalf("目标系统消息 = %d, want 1", got)
	}
	if got, _ := fixture.store.client.SystemMessage.Query().Where(systemmessage.UserUIDEQ(from.UID)).Count(fixture.ctx()); got != 0 {
		t.Fatalf("源账号仍保留 %d 条系统消息", got)
	}
	// 租户订单跟着下单人走。
	orderRow, err := fixture.store.client.NamespaceOrder.Query().Where(namespaceorder.UIDEQ(order.UID)).Only(fixture.ctx())
	if err != nil || orderRow.UserUID != to.UID {
		t.Fatalf("订单下单人 = %v err=%v, want %s", orderRow, err, to.UID)
	}
	// 邀请记录：受邀人字段迁到邮箱账号，邀请人字段也迁到邮箱账号。
	receivedRow, err := fixture.store.client.MemberInvite.Query().Where(memberinvite.UIDEQ(received.UID)).Only(fixture.ctx())
	if err != nil || receivedRow.UserUID != to.UID {
		t.Fatalf("受邀邀请 user_uid = %v err=%v, want %s", receivedRow, err, to.UID)
	}
	sentRow, err := fixture.store.client.MemberInvite.Query().Where(memberinvite.UIDEQ(sent.UID)).Only(fixture.ctx())
	if err != nil || sentRow.InviterUID != to.UID {
		t.Fatalf("邀请人字段 = %v err=%v, want %s", sentRow, err, to.UID)
	}
	// 第三方绑定随账号迁移。
	linkRow, err := fixture.store.client.UserOAuth.Query().Where(useroauth.OpenIDEQ(fixture.prefix + "-github")).Only(fixture.ctx())
	if err != nil || linkRow.UserUID != to.UID {
		t.Fatalf("第三方绑定 user_uid = %v err=%v, want %s", linkRow, err, to.UID)
	}

	// 重复合并：源账号已经软删，按已经合并过处理，不报错也不产生重复行。
	if err := fixture.mergeAccounts(from.UID, to.UID); err != nil {
		t.Fatalf("重复 MergeAccounts 应幂等成功: %v", err)
	}
	if got, _ := fixture.store.client.UserAlertPage.Query().Where(useralertpage.UserUIDEQ(to.UID)).Count(fixture.ctx()); got != 2 {
		t.Fatalf("重复合并后目标告警页选择 = %d, want 2", got)
	}
	if got, _ := fixture.store.client.AlertEventColumn.Query().Where(alerteventcolumn.UserUIDEQ(to.UID)).Count(fixture.ctx()); got != 2 {
		t.Fatalf("重复合并后目标告警列配置 = %d, want 2", got)
	}
	// 源账号根本不存在时同样按已经合并过处理，供并发提交兜底。
	if err := fixture.mergeAccounts(fixture.prefix+"-missing", to.UID); err != nil {
		t.Fatalf("源账号不存在时 MergeAccounts 应幂等成功: %v", err)
	}
}
