package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// User 是平台级用户，不隶属于任何租户。
// SystemRole 只表达系统权限，和租户内的角色完全隔离。
type User struct {
	ent.Schema
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("name").Default(""),
		field.String("nickname").Default(""),
		field.String("email").NotEmpty(),
		field.String("avatar").Default(""),
		// phone 用于短信通知；留空表示未配置。
		field.String("phone").Default(""),
		field.String("remark").Default(""),
		field.Enum("status").Values("ACTIVE", "BANNED").Default("ACTIVE"),
		field.String("password_hash").Sensitive().Default(""),
		field.Enum("system_role").Values("SUPER_ADMIN", "SYSTEM_ADMIN", "USER").Default("USER"),
		field.String("authorized_by").Default(""),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("email").Unique(),
		// 平台只能有一个超级管理员：第一个注册用户。
		index.Fields("system_role").Unique().Annotations(entsql.IndexWhere("system_role = 'SUPER_ADMIN'")),
	}
}
