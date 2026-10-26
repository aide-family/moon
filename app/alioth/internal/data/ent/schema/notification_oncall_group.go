package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// NotificationOnCallGroup 把 OnCall 绑定到一个通知组，并保存该组的接收时段。
type NotificationOnCallGroup struct {
	ent.Schema
}

func (NotificationOnCallGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (NotificationOnCallGroup) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("oncall_uid").NotEmpty(),
		field.String("group_uid").NotEmpty(),
		// time_windows 是该绑定接收时段的 JSON 数组。空数组表示全天候接收。
		field.String("time_windows").Default("[]"),
	}
}

func (NotificationOnCallGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("oncall_uid", "group_uid").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("oncall_uid"),
		index.Fields("group_uid"),
	}
}
