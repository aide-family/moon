package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// StatusPage 是对外公开的探测状态页配置。token 是不可猜测的公开访问凭证，
// 重新生成后旧链接立即失效；页面只展示目标组名称与状态，不展示目标地址与标签。
type StatusPage struct {
	ent.Schema
}

func (StatusPage) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (StatusPage) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("name").NotEmpty(),
		field.String("title").NotEmpty(),
		field.String("description").Default(""),
		// token 是 32 位随机字符串，公开访问凭证。
		field.String("token").NotEmpty(),
		field.Enum("status").Values("ACTIVE", "DISABLED").Default("ACTIVE"),
		// target_group_uids 是要展示的探测目标组，至少一个。
		field.Strings("target_group_uids").Default([]string{}),
		field.String("created_by").Default(""),
	}
}

func (StatusPage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("token").Unique(),
		index.Fields("namespace_uid", "name").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("namespace_uid", "status"),
	}
}
