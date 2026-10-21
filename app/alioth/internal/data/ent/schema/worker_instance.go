package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// WorkerInstance 保存 Megrez 与 Merak 的最近心跳，支持多 Alioth 副本共享实例状态。
type WorkerInstance struct {
	ent.Schema
}

func (WorkerInstance) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("service").Values("MEGREZ", "MERAK"),
		field.String("instance_uid").NotEmpty(),
		field.String("hostname").Default(""),
		field.Bool("busy").Default(false),
		field.String("task_uid").Default(""),
		field.String("task_name").Default(""),
		field.Time("task_started_at").Optional().Nillable(),
		field.Time("last_seen_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (WorkerInstance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("service", "instance_uid").Unique(),
		index.Fields("last_seen_at"),
	}
}
