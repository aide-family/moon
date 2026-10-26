package data

import (
	"context"
	"log/slog"
	"strconv"
	"testing"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/member"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/memberrole"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/namespace"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/role"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/rolepermission"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/user"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/useroauth"
	"github.com/aide-family/Alioth/app/alioth/pkg/id"
	"github.com/aide-family/Alioth/app/alioth/pkg/perm"
	"github.com/aide-family/Alioth/pkg/merr"
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

// cleanup 按本测试记录的 UID 清理：先成员/角色/角色授权，再租户，最后用户和第三方绑定。
func (fixture *mergeFixture) cleanup() {
	ctx := context.Background()
	client := fixture.store.client
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
