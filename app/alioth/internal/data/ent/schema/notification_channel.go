package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// NotificationChannel 是一条通知渠道。密码和签名加密后入库。
type NotificationChannel struct {
	ent.Schema
}

func (NotificationChannel) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (NotificationChannel) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("name").NotEmpty(),
		field.Enum("type").Values("SMTP", "WEBHOOK", "SMS"),
		field.Enum("app").Values("OTHER", "DINGTALK", "WECHAT", "FEISHU", "SMS_ALICLOUD").Default("OTHER"),
		field.Enum("status").Values("ACTIVE", "DISABLED").Default("ACTIVE"),
		field.String("endpoint").Default(""),
		field.String("username").Default(""),
		field.String("password_cipher").Sensitive().Default(""),
		field.String("secret_cipher").Sensitive().Default(""),
		field.String("from_address").Default(""),
		field.String("remark").Default(""),
	}
}

func (NotificationChannel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "name").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("namespace_uid", "status"),
	}
}
