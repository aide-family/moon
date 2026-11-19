package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ApplicationVersionComponent 是版本内的一个服务（组件），由若干 Kubernetes
// 资源组成。版本内容按组件分组，sort_order 决定界面展示顺序。
type ApplicationVersionComponent struct {
	ent.Schema
}

func (ApplicationVersionComponent) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (ApplicationVersionComponent) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("application_uid").NotEmpty(),
		field.String("version_uid").NotEmpty(),
		field.String("name").NotEmpty(),
		field.String("display_name").Default(""),
		field.Int("sort_order").Default(0),
	}
}

func (ApplicationVersionComponent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("version_uid", "name").Unique(),
		index.Fields("namespace_uid", "version_uid"),
	}
}
