package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// NamespaceOrder 是一次「先下单、后虚拟支付」的租户创建记录。
// 下单时快照套餐名称、价格、授权时长和人数上限，后续改套餐不影响已下的订单；
// 支付成功时在同一个事务里创建租户并把订单置为 PAID。
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
		// namespace_uid 在支付成功后才写入；待支付和已取消时为空。
		field.String("namespace_uid").Default(""),
		field.String("namespace_name").NotEmpty(),
		field.String("namespace_remark").MaxRuneLen(200).Default(""),
		field.String("package_uid").NotEmpty(),
		// package_name 是下单时的套餐名称快照。
		field.String("package_name").NotEmpty(),
		// price_cents 是下单时的价格快照，单位为人民币分；0 表示免费。
		field.Int64("price_cents").Default(0).NonNegative(),
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
