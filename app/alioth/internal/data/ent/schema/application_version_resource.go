package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ApplicationVersionResource 是版本内一个服务（组件）下的一条 Kubernetes 资源。
// manifest 存 canonical YAML；item 在版本内唯一；content_hash 供并发编辑提示。
type ApplicationVersionResource struct {
	ent.Schema
}

func (ApplicationVersionResource) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (ApplicationVersionResource) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("application_uid").NotEmpty(),
		field.String("version_uid").NotEmpty(),
		field.String("component_uid").NotEmpty(),
		field.String("component").NotEmpty(),
		field.String("item").NotEmpty(),
		field.Text("manifest").Default(""),
		field.String("api_version").Default(""),
		field.String("kind").Default(""),
		field.String("name").Default(""),
		field.String("namespace").Default(""),
		field.String("scope").Default("UNKNOWN"),
		field.Bool("cluster_scoped").Default(false),
		// images 是资源里扫描到的镜像列表 JSON；secret_keys 是 Secret 键名 JSON。
		field.Text("images").Default("[]"),
		field.Text("secret_keys").Default("[]"),
		field.String("content_hash").Default(""),
		field.Int("sort_order").Default(0),
	}
}

func (ApplicationVersionResource) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("version_uid", "item").Unique(),
		index.Fields("namespace_uid", "version_uid"),
		index.Fields("component_uid"),
	}
}
