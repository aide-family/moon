package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Board 是一块「大盘」：把多个 Grafana 面板链接按布局组合成一个整板。
// 只保存展示元数据与链接，不保存 Grafana 凭据，也不代理 Grafana 请求。
type Board struct {
	ent.Schema
}

func (Board) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (Board) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("name").NotEmpty(),
		field.String("description").Default(""),
		field.Enum("status").Values("ACTIVE", "DISABLED").Default("ACTIVE"),
		// created_by 是创建人用户标识，只用于展示与审计。
		field.String("created_by").Default(""),
	}
}

func (Board) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "name").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("namespace_uid", "status"),
	}
}
