package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// NotificationTask 是一条待投递或已投递的通知任务。
// (event_uid, phase) 唯一挡住同一事件的重复 FIRING 或恢复任务。试发的 event_uid 为空，不占用这条约束。
type NotificationTask struct {
	ent.Schema
}

func (NotificationTask) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (NotificationTask) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("event_uid").Default(""),
		field.Enum("phase").Values("FIRING", "RECOVERED", "TEST").Default("TEST"),
		field.Enum("status").Values("PENDING", "SENDING", "SENT", "FAILED", "CANCELLED").Default("PENDING"),
		field.String("rule_uid").Default(""),
		field.String("group_uid").Default(""),
		field.String("channel_uid").NotEmpty(),
		field.String("template_uid").Default(""),
		field.String("receivers").Default(""),
		field.String("subject").Default(""),
		field.String("body").Default(""),
		field.String("params").Default(""),
		field.Int("retries_left").Default(3),
		// 失败后的下一次可领取时间。为空表示立即可领取。
		field.Time("next_attempt_at").Optional().Nillable(),
		field.String("last_error").Default(""),
	}
}

func (NotificationTask) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("event_uid", "phase", "group_uid").Unique().Annotations(entsql.IndexWhere("event_uid <> ''")),
		index.Fields("status"),
		index.Fields("status", "next_attempt_at"),
		index.Fields("namespace_uid", "status"),
	}
}
