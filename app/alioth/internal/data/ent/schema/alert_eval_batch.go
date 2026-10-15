package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AlertEvalBatch 记住某条规则的一个评测批次已经回写。
// 这张表不分区，(rule_uid, eval_slot) 唯一。
type AlertEvalBatch struct {
	ent.Schema
}

func (AlertEvalBatch) Fields() []ent.Field {
	return []ent.Field{
		field.String("rule_uid").NotEmpty(),
		field.Time("eval_slot").Default(time.Now),
	}
}

func (AlertEvalBatch) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("rule_uid", "eval_slot").Unique(),
	}
}
