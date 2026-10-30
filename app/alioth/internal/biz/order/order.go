package order

import (
	"context"
	"strings"

	"github.com/google/wire"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/page"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

var ProviderSet = wire.NewSet(NewPurchasing)

// Purchasing 负责租户购买流程：选套餐、下单、虚拟支付和订单查询。
// 支付没有对接渠道，PayOrder 成功即视为付款完成，由仓储在同一个事务里创建租户并绑定套餐。
// 公开套餐列表和下单都只要求登录；支付和取消只能操作自己的订单。
type Purchasing struct {
	orders   repository.NamespaceOrderRepo
	features repository.FeatureRepo
	access   access.Checker
}

func NewPurchasing(orders repository.NamespaceOrderRepo, features repository.FeatureRepo, checker access.Checker) *Purchasing {
	return &Purchasing{orders: orders, features: features, access: checker}
}

// ListPublicPackages 返回创建租户时可选的功能套餐：公开且启用，按价格升序。
// 内测套餐不会出现在这里，由平台在租户功能授权里分配。
func (purchasing *Purchasing) ListPublicPackages(ctx context.Context) ([]bo.FeaturePackage, error) {
	if _, ok := actor.UserID(ctx); !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	items, err := purchasing.features.ListPublicPackages(ctx)
	return items, purchasing.access.Wrap(err)
}

// CreateOrder 校验套餐仍公开可售并快照价格、授权时长和人数上限，落一笔待支付订单。
// 下单不创建租户；租户要等支付成功才创建，避免出现没有付款的空租户。
func (purchasing *Purchasing) CreateOrder(ctx context.Context, name, remark, packageUID string) (*bo.NamespaceOrder, error) {
	userUID, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, merr.Invalid("invalid request")
	}
	pkg, err := purchasing.features.GetPackage(ctx, packageUID)
	if err != nil {
		return nil, purchasing.access.Wrap(err)
	}
	if pkg.Status != "ACTIVE" || pkg.Visibility != perm.PackageVisibilityPublic {
		return nil, merr.Client("package is not available")
	}
	item, err := purchasing.orders.CreateOrder(ctx, bo.NamespaceOrder{
		UserUID:         userUID,
		NamespaceName:   name,
		NamespaceRemark: strings.TrimSpace(remark),
		PackageUID:      pkg.UID,
		PackageName:     pkg.Name,
		PriceCents:      pkg.PriceCents,
		DurationDays:    pkg.DurationDays,
		MaxMembers:      pkg.MaxMembers,
	})
	return item, purchasing.access.Wrap(err)
}

// PayOrder 是虚拟支付。仓储保证「创建租户 + 绑定套餐 + 订单置为已支付」在同一个事务里，
// 重复支付已支付订单会返回当时创建的租户，不会创建第二个租户。
func (purchasing *Purchasing) PayOrder(ctx context.Context, uid string) (*bo.NamespaceOrder, *bo.Namespace, error) {
	userUID, ok := actor.UserID(ctx)
	if !ok {
		return nil, nil, merr.Unauthenticated("missing token")
	}
	item, namespace, err := purchasing.orders.PayOrder(ctx, uid, userUID)
	return item, namespace, purchasing.access.Wrap(err)
}

// CancelOrder 取消自己的待支付订单；已支付订单不能取消。
func (purchasing *Purchasing) CancelOrder(ctx context.Context, uid string) (*bo.NamespaceOrder, error) {
	userUID, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	item, err := purchasing.orders.CancelOrder(ctx, uid, userUID)
	return item, purchasing.access.Wrap(err)
}

// ListOrders 分页列出当前用户自己的订单，status 为空表示全部。
func (purchasing *Purchasing) ListOrders(ctx context.Context, pageNo, pageSize int32, status string) ([]bo.NamespaceOrder, int, int, int, error) {
	userUID, ok := actor.UserID(ctx)
	if !ok {
		return nil, 0, 0, 0, merr.Unauthenticated("missing token")
	}
	if !validOrderStatus(status) {
		return nil, 0, 0, 0, merr.Invalid("invalid status")
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := purchasing.orders.ListOrders(ctx, userUID, p, s, status)
	return items, total, p, s, purchasing.access.Wrap(err)
}

// ListNamespaceOrders 是平台按租户查询订单，需要 system:namespace:read。
// 它不读 X-Namespace，按路径上的租户标识定位。
func (purchasing *Purchasing) ListNamespaceOrders(ctx context.Context, namespaceUID string, pageNo, pageSize int32) ([]bo.NamespaceOrder, int, int, int, error) {
	if _, err := purchasing.access.RequireSystem(ctx, perm.SystemNamespaceRead); err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := purchasing.orders.ListNamespaceOrders(ctx, namespaceUID, p, s)
	return items, total, p, s, purchasing.access.Wrap(err)
}

func validOrderStatus(status string) bool {
	switch status {
	case "", bo.NamespaceOrderPending, bo.NamespaceOrderPaid, bo.NamespaceOrderCancelled:
		return true
	default:
		return false
	}
}
