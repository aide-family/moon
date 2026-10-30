package repository

import (
	"context"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// NamespaceOrderRepo 读写租户创建订单。
// 订单只记录下单时的套餐快照；支付成功时在一个事务里创建租户、绑定套餐并置为已支付。
type NamespaceOrderRepo interface {
	// CreateOrder 落一笔待支付订单。快照字段由调用方从套餐读好后填入。
	CreateOrder(ctx context.Context, order bo.NamespaceOrder) (*bo.NamespaceOrder, error)
	GetOrder(ctx context.Context, uid string) (*bo.NamespaceOrder, error)
	// ListOrders 按创建时间倒序列出下单人自己的订单，status 为空表示全部。
	ListOrders(ctx context.Context, userUID string, page, pageSize int, status string) ([]bo.NamespaceOrder, int, error)
	// ListNamespaceOrders 按创建时间倒序列出某个租户的订单，供平台查询。
	ListNamespaceOrders(ctx context.Context, namespaceUID string, page, pageSize int) ([]bo.NamespaceOrder, int, error)
	// PayOrder 是虚拟支付：校验订单属于 userUID 且待支付后，创建租户、绑定套餐并置为 PAID。
	// 已支付的订单原样返回已创建的租户，重复支付不会创建第二个租户。
	PayOrder(ctx context.Context, uid, userUID string) (*bo.NamespaceOrder, *bo.Namespace, error)
	// CancelOrder 把属于 userUID 的待支付订单置为 CANCELLED；已支付订单不能取消。
	CancelOrder(ctx context.Context, uid, userUID string) (*bo.NamespaceOrder, error)
}
