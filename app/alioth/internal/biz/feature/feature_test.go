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
	pkg               *bo.FeaturePackage
	state             *bo.NamespaceFeature
	set               bool
	args              setNamespaceFeatureArgs
	created           int32
	createdMaxMembers int32
	createdPriceCents int64
	updated           int32
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

func (s *featureRepoStub) CreatePackage(_ context.Context, pkg bo.FeaturePackage) (*bo.FeaturePackage, error) {
	s.created = pkg.DurationDays
	s.createdMaxMembers = pkg.MaxMembers
	s.createdPriceCents = pkg.PriceCents
	return &bo.FeaturePackage{
		Name: pkg.Name, Remark: pkg.Remark, DurationDays: pkg.DurationDays,
		Visibility: pkg.Visibility, MaxMembers: pkg.MaxMembers,
		PriceCents: pkg.PriceCents, Features: pkg.Features,
	}, nil
}

func (s *featureRepoStub) UpdatePackage(_ context.Context, pkg bo.FeaturePackage) (*bo.FeaturePackage, error) {
	s.updated = pkg.DurationDays
	return &bo.FeaturePackage{
		UID: pkg.UID, Name: pkg.Name, Remark: pkg.Remark,
		DurationDays: pkg.DurationDays, Visibility: pkg.Visibility,
		MaxMembers: pkg.MaxMembers, PriceCents: pkg.PriceCents, Features: pkg.Features,
	}, nil
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

// TestUpdateNamespaceFeaturesAcceptsLegacyChannelCodes 确认旧客户端或旧快照提交的
// 拆分前渠道编码会先展开成当前能力再和套餐能力比较：旧的 notify_channel_webhook /
// notify_channel_sms 覆盖套餐里的四个 Webhook 应用能力和两个短信提供商能力，
// 不会误报「缺少套餐能力」。
func TestUpdateNamespaceFeaturesAcceptsLegacyChannelCodes(t *testing.T) {
	packageFeatures := []string{
		perm.CapabilityNotifyChannelSMTP,
		perm.CapabilityNotifyChannelWebhookOther,
		perm.CapabilityNotifyChannelWebhookDingTalk,
		perm.CapabilityNotifyChannelWebhookWeChat,
		perm.CapabilityNotifyChannelWebhookFeishu,
		perm.CapabilityNotifyChannelSMSAliCloud,
		perm.CapabilityNotifyChannelSMSTencent,
	}
	repo := &featureRepoStub{pkg: &bo.FeaturePackage{
		UID: "package-1", Features: packageFeatures, DurationDays: 0,
	}}
	legacyPayload := []string{
		perm.CapabilityNotifyChannelSMTP,
		"notify_channel_webhook",
		"notify_channel_sms",
	}
	if _, err := provisioningFor(repo).UpdateNamespaceFeatures(systemContext(), "namespace-1", "package-1", legacyPayload, 0); err != nil {
		t.Fatalf("legacy channel codes rejected: %v", err)
	}
	if !repo.set {
		t.Fatal("repository was not called")
	}
	want := perm.NormalizeCapabilities(legacyPayload)
	if !slices.Equal(repo.args.capabilities, want) {
		t.Fatalf("capabilities = %v, want normalized %v", repo.args.capabilities, want)
	}
}

// TestUpdateNamespaceFeaturesOnlyRequiresCurrentSnapshot 确认管理员后来给套餐加能力时，
// 已生效租户保存授权只需包含自己的套餐快照（M22 权益快照语义），
// 不能被套餐新增的能力反向卡住。
func TestUpdateNamespaceFeaturesOnlyRequiresCurrentSnapshot(t *testing.T) {
	repo := &featureRepoStub{
		pkg: &bo.FeaturePackage{
			UID:          "package-1",
			Features:     []string{perm.CapabilityAlertRealtime, perm.CapabilityMember},
			DurationDays: 0,
		},
		state: &bo.NamespaceFeature{
			NamespaceUID:    "namespace-1",
			PackageUID:      "package-1",
			PackageFeatures: []string{perm.CapabilityAlertRealtime},
		},
	}
	_, err := provisioningFor(repo).UpdateNamespaceFeatures(
		systemContext(), "namespace-1", "package-1",
		[]string{perm.CapabilityAlertRealtime, perm.CapabilityMember}, 0,
	)
	if err != nil {
		t.Fatalf("tenant snapshot subset rejected: %v", err)
	}
	if !repo.set {
		t.Fatal("repository was not called")
	}
}

func TestPackageSaleAttributesPassedThrough(t *testing.T) {
	repo := &featureRepoStub{}
	provisioning := provisioningFor(repo)
	pkg := bo.FeaturePackage{
		Name: "专业版", Remark: "remark", DurationDays: 90,
		Visibility: perm.PackageVisibilityPublic, MaxMembers: 20, PriceCents: 9900,
		Features: []string{perm.CapabilityMember},
	}
	if _, err := provisioning.CreatePackage(systemContext(), pkg); err != nil {
		t.Fatalf("create package: %v", err)
	}
	if repo.created != 90 {
		t.Fatalf("created durationDays = %d, want 90", repo.created)
	}
	if repo.createdMaxMembers != 20 || repo.createdPriceCents != 9900 {
		t.Fatalf("created maxMembers/price = %d/%d, want 20/9900",
			repo.createdMaxMembers, repo.createdPriceCents)
	}
	repo.pkg = &bo.FeaturePackage{UID: "package-1", Name: "name", Features: []string{perm.CapabilityMember}}
	pkg.UID = "package-1"
	pkg.Name = "name"
	pkg.DurationDays = 7
	if _, err := provisioning.UpdatePackage(systemContext(), pkg); err != nil {
		t.Fatalf("update package: %v", err)
	}
	if repo.updated != 7 {
		t.Fatalf("updated durationDays = %d, want 7", repo.updated)
	}
}

func TestCreatePackageRejectsInvalidAttributes(t *testing.T) {
	provisioning := provisioningFor(&featureRepoStub{})
	valid := bo.FeaturePackage{
		Name: "name", Visibility: perm.PackageVisibilityPublic,
		Features: []string{perm.CapabilityMember},
	}
	cases := []struct {
		name string
		pkg  bo.FeaturePackage
	}{
		{name: "negative duration", pkg: func() bo.FeaturePackage {
			item := valid
			item.DurationDays = -1
			return item
		}()},
		{name: "missing visibility", pkg: func() bo.FeaturePackage {
			item := valid
			item.Visibility = ""
			return item
		}()},
		{name: "negative max members", pkg: func() bo.FeaturePackage {
			item := valid
			item.MaxMembers = -1
			return item
		}()},
		{name: "negative price", pkg: func() bo.FeaturePackage {
			item := valid
			item.PriceCents = -1
			return item
		}()},
	}
	for _, item := range cases {
		if _, err := provisioning.CreatePackage(systemContext(), item.pkg); err == nil {
			t.Fatalf("%s accepted", item.name)
		}
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
