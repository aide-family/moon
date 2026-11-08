package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// InspectionReport 是一次执行渲染出的报告快照，按格式各存一份。
// 分享只存 token 的 SHA-256，不保存可反推的原文。
type InspectionReport struct {
	ent.Schema
}

func (InspectionReport) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (InspectionReport) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("run_uid").NotEmpty(),
		field.Enum("format").Values("HTML", "MARKDOWN"),
		field.String("title").Default(""),
		field.Text("content").Default(""),
		field.Int("size").Default(0),
		field.String("checksum").Default(""),
		field.String("share_token_hash").Default(""),
		field.Bool("share_enabled").Default(false),
		field.Time("share_expires_at").Optional().Nillable(),
		field.String("share_password_hash").Default(""),
		field.String("share_created_by").Default(""),
		field.Time("share_created_at").Optional().Nillable(),
		field.Int("share_view_count").Default(0),
		field.Time("share_last_viewed_at").Optional().Nillable(),
	}
}

func (InspectionReport) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("run_uid", "format").Unique(),
		index.Fields("namespace_uid", "created_at"),
		index.Fields("share_token_hash").Unique().Annotations(entsql.IndexWhere("share_token_hash <> ''")),
	}
}
