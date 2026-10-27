package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Cluster 是一个租户接入的 Kubernetes 集群。
// kubeconfig 加密后保存，接口只返回 API Server、版本等元数据。
type Cluster struct {
	ent.Schema
}

func (Cluster) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (Cluster) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("name").NotEmpty(),
		// server 是 kubeconfig 里的 API Server 地址，用于列表展示与区分同名集群。
		field.String("server").Default(""),
		// version 是最近一次成功探测到的 Kubernetes 版本。
		field.String("version").Default(""),
		field.String("cluster_name").Default(""),
		field.String("context_name").Default(""),
		// kubeconfig_cipher 保存加密后的 kubeconfig 原文，从不直接返回给调用方。
		// 删除集群时会清空，避免软删除后仍保留可用于连接的凭据。
		field.String("kubeconfig_cipher").Sensitive().Default(""),
		field.String("remark").Default(""),
		// last_checked_at 是最近一次连通性检测时间；last_error 为空表示最近一次成功。
		field.Int64("last_checked_at").Default(0),
		field.String("last_error").Default(""),
	}
}

func (Cluster) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "name").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("namespace_uid"),
	}
}
