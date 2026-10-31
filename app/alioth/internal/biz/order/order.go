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
		Currency:        pkg.Currency,
		DurationDays:    pkg.DurationDays,
		MaxMembers:      pkg.MaxMembers,
	})
	return item, purchasing.access.Wrap(err)
}

// CreateUpgradeOrder 为当前租户下升级套餐的订单。升级只能选公开且启用的套餐，
// 当前套餐本身不算升级；支付后从支付时刻按新套餐时长重新计算到期时间。
func (purchasing *Purchasing) CreateUpgradeOrder(ctx context.Context, packageUID string) (*bo.NamespaceOrder, error) {
	userUID, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	namespaceUID, err := purchasing.access.Allow(ctx, perm.NamespaceWrite)
	if err != nil {
		return nil, err
	}
	current, err := purchasing.features.NamespaceFeatureState(ctx, namespaceUID)
	if err != nil {
		return nil, purchasing.access.Wrap(err)
	}
	pkg, err := purchasing.features.GetPackage(ctx, packageUID)
	if err != nil {
		return nil, purchasing.access.Wrap(err)
	}
	if pkg.Status != "ACTIVE" || pkg.Visibility != perm.PackageVisibilityPublic {
		return nil, merr.Client("package is not available")
	}
	if current.PackageUID == pkg.UID {
		return nil, merr.Client("package is already in use")
	}
	namespace, err := purchasing.access.CurrentNamespace(ctx)
	if err != nil {
		return nil, err
	}
	if namespace == nil {
		return nil, merr.Client("namespace is required")
	}
	item, err := purchasing.orders.CreateOrder(ctx, bo.NamespaceOrder{
		UserUID:         userUID,
		Kind:            bo.NamespaceOrderUpgrade,
		NamespaceUID:    namespaceUID,
		NamespaceName:   namespace.Name,
		NamespaceRemark: namespace.Remark,
		PackageUID:      pkg.UID,
		PackageName:     pkg.Name,
		PriceCents:      pkg.PriceCents,
		Currency:        pkg.Currency,
		DurationDays:    pkg.DurationDays,
		MaxMembers:      pkg.MaxMembers,
	})
	return item, purchasing.access.Wrap(err)
}

// CreateRenewalOrder 为当前租户下续期订单：套餐不变，支付后按当前套餐时长延长授权。
// 永久套餐没有续期概念，直接拒绝。
func (purchasing *Purchasing) CreateRenewalOrder(ctx context.Context) (*bo.NamespaceOrder, error) {
	userUID, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	namespaceUID, err := purchasing.access.Allow(ctx, perm.NamespaceWrite)
	if err != nil {
		return nil, err
	}
	current, err := purchasing.features.NamespaceFeatureState(ctx, namespaceUID)
	if err != nil {
		return nil, purchasing.access.Wrap(err)
	}
	if current.PackageUID == "" {
		return nil, merr.Client("package is not available")
	}
	pkg, err := purchasing.features.GetPackage(ctx, current.PackageUID)
	if err != nil {
		return nil, purchasing.access.Wrap(err)
	}
	if pkg.Status != "ACTIVE" {
		return nil, merr.Client("package is not available")
	}
	if pkg.DurationDays <= 0 {
		return nil, merr.Client("package is permanent")
	}
	namespace, err := purchasing.access.CurrentNamespace(ctx)
	if err != nil {
		return nil, err
	}
	if namespace == nil {
		return nil, merr.Client("namespace is required")
	}
	item, err := purchasing.orders.CreateOrder(ctx, bo.NamespaceOrder{
		UserUID:         userUID,
		Kind:            bo.NamespaceOrderRenew,
		NamespaceUID:    namespaceUID,
		NamespaceName:   namespace.Name,
		NamespaceRemark: namespace.Remark,
		PackageUID:      pkg.UID,
		PackageName:     pkg.Name,
		PriceCents:      pkg.PriceCents,
		Currency:        pkg.Currency,
		DurationDays:    pkg.DurationDays,
		MaxMembers:      pkg.MaxMembers,
	})
	return item, purchasing.access.Wrap(err)
}

// CurrentPackage 返回当前租户的套餐、价格、人数上限和到期时间，
// 供租户页展示升级 / 续期入口。租户停用后仍可读取，便于解释为什么不能升级。
func (purchasing *Purchasing) CurrentPackage(ctx context.Context, namespaceUID string) (*bo.NamespacePackage, error) {
	ns, err := purchasing.access.AllowAnyStatus(ctx, perm.NamespaceRead)
	if err != nil {
		return nil, err
	}
	if ns != namespaceUID {
		return nil, merr.Client("namespace mismatch")
	}
	state, err := purchasing.features.NamespaceFeatureState(ctx, namespaceUID)
	if err != nil {
		return nil, purchasing.access.Wrap(err)
	}
	item := &bo.NamespacePackage{
		NamespaceUID: namespaceUID,
		PackageUID:   state.PackageUID,
		Name:         state.PackageName,
		DurationDays: state.DurationDays,
		ExpiresAt:    state.ExpiresAt,
		Expired:      state.Expired,
	}
	if state.PackageUID != "" {
		pkg, err := purchasing.features.GetPackage(ctx, state.PackageUID)
		if err != nil && !merr.IsNotFound(err) {
			return nil, purchasing.access.Wrap(err)
		}
		if err == nil {
			item.Name = pkg.Name
			item.Status = pkg.Status
			item.Visibility = pkg.Visibility
			item.MaxMembers = pkg.MaxMembers
			item.PriceCents = pkg.PriceCents
			item.Currency = pkg.Currency
			item.DurationDays = pkg.DurationDays
		}
	}
	return item, nil
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
