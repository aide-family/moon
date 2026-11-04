package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AuthOAuthProvider 是单个第三方登录来源的后台配置。
// 没有这一行时回落到配置文件里的同名来源；有这一行时以库里的为准：
// enabled 决定是否对外暴露，client id / secret 都不为空才算配置完成。
// secret 只存密文，接口只回显「是否已配置」。
type AuthOAuthProvider struct {
	ent.Schema
}

func (AuthOAuthProvider) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (AuthOAuthProvider) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		// provider 是稳定标识，取值为 GITHUB、GITEE、FEISHU、DINGTALK。
		field.String("provider").NotEmpty(),
		field.Bool("enabled").Default(false),
		field.String("client_id").Default(""),
		field.String("client_secret_cipher").Sensitive().Default(""),
		field.String("updated_by").Default(""),
	}
}

func (AuthOAuthProvider) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("provider").Unique(),
	}
}
