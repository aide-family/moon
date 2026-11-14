package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type AlertRuleGroup struct {
	ent.Schema
}

func (AlertRuleGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (AlertRuleGroup) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("name").NotEmpty(),
		// 上级规则组标识；空字符串表示顶级规则组。
		field.String("parent_uid").Default(""),
		field.Enum("status").Values("ACTIVE", "DISABLED").Default("ACTIVE"),
	}
}

func (AlertRuleGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "name").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("namespace_uid"),
		index.Fields("namespace_uid", "parent_uid"),
	}
}
