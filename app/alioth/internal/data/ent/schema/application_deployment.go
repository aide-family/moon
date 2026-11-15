package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ApplicationDeployment 是一次多集群部署的聚合。M36 先支持单集群，
// 数据结构从一开始保留集群列表与每集群实例。
type ApplicationDeployment struct {
	ent.Schema
}

func (ApplicationDeployment) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (ApplicationDeployment) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.String("name").NotEmpty(),
		field.String("application_uid").NotEmpty(),
		field.String("application_name").Default(""),
		field.String("version_uid").NotEmpty(),
		field.String("version").Default(""),
		field.String("target_namespace").NotEmpty(),
		// clusters 是集群快照 JSON：[{"uid","name"}]。
		field.Text("clusters").Default("[]"),
		field.Text("values").Default("{}"),
		field.String("secret_values_cipher").Sensitive().Default(""),
		field.Text("images").Default("[]"),
		field.Enum("status").Values("PENDING", "DEPLOYING", "RUNNING", "PARTIAL", "FAILED", "UNINSTALLING", "UNINSTALLED").Default("PENDING"),
		field.String("message").Default(""),
		field.String("created_by").Default(""),
		field.Time("deployed_at").Optional().Nillable(),
	}
}

func (ApplicationDeployment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "name").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("namespace_uid", "created_at"),
		index.Fields("status"),
	}
}
