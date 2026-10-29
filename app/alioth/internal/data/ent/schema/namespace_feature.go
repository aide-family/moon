package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// NamespaceFeature 是租户相对套餐的功能差异。
// enabled=true 表示在套餐之外为租户加开功能，false 表示从套餐里减掉功能；
// 没有记录表示跟随套餐。有效功能由套餐功能叠加差异计算，不落表。
type NamespaceFeature struct {
	ent.Schema
}

func (NamespaceFeature) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (NamespaceFeature) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("feature").NotEmpty(),
		field.Bool("enabled").Default(true),
	}
}

func (NamespaceFeature) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "feature").Unique(),
	}
}
