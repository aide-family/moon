package namespace

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/internal/mailer"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

type userRepoStub struct {
	repository.UserRepo
	account *bo.Account
}

func (s userRepoStub) UserByUID(context.Context, string) (*bo.Account, error) {
	return s.account, nil
}

type namespaceRepoStub struct {
	repository.NamespaceRepo
	current   *bo.Namespace
	updated   *bo.Namespace
	setStatus string
	setReason string
}

func (s *namespaceRepoStub) GetNamespace(context.Context, string) (*bo.Namespace, error) {
	if s.current == nil {
		return nil, merr.NotFound("not found")
	}
	return s.current, nil
}

func (s *namespaceRepoStub) SetNamespaceStatus(_ context.Context, uid, status, reason string) (*bo.Namespace, error) {
	s.setStatus = status
	s.setReason = reason
	if s.updated == nil {
		return &bo.Namespace{UID: uid, Status: status, DisableReason: reason}, nil
	}
	return s.updated, nil
}

type featureRepoStub struct {
	repository.FeatureRepo
	archived string
}

func (s *featureRepoStub) ArchiveNamespace(_ context.Context, uid string) error {
	s.archived = uid
	return nil
}

func tenancyFor(systemRole string, namespaces *namespaceRepoStub, features *featureRepoStub) *Tenancy {
	return &Tenancy{
		namespaces: namespaces,
		features:   features,
		access: access.Checker{
			Users: userRepoStub{account: &bo.Account{UID: "user-1", Status: "ACTIVE", SystemRole: systemRole}},
			Log:   slog.New(slog.DiscardHandler),
		},
		mail: &mailer.Sender{},
		log:  slog.New(slog.DiscardHandler),
	}
}

func platformContext() context.Context {
	return actor.WithUser(context.Background(), "user-1")
}

// TestTenantCannotChangeOwnStatus 确认租户侧启停接口固定拒绝，租户不能自行解封。
func TestTenantCannotChangeOwnStatus(t *testing.T) {
	tenancy := tenancyFor(perm.SystemRoleSuperAdmin, &namespaceRepoStub{}, &featureRepoStub{})
	_, err := tenancy.UpdateNamespaceStatus(platformContext(), "namespace-1", "ACTIVE")
	if reason := kerrors.Reason(err); reason != merr.ReasonForbidden {
		t.Fatalf("tenant status change reason = %q, want forbidden (err = %v)", reason, err)
	}
	if message := kerrors.FromError(err).Message; message != "namespace status is managed by platform" {
		t.Fatalf("message = %q", message)
	}
}

// TestSetSystemNamespaceStatusDisabled 确认停用写入原因并清空启用原因。
func TestSetSystemNamespaceStatusDisabled(t *testing.T) {
	namespaces := &namespaceRepoStub{
		current: &bo.Namespace{UID: "namespace-1", Name: "租户一", Status: "ACTIVE", OwnerEmail: "owner@example.com"},
	}
	tenancy := tenancyFor(perm.SystemRoleSuperAdmin, namespaces, &featureRepoStub{})
	item, err := tenancy.SetSystemNamespaceStatus(platformContext(), "namespace-1", "DISABLED", "  欠费  ")
	if err != nil {
		t.Fatalf("disable namespace: %v", err)
	}
	if namespaces.setStatus != "DISABLED" || namespaces.setReason != "欠费" {
		t.Fatalf("status = %q, reason = %q", namespaces.setStatus, namespaces.setReason)
	}
	if item.Status != "DISABLED" {
		t.Fatalf("item status = %q", item.Status)
	}

	namespaces.current = &bo.Namespace{UID: "namespace-1", Name: "租户一", Status: "DISABLED"}
	if _, err := tenancy.SetSystemNamespaceStatus(platformContext(), "namespace-1", "ACTIVE", "旧原因"); err != nil {
		t.Fatalf("enable namespace: %v", err)
	}
	if namespaces.setStatus != "ACTIVE" || namespaces.setReason != "" {
		t.Fatalf("enable status = %q, reason = %q, want reason cleared", namespaces.setStatus, namespaces.setReason)
	}
}

// TestSetSystemNamespaceStatusRequiresReason 确认停用必须填写原因，启用不要求。
func TestSetSystemNamespaceStatusRequiresReason(t *testing.T) {
	namespaces := &namespaceRepoStub{current: &bo.Namespace{UID: "namespace-1", Status: "ACTIVE"}}
	tenancy := tenancyFor(perm.SystemRoleSuperAdmin, namespaces, &featureRepoStub{})
	_, err := tenancy.SetSystemNamespaceStatus(platformContext(), "namespace-1", "DISABLED", "   ")
	if err == nil || kerrors.FromError(err).Message != "reason is required" {
		t.Fatalf("blank reason err = %v, want reason is required", err)
	}
	long := strings.Repeat("原", 501)
	if _, err := tenancy.SetSystemNamespaceStatus(platformContext(), "namespace-1", "DISABLED", long); err == nil {
		t.Fatal("over-long reason accepted")
	}
}

// TestSetSystemNamespaceStatusRequiresSystemWrite 确认只有超级管理员能停用/启用租户。
func TestSetSystemNamespaceStatusRequiresSystemWrite(t *testing.T) {
	namespaces := &namespaceRepoStub{current: &bo.Namespace{UID: "namespace-1", Status: "ACTIVE"}}
	for _, systemRole := range []string{perm.SystemRoleSystemAdmin, perm.SystemRoleReadOnly, perm.SystemRoleUser} {
		tenancy := tenancyFor(systemRole, namespaces, &featureRepoStub{})
		_, err := tenancy.SetSystemNamespaceStatus(platformContext(), "namespace-1", "DISABLED", "原因")
		if reason := kerrors.Reason(err); reason != merr.ReasonForbidden {
			t.Fatalf("%s disable namespace reason = %q, want forbidden (err = %v)", systemRole, reason, err)
		}
	}
}

// TestArchiveNamespace 确认归档调用仓储软删除，且只有超级管理员可以执行。
func TestArchiveNamespace(t *testing.T) {
	features := &featureRepoStub{}
	tenancy := tenancyFor(perm.SystemRoleSuperAdmin, &namespaceRepoStub{}, features)
	if err := tenancy.ArchiveNamespace(platformContext(), "namespace-1"); err != nil {
		t.Fatalf("archive namespace: %v", err)
	}
	if features.archived != "namespace-1" {
		t.Fatalf("archived = %q, want namespace-1", features.archived)
	}
	denied := tenancyFor(perm.SystemRoleSystemAdmin, &namespaceRepoStub{}, &featureRepoStub{})
	if err := denied.ArchiveNamespace(platformContext(), "namespace-1"); kerrors.Reason(err) != merr.ReasonForbidden {
		t.Fatalf("system admin archive namespace: %v", err)
	}
}
