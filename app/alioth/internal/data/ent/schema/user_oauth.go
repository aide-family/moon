package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type UserOAuth struct {
	ent.Schema
}

func (UserOAuth) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (UserOAuth) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("user_uid").NotEmpty(),
		field.Enum("provider").Values("GITHUB", "GITEE", "FEISHU"),
		field.String("open_id").NotEmpty(),
		field.String("email").Default(""),
	}
}

func (UserOAuth) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("provider", "open_id").Unique(),
	}
}
