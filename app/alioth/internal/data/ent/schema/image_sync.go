package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ImageSync 是一次镜像同步任务。源可以是已接入的 Registry（CONNECTION），
// 也可以是外部 OCI 引用（EXTERNAL）。任务由调度器按租约认领执行。
type ImageSync struct {
	ent.Schema
}

func (ImageSync) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (ImageSync) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("namespace_uid").NotEmpty(),
		field.Enum("source_kind").Values("CONNECTION", "EXTERNAL").Default("CONNECTION"),
		field.String("source_registry_uid").Default(""),
		field.String("source_registry_name").Default(""),
		field.String("source_endpoint").Default(""),
		field.String("source_image").Default(""),
		field.String("source_username").Default(""),
		field.String("source_repository").Default(""),
		field.String("source_tag").Default(""),
		field.String("source_digest").Default(""),
		// source_platforms 是 JSON 数组，例如 ["linux/amd64","linux/arm64/v8"]。
		field.String("source_platforms").Default("[]"),
		// source_credential_cipher 只在 EXTERNAL 源使用，保存加密后的临时用户名/密码。
		field.String("source_credential_cipher").Sensitive().Default(""),
		field.Bool("source_allow_insecure_tls").Default(false),
		field.String("target_registry_uid").NotEmpty(),
		field.String("target_registry_name").Default(""),
		field.String("target_endpoint").Default(""),
		field.String("target_repository").NotEmpty(),
		field.String("target_tag").NotEmpty(),
		field.String("target_digest").Default(""),
		field.Enum("mode").Values("ALL_PLATFORMS", "SELECTED_PLATFORMS").Default("ALL_PLATFORMS"),
		field.Bool("overwrite").Default(false),
		field.Enum("status").Values("PENDING", "RUNNING", "SUCCEEDED", "FAILED", "CANCELED").Default("PENDING"),
		field.String("stage").Default(""),
		field.Int("copied_blobs").Default(0),
		field.Int("total_blobs").Default(0),
		field.String("message").Default(""),
		field.Int("attempt").Default(0),
		field.String("operator_uid").Default(""),
		field.String("operator_name").Default(""),
		field.String("lease_owner").Default(""),
		field.Time("lease_until").Optional().Nillable(),
		field.Time("started_at").Optional().Nillable(),
		field.Time("finished_at").Optional().Nillable(),
	}
}

func (ImageSync) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("namespace_uid", "created_at"),
		index.Fields("status"),
		index.Fields("source_registry_uid"),
		index.Fields("target_registry_uid"),
		// 同一源/目标的活动任务只允许一条，防止重复排队。
		index.Fields("namespace_uid", "source_kind", "source_image", "target_registry_uid", "target_repository", "target_tag").
			Unique().
			Annotations(entsql.IndexWhere("status IN ('PENDING','RUNNING')")),
	}
}
