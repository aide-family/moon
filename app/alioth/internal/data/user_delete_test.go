package data

import (
	"context"
	"log/slog"
	"testing"
	"time"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	"github.com/aide-family/moon/app/alioth/internal/data/ent/alerteventcolumn"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/alertpage"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/member"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/memberinvite"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/memberrole"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/namespace"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/systemmessage"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/user"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/useralertpage"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/useroauth"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

// userDeleteRepo 返回读写真库的用户仓储，复用 user_merge_test.go 的 fixture 和测试库开关。
func userDeleteRepo(fixture *mergeFixture) *userRepo {
	return &userRepo{newBase(fixture.store, slog.Default())}
}

// assertClientMessage 断言错误是客户端错误且文案符合预期。
func assertClientMessage(t *testing.T, err error, message string) {
	t.Helper()
	if err == nil {
		t.Fatalf("want client error %q, got nil", message)
	}
	if got := kerrors.FromError(err).Message; got != message {
		t.Fatalf("error message = %q, want %q", got, message)
	}
}

// TestDeleteUserCleansPersonalData 覆盖删除错误用户：软删用户，并清理第三方绑定、
// 成员与成员角色、告警页个人选择、告警列配置、系统消息和邀请记录（受邀人和邀请人）；
// 用户创建过的 alert_page 是租户配置，必须保留。
func TestDeleteUserCleansPersonalData(t *testing.T) {
	fixture := newMergeFixture(t)
	target := fixture.createUser("delete", user.SystemRoleUSER)
	owner := fixture.createUser("owner", user.SystemRoleUSER)
	space, roles := fixture.createNamespace("delete", owner.UID)
	memberRow := fixture.addMember(space, target.UID, member.StatusJOINED, roles[perm.RoleMember])
	fixture.linkOAuth(target.UID, "GITHUB", fixture.prefix+"-github")

	// 目标用户创建的告警页面要保留；另一个页面用来造个人选择。
	ownedPage := fixture.createAlertPage(space.UID, target.UID, "owned")
	selectedPage := fixture.createAlertPage(space.UID, owner.UID, "selected")
	fixture.selectAlertPage(space.UID, target.UID, selectedPage.UID, 1)
	fixture.saveAlertEventColumn(space.UID, target.UID, "mine", `[{"key":"mine"}]`)
	fixture.sendSystemMessage(target.UID, "message")
	received := fixture.createMemberInvite(space.UID, fixture.prefix+"-received@test.example.com", target.UID, owner.UID, "received")
	sent := fixture.createMemberInvite(space.UID, fixture.prefix+"-sent@test.example.com", "", target.UID, "sent")

	if err := userDeleteRepo(fixture).DeleteUser(fixture.ctx(), target.UID); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}

	// 用户软删，按 UID 查不到。
	if row := fixture.reloadUser(target.UID); row.DeletedAt == nil {
		t.Fatal("用户没有软删")
	}
	if account, err := userDeleteRepo(fixture).UserByUID(fixture.ctx(), target.UID); !merr.IsNotFound(err) {
		t.Fatalf("软删用户仍能按 UID 查到: %+v err=%v", account, err)
	}
	// 成员软删，成员角色行清空。
	memberRowAfter, err := fixture.store.client.Member.Query().Where(member.UIDEQ(memberRow.UID)).Only(fixture.ctx())
	if err != nil {
		t.Fatalf("查询成员: %v", err)
	}
	if memberRowAfter.DeletedAt == nil {
		t.Fatal("成员没有软删")
	}
	if got, _ := fixture.store.client.MemberRole.Query().Where(memberrole.MemberUIDEQ(memberRow.UID)).Count(fixture.ctx()); got != 0 {
		t.Fatalf("成员角色行 = %d, want 0", got)
	}
	// 第三方绑定、告警页选择、告警列配置、系统消息和邀请记录都被清理。
	if got, _ := fixture.store.client.UserOAuth.Query().Where(useroauth.UserUIDEQ(target.UID)).Count(fixture.ctx()); got != 0 {
		t.Fatalf("第三方绑定 = %d, want 0", got)
	}
	if got, _ := fixture.store.client.UserAlertPage.Query().Where(useralertpage.UserUIDEQ(target.UID)).Count(fixture.ctx()); got != 0 {
		t.Fatalf("告警页选择 = %d, want 0", got)
	}
	if got, _ := fixture.store.client.AlertEventColumn.Query().Where(alerteventcolumn.UserUIDEQ(target.UID)).Count(fixture.ctx()); got != 0 {
		t.Fatalf("告警列配置 = %d, want 0", got)
	}
	if got, _ := fixture.store.client.SystemMessage.Query().Where(systemmessage.UserUIDEQ(target.UID)).Count(fixture.ctx()); got != 0 {
		t.Fatalf("系统消息 = %d, want 0", got)
	}
	for _, inviteUID := range []string{received.UID, sent.UID} {
		if got, _ := fixture.store.client.MemberInvite.Query().Where(memberinvite.UIDEQ(inviteUID)).Count(fixture.ctx()); got != 0 {
			t.Fatalf("邀请 %s 没有清理", inviteUID)
		}
	}
	// 告警页面保留，归属仍是原创建人（账号行已软删，但页面是租户配置）。
	page, err := fixture.store.client.AlertPage.Query().Where(alertpage.UIDEQ(ownedPage.UID)).Only(fixture.ctx())
	if err != nil {
		t.Fatalf("用户创建的告警页面被删除: %v", err)
	}
	if page.OwnerUID != target.UID {
		t.Fatalf("告警页面归属 = %s, want %s", page.OwnerUID, target.UID)
	}
}

// TestDeleteUserRejectsNamespaceOwner 确认仍是未删除租户负责人的用户不能删除，
// 避免租户失去负责人；用户行保持未删除。
func TestDeleteUserRejectsNamespaceOwner(t *testing.T) {
	fixture := newMergeFixture(t)
	owner := fixture.createUser("owner", user.SystemRoleUSER)
	space, _ := fixture.createNamespace("owned", owner.UID)

	err := userDeleteRepo(fixture).DeleteUser(fixture.ctx(), owner.UID)
	assertClientMessage(t, err, "cannot delete a namespace owner")
	if row := fixture.reloadUser(owner.UID); row.DeletedAt != nil {
		t.Fatal("拒绝删除后用户不应被软删")
	}
	// 租户归档后不再是"未删除租户"，这条限制随之解除。
	if _, err := fixture.store.client.Namespace.Update().Where(namespace.UIDEQ(space.UID)).SetDeletedAt(time.Now()).Save(fixture.ctx()); err != nil {
		t.Fatalf("归档租户: %v", err)
	}
	if err := userDeleteRepo(fixture).DeleteUser(fixture.ctx(), owner.UID); err != nil {
		t.Fatalf("租户归档后应允许删除负责人: %v", err)
	}
}

// TestDeleteUserRejectsOrders 确认还有租户订单的用户不能删除，避免订单找不到下单人。
func TestDeleteUserRejectsOrders(t *testing.T) {
	fixture := newMergeFixture(t)
	target := fixture.createUser("buyer", user.SystemRoleUSER)
	owner := fixture.createUser("owner", user.SystemRoleUSER)
	space, _ := fixture.createNamespace("orders", owner.UID)
	fixture.createNamespaceOrder(target.UID, space.UID, "order")

	err := userDeleteRepo(fixture).DeleteUser(fixture.ctx(), target.UID)
	assertClientMessage(t, err, "cannot delete a user with orders")
	if row := fixture.reloadUser(target.UID); row.DeletedAt != nil {
		t.Fatal("拒绝删除后用户不应被软删")
	}
}

// TestDeleteUserMissingUser 覆盖并发删除：用户已经不存在时数据层返回 NotFound，
// biz 层在校验目标用户时也会先拿到 NotFound，不会删除失败。
func TestDeleteUserMissingUser(t *testing.T) {
	// 数据层删除不存在的用户返回 NotFound，供 biz 层在用户已被并发删除时映射成客户端错误。
	fixture := newMergeFixture(t)
	err := userDeleteRepo(fixture).DeleteUser(context.Background(), fixture.prefix+"-missing")
	if !merr.IsNotFound(err) {
		t.Fatalf("删除不存在的用户应返回 NotFound，实际 %v", err)
	}
}
