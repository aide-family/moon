package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// NotificationGroup 是租户内的接收方集合。
type NotificationGroup struct {
	ent.Schema
}

func (NotificationGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (NotificationGroup) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("name").NotEmpty(),
		field.Enum("status").Values("ACTIVE", "DISABLED").Default("ACTIVE"),
		field.String("channel_uid").NotEmpty(),
		field.String("template_uid").NotEmpty(),
		field.String("receivers").Default(""),
		field.String("remark").Default(""),
	}
}

func (NotificationGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "name").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("namespace_uid", "status"),
	}
}
