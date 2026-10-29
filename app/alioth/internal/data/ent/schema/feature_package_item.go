package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// FeaturePackageItem 是套餐里的一个功能。同一套餐同一功能只能有一条记录。
type FeaturePackageItem struct {
	ent.Schema
}

func (FeaturePackageItem) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (FeaturePackageItem) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("package_uid").NotEmpty(),
		// 功能编码由 perm 包维护，不在这里用枚举锁定，便于功能目录调整。
		field.String("feature").NotEmpty(),
	}
}

func (FeaturePackageItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("package_uid", "feature").Unique(),
	}
}
