package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// OnCallRotationMember 是值班轮换里的一个参与成员，position 决定轮转顺序。
// 成员随轮换整体保存，被移除的成员直接删除，不保留历史。
type OnCallRotationMember struct {
	ent.Schema
}

func (OnCallRotationMember) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (OnCallRotationMember) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("rotation_uid").NotEmpty(),
		field.String("member_uid").NotEmpty(),
		// position 从 0 开始，决定班次轮转顺序。
		field.Int("position"),
	}
}

func (OnCallRotationMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("rotation_uid", "member_uid").Unique(),
		index.Fields("namespace_uid", "rotation_uid"),
	}
}
