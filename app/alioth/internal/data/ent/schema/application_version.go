package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ApplicationVersion 是一个不可变的发布包版本。spec 存任意 Kubernetes 资源清单
// 的 canonical JSON；Secret 值单独加密在 secrets_cipher。
type ApplicationVersion struct {
	ent.Schema
}

func (ApplicationVersion) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (ApplicationVersion) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("application_uid").NotEmpty(),
		field.String("version").NotEmpty(),
		field.Enum("status").Values("DRAFT", "PUBLISHED", "DEPRECATED").Default("DRAFT"),
		field.String("changelog").Default(""),
		// spec 形如 {"resources":[{"item":"...","manifest":{...}}]}，不含 Secret 值。
		field.Text("spec").Default(""),
		field.String("secrets_cipher").Sensitive().Default(""),
		field.Text("validation").Default(""),
		field.Time("published_at").Optional().Nillable(),
	}
}

func (ApplicationVersion) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("application_uid", "version").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("namespace_uid", "application_uid", "status"),
	}
}
