package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SFTPAudit 只追加。记录每一次 SFTP 操作。
type SFTPAudit struct {
	ent.Schema
}

func (SFTPAudit) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (SFTPAudit) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("agent_uid").Default(""),
		field.String("operator_uid").Default(""),
		field.String("path").Default(""),
		field.String("action").Default(""),
		field.String("result").Default("PENDING"),
		field.String("detail").Default(""),
	}
}

func (SFTPAudit) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid"),
		index.Fields("agent_uid"),
	}
}
