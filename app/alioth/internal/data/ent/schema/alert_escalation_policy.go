package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AlertEscalationPolicy 是一条告警升级策略：命中的 FIRING 事件在指定分钟数后
// 按步骤向通知组生成升级通知任务；步骤在 AlertEscalationStep 里按顺序保存。
type AlertEscalationPolicy struct {
	ent.Schema
}

func (AlertEscalationPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (AlertEscalationPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("name").NotEmpty(),
		field.String("description").Default(""),
		// matcher 是 JSON：labels（键到完整匹配正则）、level_uids、datasource_uids、rule_uids。
		// 四个条件之间是 AND，空条件不限制；语义与维护窗口一致。
		field.String("matcher").Default("{}"),
		field.Enum("status").Values("ACTIVE", "DISABLED").Default("ACTIVE"),
		field.String("created_by").Default(""),
	}
}

func (AlertEscalationPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "name").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("namespace_uid", "status"),
	}
}
