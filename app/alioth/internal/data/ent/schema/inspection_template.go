package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// InspectionTemplate 是报告版式：一份模板两段正文，HTML 与 Markdown 共用同一份数据契约。
// 内置模板的 namespace_uid 为空，只读；租户复制后成为自己的模板。
type InspectionTemplate struct {
	ent.Schema
}

func (InspectionTemplate) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (InspectionTemplate) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").Default(""),
		field.String("name").NotEmpty(),
		field.String("description").Default(""),
		field.Text("html_body").Default(""),
		field.Text("markdown_body").Default(""),
		field.Bool("is_builtin").Default(false),
		field.Enum("status").Values("ACTIVE", "DISABLED").Default("ACTIVE"),
		field.String("created_by").Default(""),
	}
}

func (InspectionTemplate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "name").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("is_builtin"),
	}
}
