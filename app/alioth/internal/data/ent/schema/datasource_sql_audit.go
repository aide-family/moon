package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// DatasourceSQLAudit 只追加，记录数据源上的建表、改表、删表。
// 语句在用户数据库里执行，审计保留实际执行的 SQL、结果和操作人。
type DatasourceSQLAudit struct {
	ent.Schema
}

func (DatasourceSQLAudit) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (DatasourceSQLAudit) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("datasource_uid").NotEmpty(),
		field.String("operator_uid").Default(""),
		field.Enum("action").Values("CREATE_TABLE", "ALTER_TABLE", "DROP_TABLE"),
		field.Enum("status").Values("SUCCESS", "FAILED"),
		field.String("database").Default(""),
		field.String("schema").Default(""),
		field.String("table").Default(""),
		field.String("statement").Default(""),
		// 失败原因只保留第一行摘要，避免把上游整段错误写进审计。
		field.String("error_message").Default(""),
	}
}

func (DatasourceSQLAudit) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "datasource_uid", "created_at"),
	}
}
