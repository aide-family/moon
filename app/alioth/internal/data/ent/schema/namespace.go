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
		field.String("logo").Default(""),
		field.String("remark").Default(""),
		field.Enum("status").Values("ACTIVE", "DISABLED").Default("ACTIVE"),
	}
}

func (Namespace) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("name").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}
