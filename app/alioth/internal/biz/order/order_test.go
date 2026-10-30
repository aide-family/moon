package order

import (
	"context"
	"log/slog"
	"testing"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
)

type featureRepoStub struct {
	repository.FeatureRepo
	pkg    *bo.FeaturePackage
	public []bo.FeaturePackage
}

func (s *featureRepoStub) GetPackage(context.Context, string) (*bo.FeaturePackage, error) {
	return s.pkg, nil
}

func (s *featureRepoStub) ListPublicPackages(context.Context) ([]bo.FeaturePackage, error) {
	return s.public, nil
}

type orderRepoStub struct {
	repository.NamespaceOrderRepo
	created    *bo.NamespaceOrder
	payUID     string
	payUser    string
	cancelUID  string
	listStatus string
	listUser   string
}

func (s *orderRepoStub) CreateOrder(_ context.Context, order bo.NamespaceOrder) (*bo.NamespaceOrder, error) {
	order.UID = "order-1"
	s.created = &order
	return &order, nil
}

func (s *orderRepoStub) PayOrder(_ context.Context, uid, userUID string) (*bo.NamespaceOrder, *bo.Namespace, error) {
	s.payUID = uid
	s.payUser = userUID
	return &bo.NamespaceOrder{UID: uid, Status: bo.NamespaceOrderPaid}, &bo.Namespace{UID: "namespace-1"}, nil
}

func (s *orderRepoStub) CancelOrder(_ context.Context, uid, userUID string) (*bo.NamespaceOrder, error) {
	s.cancelUID = uid
	return &bo.NamespaceOrder{UID: uid, Status: bo.NamespaceOrderCancelled}, nil
}

func (s *orderRepoStub) ListOrders(_ context.Context, userUID string, _, _ int, status string) ([]bo.NamespaceOrder, int, error) {
	s.listUser = userUID
	s.listStatus = status
	return nil, 0, nil
}

func purchasingFor(features repository.FeatureRepo, orders repository.NamespaceOrderRepo) *Purchasing {
	return &Purchasing{
		features: features,
		orders:   orders,
		access: access.Checker{
			Log: slog.New(slog.DiscardHandler),
		},
	}
}

func purchaseContext() context.Context {
	return actor.WithUser(context.Background(), "user-1")
}

func TestCreateOrderSnapshotsPackage(t *testing.T) {
	orders := &orderRepoStub{}
	purchasing := purchasingFor(&featureRepoStub{pkg: &bo.FeaturePackage{
		UID: "package-1", Name: "专业版", Status: "ACTIVE", Visibility: perm.PackageVisibilityPublic,
		DurationDays: 30, MaxMembers: 20, PriceCents: 9900,
	}}, orders)
	if _, err := purchasing.CreateOrder(purchaseContext(), " 租户A ", " 备注 ", "package-1"); err != nil {
		t.Fatalf("create order: %v", err)
	}
	if orders.created == nil {
		t.Fatal("order repository was not called")
	}
	created := orders.created
	if created.UserUID != "user-1" || created.NamespaceName != "租户A" || created.NamespaceRemark != "备注" {
		t.Fatalf("order identity fields = %+v", created)
	}
	if created.PackageName != "专业版" || created.PriceCents != 9900 ||
		created.DurationDays != 30 || created.MaxMembers != 20 {
		t.Fatalf("order package snapshot = %+v", created)
	}
}

func TestCreateOrderRejectsBetaPackage(t *testing.T) {
	orders := &orderRepoStub{}
	purchasing := purchasingFor(&featureRepoStub{pkg: &bo.FeaturePackage{
		UID: "package-1", Status: "ACTIVE", Visibility: perm.PackageVisibilityBeta,
	}}, orders)
	_, err := purchasing.CreateOrder(purchaseContext(), "tenant", "", "package-1")
	if err == nil || kerrors.FromError(err).Message != "package is not available" {
		t.Fatalf("err = %v, want package is not available", err)
	}
	if orders.created != nil {
		t.Fatal("beta package order was persisted")
	}
}

func TestPayOrderRequiresLogin(t *testing.T) {
	purchasing := purchasingFor(&featureRepoStub{}, &orderRepoStub{})
	if _, _, err := purchasing.PayOrder(context.Background(), "order-1"); err == nil {
		t.Fatal("anonymous pay accepted")
	}
}

func TestPayAndCancelOrderPassOrderAndUser(t *testing.T) {
	orders := &orderRepoStub{}
	purchasing := purchasingFor(&featureRepoStub{}, orders)
	if _, _, err := purchasing.PayOrder(purchaseContext(), "order-1"); err != nil {
		t.Fatalf("pay order: %v", err)
	}
	if orders.payUID != "order-1" || orders.payUser != "user-1" {
		t.Fatalf("pay args = %q/%q", orders.payUID, orders.payUser)
	}
	if _, err := purchasing.CancelOrder(purchaseContext(), "order-2"); err != nil {
		t.Fatalf("cancel order: %v", err)
	}
	if orders.cancelUID != "order-2" {
		t.Fatalf("cancel uid = %q", orders.cancelUID)
	}
}

func TestListOrdersRejectsInvalidStatus(t *testing.T) {
	purchasing := purchasingFor(&featureRepoStub{}, &orderRepoStub{})
	if _, _, _, _, err := purchasing.ListOrders(purchaseContext(), 0, 0, "UNKNOWN"); err == nil {
		t.Fatal("invalid status accepted")
	}
	if _, _, _, _, err := purchasing.ListOrders(purchaseContext(), 0, 0, ""); err != nil {
		t.Fatalf("empty status rejected: %v", err)
	}
}

func TestListPublicPackagesRequiresLogin(t *testing.T) {
	purchasing := purchasingFor(&featureRepoStub{public: []bo.FeaturePackage{{UID: "package-1"}}}, &orderRepoStub{})
	if _, err := purchasing.ListPublicPackages(context.Background()); err == nil {
		t.Fatal("anonymous public package list accepted")
	}
	items, err := purchasing.ListPublicPackages(purchaseContext())
	if err != nil {
		t.Fatalf("list public packages: %v", err)
	}
	if len(items) != 1 || items[0].UID != "package-1" {
		t.Fatalf("public packages = %+v", items)
	}
}
