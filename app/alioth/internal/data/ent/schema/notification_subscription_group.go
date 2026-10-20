package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// NotificationSubscriptionGroup 把订阅绑定到通知组。
type NotificationSubscriptionGroup struct {
	ent.Schema
}

func (NotificationSubscriptionGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (NotificationSubscriptionGroup) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("subscription_uid").NotEmpty(),
		field.String("group_uid").NotEmpty(),
	}
}

func (NotificationSubscriptionGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("subscription_uid", "group_uid").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("subscription_uid"),
		index.Fields("group_uid"),
	}
}
