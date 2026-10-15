package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type EmailLoginCode struct {
	ent.Schema
}

func (EmailLoginCode) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (EmailLoginCode) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("email").NotEmpty(),
		field.String("code_hash").Sensitive().NotEmpty(),
		field.Time("expires_at"),
		field.Time("used_at").Optional().Nillable(),
	}
}

func (EmailLoginCode) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("email"),
	}
}
