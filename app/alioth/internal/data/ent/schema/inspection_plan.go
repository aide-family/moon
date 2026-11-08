package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// InspectionPlan 是一次周期巡检的配置：检查项、报告模板、调度与通知组。
// 执行记录与报告是独立快照，模板或计划后续修改不回写历史。
type InspectionPlan struct {
	ent.Schema
}

func (InspectionPlan) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (InspectionPlan) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("name").NotEmpty(),
		field.String("description").Default(""),
		field.Enum("status").Values("ACTIVE", "PAUSED").Default("ACTIVE"),
		// cron 是 5 段表达式：分 时 日 月 周。
		field.String("cron").Default(""),
		field.String("timezone").Default("Asia/Shanghai"),
		field.Int("timeout_seconds").Default(300),
		field.String("template_uid").Default(""),
		// output_formats 取值为 HTML / MARKDOWN，至少一个。
		field.Strings("output_formats").Default([]string{}),
		field.String("title_template").Default(""),
		field.Strings("notification_group_uids").Default([]string{}),
		field.Enum("notify_policy").Values("ALWAYS", "ON_ABNORMAL").Default("ON_ABNORMAL"),
		field.String("labels").Default("{}"),
		field.String("remark").Default(""),
		field.String("created_by").Default(""),
		// next_run_at 由保存、启停与执行完成时重算；暂停时保留旧值不调度。
		field.Time("next_run_at").Optional().Nillable(),
		field.Time("last_run_at").Optional().Nillable(),
	}
}

func (InspectionPlan) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "name").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("namespace_uid", "status"),
		index.Fields("status", "next_run_at"),
	}
}
