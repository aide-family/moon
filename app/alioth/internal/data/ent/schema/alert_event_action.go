package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AlertEventAction 是告警事件的处置时间线。只追加、不修改，租户不能删除。
// 事件归档到历史分库后时间线仍留在主库，按 event_uid 查询。
type AlertEventAction struct {
	ent.Schema
}

func (AlertEventAction) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (AlertEventAction) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("event_uid").NotEmpty(),
		// action 取值为 FIRING、NOTIFIED、ACKED、UNACKED、INTERVENED、SUPPRESSED、
		// MAINTENANCE、RECOVERED 或 RECOVERED_BY_MANUAL；用字符串保留扩展空间。
		field.String("action").NotEmpty(),
		field.String("actor_uid").Default(""),
		field.String("actor_name").Default(""),
		field.String("detail").Default(""),
	}
}

func (AlertEventAction) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("event_uid", "created_at"),
		index.Fields("namespace_uid", "event_uid"),
	}
}
