package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Datasource struct {
	ent.Schema
}

func (Datasource) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (Datasource) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("name").NotEmpty(),
		field.Enum("type").Values("METRICS", "LOGS", "TRACE"),
		field.Enum("driver").Values("METRICS_PROMETHEUS", "METRICS_VICTORIA_METRICS", "LOGS_ELASTICSEARCH", "LOGS_LOKI", "TRACE_JAEGER"),
		field.String("endpoint").NotEmpty(),
		field.String("username").Default(""),
		field.String("password_cipher").Sensitive().Default(""),
		field.String("secret_cipher").Sensitive().Default(""),
		field.Enum("status").Values("ACTIVE", "DISABLED").Default("ACTIVE"),
		field.String("level_uid").Default(""),
		field.String("labels").Default("{}"),
		field.String("remark").Default(""),
	}
}

func (Datasource) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "name").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("namespace_uid"),
	}
}
