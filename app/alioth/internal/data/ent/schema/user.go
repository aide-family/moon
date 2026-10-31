package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// User 是平台级用户，不隶属于任何租户。
// SystemRole 只表达系统权限，和租户内的角色完全隔离。
// 用户删除是软删：绑定、成员和配置由数据层清理，行本身保留供审计追溯。
type User struct {
	ent.Schema
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
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
		// SYSTEM_READONLY 是可查看平台全部信息但没有任何写权限的只读管理员。
		field.Enum("system_role").Values("SUPER_ADMIN", "SYSTEM_ADMIN", "SYSTEM_READONLY", "USER").Default("USER"),
		field.String("authorized_by").Default(""),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		// 软删用户不再占用邮箱：删除错误账号后，同一个邮箱可以重新注册。
		index.Fields("email").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		// 平台只能有一个超级管理员：第一个注册用户。
		index.Fields("system_role").Unique().Annotations(entsql.IndexWhere("system_role = 'SUPER_ADMIN'")),
	}
}
