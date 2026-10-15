package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type User struct {
	ent.Schema
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("name").Default(""),
		field.String("nickname").Default(""),
		field.String("email").NotEmpty(),
		field.String("avatar").Default(""),
		field.String("remark").Default(""),
		field.Enum("status").Values("ACTIVE", "BANNED").Default("ACTIVE"),
		field.String("password_hash").Sensitive().Default(""),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("email").Unique(),
	}
}
