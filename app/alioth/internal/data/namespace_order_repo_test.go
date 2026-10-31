package data

import (
	"context"
	"log/slog"
	"testing"
	"time"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/featurepackage"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/featurepackageitem"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/member"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/memberinvite"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/memberrole"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/namespace"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/namespacefeature"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/namespaceorder"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/role"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/rolepermission"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/systemmessage"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/user"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
)

// TestNamespaceOrderPaymentFlow 用真库覆盖「下单 -> 支付 -> 创建租户」：
// 订单快照、支付事务、重复支付幂等、取消、本人订单列表和平台按租户查询。
func TestNamespaceOrderPaymentFlow(t *testing.T) {
	source := testDSN(t)
	ctx := context.Background()
	if err := Migrate(ctx, source); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	store, cleanup, err := NewDependencies(
		&conf.Data{Database: &conf.Data_Database{Source: source}},
		&conf.Auth{EncryptKey: "dev-only-encrypt-key"},
		slog.Default(),
	)
	if err != nil {
		t.Fatalf("dependencies: %v", err)
	}
	t.Cleanup(cleanup)
	logger := slog.Default()
	features := NewFeatureRepo(store, logger)
	members := NewMemberRepo(store, logger)
	orders := NewNamespaceOrderRepo(store, logger)

	stamp := time.Now().Format("150405.000000000")
	userUID := "it-order-user-" + stamp
	pkg, err := features.CreatePackage(ctx, bo.FeaturePackage{
		Name: "it-order-pkg-" + stamp, Remark: "it", DurationDays: 30,
		Visibility: perm.PackageVisibilityPublic, MaxMembers: 3, PriceCents: 9900,
		Features: []string{perm.CapabilityAlertRealtime},
	})
	if err != nil {
		t.Fatalf("create package: %v", err)
	}
	namespaceName := "it-order-ns-" + stamp
	t.Cleanup(func() {
		// 真库测试自清理，避免污染下一次运行。
		_, _ = store.client.NamespaceOrder.Delete().Where(namespaceorder.UserUIDEQ(userUID)).Exec(ctx)
		nsRows, _ := store.client.Namespace.Query().Where(namespace.NameEQ(namespaceName)).All(ctx)
		for _, row := range nsRows {
			_, _ = store.client.MemberRole.Delete().Where(memberrole.NamespaceUIDEQ(row.UID)).Exec(ctx)
			_, _ = store.client.Member.Delete().Where(member.NamespaceUIDEQ(row.UID)).Exec(ctx)
			_, _ = store.client.RolePermission.Delete().Where(rolepermission.NamespaceUIDEQ(row.UID)).Exec(ctx)
			_, _ = store.client.Role.Delete().Where(role.NamespaceUIDEQ(row.UID)).Exec(ctx)
			_, _ = store.client.Namespace.Delete().Where(namespace.UIDEQ(row.UID)).Exec(ctx)
		}
		_, _ = store.client.FeaturePackageItem.Delete().Where(featurepackageitem.PackageUIDEQ(pkg.UID)).Exec(ctx)
		_, _ = store.client.FeaturePackage.Delete().Where(featurepackage.UIDEQ(pkg.UID)).Exec(ctx)
	})

	order, err := orders.CreateOrder(ctx, bo.NamespaceOrder{
		UserUID: userUID, NamespaceName: namespaceName, NamespaceRemark: "it",
		PackageUID: pkg.UID, PackageName: pkg.Name, PriceCents: pkg.PriceCents,
		DurationDays: pkg.DurationDays, MaxMembers: pkg.MaxMembers,
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if order.Status != bo.NamespaceOrderPending || order.NamespaceUID != "" || !order.PaidAt.IsZero() {
		t.Fatalf("new order = %+v, want PENDING without namespace", order)
	}

	// 支付前不得存在租户。
	if count, err := store.client.Namespace.Query().Where(namespace.NameEQ(namespaceName)).Count(ctx); err != nil || count != 0 {
		t.Fatalf("namespace count before pay = %d, err = %v", count, err)
	}

	paid, ns, err := orders.PayOrder(ctx, order.UID, userUID)
	if err != nil {
		t.Fatalf("pay order: %v", err)
	}
	if paid.Status != bo.NamespaceOrderPaid || paid.NamespaceUID != ns.UID || paid.PaidAt.IsZero() {
		t.Fatalf("paid order = %+v, want PAID with namespace %q", paid, ns.UID)
	}
	if ns.PackageUID != pkg.UID {
		t.Fatalf("namespace package = %q, want %q", ns.PackageUID, pkg.UID)
	}
	state, err := features.NamespaceFeatureState(ctx, ns.UID)
	if err != nil {
		t.Fatalf("namespace feature state: %v", err)
	}
	if state.PackageUID != pkg.UID || state.Expired {
		t.Fatalf("feature state = %+v", state)
	}
	if delta := time.Until(state.ExpiresAt); delta < 29*24*time.Hour || delta > 31*24*time.Hour {
		t.Fatalf("expires in %v, want about 30 days", delta)
	}
	seats, err := members.CountJoined(ctx, ns.UID)
	if err != nil {
		t.Fatalf("count joined: %v", err)
	}
	if seats != 1 {
		t.Fatalf("owner seats = %d, want 1", seats)
	}

	// 重复支付返回同一个租户，不重复创建。
	againOrder, againNS, err := orders.PayOrder(ctx, order.UID, userUID)
	if err != nil {
		t.Fatalf("pay order again: %v", err)
	}
	if againOrder.NamespaceUID != ns.UID || againNS.UID != ns.UID {
		t.Fatalf("replay pay = %q/%q, want %q", againOrder.NamespaceUID, againNS.UID, ns.UID)
	}
	if count, err := store.client.Namespace.Query().Where(namespace.NameEQ(namespaceName)).Count(ctx); err != nil || count != 1 {
		t.Fatalf("namespace count after replay = %d, err = %v", count, err)
	}

	// 别人的订单不能支付，也不能取消。
	otherOrder, err := orders.CreateOrder(ctx, bo.NamespaceOrder{
		UserUID: userUID + "-other", NamespaceName: namespaceName + "-other",
		PackageUID: pkg.UID, PackageName: pkg.Name, PriceCents: pkg.PriceCents,
		DurationDays: pkg.DurationDays, MaxMembers: pkg.MaxMembers,
	})
	if err != nil {
		t.Fatalf("create other order: %v", err)
	}
	if _, _, err := orders.PayOrder(ctx, otherOrder.UID, userUID); err == nil || !kerrors.IsNotFound(err) {
		t.Fatalf("pay other order err = %v, want not found", err)
	}
	if _, err := orders.CancelOrder(ctx, otherOrder.UID, userUID); err == nil || !kerrors.IsNotFound(err) {
		t.Fatalf("cancel other order err = %v, want not found", err)
	}

	// 取消待支付订单；已支付订单不能取消。
	cancelled, err := orders.CancelOrder(ctx, otherOrder.UID, otherOrder.UserUID)
	if err != nil {
		t.Fatalf("cancel order: %v", err)
	}
	if cancelled.Status != bo.NamespaceOrderCancelled {
		t.Fatalf("cancelled order status = %q", cancelled.Status)
	}
	if _, _, err := orders.PayOrder(ctx, cancelled.UID, cancelled.UserUID); err == nil ||
		kerrors.FromError(err).Message != "order is cancelled" {
		t.Fatalf("pay cancelled order err = %v", err)
	}
	if _, err := orders.CancelOrder(ctx, order.UID, userUID); err == nil ||
		kerrors.FromError(err).Message != "paid order cannot be cancelled" {
		t.Fatalf("cancel paid order err = %v", err)
	}

	// 本人订单列表按创建时间倒序；平台按租户查订单只返回该租户的已支付订单。
	items, total, err := orders.ListOrders(ctx, userUID, 1, 10, "")
	if err != nil {
		t.Fatalf("list orders: %v", err)
	}
	if total != 1 || len(items) != 1 || items[0].UID != order.UID {
		t.Fatalf("own orders = %d/%v, want only paid order %q", total, items, order.UID)
	}
	nsItems, nsTotal, err := orders.ListNamespaceOrders(ctx, ns.UID, 1, 10)
	if err != nil {
		t.Fatalf("list namespace orders: %v", err)
	}
	if nsTotal != 1 || len(nsItems) != 1 || nsItems[0].UID != order.UID {
		t.Fatalf("namespace orders = %d/%v", nsTotal, nsItems)
	}
}

// TestMemberCapacityLimit 用真库验证套餐人数上限：
// JOINED 成员数达到上限后确认邀请被拒绝；把租户换到上限更小的套餐后同样拒绝。
func TestMemberCapacityLimit(t *testing.T) {
	source := testDSN(t)
	ctx := context.Background()
	if err := Migrate(ctx, source); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	store, cleanup, err := NewDependencies(
		&conf.Data{Database: &conf.Data_Database{Source: source}},
		&conf.Auth{EncryptKey: "dev-only-encrypt-key"},
		slog.Default(),
	)
	if err != nil {
		t.Fatalf("dependencies: %v", err)
	}
	t.Cleanup(cleanup)
	logger := slog.Default()
	features := NewFeatureRepo(store, logger)
	namespaces := NewNamespaceRepo(store, logger)
	invites := NewInviteRepo(store, logger)
	members := NewMemberRepo(store, logger)

	stamp := time.Now().Format("150405.000000000")
	ownerUID := "it-seat-owner-" + stamp
	userA := "it-seat-a-" + stamp
	userB := "it-seat-b-" + stamp
	emailA := userA + "@example.com"
	emailB := userB + "@example.com"
	pkg, err := features.CreatePackage(ctx, bo.FeaturePackage{
		Name: "it-seat-pkg-" + stamp, DurationDays: 0,
		Visibility: perm.PackageVisibilityPublic, MaxMembers: 2,
		Features: []string{perm.CapabilityMember},
	})
	if err != nil {
		t.Fatalf("create package: %v", err)
	}
	smallPkg, err := features.CreatePackage(ctx, bo.FeaturePackage{
		Name: "it-seat-small-" + stamp, DurationDays: 0,
		Visibility: perm.PackageVisibilityBeta, MaxMembers: 1,
		Features: []string{perm.CapabilityMember},
	})
	if err != nil {
		t.Fatalf("create small package: %v", err)
	}
	nsRow, err := namespaces.CreateNamespace(ctx, "it-seat-ns-"+stamp, "it", ownerUID, pkg.UID)
	if err != nil {
		t.Fatalf("create namespace: %v", err)
	}
	namespaceUID := nsRow.UID
	if _, err := store.client.User.Create().SetUID(userA).SetEmail(emailA).Save(ctx); err != nil {
		t.Fatalf("create user a: %v", err)
	}
	if _, err := store.client.User.Create().SetUID(userB).SetEmail(emailB).Save(ctx); err != nil {
		t.Fatalf("create user b: %v", err)
	}
	t.Cleanup(func() {
		_, _ = store.client.SystemMessage.Delete().Where(systemmessage.NamespaceUIDEQ(namespaceUID)).Exec(ctx)
		_, _ = store.client.MemberInvite.Delete().Where(memberinvite.NamespaceUIDEQ(namespaceUID)).Exec(ctx)
		_, _ = store.client.MemberRole.Delete().Where(memberrole.NamespaceUIDEQ(namespaceUID)).Exec(ctx)
		_, _ = store.client.Member.Delete().Where(member.NamespaceUIDEQ(namespaceUID)).Exec(ctx)
		_, _ = store.client.RolePermission.Delete().Where(rolepermission.NamespaceUIDEQ(namespaceUID)).Exec(ctx)
		_, _ = store.client.Role.Delete().Where(role.NamespaceUIDEQ(namespaceUID)).Exec(ctx)
		_, _ = store.client.Namespace.Delete().Where(namespace.UIDEQ(namespaceUID)).Exec(ctx)
		_, _ = store.client.User.Delete().Where(user.UIDIn(userA, userB)).Exec(ctx)
		for _, uid := range []string{pkg.UID, smallPkg.UID} {
			_, _ = store.client.FeaturePackageItem.Delete().Where(featurepackageitem.PackageUIDEQ(uid)).Exec(ctx)
			_, _ = store.client.FeaturePackage.Delete().Where(featurepackage.UIDEQ(uid)).Exec(ctx)
		}
	})
	// 新租户不再预置角色；会员名额测试自己创建一个角色用于邀请。
	memberRole, err := store.client.Role.Create().
		SetUID(id.New()).
		SetNamespaceUID(namespaceUID).
		SetName("it-seat-member").
		SetStatus(role.StatusACTIVE).
		Save(ctx)
	if err != nil {
		t.Fatalf("member role: %v", err)
	}

	// 负责人已占 1 个名额：再确认 1 位成员后到达上限 2。
	inviteA, err := invites.UpsertInvite(ctx, bo.MemberInvite{
		NamespaceUID: namespaceUID, Email: emailA, RoleUID: memberRole.UID,
		InviterUID: ownerUID, ExpiresAt: time.Now().Add(time.Hour),
	}, "it-seat-hash-a-"+stamp)
	if err != nil {
		t.Fatalf("invite a: %v", err)
	}
	if _, err := invites.AcceptInvite(ctx, "it-seat-hash-a-"+stamp, inviteA.UID, userA, emailA); err != nil {
		t.Fatalf("accept a: %v", err)
	}
	if seats, err := members.CountJoined(ctx, namespaceUID); err != nil || seats != 2 {
		t.Fatalf("seats after a = %d, err = %v", seats, err)
	}

	inviteB, err := invites.UpsertInvite(ctx, bo.MemberInvite{
		NamespaceUID: namespaceUID, Email: emailB, RoleUID: memberRole.UID,
		InviterUID: ownerUID, ExpiresAt: time.Now().Add(time.Hour),
	}, "it-seat-hash-b-"+stamp)
	if err != nil {
		t.Fatalf("invite b: %v", err)
	}
	if _, err := invites.AcceptInvite(ctx, "it-seat-hash-b-"+stamp, inviteB.UID, userB, emailB); err == nil ||
		kerrors.FromError(err).Message != "member limit reached" {
		t.Fatalf("accept b err = %v, want member limit reached", err)
	}
	if seats, err := members.CountJoined(ctx, namespaceUID); err != nil || seats != 2 {
		t.Fatalf("seats after rejected b = %d, err = %v", seats, err)
	}

	// 换到上限 1 的套餐：已有 2 名成员保留，新的确认继续被拒绝。
	if err := features.SetNamespaceFeatureState(ctx, namespaceUID, smallPkg.UID, []string{perm.CapabilityMember}, time.Time{}); err != nil {
		t.Fatalf("switch package: %v", err)
	}
	if limit, err := features.NamespaceMemberLimit(ctx, namespaceUID); err != nil || limit != 1 {
		t.Fatalf("limit after switch = %d, err = %v", limit, err)
	}
	if _, err := invites.AcceptInvite(ctx, "it-seat-hash-b-"+stamp, inviteB.UID, userB, emailB); err == nil ||
		kerrors.FromError(err).Message != "member limit reached" {
		t.Fatalf("accept b after downgrade err = %v, want member limit reached", err)
	}
}

// TestNamespaceUpgradeAndRenewalFlow 用真库验证升级 / 续期的支付事务：
// 升级切换套餐并重算到期时间，续期在原到期时间上累加，永久套餐不能续期。
func TestNamespaceUpgradeAndRenewalFlow(t *testing.T) {
	source := testDSN(t)
	ctx := context.Background()
	if err := Migrate(ctx, source); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	store, cleanup, err := NewDependencies(
		&conf.Data{Database: &conf.Data_Database{Source: source}},
		&conf.Auth{EncryptKey: "dev-only-encrypt-key"},
		slog.Default(),
	)
	if err != nil {
		t.Fatalf("dependencies: %v", err)
	}
	t.Cleanup(cleanup)
	logger := slog.Default()
	features := NewFeatureRepo(store, logger)
	namespaces := NewNamespaceRepo(store, logger)
	orders := NewNamespaceOrderRepo(store, logger)

	stamp := time.Now().Format("150405.000000000")
	userUID := "it-upgrade-user-" + stamp
	ownerUID := "it-upgrade-owner-" + stamp
	basePkg, err := features.CreatePackage(ctx, bo.FeaturePackage{
		Name: "it-upgrade-base-" + stamp, DurationDays: 30,
		Visibility: perm.PackageVisibilityPublic, MaxMembers: 2, PriceCents: 1000,
		Features: []string{perm.CapabilityAlertRealtime},
	})
	if err != nil {
		t.Fatalf("create base package: %v", err)
	}
	upgradePkg, err := features.CreatePackage(ctx, bo.FeaturePackage{
		Name: "it-upgrade-pro-" + stamp, DurationDays: 30,
		Visibility: perm.PackageVisibilityPublic, MaxMembers: 5, PriceCents: 9900,
		Features: []string{perm.CapabilityAlertRealtime, perm.CapabilityMember},
	})
	if err != nil {
		t.Fatalf("create upgrade package: %v", err)
	}
	permanentPkg, err := features.CreatePackage(ctx, bo.FeaturePackage{
		Name: "it-upgrade-permanent-" + stamp, DurationDays: 0,
		Visibility: perm.PackageVisibilityPublic, MaxMembers: 0, PriceCents: 19900,
		Features: []string{perm.CapabilityAlertRealtime},
	})
	if err != nil {
		t.Fatalf("create permanent package: %v", err)
	}
	nsRow, err := namespaces.CreateNamespace(ctx, "it-upgrade-ns-"+stamp, "it", ownerUID, basePkg.UID)
	if err != nil {
		t.Fatalf("create namespace: %v", err)
	}
	t.Cleanup(func() {
		_, _ = store.client.NamespaceOrder.Delete().Where(namespaceorder.UserUIDEQ(userUID)).Exec(ctx)
		_, _ = store.client.MemberRole.Delete().Where(memberrole.NamespaceUIDEQ(nsRow.UID)).Exec(ctx)
		_, _ = store.client.Member.Delete().Where(member.NamespaceUIDEQ(nsRow.UID)).Exec(ctx)
		_, _ = store.client.RolePermission.Delete().Where(rolepermission.NamespaceUIDEQ(nsRow.UID)).Exec(ctx)
		_, _ = store.client.Role.Delete().Where(role.NamespaceUIDEQ(nsRow.UID)).Exec(ctx)
		_, _ = store.client.Namespace.Delete().Where(namespace.UIDEQ(nsRow.UID)).Exec(ctx)
		for _, uid := range []string{basePkg.UID, upgradePkg.UID, permanentPkg.UID} {
			_, _ = store.client.FeaturePackageItem.Delete().Where(featurepackageitem.PackageUIDEQ(uid)).Exec(ctx)
			_, _ = store.client.FeaturePackage.Delete().Where(featurepackage.UIDEQ(uid)).Exec(ctx)
		}
	})

	beforeUpgrade, err := features.NamespaceFeatureState(ctx, nsRow.UID)
	if err != nil {
		t.Fatalf("base state: %v", err)
	}
	upgrade, err := orders.CreateOrder(ctx, bo.NamespaceOrder{
		UserUID: userUID, Kind: bo.NamespaceOrderUpgrade, NamespaceUID: nsRow.UID,
		NamespaceName: nsRow.Name, PackageUID: upgradePkg.UID, PackageName: upgradePkg.Name,
		PriceCents: upgradePkg.PriceCents, DurationDays: upgradePkg.DurationDays, MaxMembers: upgradePkg.MaxMembers,
	})
	if err != nil {
		t.Fatalf("create upgrade order: %v", err)
	}
	paidUpgrade, upgradedNS, err := orders.PayOrder(ctx, upgrade.UID, userUID)
	if err != nil {
		t.Fatalf("pay upgrade: %v", err)
	}
	if paidUpgrade.Kind != bo.NamespaceOrderUpgrade || upgradedNS.PackageUID != upgradePkg.UID {
		t.Fatalf("upgraded order/namespace = %+v / %+v", paidUpgrade, upgradedNS)
	}
	afterUpgrade, err := features.NamespaceFeatureState(ctx, nsRow.UID)
	if err != nil {
		t.Fatalf("upgraded state: %v", err)
	}
	if afterUpgrade.PackageUID != upgradePkg.UID {
		t.Fatalf("package after upgrade = %q", afterUpgrade.PackageUID)
	}
	if !afterUpgrade.ExpiresAt.After(beforeUpgrade.ExpiresAt) {
		t.Fatalf("upgrade expiry = %v, want later than %v", afterUpgrade.ExpiresAt, beforeUpgrade.ExpiresAt)
	}
	if count, err := store.client.Namespace.Query().Where(namespace.NameEQ(nsRow.Name)).Count(ctx); err != nil || count != 1 {
		t.Fatalf("namespace count after upgrade = %d, err = %v", count, err)
	}

	renewal, err := orders.CreateOrder(ctx, bo.NamespaceOrder{
		UserUID: userUID, Kind: bo.NamespaceOrderRenew, NamespaceUID: nsRow.UID,
		NamespaceName: nsRow.Name, PackageUID: upgradePkg.UID, PackageName: upgradePkg.Name,
		PriceCents: upgradePkg.PriceCents, DurationDays: upgradePkg.DurationDays, MaxMembers: upgradePkg.MaxMembers,
	})
	if err != nil {
		t.Fatalf("create renewal order: %v", err)
	}
	if _, renewedNS, err := orders.PayOrder(ctx, renewal.UID, userUID); err != nil {
		t.Fatalf("pay renewal: %v", err)
	} else if renewedNS.PackageUID != upgradePkg.UID {
		t.Fatalf("package after renewal = %q", renewedNS.PackageUID)
	}
	afterRenewal, err := features.NamespaceFeatureState(ctx, nsRow.UID)
	if err != nil {
		t.Fatalf("renewed state: %v", err)
	}
	wantRenewal := afterUpgrade.ExpiresAt.AddDate(0, 0, int(upgradePkg.DurationDays))
	if delta := afterRenewal.ExpiresAt.Sub(wantRenewal); delta < -time.Minute || delta > time.Minute {
		t.Fatalf("renewed expiry = %v, want about %v", afterRenewal.ExpiresAt, wantRenewal)
	}

	// 升级到永久套餐后，续期订单在支付时被拒绝。
	permanentOrder, err := orders.CreateOrder(ctx, bo.NamespaceOrder{
		UserUID: userUID, Kind: bo.NamespaceOrderUpgrade, NamespaceUID: nsRow.UID,
		NamespaceName: nsRow.Name, PackageUID: permanentPkg.UID, PackageName: permanentPkg.Name,
		PriceCents: permanentPkg.PriceCents, DurationDays: permanentPkg.DurationDays, MaxMembers: permanentPkg.MaxMembers,
	})
	if err != nil {
		t.Fatalf("create permanent order: %v", err)
	}
	if _, permanentNS, err := orders.PayOrder(ctx, permanentOrder.UID, userUID); err != nil {
		t.Fatalf("pay permanent upgrade: %v", err)
	} else if !permanentNS.PackageExpiresAt.IsZero() {
		t.Fatalf("permanent package expiry = %v", permanentNS.PackageExpiresAt)
	}
	renewPermanent, err := orders.CreateOrder(ctx, bo.NamespaceOrder{
		UserUID: userUID, Kind: bo.NamespaceOrderRenew, NamespaceUID: nsRow.UID,
		NamespaceName: nsRow.Name, PackageUID: permanentPkg.UID, PackageName: permanentPkg.Name,
		PriceCents: permanentPkg.PriceCents, DurationDays: 0, MaxMembers: permanentPkg.MaxMembers,
	})
	if err != nil {
		t.Fatalf("create permanent renewal order: %v", err)
	}
	if _, _, err := orders.PayOrder(ctx, renewPermanent.UID, userUID); err == nil ||
		kerrors.FromError(err).Message != "package is permanent" {
		t.Fatalf("pay permanent renewal err = %v, want package is permanent", err)
	}
}

// TestPackageEditsDoNotChangeExistingTenant 用真库验证租户授权快照：
// 管理员修改套餐内容不影响已生效租户；租户仍可加开/关闭套餐外的能力。
func TestPackageEditsDoNotChangeExistingTenant(t *testing.T) {
	source := testDSN(t)
	ctx := context.Background()
	if err := Migrate(ctx, source); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	store, cleanup, err := NewDependencies(
		&conf.Data{Database: &conf.Data_Database{Source: source}},
		&conf.Auth{EncryptKey: "dev-only-encrypt-key"},
		slog.Default(),
	)
	if err != nil {
		t.Fatalf("dependencies: %v", err)
	}
	t.Cleanup(cleanup)
	logger := slog.Default()
	features := NewFeatureRepo(store, logger)
	namespaces := NewNamespaceRepo(store, logger)

	stamp := time.Now().Format("150405.000000000")
	ownerUID := "it-snapshot-owner-" + stamp
	pkg, err := features.CreatePackage(ctx, bo.FeaturePackage{
		Name: "it-snapshot-pkg-" + stamp, DurationDays: 30,
		Visibility: perm.PackageVisibilityPublic, MaxMembers: 5,
		Features: []string{perm.CapabilityAlertRealtime, perm.CapabilityDatasourcePrometheus},
	})
	if err != nil {
		t.Fatalf("create package: %v", err)
	}
	nsRow, err := namespaces.CreateNamespace(ctx, "it-snapshot-ns-"+stamp, "it", ownerUID, pkg.UID)
	if err != nil {
		t.Fatalf("create namespace: %v", err)
	}
	t.Cleanup(func() {
		_, _ = store.client.NamespaceFeature.Delete().Where(namespacefeature.NamespaceUIDEQ(nsRow.UID)).Exec(ctx)
		_, _ = store.client.MemberRole.Delete().Where(memberrole.NamespaceUIDEQ(nsRow.UID)).Exec(ctx)
		_, _ = store.client.Member.Delete().Where(member.NamespaceUIDEQ(nsRow.UID)).Exec(ctx)
		_, _ = store.client.RolePermission.Delete().Where(rolepermission.NamespaceUIDEQ(nsRow.UID)).Exec(ctx)
		_, _ = store.client.Role.Delete().Where(role.NamespaceUIDEQ(nsRow.UID)).Exec(ctx)
		_, _ = store.client.Namespace.Delete().Where(namespace.UIDEQ(nsRow.UID)).Exec(ctx)
		_, _ = store.client.FeaturePackageItem.Delete().Where(featurepackageitem.PackageUIDEQ(pkg.UID)).Exec(ctx)
		_, _ = store.client.FeaturePackage.Delete().Where(featurepackage.UIDEQ(pkg.UID)).Exec(ctx)
	})

	// 管理员把套餐能力改成历史告警：已有租户必须保持购买时的快照。
	pkg.Features = []string{perm.CapabilityAlertHistory}
	if _, err := features.UpdatePackage(ctx, *pkg); err != nil {
		t.Fatalf("update package: %v", err)
	}
	state, err := features.NamespaceFeatureState(ctx, nsRow.UID)
	if err != nil {
		t.Fatalf("feature state: %v", err)
	}
	if !sameStrings(state.PackageFeatures, []string{perm.CapabilityAlertRealtime, perm.CapabilityDatasourcePrometheus}) {
		t.Fatalf("tenant snapshot changed by package edit: %v", state.PackageFeatures)
	}
	capabilities, err := features.EffectiveCapabilities(ctx, nsRow.UID)
	if err != nil {
		t.Fatalf("effective capabilities: %v", err)
	}
	if !sameStrings(capabilities, []string{perm.CapabilityAlertRealtime, perm.CapabilityDatasourcePrometheus}) {
		t.Fatalf("tenant capabilities changed by package edit: %v", capabilities)
	}

	// 同一套餐只调整额外能力时，不能把套餐快照改成新内容。
	if err := features.SetNamespaceFeatureState(ctx, nsRow.UID, pkg.UID,
		[]string{perm.CapabilityAlertHistory}, time.Time{}); err == nil {
		t.Fatal("downgraded same-package assignment accepted")
	}
	// 加开套餐外能力可以：快照不变，历史告警只作为额外能力。
	if err := features.SetNamespaceFeatureState(ctx, nsRow.UID, pkg.UID,
		[]string{perm.CapabilityAlertRealtime, perm.CapabilityDatasourcePrometheus, perm.CapabilityAlertHistory},
		time.Time{}); err != nil {
		t.Fatalf("add extra capability: %v", err)
	}
	state, err = features.NamespaceFeatureState(ctx, nsRow.UID)
	if err != nil {
		t.Fatalf("feature state after extra: %v", err)
	}
	if !sameStrings(state.PackageFeatures, []string{perm.CapabilityAlertRealtime, perm.CapabilityDatasourcePrometheus}) {
		t.Fatalf("snapshot changed by extra capability: %v", state.PackageFeatures)
	}
	if !sameStrings(state.ExtraFeatures, []string{perm.CapabilityAlertHistory}) {
		t.Fatalf("extra capabilities = %v", state.ExtraFeatures)
	}

	// 新租户购买同一套餐时使用套餐当前内容（历史告警）。
	newNS, err := namespaces.CreateNamespace(ctx, "it-snapshot-new-"+stamp, "it", ownerUID, pkg.UID)
	if err != nil {
		t.Fatalf("create new namespace: %v", err)
	}
	t.Cleanup(func() {
		_, _ = store.client.NamespaceFeature.Delete().Where(namespacefeature.NamespaceUIDEQ(newNS.UID)).Exec(ctx)
		_, _ = store.client.MemberRole.Delete().Where(memberrole.NamespaceUIDEQ(newNS.UID)).Exec(ctx)
		_, _ = store.client.Member.Delete().Where(member.NamespaceUIDEQ(newNS.UID)).Exec(ctx)
		_, _ = store.client.RolePermission.Delete().Where(rolepermission.NamespaceUIDEQ(newNS.UID)).Exec(ctx)
		_, _ = store.client.Role.Delete().Where(role.NamespaceUIDEQ(newNS.UID)).Exec(ctx)
		_, _ = store.client.Namespace.Delete().Where(namespace.UIDEQ(newNS.UID)).Exec(ctx)
	})
	newCapabilities, err := features.EffectiveCapabilities(ctx, newNS.UID)
	if err != nil {
		t.Fatalf("new tenant capabilities: %v", err)
	}
	if !sameStrings(newCapabilities, []string{perm.CapabilityAlertHistory}) {
		t.Fatalf("new tenant capabilities = %v", newCapabilities)
	}
}

// TestSyncPackageToNamespaces 用真库验证「同步套餐授权」：
// 指定租户只更新选中的快照，全部租户覆盖该套餐下所有租户；额外能力保留，
// 已并入套餐的额外能力会被清理，未使用该套餐的租户计入 skipped。
func TestSyncPackageToNamespaces(t *testing.T) {
	source := testDSN(t)
	ctx := context.Background()
	if err := Migrate(ctx, source); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	store, cleanup, err := NewDependencies(
		&conf.Data{Database: &conf.Data_Database{Source: source}},
		&conf.Auth{EncryptKey: "dev-only-encrypt-key"},
		slog.Default(),
	)
	if err != nil {
		t.Fatalf("dependencies: %v", err)
	}
	t.Cleanup(cleanup)
	logger := slog.Default()
	features := NewFeatureRepo(store, logger)
	namespaces := NewNamespaceRepo(store, logger)

	stamp := time.Now().Format("150405.000000000")
	ownerUID := "it-sync-owner-" + stamp
	pkg, err := features.CreatePackage(ctx, bo.FeaturePackage{
		Name: "it-sync-pkg-" + stamp, DurationDays: 30,
		Visibility: perm.PackageVisibilityPublic, MaxMembers: 5,
		Features: []string{perm.CapabilityAlertRealtime},
	})
	if err != nil {
		t.Fatalf("create package: %v", err)
	}
	newNamespace := func(name string) *bo.Namespace {
		t.Helper()
		ns, err := namespaces.CreateNamespace(ctx, name, "it", ownerUID, pkg.UID)
		if err != nil {
			t.Fatalf("create namespace %s: %v", name, err)
		}
		return ns
	}
	nsOne := newNamespace("it-sync-one-" + stamp)
	nsTwo := newNamespace("it-sync-two-" + stamp)
	otherNS, err := namespaces.CreateNamespace(ctx, "it-sync-other-"+stamp, "it", ownerUID,
		basePackageUID(ctx, t, store))
	if err != nil {
		t.Fatalf("create other namespace: %v", err)
	}
	t.Cleanup(func() {
		for _, uid := range []string{nsOne.UID, nsTwo.UID, otherNS.UID} {
			_, _ = store.client.NamespaceFeature.Delete().Where(namespacefeature.NamespaceUIDEQ(uid)).Exec(ctx)
			_, _ = store.client.MemberRole.Delete().Where(memberrole.NamespaceUIDEQ(uid)).Exec(ctx)
			_, _ = store.client.Member.Delete().Where(member.NamespaceUIDEQ(uid)).Exec(ctx)
			_, _ = store.client.RolePermission.Delete().Where(rolepermission.NamespaceUIDEQ(uid)).Exec(ctx)
			_, _ = store.client.Role.Delete().Where(role.NamespaceUIDEQ(uid)).Exec(ctx)
			_, _ = store.client.Namespace.Delete().Where(namespace.UIDEQ(uid)).Exec(ctx)
		}
		_, _ = store.client.FeaturePackageItem.Delete().Where(featurepackageitem.PackageUIDEQ(pkg.UID)).Exec(ctx)
		_, _ = store.client.FeaturePackage.Delete().Where(featurepackage.UIDEQ(pkg.UID)).Exec(ctx)
	})

	// nsOne 额外开通历史告警，作为套餐外能力保留。
	if err := features.SetNamespaceFeatureState(ctx, nsOne.UID, pkg.UID,
		[]string{perm.CapabilityAlertRealtime, perm.CapabilityAlertHistory}, time.Time{}); err != nil {
		t.Fatalf("add extra capability: %v", err)
	}
	// 套餐内容改为 Prometheus：同步前只有新租户会得到新内容。
	pkg.Features = []string{perm.CapabilityDatasourcePrometheus}
	if _, err := features.UpdatePackage(ctx, *pkg); err != nil {
		t.Fatalf("update package: %v", err)
	}
	synced, skipped, err := features.SyncPackageToNamespaces(ctx, pkg.UID, []string{nsOne.UID, otherNS.UID})
	if err != nil {
		t.Fatalf("sync selected: %v", err)
	}
	if synced != 1 || skipped != 1 {
		t.Fatalf("sync selected = %d/%d, want 1/1", synced, skipped)
	}
	oneState, err := features.NamespaceFeatureState(ctx, nsOne.UID)
	if err != nil {
		t.Fatalf("ns one state: %v", err)
	}
	if !sameStrings(oneState.PackageFeatures, []string{perm.CapabilityDatasourcePrometheus}) {
		t.Fatalf("ns one snapshot = %v", oneState.PackageFeatures)
	}
	if !sameStrings(oneState.ExtraFeatures, []string{perm.CapabilityAlertHistory}) {
		t.Fatalf("ns one extra = %v", oneState.ExtraFeatures)
	}
	twoState, err := features.NamespaceFeatureState(ctx, nsTwo.UID)
	if err != nil {
		t.Fatalf("ns two state: %v", err)
	}
	if !sameStrings(twoState.PackageFeatures, []string{perm.CapabilityAlertRealtime}) {
		t.Fatalf("ns two snapshot changed without sync: %v", twoState.PackageFeatures)
	}

	// 全部同步：两个使用该套餐的租户都更新；其他套餐的租户不受影响。
	synced, skipped, err = features.SyncPackageToNamespaces(ctx, pkg.UID, nil)
	if err != nil {
		t.Fatalf("sync all: %v", err)
	}
	if synced != 2 || skipped != 0 {
		t.Fatalf("sync all = %d/%d, want 2/0", synced, skipped)
	}
	twoState, err = features.NamespaceFeatureState(ctx, nsTwo.UID)
	if err != nil {
		t.Fatalf("ns two state after all: %v", err)
	}
	if !sameStrings(twoState.PackageFeatures, []string{perm.CapabilityDatasourcePrometheus}) {
		t.Fatalf("ns two snapshot after all = %v", twoState.PackageFeatures)
	}
	otherState, err := features.NamespaceFeatureState(ctx, otherNS.UID)
	if err != nil {
		t.Fatalf("other state: %v", err)
	}
	if sameStrings(otherState.PackageFeatures, []string{perm.CapabilityDatasourcePrometheus}) {
		t.Fatal("other package tenant was synced")
	}

	// 套餐把历史告警并入后，nsOne 的额外能力被清理，不重复展示。
	pkg.Features = []string{perm.CapabilityDatasourcePrometheus, perm.CapabilityAlertHistory}
	if _, err := features.UpdatePackage(ctx, *pkg); err != nil {
		t.Fatalf("update package again: %v", err)
	}
	if _, _, err := features.SyncPackageToNamespaces(ctx, pkg.UID, []string{nsOne.UID}); err != nil {
		t.Fatalf("sync merged extra: %v", err)
	}
	oneState, err = features.NamespaceFeatureState(ctx, nsOne.UID)
	if err != nil {
		t.Fatalf("ns one state after merge: %v", err)
	}
	if len(oneState.ExtraFeatures) != 0 {
		t.Fatalf("extra after merge = %v, want none", oneState.ExtraFeatures)
	}
	if !sameStrings(oneState.PackageFeatures, []string{
		perm.CapabilityDatasourcePrometheus, perm.CapabilityAlertHistory,
	}) {
		t.Fatalf("ns one snapshot after merge = %v", oneState.PackageFeatures)
	}
}
