package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AlertHistoryEvent 是历史告警的分片索引。
// 事件恢复后写入租户库的周表，这里只保存事件标识和分片位置，供详情按 UID 定位。
type AlertHistoryEvent struct {
	ent.Schema
}

func (AlertHistoryEvent) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (AlertHistoryEvent) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("event_uid").NotEmpty().Unique(),
		// database_name 是统一前缀加租户标识，table_name 是统一前缀加周一日期。
		field.String("database_name").NotEmpty(),
		field.String("table_name").NotEmpty(),
		field.Time("week_start"),
	}
}

func (AlertHistoryEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "week_start"),
	}
}
