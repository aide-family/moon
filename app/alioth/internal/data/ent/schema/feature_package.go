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
		field.String("remark").MaxLen(5000).Default(""),
		// 基础套餐是新租户的默认套餐，只能有一个：改名和删除都被禁止，功能与备注可改。
		field.Bool("is_base").Default(false),
		field.Enum("status").Values("ACTIVE", "DISABLED").Default("ACTIVE"),
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
