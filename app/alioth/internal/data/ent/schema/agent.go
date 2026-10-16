package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Agent 是一条边缘机器记录。令牌只存哈希，SSH 凭据加密后入库。
type Agent struct {
	ent.Schema
}

func (Agent) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (Agent) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("name").Default(""),
		field.String("hostname").Default(""),
		field.String("machine_id").Default(""),
		field.String("local_ip").Default(""),
		field.String("version").Default(""),
		field.String("collect_json").Default(""),
		field.Bool("online").Default(false),
		field.String("token_hash").Sensitive().Default(""),
		field.Enum("status").Values("ACTIVE", "REVOKED").Default("ACTIVE"),
		field.String("ssh_host").Default(""),
		field.Int("ssh_port").Default(22),
		field.String("ssh_user").Default(""),
		field.String("ssh_password_cipher").Sensitive().Default(""),
		field.String("ssh_private_key_cipher").Sensitive().Default(""),
		field.String("remark").Default(""),
	}
}

func (Agent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "machine_id").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL AND machine_id <> ''")),
		index.Fields("namespace_uid", "status"),
		index.Fields("token_hash"),
	}
}
