package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AlertEventColumn 保存用户在实时告警或历史告警明细上自定义的列。
// scope 为 mine 表示实时告警的「我的告警」，history 表示历史告警明细，
// 其余是告警页面 UID；一个用户在一个视图上只有一条配置。
type AlertEventColumn struct {
	ent.Schema
}

func (AlertEventColumn) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (AlertEventColumn) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("user_uid").NotEmpty(),
		field.String("scope").NotEmpty(),
		// columns_json 是列配置 JSON 数组，元素包含 key、type、label。
		field.String("columns_json").Default("[]"),
	}
}

func (AlertEventColumn) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "user_uid", "scope").Unique(),
		index.Fields("namespace_uid", "user_uid"),
	}
}
