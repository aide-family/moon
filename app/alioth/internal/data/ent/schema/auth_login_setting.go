package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AuthLoginSetting 是登录方式的平台级开关，全平台只有一行。
// 邮箱验证码与第三方（OAuth2）登录各自独立开关：第三方来源的客户端凭据仍来自配置，
// 这里的开关决定是否对外暴露与是否允许发起登录。两种都关闭会让所有人都登不进来，
// 由 biz 层拒绝这种保存。
type AuthLoginSetting struct {
	ent.Schema
}

func (AuthLoginSetting) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (AuthLoginSetting) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		// 是否允许邮箱验证码登录，默认开启。
		field.Bool("email_enabled").Default(true),
		// 是否允许第三方登录，默认开启。
		field.Bool("oauth2_enabled").Default(true),
		field.String("updated_by").Default(""),
	}
}

func (AuthLoginSetting) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
	}
}
