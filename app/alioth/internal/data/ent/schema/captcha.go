package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Captcha struct {
	ent.Schema
}

func (Captcha) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (Captcha) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("answer_hash").Sensitive().NotEmpty(),
		field.Time("expires_at"),
		field.Time("used_at").Optional().Nillable(),
	}
}

func (Captcha) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
	}
}
