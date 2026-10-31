package data

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/featurepackage"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/namespace"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/namespaceorder"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/predicate"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/pkg/merr"
)

type namespaceOrderRepo struct{ *repoBase }

func (r *namespaceOrderRepo) CreateOrder(ctx context.Context, order bo.NamespaceOrder) (*bo.NamespaceOrder, error) {
	kind := namespaceorder.Kind(order.Kind)
	if order.Kind == "" {
		kind = namespaceorder.KindCREATE
	}
	currency := namespaceorder.Currency(order.Currency)
	if order.Currency == "" {
		currency = namespaceorder.CurrencyCNY
	}
	row, err := r.client.NamespaceOrder.Create().
		SetUID(id.New()).
		SetUserUID(order.UserUID).
		SetKind(kind).
		SetNamespaceUID(order.NamespaceUID).
		SetNamespaceName(order.NamespaceName).
		SetNamespaceRemark(order.NamespaceRemark).
		SetPackageUID(order.PackageUID).
		SetPackageName(order.PackageName).
		SetPriceCents(order.PriceCents).
		SetCurrency(currency).
		SetDurationDays(order.DurationDays).
		SetMaxMembers(order.MaxMembers).
		SetStatus(namespaceorder.StatusPENDING).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return orderView(row), nil
}

func (r *namespaceOrderRepo) GetOrder(ctx context.Context, uid string) (*bo.NamespaceOrder, error) {
	row, err := r.client.NamespaceOrder.Query().Where(namespaceorder.UIDEQ(uid)).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return orderView(row), nil
}

func (r *namespaceOrderRepo) ListOrders(ctx context.Context, userUID string, page, pageSize int, status string) ([]bo.NamespaceOrder, int, error) {
	preds := []predicate.NamespaceOrder{namespaceorder.UserUIDEQ(userUID)}
	if status != "" {
		preds = append(preds, namespaceorder.StatusEQ(namespaceorder.Status(status)))
	}
	return r.listOrders(ctx, preds, page, pageSize)
}

func (r *namespaceOrderRepo) ListNamespaceOrders(ctx context.Context, namespaceUID string, page, pageSize int) ([]bo.NamespaceOrder, int, error) {
	preds := []predicate.NamespaceOrder{namespaceorder.NamespaceUIDEQ(namespaceUID)}
	return r.listOrders(ctx, preds, page, pageSize)
}

func (r *namespaceOrderRepo) listOrders(ctx context.Context, preds []predicate.NamespaceOrder, page, pageSize int) ([]bo.NamespaceOrder, int, error) {
	query := r.client.NamespaceOrder.Query().Where(preds...)
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.
		Order(ent.Desc(namespaceorder.FieldCreatedAt), ent.Desc(namespaceorder.FieldUID)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.NamespaceOrder, 0, len(rows))
	for _, row := range rows {
		items = append(items, *orderView(row))
	}
	return items, total, nil
}

// PayOrder 是虚拟支付：订单必须属于下单人。不同订单类型在同一个事务里执行不同动作：
// CREATE 创建租户，UPGRADE 换套餐，RENEW 延长授权；随后把订单置为已支付。
// 已支付的订单原样返回，重复支付不会重复创建或重复延长。
func (r *namespaceOrderRepo) PayOrder(ctx context.Context, uid, userUID string) (*bo.NamespaceOrder, *bo.Namespace, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	order, err := tx.NamespaceOrder.Query().Where(namespaceorder.UIDEQ(uid)).Only(ctx)
	if err != nil {
		return nil, nil, r.fail(err)
	}
	if order.UserUID != userUID {
		// 不暴露别人的订单是否存在。
		return nil, nil, merr.NotFound("order not found")
	}
	switch order.Status {
	case namespaceorder.StatusPAID:
		ns, err := tx.Namespace.Query().Where(namespace.UIDEQ(order.NamespaceUID)).Only(ctx)
		if err != nil {
			return nil, nil, r.fail(err)
		}
		item := toNamespace(ns)
		return orderView(order), &item, nil
	case namespaceorder.StatusCANCELLED:
		return nil, nil, merr.Client("order is cancelled")
	}
	// 支付时重新确认套餐仍然存在且启用：下单后套餐可能已被停用或删除。
	pkg, err := tx.FeaturePackage.Query().
		Where(featurepackage.UIDEQ(order.PackageUID), featurepackage.DeletedAtIsNil(), featurepackage.StatusEQ(featurepackage.StatusACTIVE)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil, merr.Client("package is no longer available")
		}
		return nil, nil, r.fail(err)
	}
	paidAt := time.Now()
	namespaceUID := order.NamespaceUID
	switch order.Kind {
	case namespaceorder.KindUPGRADE, namespaceorder.KindRENEW:
		if err := r.payPackageChange(ctx, tx, order, pkg, paidAt); err != nil {
			return nil, nil, err
		}
	default:
		// 到期时间用订单快照的时长，而不是套餐当前时长：用户按当时的条件付款。
		namespaceUID, err = r.createNamespaceTx(ctx, tx, order.NamespaceName, order.NamespaceRemark, userUID, pkg, order.DurationDays)
		if err != nil {
			return nil, nil, err
		}
	}
	// 条件更新保证同一订单只有一个支付能成功；并发的第二次支付会回滚整个事务。
	updated, err := tx.NamespaceOrder.Update().
		Where(namespaceorder.UIDEQ(uid), namespaceorder.StatusEQ(namespaceorder.StatusPENDING)).
		SetStatus(namespaceorder.StatusPAID).
		SetNamespaceUID(namespaceUID).
		SetPaidAt(paidAt).
		Save(ctx)
	if err != nil {
		return nil, nil, r.fail(err)
	}
	if updated == 0 {
		return nil, nil, merr.Conflict("order is already paid")
	}
	ns, err := tx.Namespace.Query().Where(namespace.UIDEQ(namespaceUID)).Only(ctx)
	if err != nil {
		return nil, nil, r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, r.fail(err)
	}
	order.Status = namespaceorder.StatusPAID
	order.NamespaceUID = namespaceUID
	order.PaidAt = &paidAt
	item := toNamespace(ns)
	return orderView(order), &item, nil
}

// payPackageChange 处理套餐升级 / 续期：更新租户套餐和到期时间。
// 租户额外能力保留，升级不会清掉管理员加开的能力；角色权限不再随套餐自动补齐。
func (r *namespaceOrderRepo) payPackageChange(ctx context.Context, tx *ent.Tx, order *ent.NamespaceOrder, pkg *ent.FeaturePackage, now time.Time) error {
	ns, err := tx.Namespace.Query().Where(
		namespace.UIDEQ(order.NamespaceUID),
		namespace.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return merr.Client("namespace is no longer available")
		}
		return r.fail(err)
	}
	if ns.Status != namespace.StatusACTIVE {
		return merr.Client("namespace is disabled")
	}
	expiresAt, err := packageChangeExpiry(order, ns.PackageExpiresAt, now)
	if err != nil {
		return err
	}
	packageFeatures, err := packageSnapshotCapabilities(ctx, tx.Client(), pkg.UID)
	if err != nil {
		return r.fail(err)
	}
	update := tx.Namespace.UpdateOne(ns).SetPackageUID(order.PackageUID)
	// 升级采用新套餐的能力快照；续期套餐不变，必须保留租户原来的快照，
	// 否则管理员后来改过套餐明细，续期会顺带改变租户权益。
	if order.Kind == namespaceorder.KindUPGRADE {
		update = update.SetPackageFeatures(packageFeatures)
	}
	if expiresAt.IsZero() {
		update = update.ClearPackageExpiresAt()
	} else {
		update = update.SetPackageExpiresAt(expiresAt)
	}
	if _, err := update.Save(ctx); err != nil {
		return r.fail(err)
	}
	return nil
}

// packageChangeExpiry 计算升级 / 续期后的到期时间：
// 升级从支付时刻按新套餐时长重新计算；续期在原到期时间（未过期时）或现在的基础上累加。
// 0 天表示永久，只有升级可以买到永久套餐，续期永久套餐没有意义。
func packageChangeExpiry(order *ent.NamespaceOrder, current *time.Time, now time.Time) (time.Time, error) {
	switch order.Kind {
	case namespaceorder.KindUPGRADE:
		if order.DurationDays <= 0 {
			return time.Time{}, nil
		}
		return now.AddDate(0, 0, int(order.DurationDays)), nil
	case namespaceorder.KindRENEW:
		if order.DurationDays <= 0 {
			return time.Time{}, merr.Client("package is permanent")
		}
		base := now
		if current != nil && current.After(now) {
			base = *current
		}
		return base.AddDate(0, 0, int(order.DurationDays)), nil
	default:
		return time.Time{}, merr.Invalid("invalid request")
	}
}

// CancelOrder 只能取消自己的待支付订单。取消后订单保留，方便租户订单查询里解释结果。
func (r *namespaceOrderRepo) CancelOrder(ctx context.Context, uid, userUID string) (*bo.NamespaceOrder, error) {
	n, err := r.client.NamespaceOrder.Update().
		Where(
			namespaceorder.UIDEQ(uid),
			namespaceorder.UserUIDEQ(userUID),
			namespaceorder.StatusEQ(namespaceorder.StatusPENDING),
		).
		SetStatus(namespaceorder.StatusCANCELLED).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		// 区分「订单不存在 / 不是自己的」和「已经付过款不能取消」，前端给出不同提示。
		order, err := r.client.NamespaceOrder.Query().Where(namespaceorder.UIDEQ(uid)).Only(ctx)
		if err != nil {
			return nil, r.fail(err)
		}
		if order.UserUID != userUID {
			return nil, merr.NotFound("order not found")
		}
		if order.Status == namespaceorder.StatusPAID {
			return nil, merr.Client("paid order cannot be cancelled")
		}
		return nil, merr.Client("order is not pending")
	}
	return r.GetOrder(ctx, uid)
}

func orderView(row *ent.NamespaceOrder) *bo.NamespaceOrder {
	item := &bo.NamespaceOrder{
		UID:             row.UID,
		UserUID:         row.UserUID,
		Kind:            string(row.Kind),
		NamespaceUID:    row.NamespaceUID,
		NamespaceName:   row.NamespaceName,
		NamespaceRemark: row.NamespaceRemark,
		PackageUID:      row.PackageUID,
		PackageName:     row.PackageName,
		PriceCents:      row.PriceCents,
		Currency:        string(row.Currency),
		DurationDays:    row.DurationDays,
		MaxMembers:      row.MaxMembers,
		Status:          string(row.Status),
		CreatedAt:       row.CreatedAt,
	}
	if row.PaidAt != nil {
		item.PaidAt = *row.PaidAt
	}
	return item
}
