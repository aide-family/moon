package feature

import (
	"context"
	"log/slog"
	"slices"
	"testing"
	"time"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
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

type featureRepoStub struct {
	repository.FeatureRepo
	pkg     *bo.FeaturePackage
	state   *bo.NamespaceFeature
	set     bool
	args    setNamespaceFeatureArgs
	created int32
	updated int32
}

type setNamespaceFeatureArgs struct {
	namespaceUID string
	packageUID   string
	capabilities []string
	expiresAt    time.Time
}

func (s *featureRepoStub) GetPackage(context.Context, string) (*bo.FeaturePackage, error) {
	if s.pkg == nil {
		return nil, merr.NotFound("not found")
	}
	return s.pkg, nil
}

func (s *featureRepoStub) SetNamespaceFeatureState(_ context.Context, namespaceUID, packageUID string, capabilities []string, expiresAt time.Time) error {
	s.set = true
	s.args = setNamespaceFeatureArgs{
		namespaceUID: namespaceUID, packageUID: packageUID,
		capabilities: capabilities, expiresAt: expiresAt,
	}
	return nil
}

func (s *featureRepoStub) NamespaceFeatureState(context.Context, string) (*bo.NamespaceFeature, error) {
	if s.state == nil {
		return &bo.NamespaceFeature{}, nil
	}
	return s.state, nil
}

func (s *featureRepoStub) CreatePackage(_ context.Context, name, remark string, durationDays int32, features []string) (*bo.FeaturePackage, error) {
	s.created = durationDays
	return &bo.FeaturePackage{Name: name, Remark: remark, DurationDays: durationDays, Features: features}, nil
}

func (s *featureRepoStub) UpdatePackage(_ context.Context, uid, name, remark string, durationDays int32, features []string) (*bo.FeaturePackage, error) {
	s.updated = durationDays
	return &bo.FeaturePackage{UID: uid, Name: name, Remark: remark, DurationDays: durationDays, Features: features}, nil
}

func provisioningFor(features repository.FeatureRepo) *Provisioning {
	return &Provisioning{
		features: features,
		access: access.Checker{
			Users: userRepoStub{account: &bo.Account{UID: "user-1", Status: "ACTIVE", SystemRole: perm.SystemRoleSuperAdmin}},
			Log:   slog.New(slog.DiscardHandler),
		},
	}
}

func systemContext() context.Context {
	return actor.WithUser(context.Background(), "user-1")
}

func TestUpdateNamespaceFeaturesUsesPackageDuration(t *testing.T) {
	repo := &featureRepoStub{pkg: &bo.FeaturePackage{
		UID: "package-1", Features: []string{perm.CapabilityAlertRealtime}, DurationDays: 30,
	}}
	before := time.Now()
	capabilities := []string{perm.CapabilityAlertRealtime, perm.CapabilityMember}
	if _, err := provisioningFor(repo).UpdateNamespaceFeatures(systemContext(), "namespace-1", "package-1", capabilities, 0); err != nil {
		t.Fatalf("update namespace features: %v", err)
	}
	if !repo.set {
		t.Fatal("repository was not called")
	}
	if !slices.Equal(repo.args.capabilities, capabilities) {
		t.Fatalf("capabilities = %v, want %v", repo.args.capabilities, capabilities)
	}
	earliest := before.AddDate(0, 0, 30)
	latest := time.Now().AddDate(0, 0, 30)
	if repo.args.expiresAt.Before(earliest) || repo.args.expiresAt.After(latest) {
		t.Fatalf("expiresAt = %v, want between %v and %v", repo.args.expiresAt, earliest, latest)
	}
}

func TestUpdateNamespaceFeaturesExplicitExpiry(t *testing.T) {
	repo := &featureRepoStub{pkg: &bo.FeaturePackage{
		UID: "package-1", Features: []string{perm.CapabilityMember}, DurationDays: 30,
	}}
	expiresAt := time.Now().Add(48 * time.Hour).Unix()
	if _, err := provisioningFor(repo).UpdateNamespaceFeatures(systemContext(), "namespace-1", "package-1", []string{perm.CapabilityMember}, expiresAt); err != nil {
		t.Fatalf("update namespace features: %v", err)
	}
	if want := time.Unix(expiresAt, 0); !repo.args.expiresAt.Equal(want) {
		t.Fatalf("expiresAt = %v, want explicit %v", repo.args.expiresAt, want)
	}
}

func TestUpdateNamespaceFeaturesPermanent(t *testing.T) {
	repo := &featureRepoStub{pkg: &bo.FeaturePackage{
		UID: "package-1", Features: []string{perm.CapabilityMember}, DurationDays: 0,
	}}
	if _, err := provisioningFor(repo).UpdateNamespaceFeatures(systemContext(), "namespace-1", "package-1", []string{perm.CapabilityMember}, 0); err != nil {
		t.Fatalf("update namespace features: %v", err)
	}
	if !repo.args.expiresAt.IsZero() {
		t.Fatalf("expiresAt = %v, want zero (permanent)", repo.args.expiresAt)
	}
}

func TestUpdateNamespaceFeaturesRequiresPackageCapabilities(t *testing.T) {
	repo := &featureRepoStub{pkg: &bo.FeaturePackage{
		UID: "package-1", Features: []string{perm.CapabilityAlertRealtime}, DurationDays: 0,
	}}
	_, err := provisioningFor(repo).UpdateNamespaceFeatures(systemContext(), "namespace-1", "package-1", []string{perm.CapabilityMember}, 0)
	if err == nil || kerrors.FromError(err).Message != "features must include the package features" {
		t.Fatalf("err = %v, want package capabilities required", err)
	}
	if repo.set {
		t.Fatal("repository called with a capability subset")
	}
}

func TestUpdateNamespaceFeaturesRejectsUnknownCapability(t *testing.T) {
	repo := &featureRepoStub{pkg: &bo.FeaturePackage{UID: "package-1"}}
	_, err := provisioningFor(repo).UpdateNamespaceFeatures(systemContext(), "namespace-1", "package-1", []string{"not_a_capability"}, 0)
	if err == nil || kerrors.FromError(err).Message != "invalid capability" {
		t.Fatalf("err = %v, want invalid capability", err)
	}
}

func TestPackageDurationDaysPassedThrough(t *testing.T) {
	repo := &featureRepoStub{}
	provisioning := provisioningFor(repo)
	if _, err := provisioning.CreatePackage(systemContext(), "name", "remark", 90, []string{perm.CapabilityMember}); err != nil {
		t.Fatalf("create package: %v", err)
	}
	if repo.created != 90 {
		t.Fatalf("created durationDays = %d, want 90", repo.created)
	}
	repo.pkg = &bo.FeaturePackage{UID: "package-1", Name: "name", Features: []string{perm.CapabilityMember}}
	if _, err := provisioning.UpdatePackage(systemContext(), "package-1", "name", "remark", 7, []string{perm.CapabilityMember}); err != nil {
		t.Fatalf("update package: %v", err)
	}
	if repo.updated != 7 {
		t.Fatalf("updated durationDays = %d, want 7", repo.updated)
	}
	if _, err := provisioning.CreatePackage(systemContext(), "name", "remark", -1, nil); err == nil {
		t.Fatal("negative durationDays accepted")
	}
}

// TestListFeaturesReturnsCapabilities 确认 /v1/system/features 返回 30 个能力并带模块编码。
func TestListFeaturesReturnsCapabilities(t *testing.T) {
	items, err := provisioningFor(&featureRepoStub{}).ListFeatures(systemContext())
	if err != nil {
		t.Fatalf("list features: %v", err)
	}
	if len(items) != len(perm.Capabilities()) {
		t.Fatalf("features = %d, want %d", len(items), len(perm.Capabilities()))
	}
	for _, item := range items {
		if item.Code == "" || item.Module == "" || item.Name == "" {
			t.Fatalf("feature item incomplete: %+v", item)
		}
	}
}
