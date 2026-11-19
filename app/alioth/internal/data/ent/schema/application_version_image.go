package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ApplicationVersionImage 是发布包版本里扫描到的镜像快照，用于发布/部署校验
// 与「被谁引用」反查。
type ApplicationVersionImage struct {
	ent.Schema
}

func (ApplicationVersionImage) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (ApplicationVersionImage) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("application_uid").NotEmpty(),
		field.String("version_uid").NotEmpty(),
		field.String("resource_item").NotEmpty(),
		field.String("resource_uid").Default(""),
		field.String("component").Default(""),
		field.String("container").Default(""),
		field.String("container_path").Default(""),
		field.String("image").NotEmpty(),
		field.String("registry_uid").Default(""),
		field.String("registry_name").Default(""),
		field.String("repository").Default(""),
		field.String("tag").Default(""),
		field.String("digest").Default(""),
		field.String("project").Default(""),
		// source 取 REGISTRY / EXTERNAL；external 保留给旧客户端兼容。
		field.String("source").Default(""),
		field.String("resolved_image").Default(""),
		field.String("platforms").Default("[]"),
		field.Bool("external").Default(false),
		field.Time("resolved_at").Optional().Nillable(),
	}
}

func (ApplicationVersionImage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("version_uid", "resource_item", "container").Unique(),
		index.Fields("namespace_uid", "repository", "tag"),
		index.Fields("registry_uid"),
	}
}
