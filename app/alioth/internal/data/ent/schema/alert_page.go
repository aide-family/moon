package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AlertPage 是实时告警的标签过滤视图。公开页面可被同租户其他人选用。
type AlertPage struct {
	ent.Schema
}

func (AlertPage) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (AlertPage) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("name").NotEmpty(),
		field.String("owner_uid").NotEmpty(),
		field.String("labels").Default("{}"),
		field.Bool("public").Default(false),
		field.String("remark").Default(""),
	}
}

func (AlertPage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "name").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("namespace_uid", "owner_uid"),
		index.Fields("namespace_uid", "public"),
	}
}
