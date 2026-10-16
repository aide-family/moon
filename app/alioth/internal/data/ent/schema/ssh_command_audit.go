package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SSHCommandAudit 只追加。记录创建、更新、审核和执行结果。
type SSHCommandAudit struct {
	ent.Schema
}

func (SSHCommandAudit) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (SSHCommandAudit) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("command_uid").Default(""),
		field.String("agent_uid").Default(""),
		field.String("operator_uid").Default(""),
		field.Enum("kind").Values("CREATE", "UPDATE", "APPROVE", "REJECT", "EXECUTE"),
		field.Enum("status").Values("PENDING", "APPROVED", "REJECTED", "DONE", "FAILED").Default("PENDING"),
		field.String("command").Default(""),
		field.Int("exit_code").Default(0),
		field.String("output").Default(""),
		field.Bool("truncated").Default(false),
		field.String("detail").Default(""),
	}
}

func (SSHCommandAudit) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid"),
		index.Fields("command_uid"),
		index.Fields("agent_uid"),
	}
}
