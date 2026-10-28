package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AlertEvalBatch 记住某条规则在某个数据源上的一个评测批次已经回写。
// 这张表不分区，(rule_uid, datasource_uid, eval_slot) 唯一。
// failed 为真的批次是数据源请求失败，用于数据源可用性判断。
type AlertEvalBatch struct {
	ent.Schema
}

func (AlertEvalBatch) Fields() []ent.Field {
	return []ent.Field{
		field.String("rule_uid").NotEmpty(),
		field.String("datasource_uid").NotEmpty(),
		field.Time("eval_slot").Default(time.Now),
		// failed 为真表示本次向数据源的查询失败，failure_message 保存失败原因。
		field.Bool("failed").Default(false),
		field.String("failure_message").Default(""),
		// created_at 是回写时刻，可用性告警按这个时间统计窗口内的失败次数。
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (AlertEvalBatch) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("rule_uid", "datasource_uid", "eval_slot").Unique(),
		// 可用性判断按数据源统计窗口内的失败批次。
		index.Fields("datasource_uid", "failed", "created_at"),
	}
}
