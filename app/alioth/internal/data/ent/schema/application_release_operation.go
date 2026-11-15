package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ApplicationReleaseOperation 是某个集群实例的一次安装/升级/回滚/卸载/重新部署。
type ApplicationReleaseOperation struct {
	ent.Schema
}

func (ApplicationReleaseOperation) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (ApplicationReleaseOperation) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("deployment_uid").NotEmpty(),
		field.String("release_uid").NotEmpty(),
		field.String("cluster_name").Default(""),
		field.String("batch_uid").Default(""),
		field.Enum("action").Values("INSTALL", "UPGRADE", "ROLLBACK", "UNINSTALL", "REDEPLOY").Default("INSTALL"),
		field.Enum("status").Values("PENDING", "RUNNING", "SUCCEEDED", "FAILED", "CANCELED").Default("PENDING"),
		field.String("from_version_uid").Default(""),
		field.String("to_version_uid").Default(""),
		field.Text("values").Default("{}"),
		field.String("secret_values_cipher").Sensitive().Default(""),
		// resources 是本次应用的资源标识 JSON，用于升级裁剪与卸载。
		field.Text("resources").Default("[]"),
		field.Text("diff").Default("{}"),
		field.String("message").Default(""),
		field.String("operator_uid").Default(""),
		field.String("operator_name").Default(""),
		field.Bool("keep_persistent_volume_claims").Default(true),
		field.Bool("delete_cluster_scoped_resources").Default(false),
		field.String("lease_owner").Default(""),
		field.Time("lease_until").Optional().Nillable(),
		field.Time("started_at").Optional().Nillable(),
		field.Time("finished_at").Optional().Nillable(),
	}
}

func (ApplicationReleaseOperation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("release_uid").Unique().
			Annotations(entsql.IndexWhere("status IN ('PENDING','RUNNING')")),
		index.Fields("deployment_uid", "created_at"),
		index.Fields("batch_uid"),
		index.Fields("status", "lease_until"),
	}
}
