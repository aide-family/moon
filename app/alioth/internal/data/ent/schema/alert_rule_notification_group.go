package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AlertRuleNotificationGroup 把规则绑到通知组。
type AlertRuleNotificationGroup struct {
	ent.Schema
}

func (AlertRuleNotificationGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (AlertRuleNotificationGroup) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("rule_uid").NotEmpty(),
		field.String("group_uid").NotEmpty(),
	}
}

func (AlertRuleNotificationGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("rule_uid", "group_uid").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("rule_uid"),
	}
}
