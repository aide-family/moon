package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ProbeTarget 是探测目标组里的一个地址。spec 保存归一化后的目标串，
// 组内不允许重复；目标可以单独停用。
type ProbeTarget struct {
	ent.Schema
}

func (ProbeTarget) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (ProbeTarget) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("group_uid").NotEmpty(),
		// spec 是归一化后的目标：HTTP 为去掉默认端口、末尾 # 与 userinfo 的 URL。
		field.String("spec").NotEmpty(),
		field.String("labels").Default("{}"),
		field.Enum("status").Values("ACTIVE", "DISABLED").Default("ACTIVE"),
		field.Int("sort").Default(0),
	}
}

func (ProbeTarget) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("group_uid", "spec").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("group_uid"),
		index.Fields("namespace_uid"),
	}
}
