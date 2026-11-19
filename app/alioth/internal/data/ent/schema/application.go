package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Application 是一个应用（发布包的身份）。版本与部署实例分别落在
// application_versions 与 application_deployments。
type Application struct {
	ent.Schema
}

func (Application) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (Application) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("name").NotEmpty(),
		field.String("display_name").Default(""),
		// category 是自由字符串，由业务层限制长度；内置分类只是默认值，不再做枚举校验。
		field.String("category").Default("OTHER"),
		field.String("description").Default(""),
		field.Text("readme").Default(""),
	}
}

func (Application) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "name").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("namespace_uid", "category"),
	}
}
