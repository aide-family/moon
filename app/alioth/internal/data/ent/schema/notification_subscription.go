package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// NotificationSubscription 按标签、告警级别和数据源订阅告警事件。
type NotificationSubscription struct {
	ent.Schema
}

func (NotificationSubscription) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (NotificationSubscription) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("name").NotEmpty(),
		field.Enum("status").Values("ACTIVE", "DISABLED").Default("ACTIVE"),
		// labels 是事件标签的包含匹配条件。空对象表示不限制。
		field.String("labels").Default("{}"),
		field.Strings("level_uids").Default([]string{}),
		field.Strings("datasource_uids").Default([]string{}),
		// datasource_level_uids 按数据源级别过滤；为空表示不限制。
		field.Strings("datasource_level_uids").Default([]string{}),
		// repeat_interval_seconds 是同一告警重复通知的最小间隔，0 表示不限制。
		field.Int("repeat_interval_seconds").Default(0),
		field.String("remark").Default(""),
	}
}

func (NotificationSubscription) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "name").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("namespace_uid", "status"),
	}
}
