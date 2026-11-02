package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AlertMaintenanceWindow 是一个一次性维护窗口：窗口生效期间匹配的告警被静默，
// 不生成通知任务；窗口结束后抑制自动到期。matcher 是 JSON 对象。
type AlertMaintenanceWindow struct {
	ent.Schema
}

func (AlertMaintenanceWindow) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (AlertMaintenanceWindow) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("name").NotEmpty(),
		field.String("description").Default(""),
		// matcher 是 JSON：labels（键到完整匹配正则）、level_uids、datasource_uids、rule_uids。
		field.String("matcher").Default("{}"),
		field.Time("starts_at"),
		field.Time("ends_at"),
		field.Enum("status").Values("ACTIVE", "DISABLED").Default("ACTIVE"),
		field.String("created_by").Default(""),
	}
}

func (AlertMaintenanceWindow) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "name").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("namespace_uid", "status", "starts_at"),
	}
}
