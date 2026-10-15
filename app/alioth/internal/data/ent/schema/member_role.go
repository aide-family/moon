package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type MemberRole struct {
	ent.Schema
}

func (MemberRole) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (MemberRole) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("member_uid").NotEmpty(),
		field.String("role_uid").NotEmpty(),
	}
}

func (MemberRole) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("member_uid", "role_uid").Unique(),
		index.Fields("namespace_uid"),
	}
}
