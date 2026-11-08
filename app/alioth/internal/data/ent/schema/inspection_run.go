package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// InspectionRun 是一次执行记录。计划名称与参数在入队时快照，计划后续改名或改配置
// 不影响历史；状态机为 PENDING → RUNNING → SUCCEEDED / PARTIAL / FAILED / CANCELED。
type InspectionRun struct {
	ent.Schema
}

func (InspectionRun) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (InspectionRun) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("plan_uid").NotEmpty(),
		field.String("plan_name").Default(""),
		field.Enum("trigger").Values("SCHEDULED", "MANUAL").Default("SCHEDULED"),
		field.Enum("status").Values("PENDING", "RUNNING", "SUCCEEDED", "PARTIAL", "FAILED", "CANCELED").Default("PENDING"),
		field.Time("scheduled_at"),
		field.Time("started_at").Optional().Nillable(),
		field.Time("finished_at").Optional().Nillable(),
		field.Time("window_start").Optional().Nillable(),
		field.Time("window_end").Optional().Nillable(),
		field.Int("item_total").Default(0),
		field.Int("item_ok").Default(0),
		field.Int("item_warn").Default(0),
		field.Int("item_failed").Default(0),
		// summary 是报告摘要 JSON，列表与投递直接使用。
		field.String("summary").Default("{}"),
		field.String("error").Default(""),
		field.String("operator_uid").Default(""),
		field.String("operator_name").Default(""),
		// lease_owner / lease_until 保证同一 run 只被一个执行器持有。
		field.String("lease_owner").Default(""),
		field.Time("lease_until").Optional().Nillable(),
		field.Time("notified_at").Optional().Nillable(),
	}
}

func (InspectionRun) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "plan_uid", "scheduled_at").Unique().
			Annotations(entsql.IndexWhere("status <> 'CANCELED'")),
		index.Fields("namespace_uid", "status", "scheduled_at"),
		index.Fields("status", "lease_until"),
		index.Fields("plan_uid", "scheduled_at"),
	}
}
