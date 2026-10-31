package data

import (
	"context"
	"log/slog"
	"testing"
	"time"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/member"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/memberrole"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/namespace"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/role"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/rolepermission"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/user"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/pkg/merr"
)

// TestCreateNamespaceWithoutPresetRolesAndTransferOwner 用真库验证：
// 新租户只创建负责人成员，不创建任何预置角色；移交负责人只更新 owner_uid，
// 不增删角色或角色授权；目标必须是 JOINED 成员。
// 需要独立测试库：ALIOTH_TEST_DSN=postgres://alioth:alioth@127.0.0.1:5432/alioth_test?sslmode=disable
func TestCreateNamespaceWithoutPresetRolesAndTransferOwner(t *testing.T) {
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
	ownerUID := "it-transfer-owner-" + stamp
	targetUID := "it-transfer-target-" + stamp
	outsiderUID := "it-transfer-outsider-" + stamp
	for _, uid := range []string{ownerUID, targetUID, outsiderUID} {
		if _, err := store.client.User.Create().
			SetUID(uid).
			SetName(uid).
			SetEmail(uid + "@example.com").
			Save(ctx); err != nil {
			t.Fatalf("create user %s: %v", uid, err)
		}
	}
	basePkg, err := features.GetPackage(ctx, basePackageUID(ctx, t, store))
	if err != nil {
		t.Fatalf("base package: %v", err)
	}
	ns, err := namespaces.CreateNamespace(ctx, "it-transfer-ns-"+stamp, "it", ownerUID, basePkg.UID)
	if err != nil {
		t.Fatalf("create namespace: %v", err)
	}
	t.Cleanup(func() {
		_, _ = store.client.MemberRole.Delete().Where(memberrole.NamespaceUIDEQ(ns.UID)).Exec(ctx)
		_, _ = store.client.RolePermission.Delete().Where(rolepermission.NamespaceUIDEQ(ns.UID)).Exec(ctx)
		_, _ = store.client.Role.Delete().Where(role.NamespaceUIDEQ(ns.UID)).Exec(ctx)
		_, _ = store.client.Member.Delete().Where(member.NamespaceUIDEQ(ns.UID)).Exec(ctx)
		_, _ = store.client.Namespace.Delete().Where(namespace.UIDEQ(ns.UID)).Exec(ctx)
		_, _ = store.client.User.Delete().Where(user.UIDIn(ownerUID, targetUID, outsiderUID)).Exec(ctx)
	})

	// 新租户没有预置角色，负责人由 owner_uid 决定，并且有一名 JOINED 成员。
	roles, err := store.client.Role.Query().Where(role.NamespaceUIDEQ(ns.UID)).All(ctx)
	if err != nil {
		t.Fatalf("query roles: %v", err)
	}
	if len(roles) != 0 {
		t.Fatalf("new namespace roles = %d, want 0", len(roles))
	}
	ownerMember, err := store.client.Member.Query().Where(
		member.NamespaceUIDEQ(ns.UID),
		member.UserUIDEQ(ownerUID),
		member.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		t.Fatalf("owner member: %v", err)
	}
	if ownerMember.Status != member.StatusJOINED {
		t.Fatalf("owner member status = %s, want JOINED", ownerMember.Status)
	}

	// 目标成员加入后可以接管租户；移交只改 owner_uid。
	if _, err := store.client.Member.Create().
		SetUID(id.New()).
		SetNamespaceUID(ns.UID).
		SetUserUID(targetUID).
		SetStatus(member.StatusJOINED).
		Save(ctx); err != nil {
		t.Fatalf("create target member: %v", err)
	}
	updated, err := namespaces.TransferOwner(ctx, ns.UID, targetUID)
	if err != nil {
		t.Fatalf("transfer owner: %v", err)
	}
	if updated.OwnerUID != targetUID {
		t.Fatalf("owner uid = %q, want %q", updated.OwnerUID, targetUID)
	}
	if roles, err := store.client.Role.Query().Where(role.NamespaceUIDEQ(ns.UID)).All(ctx); err != nil || len(roles) != 0 {
		t.Fatalf("transfer created roles = %d, err = %v, want 0", len(roles), err)
	}
	if links, err := store.client.MemberRole.Query().Where(memberrole.NamespaceUIDEQ(ns.UID)).All(ctx); err != nil || len(links) != 0 {
		t.Fatalf("transfer created member roles = %d, err = %v, want 0", len(links), err)
	}

	// 不是 JOINED 成员的用户不能接管租户。
	_, err = namespaces.TransferOwner(ctx, ns.UID, outsiderUID)
	if err == nil || kerrors.FromError(err).Message != "target must be a joined member" {
		t.Fatalf("transfer to outsider err = %v, want target must be a joined member", err)
	}
	if kerrors.Reason(err) != merr.ReasonForbidden && kerrors.Reason(err) != merr.ReasonClient {
		t.Fatalf("transfer to outsider reason = %q", kerrors.Reason(err))
	}
}
