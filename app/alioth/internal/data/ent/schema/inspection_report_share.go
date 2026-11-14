package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// InspectionReportShare 是一份报告的一条公开分享链接。
// token 只保存 SHA-256，不保存可反推的原文；停用只写 revoked_at，保留记录。
type InspectionReportShare struct {
	ent.Schema
}

func (InspectionReportShare) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (InspectionReportShare) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("report_uid").NotEmpty(),
		field.String("token_hash").NotEmpty(),
		field.String("password_hash").Default(""),
		field.Time("expires_at").Optional().Nillable(),
		field.String("created_by").Default(""),
		field.Time("revoked_at").Optional().Nillable(),
		field.Int("view_count").Default(0),
		field.Time("last_viewed_at").Optional().Nillable(),
	}
}

func (InspectionReportShare) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("token_hash").Unique(),
		index.Fields("namespace_uid", "report_uid", "created_at"),
	}
}
