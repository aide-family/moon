package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// NamespaceFeature 是租户相对套餐的额外能力。
// 只存「在套餐之外为租户加开」的能力，enabled 恒为 true；没有记录表示完全跟随套餐。
// 能力只能加不能减：目标能力必须包含套餐能力，因此不存在 enabled=false 的正常数据。
// 升级前写入的 enabled=false 旧记录由 seed 清理，不再参与计算。
// 有效能力 = 套餐能力 ∪ 额外能力，不落表。
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
