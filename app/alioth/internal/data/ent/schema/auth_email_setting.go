package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AuthEmailSetting 是登录验证码与邀请邮件的 SMTP 配置，平台级单例。
// 没有这一行、或 enabled 为假时，发信回落到配置文件里的 SMTP。
// 密码只存密文，接口永远不回显。
type AuthEmailSetting struct {
	ent.Schema
}

func (AuthEmailSetting) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (AuthEmailSetting) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		// 是否使用后台配置发信；关闭时回落到服务器配置。
		field.Bool("enabled").Default(true),
		field.String("host").Default(""),
		field.Int("port").Default(587),
		field.String("username").Default(""),
		field.String("password_cipher").Sensitive().Default(""),
		field.String("from").Default(""),
		field.String("updated_by").Default(""),
	}
}

func (AuthEmailSetting) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
	}
}
