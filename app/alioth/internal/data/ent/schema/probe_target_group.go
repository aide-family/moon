package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ProbeTargetGroup 是一组共用请求参数的探测地址。告警规则按类型 + 标签匹配它，
// 同一组里的每个地址独立探测、独立产生事件。
type ProbeTargetGroup struct {
	ent.Schema
}

func (ProbeTargetGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (ProbeTargetGroup) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("name").NotEmpty(),
		// 一期只有 PROBE_HTTP；TCP / ICMP 后续追加枚举值。
		field.Enum("driver").Values("PROBE_HTTP"),
		field.Enum("status").Values("ACTIVE", "DISABLED").Default("ACTIVE"),
		// labels 是规则按标签匹配目标组时使用的 JSON 对象。
		field.String("labels").Default("{}"),
		// executor_labels 预留给边缘探测：空对象表示由集群内执行器探测。
		field.String("executor_labels").Default("{}"),
		// request 是组内地址共用的请求参数 JSON：method、expect_status、timeout_ms、
		// follow_redirects、skip_tls_verify、verify_cert、headers、body_keyword。
		field.String("request").Default("{}"),
		field.String("remark").Default(""),
	}
}

func (ProbeTargetGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "name").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("namespace_uid", "status"),
	}
}
