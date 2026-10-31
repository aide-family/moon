package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// NamespaceOrder 是一次「先下单、后虚拟支付」的租户购买记录。
// kind 区分新建租户、套餐升级和套餐续期；价格、时长和人数上限都是下单时的快照。
// 支付成功时在同一个事务里创建租户或更新租户套餐，并把订单置为 PAID。
type NamespaceOrder struct {
	ent.Schema
}

func (NamespaceOrder) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (NamespaceOrder) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		// user_uid 是下单人，只有本人可以支付和取消自己的订单。
		field.String("user_uid").NotEmpty(),
		// kind 决定支付时做什么：CREATE 创建租户，UPGRADE 换套餐，RENEW 延长授权。
		field.Enum("kind").Values("CREATE", "UPGRADE", "RENEW").Default("CREATE"),
		// namespace_uid：CREATE 在支付成功后写入；UPGRADE / RENEW 在下单时写入目标租户。
		field.String("namespace_uid").Default(""),
		field.String("namespace_name").NotEmpty(),
		field.String("namespace_remark").MaxRuneLen(200).Default(""),
		field.String("package_uid").NotEmpty(),
		// package_name 是下单时的套餐名称快照。
		field.String("package_name").NotEmpty(),
		// price_cents 和 currency 是下单时的价格与币种快照，0 表示免费。
		field.Int64("price_cents").Default(0).NonNegative(),
		field.Enum("currency").Values("CNY", "USD", "EUR", "JPY").Default("CNY"),
		// duration_days 是下单时的授权时长快照；0 表示永久。
		field.Int32("duration_days").Default(0).NonNegative(),
		// max_members 是下单时的人数上限快照；0 表示不限制。
		field.Int32("max_members").Default(0).NonNegative(),
		field.Enum("status").Values("PENDING", "PAID", "CANCELLED").Default("PENDING"),
		// paid_at 是虚拟支付完成时间；未支付时为空。
		field.Time("paid_at").Optional().Nillable(),
	}
}

func (NamespaceOrder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		// 订单页按用户和创建时间倒序分页。
		index.Fields("user_uid"),
		// 平台按租户查订单。
		index.Fields("namespace_uid"),
	}
}
