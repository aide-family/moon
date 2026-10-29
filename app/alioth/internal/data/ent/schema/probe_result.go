package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ProbeResult 是一次探测的结果。(target_uid, slot) 唯一，保证回写幂等；
// 判定窗口和目标组详情都读这张表，按保留策略滚动清理。
type ProbeResult struct {
	ent.Schema
}

func (ProbeResult) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (ProbeResult) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("rule_uid").NotEmpty(),
		field.String("group_uid").NotEmpty(),
		field.String("target_uid").NotEmpty(),
		// slot 是这一轮的计划时刻，按探测周期对齐。
		field.Time("slot"),
		field.Time("probed_at"),
		field.Bool("ok").Default(false),
		// value 是与规则观测指标对应的数值：up 取 1/0，耗时为毫秒，状态码为整数，
		// 证书剩余天数为天数。
		field.String("value").Default(""),
		field.Int("duration_ms").Default(0),
		field.Int("status_code").Default(0),
		field.String("error_kind").Default(""),
		field.String("error_message").Default(""),
		field.Int("dns_ms").Default(0),
		field.Int("tcp_ms").Default(0),
		field.Int("tls_ms").Default(0),
		field.Int("ttfb_ms").Default(0),
		field.Int("cert_days").Default(0),
		field.String("resolved_ip").Default(""),
		field.String("executor").Default(""),
	}
}

func (ProbeResult) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("target_uid", "slot").Unique(),
		index.Fields("rule_uid", "group_uid", "target_uid", "probed_at"),
		index.Fields("namespace_uid", "probed_at"),
	}
}
