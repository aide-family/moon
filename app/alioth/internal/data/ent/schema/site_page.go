package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SitePage 是官网的一个页面，按 (slug, locale) 唯一，未删除记录上不允许重复。
// 草稿与发布快照分两套字段：官网只读 published_*，后台编辑 draft_*，发布时把草稿
// 整份复制成快照，因此「保存但没发布」不会影响线上内容，下架也不丢已发布内容。
type SitePage struct {
	ent.Schema
}

func (SitePage) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}, SoftDelete{}}
}

func (SitePage) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		// slug 只允许小写字母、数字与连字符；首页固定为 home，不可下架或删除。
		field.String("slug").NotEmpty(),
		field.String("locale").NotEmpty(),
		field.Enum("state").Values("DRAFT", "PUBLISHED", "ARCHIVED").Default("DRAFT"),
		field.Bool("show_in_nav").Default(false),
		// sort_order 在同一语言内唯一，由 biz 层保证。
		field.Int("sort_order").Default(0),
		// 草稿：后台编辑的就是这一套。
		field.String("draft_title").Default(""),
		field.String("draft_summary").Default(""),
		field.String("draft_body_md").Default(""),
		field.String("draft_cover_image_url").Default(""),
		field.String("draft_seo_title").Default(""),
		field.String("draft_seo_description").Default(""),
		// 发布快照：官网唯一读取的内容，只由 PublishSitePage 写入。
		field.String("published_title").Default(""),
		field.String("published_summary").Default(""),
		field.String("published_body_md").Default(""),
		field.String("published_cover_image_url").Default(""),
		field.String("published_seo_title").Default(""),
		field.String("published_seo_description").Default(""),
		field.Time("published_at").Optional().Nillable(),
		field.String("published_by").Default(""),
	}
}

func (SitePage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		// 软删的行不占地址：删掉页面后可以用同一个 slug 重新新建。
		index.Fields("slug", "locale").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("locale", "state"),
	}
}
