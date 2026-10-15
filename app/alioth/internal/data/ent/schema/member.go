package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Member struct {
	ent.Schema
}

func (Member) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (Member) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("user_uid").NotEmpty(),
		field.Enum("status").Values("JOINED", "INVITED", "EXPIRED").Default("JOINED"),
	}
}

func (Member) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "user_uid").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}
