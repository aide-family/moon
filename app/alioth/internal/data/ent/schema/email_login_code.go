package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type EmailLoginCode struct {
	ent.Schema
}

func (EmailLoginCode) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (EmailLoginCode) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("email").NotEmpty(),
		field.String("code_hash").Sensitive().NotEmpty(),
		field.Time("expires_at"),
		field.Time("used_at").Optional().Nillable(),
		// purpose 区分验证码用途（login/register/bind/change_email）。
		// 消费时必须匹配，避免绑定码被拿去登录；升级前的历史行是空串，按通配处理。
		field.String("purpose").Default(""),
		// attempts 记录该码已经验证失败的次数；达到阈值后立即作废，必须重新发码。
		field.Int("attempts").Default(0),
		// bind_token_hash 是发码时使用的第三方登录绑定凭证摘要，
		// 用于限制同一凭证可发往的不同邮箱数量，避免绑定发码被当成邮件轰炸入口。
		field.String("bind_token_hash").Default(""),
	}
}

func (EmailLoginCode) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("email"),
	}
}
