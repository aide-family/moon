package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SystemMessage 是发给平台用户的一条系统消息。
// 邀请创建、同意和拒绝都会写入消息；未注册邮箱先按 email 匹配，注册后即可看到。
type SystemMessage struct {
	ent.Schema
}

func (SystemMessage) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (SystemMessage) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("user_uid").Default(""),
		field.String("email").Default(""),
		field.Enum("kind").Values("INVITE_RECEIVED", "INVITE_ACCEPTED", "INVITE_DECLINED", "MEMBER_REMOVED"),
		field.String("actor_name").Default(""),
		field.String("namespace_uid").Default(""),
		field.String("namespace_name").Default(""),
		field.String("role_name").Default(""),
		field.String("invite_uid").Default(""),
		field.Time("read_at").Optional().Nillable(),
	}
}

func (SystemMessage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("user_uid"),
		index.Fields("email"),
		index.Fields("invite_uid"),
		index.Fields("created_at"),
	}
}
