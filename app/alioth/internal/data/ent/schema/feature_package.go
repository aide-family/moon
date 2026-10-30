package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// FeaturePackage 是平台级的功能套餐。功能明细在 FeaturePackageItem，
// 不在这张表里存数组，避免套餐功能无法用唯一约束去重。
type FeaturePackage struct {
	ent.Schema
}

func (FeaturePackage) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (FeaturePackage) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		// 套餐名称在未删除记录上唯一。
		field.String("name").NotEmpty(),
		field.String("remark").MaxRuneLen(5000).Default(""),
		// 基础套餐是新租户的默认套餐，只能有一个：改名和删除都被禁止，功能与备注可改。
		field.Bool("is_base").Default(false),
		field.Enum("status").Values("ACTIVE", "DISABLED").Default("ACTIVE"),
		// duration_days 是租户使用该套餐时的默认授权时长；0 表示永久。
		// 套餐只给默认值，真正到期时间落在 namespaces.package_expires_at。
		field.Int32("duration_days").Default(0).NonNegative(),
		// visibility 决定套餐出现在哪里：PUBLIC 可被用户在创建租户时选择，
		// BETA 只由平台管理员在租户功能授权里分配。默认内测，避免误上架。
		field.Enum("visibility").Values("PUBLIC", "BETA").Default("BETA"),
		// max_members 是使用该套餐的租户可容纳的 JOINED 成员数上限；0 表示不限制。
		// 容量跟着租户当前套餐实时计算，不在 namespaces 上冗余。
		field.Int32("max_members").Default(0).NonNegative(),
		// price_cents 是套餐价格，单位为人民币分；0 表示免费。
		// 下单时快照到订单，之后改价不影响已下的订单。
		field.Int64("price_cents").Default(0).NonNegative(),
	}
}

func (FeaturePackage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("name").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		// 部分唯一索引保证未删除的基础套餐最多一个，与超级管理员唯一性同一手段。
		index.Fields("is_base").Unique().Annotations(entsql.IndexWhere("is_base = true AND deleted_at IS NULL")),
	}
}
