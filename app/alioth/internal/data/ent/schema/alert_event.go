package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AlertEvent 是一条告警事件。表在迁移后按 created_at 做周分区。
// 分区表的主键必须包含分区键，因此当前 FIRING 和评测批次的唯一约束放在另外两张窄表上。
type AlertEvent struct {
	ent.Schema
}

func (AlertEvent) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (AlertEvent) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("rule_uid").NotEmpty(),
		field.String("datasource_uid").Default(""),
		field.String("fingerprint").NotEmpty(),
		field.Enum("status").Values("FIRING", "RECOVERED", "RECOVERED_BY_MANUAL"),
		field.String("expr").Default(""),
		field.String("labels").Default(""),
		field.String("summary").Default(""),
		field.String("description").Default(""),
		field.String("sample_value").Default(""),
		field.String("level_uid").Default(""),
		field.Time("starts_at"),
		field.Time("ends_at").Optional().Nillable(),
		field.Time("eval_slot"),
		field.Bool("suppressed").Default(false),
		field.String("intervened_by").Default(""),
		field.Time("intervened_at").Optional().Nillable(),
	}
}

func (AlertEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid", "created_at").Unique(),
		index.Fields("namespace_uid", "status"),
		index.Fields("rule_uid"),
		index.Fields("fingerprint"),
	}
}
