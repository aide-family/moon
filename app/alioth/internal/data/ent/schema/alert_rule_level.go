package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type AlertRuleLevel struct {
	ent.Schema
}

func (AlertRuleLevel) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (AlertRuleLevel) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("rule_uid").NotEmpty(),
		field.String("condition").NotEmpty(),
		field.Enum("sample").Values("FOR", "MAX", "MIN"),
		field.String("threshold").Default(""),
		field.Int("for_seconds").Default(0),
		field.String("level_uid").NotEmpty(),
	}
}

func (AlertRuleLevel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("rule_uid"),
		index.Fields("namespace_uid"),
	}
}
