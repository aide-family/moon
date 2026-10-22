package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// NotificationPreset 是系统预置的通知模板，不属于任何租户。
// 用户从预设创建模板后会得到一条可编辑的 NotificationTemplate 副本。
type NotificationPreset struct {
	ent.Schema
}

func (NotificationPreset) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (NotificationPreset) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("name").NotEmpty(),
		field.Enum("type").Values("EMAIL", "WEBHOOK", "SMS_ALICLOUD"),
		// app 区分 Webhook 预设的应用；OTHER 是通用模板。
		field.Enum("app").Values("OTHER", "DINGTALK", "WECHAT", "FEISHU", "SMS_ALICLOUD").Default("OTHER"),
		field.String("subject").Default(""),
		field.String("body").Default(""),
		field.String("description").Default(""),
		field.Int("sort").Default(0),
		field.Enum("status").Values("ACTIVE", "DISABLED").Default("ACTIVE"),
	}
}

func (NotificationPreset) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("name").Unique(),
		index.Fields("status", "sort"),
	}
}
