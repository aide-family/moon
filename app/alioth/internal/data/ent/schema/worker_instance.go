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
		// tasks 是本批领取的「规则 × 数据源」组合快照，存 JSON 文本；
		// 每次心跳整批覆盖，后台工作进程页面只做展示，不参与查询条件。
		field.String("tasks").Default("[]"),
		// round_started_at 是本批领取时间，没有领取批次时为空。
		field.Time("round_started_at").Optional().Nillable(),
		// next_poll_at 是本实例下一次轮询时间，由天权自己上报。
		field.Time("next_poll_at").Optional().Nillable(),
		// poll_interval_seconds 是本实例的轮询周期，用于页面倒计时。
		field.Int32("poll_interval_seconds").Default(0),
		field.Time("last_seen_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (WorkerInstance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("service", "instance_uid").Unique(),
		index.Fields("last_seen_at"),
	}
}
