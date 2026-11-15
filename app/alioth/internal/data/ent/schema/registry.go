package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Registry 是一个租户接入的镜像仓库连接。密码只保存密文，接口从不回传。
// driver 决定读取元数据时使用 Harbor 专有接口还是标准 OCI Distribution 接口。
type Registry struct {
	ent.Schema
}

func (Registry) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (Registry) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("name").NotEmpty(),
		field.Enum("driver").Values("HARBOR", "DOCKER_V2").Default("HARBOR"),
		field.String("endpoint").NotEmpty(),
		field.String("username").Default(""),
		// password_cipher 保存加密后的密码 / robot token，任何响应都不返回。
		field.String("password_cipher").Sensitive().Default(""),
		// allow_insecure_tls 只用于内网自签证书，默认关闭。
		field.Bool("allow_insecure_tls").Default(false),
		field.String("remark").Default(""),
		field.Int64("last_checked_at").Default(0),
		field.String("last_error").Default(""),
	}
}

func (Registry) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "name").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("namespace_uid"),
	}
}
