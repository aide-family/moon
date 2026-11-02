package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AlertEscalationStep 是升级策略的一步：事件触发后第 delay_minutes 分钟，
// 向 notification_group_uids 里的通知组生成升级通知任务。
// 步骤随策略整体保存，被移除的步骤直接删除，不保留历史。
type AlertEscalationStep struct {
	ent.Schema
}

func (AlertEscalationStep) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (AlertEscalationStep) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("policy_uid").NotEmpty(),
		// step_order 从 0 开始，同一策略内唯一。
		field.Int("step_order"),
		// delay_minutes 是距事件 starts_at 的累计分钟数，必须递增，最后一步不得为 0。
		field.Int("delay_minutes"),
		// notification_group_uids 是本步投递的通知组，至少一个。
		field.Strings("notification_group_uids").Default([]string{}),
	}
}

func (AlertEscalationStep) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("policy_uid", "step_order").Unique(),
		index.Fields("namespace_uid", "policy_uid"),
	}
}
