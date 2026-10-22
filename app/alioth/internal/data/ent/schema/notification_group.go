package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// NotificationGroup 是租户内的接收方集合。
type NotificationGroup struct {
	ent.Schema
}

func (NotificationGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (NotificationGroup) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("name").NotEmpty(),
		field.Enum("status").Values("ACTIVE", "DISABLED").Default("ACTIVE"),
		// channel_uids 和 template_uids 一一对应：邮件渠道、短信渠道和每个 Webhook 接收者各一项。
		field.Strings("channel_uids").Default([]string{}),
		field.Strings("template_uids").Default([]string{}),
		// email_receivers 是手填邮箱，sms_receivers 是手填手机号。
		field.String("email_receivers").Default(""),
		field.String("sms_receivers").Default(""),
		// email_member_uids 选择成员邮箱，sms_member_uids 选择成员手机号。
		field.Strings("email_member_uids").Default([]string{}),
		field.Strings("sms_member_uids").Default([]string{}),
		field.String("remark").Default(""),
		// 以下是旧版单渠道字段，只为兼容历史数据和旧客户端保留。
		field.String("channel_uid").Default(""),
		field.String("template_uid").Default(""),
		field.String("receivers").Default(""),
		field.Strings("member_uids").Default([]string{}),
		field.Strings("webhook_channel_uids").Default([]string{}),
		field.String("webhook_template_uid").Default(""),
	}
}

func (NotificationGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "name").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("namespace_uid", "status"),
	}
}
