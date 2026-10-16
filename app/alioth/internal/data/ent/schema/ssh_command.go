package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SSHCommand 是一条可执行命令。只有审核通过后才可执行。
type SSHCommand struct {
	ent.Schema
}

func (SSHCommand) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (SSHCommand) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("name").NotEmpty(),
		field.String("command").NotEmpty(),
		field.String("workdir").Default(""),
		field.String("env").Default(""),
		field.Enum("review_status").Values("PENDING", "APPROVED", "REJECTED").Default("PENDING"),
		field.String("remark").Default(""),
	}
}

func (SSHCommand) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "name").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("namespace_uid", "review_status"),
	}
}
