package member

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

func TestInviteTokenIsOpaqueAndHashed(t *testing.T) {
	first, err := inviteToken()
	if err != nil {
		t.Fatal(err)
	}
	second, err := inviteToken()
	if err != nil {
		t.Fatal(err)
	}
	if len(first) < 40 || first == second {
		t.Fatalf("invite token is not random enough: %q %q", first, second)
	}
	if hashToken(first) == first {
		t.Fatal("invite token is stored in plaintext")
	}
	if optionalTokenHash("") != "" {
		t.Fatal("empty token must not produce a digest")
	}
	if optionalTokenHash(" ") != "" {
		t.Fatal("blank token must not produce a digest")
	}
	if optionalTokenHash(first) != hashToken(first) {
		t.Fatal("token digest is not stable")
	}
}

// ---- 角色分配防越权用例 ----

type memberRepoStub struct {
	repository.MemberRepo
	updated []string
}

func (s *memberRepoStub) UpdateMemberRoles(_ context.Context, _, _ string, roleUIDs []string) (*bo.Member, error) {
	s.updated = roleUIDs
	return &bo.Member{UID: "member-1"}, nil
}

type memberInviteRepoStub struct {
	repository.InviteRepo
	created *bo.MemberInvite
}

func (s *memberInviteRepoStub) UpsertInvite(_ context.Context, invite bo.MemberInvite, _ string) (*bo.MemberInvite, error) {
	invite.UID = "invite-1"
	s.created = &invite
	return &invite, nil
}

type memberRoleRepoStub struct {
	repository.RoleRepo
	roles       map[string]bo.Role
	permissions map[string]struct{}
}

func (s memberRoleRepoStub) GetRole(_ context.Context, _, uid string) (*bo.Role, error) {
	item, ok := s.roles[uid]
	if !ok {
		return nil, merr.NotFound("not found")
	}
	return &item, nil
}

func (s memberRoleRepoStub) Permissions(context.Context, string, string) (map[string]struct{}, error) {
	return s.permissions, nil
}

type memberUserRepoStub struct {
	repository.UserRepo
	account *bo.Account
}

func (s memberUserRepoStub) UserByUID(context.Context, string) (*bo.Account, error) {
	return s.account, nil
}

type memberNamespaceRepoStub struct {
	repository.NamespaceRepo
	ownerUID string
}

func (s memberNamespaceRepoStub) GetNamespace(_ context.Context, uid string) (*bo.Namespace, error) {
	return &bo.Namespace{UID: uid, Name: "租户一", Status: "ACTIVE", OwnerUID: s.ownerUID}, nil
}

type memberFeatureRepoStub struct {
	repository.FeatureRepo
	capabilities []string
	modules      []string
}

func (s memberFeatureRepoStub) EffectiveCapabilities(context.Context, string) ([]string, error) {
	return s.capabilities, nil
}

func (s memberFeatureRepoStub) EffectiveModules(context.Context, string) ([]string, error) {
	return s.modules, nil
}

func (s memberFeatureRepoStub) NamespaceMemberLimit(context.Context, string) (int32, error) {
	return 0, nil
}

// membershipForCaller 构造成员用例：ownerUID 为当前租户负责人（空表示调用者不是负责人），
// held 是调用者当前持有的权限。
func membershipForCaller(ownerUID string, held []string, roles map[string]bo.Role) (*Membership, *memberRepoStub, *memberInviteRepoStub) {
	members := &memberRepoStub{}
	invites := &memberInviteRepoStub{}
	granted := make(map[string]struct{}, len(held))
	for _, name := range held {
		granted[name] = struct{}{}
	}
	roleRepo := memberRoleRepoStub{roles: roles, permissions: granted}
	features := memberFeatureRepoStub{capabilities: allCapabilityCodes(), modules: allModuleCodes()}
	checker := access.Checker{
		Users:      memberUserRepoStub{account: &bo.Account{UID: "user-1", Status: "ACTIVE", SystemRole: perm.SystemRoleUser}},
		Namespaces: memberNamespaceRepoStub{ownerUID: ownerUID},
		Roles:      roleRepo,
		Features:   features,
		Log:        slog.New(slog.DiscardHandler),
	}
	return &Membership{
		members:  members,
		invites:  invites,
		users:    memberUserRepoStub{account: &bo.Account{UID: "user-1", Status: "ACTIVE", SystemRole: perm.SystemRoleUser}},
		roles:    roleRepo,
		features: features,
		access:   checker,
		log:      slog.New(slog.DiscardHandler),
	}, members, invites
}

func allCapabilityCodes() []string {
	items := perm.Capabilities()
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.Code)
	}
	return out
}

func allModuleCodes() []string {
	items := perm.Features()
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.Code)
	}
	return out
}

func memberContext() context.Context {
	ctx := actor.WithUser(context.Background(), "user-1")
	return actor.WithNamespace(ctx, "namespace-1")
}

// callerPermissions 是一个只持有成员菜单权限的非负责人调用者。
func callerPermissions() []string {
	return []string{perm.MemberRead, perm.MemberWrite, perm.MenuPermission(perm.CapabilityMember)}
}

// TestInviteMemberRejectsRoleBeyondOwn 确认非负责人不能邀请持有超出自己权限的角色。
func TestInviteMemberRejectsRoleBeyondOwn(t *testing.T) {
	roles := map[string]bo.Role{
		"role-admin":  {UID: "role-admin", Name: "ops", Status: "ACTIVE", Permissions: []string{perm.DatasourceWrite, perm.MenuPermission("datasource")}},
		"role-subset": {UID: "role-subset", Name: "member", Status: "ACTIVE", Permissions: []string{perm.MemberRead}},
	}
	membership, _, invites := membershipForCaller("", callerPermissions(), roles)
	if _, _, err := membership.InviteMember(memberContext(), "a@example.com", "role-admin"); err == nil ||
		!strings.Contains(kerrors.FromError(err).Message, "cannot grant permissions beyond your own") {
		t.Fatalf("invite beyond own err = %v, want subset rejection", err)
	}
	if invites.created != nil {
		t.Fatal("invite created despite subset rejection")
	}
	if _, _, err := membership.InviteMember(memberContext(), "a@example.com", "role-subset"); err != nil {
		t.Fatalf("invite subset role: %v", err)
	}
	if invites.created == nil {
		t.Fatal("subset role invite not created")
	}
}

// TestUpdateMemberRolesRejectsRoleBeyondOwn 确认非负责人不能分配超出自己权限的角色。
func TestUpdateMemberRolesRejectsRoleBeyondOwn(t *testing.T) {
	roles := map[string]bo.Role{
		"role-admin": {UID: "role-admin", Name: "ops", Status: "ACTIVE", Permissions: []string{perm.DatasourceWrite}},
	}
	membership, members, _ := membershipForCaller("", callerPermissions(), roles)
	if _, err := membership.UpdateMemberRoles(memberContext(), "member-2", []string{"role-admin"}); err == nil ||
		!strings.Contains(kerrors.FromError(err).Message, "cannot grant permissions beyond your own") {
		t.Fatalf("assign beyond own err = %v, want subset rejection", err)
	}
	if len(members.updated) != 0 {
		t.Fatalf("roles written despite rejection: %v", members.updated)
	}
}

// TestOwnerCanAssignAnyRole 确认负责人不受子集限制，即使自己没有角色授权。
func TestOwnerCanAssignAnyRole(t *testing.T) {
	roles := map[string]bo.Role{
		"role-admin": {UID: "role-admin", Name: "ops", Status: "ACTIVE", Permissions: []string{perm.DatasourceWrite, perm.MenuPermission("datasource")}},
	}
	membership, members, invites := membershipForCaller("user-1", nil, roles)
	if _, err := membership.UpdateMemberRoles(memberContext(), "member-2", []string{"role-admin"}); err != nil {
		t.Fatalf("owner assign role: %v", err)
	}
	if len(members.updated) != 1 || members.updated[0] != "role-admin" {
		t.Fatalf("owner assign updated = %v", members.updated)
	}
	if _, _, err := membership.InviteMember(memberContext(), "a@example.com", "role-admin"); err != nil {
		t.Fatalf("owner invite any role: %v", err)
	}
	if invites.created == nil {
		t.Fatal("owner invite not created")
	}
}

// TestInviteMemberRejectsDisabledRole 确认停用角色仍不能用于邀请。
func TestInviteMemberRejectsDisabledRole(t *testing.T) {
	roles := map[string]bo.Role{
		"role-disabled": {UID: "role-disabled", Name: "ops", Status: "DISABLED", Permissions: []string{perm.MemberRead}},
	}
	membership, _, _ := membershipForCaller("", callerPermissions(), roles)
	if _, _, err := membership.InviteMember(memberContext(), "a@example.com", "role-disabled"); err == nil ||
		kerrors.FromError(err).Message != "invalid request" {
		t.Fatalf("invite disabled role err = %v, want invalid request", err)
	}
}
