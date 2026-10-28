package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AlertEventFiring 记住当前仍在触发的 fingerprint。
// 这张表不分区，用来挡住同一个租户里的第二条当前 FIRING。
type AlertEventFiring struct {
	ent.Schema
}

func (AlertEventFiring) Fields() []ent.Field {
	return []ent.Field{
		field.String("namespace_uid").NotEmpty(),
		field.String("fingerprint").NotEmpty(),
		field.String("event_uid").NotEmpty(),
		// 数据源可用性事件没有关联规则，rule_uid 为空。
		field.String("rule_uid").Default(""),
	}
}

func (AlertEventFiring) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("namespace_uid", "fingerprint").Unique(),
		index.Fields("event_uid").Unique(),
		index.Fields("rule_uid"),
	}
}
