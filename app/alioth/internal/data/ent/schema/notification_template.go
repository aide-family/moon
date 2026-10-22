package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// NotificationTemplate 按消息类型保存正文。
type NotificationTemplate struct {
	ent.Schema
}

func (NotificationTemplate) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (NotificationTemplate) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("name").NotEmpty(),
		field.Enum("type").Values("EMAIL", "WEBHOOK", "SMS_ALICLOUD"),
		// app 区分 Webhook 模板的接收应用；OTHER 是通用模板，EMAIL 和短信固定 OTHER。
		field.Enum("app").Values("OTHER", "DINGTALK", "WECHAT", "FEISHU", "SMS_ALICLOUD").Default("OTHER"),
		field.Enum("status").Values("ACTIVE", "DISABLED").Default("ACTIVE"),
		field.String("subject").Default(""),
		field.String("body").Default(""),
		field.String("remark").Default(""),
	}
}

func (NotificationTemplate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "name").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("namespace_uid", "status"),
	}
}
