package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Namespace struct {
	ent.Schema
}

func (Namespace) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (Namespace) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("name").NotEmpty(),
		field.String("secret").Sensitive().Default(""),
		field.String("owner_uid").NotEmpty(),
		// package_uid 指向租户当前使用的基础/自定义功能套餐；为空只出现在升级前的存量租户，
		// seed 会把它们迁到基础套餐。
		field.String("package_uid").Default(""),
		field.String("logo").Default(""),
		field.String("remark").Default(""),
		field.Enum("status").Values("ACTIVE", "DISABLED").Default("ACTIVE"),
		// package_expires_at 是套餐到期时间；NULL 表示永久。到期后能力按停用处理，但不清空配置。
		field.Time("package_expires_at").Optional().Nillable(),
		// disable_reason 是平台停用租户的原因，启用时清空。
		field.String("disable_reason").MaxLen(500).Default(""),
	}
}

func (Namespace) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("name").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}
