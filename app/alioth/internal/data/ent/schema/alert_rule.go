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
		field.Enum("driver").Values("METRICS_PROMETHEUS", "METRICS_VICTORIA_METRICS", "LOGS_ELASTICSEARCH", "LOGS_LOKI", "LOGS_VLOGS", "TRACE_JAEGER", "PROBE_HTTP"),
		field.Enum("status").Values("ACTIVE", "DISABLED").Default("ACTIVE"),
		field.String("expr").Default(""),
		field.String("labels").Default(""),
		field.String("datasource_uid").Default(""),
		field.String("datasource_labels").Default("{}"),
		field.String("summary").Default(""),
		field.String("description").Default(""),
		// variables 是自定义查询变量的 JSON 数组文本，默认空数组。
		field.String("variables").Default("[]"),
		// probe_group_labels 是探测规则的目标组标签匹配器 JSON，查询规则留空对象。
		field.String("probe_group_labels").Default("{}"),
		// probe_interval_seconds 是探测周期，查询规则为 0。
		field.Int("probe_interval_seconds").Default(0),
		// probe_metric 是探测规则比较的观测指标，查询规则不使用。
		field.Enum("probe_metric").Values("UP", "LATENCY_MS", "STATUS_CODE", "CERT_DAYS").Default("UP"),
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
