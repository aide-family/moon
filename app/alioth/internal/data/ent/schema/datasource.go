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
		field.Enum("driver").Values("METRICS_PROMETHEUS", "METRICS_VICTORIA_METRICS", "LOGS_ELASTICSEARCH", "LOGS_LOKI", "LOGS_VLOGS", "TRACE_JAEGER"),
		field.String("endpoint").NotEmpty(),
		// Elasticsearch 的索引模式和时间字段。其他驱动留空，不使用这两个字段。
		field.String("index_pattern").Default(""),
		field.String("time_field").Default(""),
		field.String("username").Default(""),
		field.String("password_cipher").Sensitive().Default(""),
		field.String("secret_cipher").Sensitive().Default(""),
		field.Enum("status").Values("ACTIVE", "DISABLED").Default("ACTIVE"),
		field.String("level_uid").Default(""),
		field.String("labels").Default("{}"),
		field.String("remark").Default(""),
		// 可用性告警：统计窗口内向数据源发起的规则评测请求失败达到阈值时，
		// 生成一条数据源级别的告警事件，并直接通知绑定的通知组。
		field.Bool("availability_alert_enabled").Default(false),
		field.Int("failure_window_minutes").Default(5),
		field.Int("failure_threshold").Default(3),
		field.Strings("notification_group_uids").Default([]string{}),
	}
}

func (Datasource) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "name").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		// 同一租户、同一类型、同一地址和索引模式只允许一个启用中的数据源。
		// 停用的历史数据源保留，便于用户查看或删除，但不能与启用项同时评测。
		index.Fields("namespace_uid", "type", "endpoint", "index_pattern").Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL AND status = 'ACTIVE'")),
		index.Fields("namespace_uid"),
	}
}
