package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ApplicationVersion 是发布包版本。草稿可编辑，上架后内容不可变；内容按服务
// （组件）拆分到 application_version_components/resources。spec 保留为旧客户端
// 兼容的 canonical JSON，Secret 值单独加密在 secrets_cipher。
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
		// revision 是版本内容的乐观锁版本号；composition_version 为 1 时组件与
		// 资源来自 application_version_components / application_version_resources，
		// 否则读 spec 兼容旧数据。
		field.Int("revision").Default(1),
		field.Int("composition_version").Default(0),
		field.Time("status_changed_at").Optional().Nillable(),
		field.Time("deprecated_at").Optional().Nillable(),
		field.String("copied_from_version_uid").Default(""),
	}
}

func (ApplicationVersion) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("application_uid", "version").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("namespace_uid", "application_uid", "status"),
	}
}
