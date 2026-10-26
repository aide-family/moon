package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// NotificationTask 是一条待投递或已投递的通知任务。
// (event_uid, phase, group_uid, channel_uid) 唯一挡住同一事件、同一通知组、同一渠道的重复任务。
// 试发的 event_uid 为空，不占用这条约束。
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
		// fingerprint 是事件指纹，用于同一告警的重复通知间隔判断。
		field.String("fingerprint").Default(""),
		field.String("subscription_uid").Default(""),
		field.String("group_uid").Default(""),
		field.String("channel_uid").NotEmpty(),
		field.String("template_uid").Default(""),
		// 投递时的发送方快照，避免渠道配置后续变更影响任务审计。
		field.String("sender").Default(""),
		field.String("receivers").Default(""),
		field.String("subject").Default(""),
		field.String("body").Default(""),
		// 最终渲染后的发送正文；body 保留原始模板以支持安全重试。
		field.String("content").Default(""),
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
		index.Fields("event_uid", "phase", "group_uid", "channel_uid").Unique().Annotations(entsql.IndexWhere("event_uid <> ''")),
		index.Fields("status"),
		index.Fields("status", "next_attempt_at"),
		index.Fields("namespace_uid", "status"),
		// 同一指纹的重复通知间隔判断和恢复前的已发送检查都按这三个字段查询。
		index.Fields("namespace_uid", "fingerprint", "phase"),
	}
}
