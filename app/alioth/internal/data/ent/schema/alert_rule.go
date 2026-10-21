package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type AlertRule struct {
	ent.Schema
}

func (AlertRule) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (AlertRule) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("group_uid").NotEmpty(),
		field.String("name").NotEmpty(),
		field.Enum("driver").Values("METRICS_PROMETHEUS", "METRICS_VICTORIA_METRICS", "LOGS_ELASTICSEARCH", "LOGS_LOKI", "LOGS_VLOGS", "TRACE_JAEGER"),
		field.Enum("status").Values("ACTIVE", "DISABLED").Default("ACTIVE"),
		field.String("expr").Default(""),
		field.String("labels").Default(""),
		field.String("datasource_uid").Default(""),
		field.String("datasource_labels").Default("{}"),
		field.String("summary").Default(""),
		field.String("description").Default(""),
		field.Time("next_eval_at").Optional().Nillable(),
		field.Time("scheduled_slot").Optional().Nillable(),
	}
}

func (AlertRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "name").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("namespace_uid", "status"),
		index.Fields("group_uid"),
	}
}
