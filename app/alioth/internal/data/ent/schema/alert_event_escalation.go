package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AlertEventEscalation 记录「某事件在某策略的某一步」的升级发送状态。
// 调度器据它补齐 PENDING、按时生成通知任务，并防止同一步重复触发。
type AlertEventEscalation struct {
	ent.Schema
}

func (AlertEventEscalation) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (AlertEventEscalation) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("event_uid").NotEmpty(),
		field.String("policy_uid").NotEmpty(),
		field.Int("step_order"),
		// due_at 是事件 starts_at + 该步 delay_minutes。
		field.Time("due_at"),
		field.Enum("status").Values("PENDING", "SENT", "SKIPPED", "CANCELLED").Default("PENDING"),
		// sent_at 是实际生成通知任务的时间；未发送为空。
		field.Time("sent_at").Optional().Nillable(),
	}
}

func (AlertEventEscalation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("event_uid", "policy_uid", "step_order").Unique(),
		index.Fields("namespace_uid", "event_uid"),
		// 调度器按状态 + 到期时间扫描待发送与待取消的行。
		index.Fields("status", "due_at"),
	}
}
