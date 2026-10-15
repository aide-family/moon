package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type RolePermission struct {
	ent.Schema
}

func (RolePermission) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (RolePermission) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("role_uid").NotEmpty(),
		field.String("permission_uid").NotEmpty(),
	}
}

func (RolePermission) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("role_uid", "permission_uid").Unique(),
		index.Fields("namespace_uid"),
	}
}
