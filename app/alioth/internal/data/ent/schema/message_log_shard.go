package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// MessageLogShard 是消息日志的分片索引。
// 日志归档到租户库的周表后，主库只保留日志标识和分片位置，供详情按 UID 定位。
type MessageLogShard struct {
	ent.Schema
}

func (MessageLogShard) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (MessageLogShard) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("message_uid").NotEmpty().Unique(),
		// database_name 是统一前缀加租户标识，table_name 是统一前缀加周一日期。
		field.String("database_name").NotEmpty(),
		field.String("table_name").NotEmpty(),
		field.Time("week_start"),
	}
}

func (MessageLogShard) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "week_start"),
	}
}
