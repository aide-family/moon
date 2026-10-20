package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// MemberInvite 是平台发给用户的一条租户邀请消息。
// 邀请只创建 INVITED 成员，用户在平台点击同意后才转为 JOINED。
// 邮箱未注册时 UserUID 为空，注册或登录后按邮箱匹配。
type MemberInvite struct {
	ent.Schema
}

func (MemberInvite) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (MemberInvite) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("email").NotEmpty(),
		field.String("user_uid").Default(""),
		field.String("role_uid").NotEmpty(),
		field.String("inviter_uid").NotEmpty(),
		field.String("token_hash").NotEmpty(),
		field.Enum("status").Values("PENDING", "ACCEPTED", "DECLINED", "REVOKED", "EXPIRED").Default("PENDING"),
		field.Time("expires_at"),
		field.Time("responded_at").Optional().Nillable(),
	}
}

func (MemberInvite) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("token_hash").Unique(),
		index.Fields("user_uid"),
		index.Fields("email"),
		// 同一租户同一邮箱只能有一条待确认邀请；重复邀请走刷新令牌和重发邮件。
		index.Fields("namespace_uid", "email").Unique().Annotations(entsql.IndexWhere("status = 'PENDING'")),
	}
}
