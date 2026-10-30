package bo

import "time"

// 租户订单状态。待支付订单可以支付或取消；支付后创建租户并绑定套餐；
// 已取消订单保留记录但不占用任何资源。
const (
	NamespaceOrderPending   = "PENDING"
	NamespaceOrderPaid      = "PAID"
	NamespaceOrderCancelled = "CANCELLED"
)

// NamespaceOrder 是一笔租户创建订单。价格、授权时长和人数上限都是下单时的
// 套餐快照，后续修改套餐不影响已经创建的订单。
// NamespaceUID 只在支付成功后写入；PaidAt 零值表示尚未支付。
type NamespaceOrder struct {
	UID             string
	UserUID         string
	NamespaceUID    string
	NamespaceName   string
	NamespaceRemark string
	PackageUID      string
	PackageName     string
	PriceCents      int64
	DurationDays    int32
	MaxMembers      int32
	Status          string
	CreatedAt       time.Time
	PaidAt          time.Time
}
