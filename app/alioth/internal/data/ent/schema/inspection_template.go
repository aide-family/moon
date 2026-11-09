package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// InspectionTemplate 是报告版式：一份模板只对应一种格式（HTML 或 Markdown）。
// 模板全部由租户自定义，不再提供系统内置模板。
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
		// format 决定使用哪段正文，以及报告输出格式。
		field.Enum("format").Values("HTML", "MARKDOWN").Default("HTML"),
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
