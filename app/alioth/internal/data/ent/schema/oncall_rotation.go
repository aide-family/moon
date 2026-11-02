package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// OnCallRotation 是一条值班轮换：按固定周期与时区在成员之间轮转，
// 绑定到 OnCall 的通知组后，投递时把当前当班成员并入接收方。
type OnCallRotation struct {
	ent.Schema
}

func (OnCallRotation) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (OnCallRotation) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("name").NotEmpty(),
		field.String("remark").Default(""),
		// timezone 是 IANA 时区名，保存规范名称；handoff_hour 按该时区解释。
		field.String("timezone").Default("Asia/Shanghai"),
		// period 决定一轮班次的长度：DAILY 每天交班，WEEKLY 每周交班。
		field.Enum("period").Values("DAILY", "WEEKLY").Default("DAILY"),
		// starts_at 是轮换起点，第一班从该时刻开始；按 UTC 保存。
		field.Time("starts_at"),
		// handoff_hour 是本地时区的交班时刻，0–23。
		field.Int("handoff_hour").Default(0),
		field.Enum("status").Values("ACTIVE", "DISABLED").Default("ACTIVE"),
	}
}

func (OnCallRotation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "name").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("namespace_uid", "status"),
	}
}
