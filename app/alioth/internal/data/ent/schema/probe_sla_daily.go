package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ProbeSlaDaily 按「目标组 × 目标 × 自然日」预聚合探测结果。原始结果只保留 7 天，
// 日聚合让 SLA 报表与状态页可用率能查询 30 天；聚合任务重跑时按唯一键 upsert。
type ProbeSlaDaily struct {
	ent.Schema
}

func (ProbeSlaDaily) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (ProbeSlaDaily) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("group_uid").NotEmpty(),
		field.String("target_uid").NotEmpty(),
		// day 是 UTC 自然日零点的日期（便于按范围比较）。
		field.Time("day"),
		field.Int("checks").Default(0),
		field.Int("ok_checks").Default(0),
		field.Int("failed_checks").Default(0),
		field.Int64("total_duration_ms").Default(0),
		field.Int("max_duration_ms").Default(0),
	}
}

func (ProbeSlaDaily) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("target_uid", "day").Unique(),
		index.Fields("namespace_uid", "group_uid", "day"),
	}
}
