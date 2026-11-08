package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// InspectionItem 是巡检计划里的一个检查项。检查项随计划整体保存，
// 不提供独立 CRUD，避免定时器读到只配置了一半的计划。
type InspectionItem struct {
	ent.Schema
}

func (InspectionItem) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (InspectionItem) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("plan_uid").NotEmpty(),
		field.Enum("kind").Values("METRICS_QUERY", "LOGS_QUERY", "PROBE_TARGET", "ALERT_EVENTS", "NOTE"),
		field.String("title").NotEmpty(),
		field.Int("sort").Default(0),
		// datasource_uid 只对指标与日志查询项有效。
		field.String("datasource_uid").Default(""),
		// query 是 PromQL / LogQL / LogsQL / ES DSL，说明项存 Markdown。
		field.Text("query").Default(""),
		field.Int("window_seconds").Default(3600),
		field.Int("step_seconds").Default(60),
		field.Enum("display").Values("METRIC", "TABLE", "TREND", "LOGS", "TEXT").Default("TABLE"),
		// params 是类型专属配置 JSON：探测目标组、告警筛选、阈值等。
		field.String("params").Default("{}"),
		field.Bool("enabled").Default(true),
	}
}

func (InspectionItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("plan_uid", "sort"),
		index.Fields("namespace_uid", "plan_uid"),
	}
}
