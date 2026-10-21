package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// MessageLog 是一次投递结果。表在迁移后按 created_at 做周分区。
type MessageLog struct {
	ent.Schema
}

func (MessageLog) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (MessageLog) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("task_uid").NotEmpty(),
		field.Enum("status").Values("SENT", "FAILED"),
		field.String("channel_uid").Default(""),
		field.String("template_uid").Default(""),
		field.String("sender").Default(""),
		field.String("receivers").Default(""),
		field.String("subject").Default(""),
		field.String("content").Default(""),
		field.String("detail").Default(""),
	}
}

func (MessageLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid", "created_at").Unique(),
		index.Fields("namespace_uid"),
		index.Fields("task_uid"),
	}
}
