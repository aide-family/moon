package role

import (
	"context"
	"log/slog"
	"slices"
	"testing"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
)

type roleRepoStub struct {
	repository.RoleRepo
	permissions map[string]struct{}
	member      bool
}

func (s roleRepoStub) Permissions(context.Context, string, string) (map[string]struct{}, error) {
	return s.permissions, nil
}

func (s roleRepoStub) IsMember(context.Context, string, string) (bool, error) {
	return s.member, nil
}

type userRepoStub struct {
	repository.UserRepo
	account *bo.Account
}

func (s userRepoStub) UserByUID(context.Context, string) (*bo.Account, error) {
	return s.account, nil
}

type featureRepoStub struct {
	repository.FeatureRepo
	features []string
}

func (s featureRepoStub) EffectiveFeatures(context.Context, string) ([]string, error) {
	return s.features, nil
}

func authorityFor(account *bo.Account, member bool, permissions ...string) *Authority {
	granted := make(map[string]struct{}, len(permissions))
	for _, name := range permissions {
		granted[name] = struct{}{}
	}
	return &Authority{
		roles:    roleRepoStub{permissions: granted, member: member},
		features: featureRepoStub{features: perm.FeatureCodes()},
		access: access.Checker{
			Users: userRepoStub{account: account},
			Log:   slog.New(slog.DiscardHandler),
		},
	}
}

func requestContext(namespace string) context.Context {
	ctx := actor.WithUser(context.Background(), "user-1")
	return actor.WithNamespace(ctx, namespace)
}

// TestMineReadOnlyForSystemAdmin 确认非成员的系统管理员拿到租户只读目录，并被标记 readOnly。
func TestMineReadOnlyForSystemAdmin(t *testing.T) {
	authority := authorityFor(&bo.Account{UID: "user-1", Status: "ACTIVE", SystemRole: perm.SystemRoleSystemAdmin}, false)
	names, _, readOnly, err := authority.Mine(requestContext("namespace-1"))
	if err != nil {
		t.Fatalf("mine: %v", err)
	}
	if !readOnly {
		t.Fatal("system admin without membership is not read-only")
	}
	if !slices.Equal(names, perm.ReadOnlyNames()) {
		t.Fatalf("read-only names = %v, want %v", names, perm.ReadOnlyNames())
	}
}

// TestMineMemberKeepsRolePermissions 确认成员身份优先：系统管理员同时是成员时按角色权限返回。
func TestMineMemberKeepsRolePermissions(t *testing.T) {
	authority := authorityFor(
		&bo.Account{UID: "user-1", Status: "ACTIVE", SystemRole: perm.SystemRoleSystemAdmin},
		true,
		perm.RoleRead,
	)
	names, _, readOnly, err := authority.Mine(requestContext("namespace-1"))
	if err != nil {
		t.Fatalf("mine: %v", err)
	}
	if readOnly {
		t.Fatal("member treated as read-only")
	}
	if !slices.Equal(names, []string{perm.RoleRead}) {
		t.Fatalf("member names = %v, want [%s]", names, perm.RoleRead)
	}
}

func TestMineEmptyForNormalUser(t *testing.T) {
	authority := authorityFor(&bo.Account{UID: "user-1", Status: "ACTIVE", SystemRole: perm.SystemRoleUser}, false)
	names, _, readOnly, err := authority.Mine(requestContext("namespace-1"))
	if err != nil {
		t.Fatalf("mine: %v", err)
	}
	if readOnly || len(names) != 0 {
		t.Fatalf("normal user names = %v, readOnly = %v", names, readOnly)
	}
}

// TestMineMemberWithoutPermissionsIsNotReadOnly 确认「成员但没有有效权限」不会被当成只读切入。
func TestMineMemberWithoutPermissionsIsNotReadOnly(t *testing.T) {
	authority := authorityFor(&bo.Account{UID: "user-1", Status: "ACTIVE", SystemRole: perm.SystemRoleSystemAdmin}, true)
	names, _, readOnly, err := authority.Mine(requestContext("namespace-1"))
	if err != nil {
		t.Fatalf("mine: %v", err)
	}
	if readOnly || len(names) != 0 {
		t.Fatalf("member without permissions names = %v, readOnly = %v", names, readOnly)
	}
}

func TestMineWithoutNamespace(t *testing.T) {
	authority := authorityFor(&bo.Account{UID: "user-1", Status: "ACTIVE", SystemRole: perm.SystemRoleSuperAdmin}, true)
	ctx := actor.WithUser(context.Background(), "user-1")
	names, _, readOnly, err := authority.Mine(ctx)
	if err != nil {
		t.Fatalf("mine: %v", err)
	}
	if readOnly || names == nil || len(names) != 0 {
		t.Fatalf("names without namespace = %v, readOnly = %v", names, readOnly)
	}
}
