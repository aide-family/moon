package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ApplicationRelease 是部署组在一个集群中的实例。
type ApplicationRelease struct {
	ent.Schema
}

func (ApplicationRelease) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (ApplicationRelease) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("deployment_uid").NotEmpty(),
		field.String("cluster_uid").NotEmpty(),
		field.String("cluster_name").Default(""),
		field.String("target_namespace").NotEmpty(),
		field.String("name").NotEmpty(),
		field.String("application_uid").NotEmpty(),
		field.String("application_name").Default(""),
		field.String("version_uid").NotEmpty(),
		field.String("version").Default(""),
		field.Enum("status").Values("PENDING", "DEPLOYING", "RUNNING", "FAILED", "UNINSTALLING", "UNINSTALLED").Default("PENDING"),
		field.Text("values").Default("{}"),
		field.String("secret_values_cipher").Sensitive().Default(""),
		field.Text("images").Default("[]"),
		field.String("message").Default(""),
		field.String("last_operation_uid").Default(""),
		field.Time("deployed_at").Optional().Nillable(),
	}
}

func (ApplicationRelease) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("deployment_uid", "cluster_uid").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("cluster_uid", "target_namespace", "name").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL AND status <> 'UNINSTALLED'")),
		index.Fields("namespace_uid", "deployment_uid"),
		index.Fields("status"),
	}
}
