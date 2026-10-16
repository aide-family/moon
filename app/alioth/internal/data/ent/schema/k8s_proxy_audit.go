package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// K8sProxyAudit 只追加。记录每一次 Kubernetes API 代理。
type K8sProxyAudit struct {
	ent.Schema
}

func (K8sProxyAudit) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (K8sProxyAudit) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("agent_uid").Default(""),
		field.String("operator_uid").Default(""),
		field.String("method").Default(""),
		field.String("resource").Default(""),
		field.String("kube_namespace").Default(""),
		field.String("result").Default("PENDING"),
		field.String("detail").Default(""),
	}
}

func (K8sProxyAudit) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid"),
		index.Fields("agent_uid"),
	}
}
