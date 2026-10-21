package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// UserAlertPage 记录用户在实时告警页选择显示的告警页面。
type UserAlertPage struct {
	ent.Schema
}

func (UserAlertPage) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (UserAlertPage) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("user_uid").NotEmpty(),
		field.String("alert_page_uid").NotEmpty(),
		// sort 是用户在实时告警页拖动后的 Tab 顺序，从 0 开始递增。
		field.Int("sort").Default(0),
	}
}

func (UserAlertPage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "user_uid", "alert_page_uid").Unique(),
		index.Fields("namespace_uid", "user_uid"),
	}
}
