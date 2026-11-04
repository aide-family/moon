package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SiteHomeContent 是官网首页的版式文案，按语言各一行。
// 版式固定在 web 端，这里只承载标题、副标、CTA 文案与三个解释块的短文本；
// 字段长度上限由 biz 层按版式的字数预算校验，避免后台自由发挥把版式撑坏。
type SiteHomeContent struct {
	ent.Schema
}

func (SiteHomeContent) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (SiteHomeContent) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		// locale 取值为 zh 或 en，同一语言一行。
		field.String("locale").NotEmpty(),
		field.String("hero_title").Default(""),
		field.String("hero_subtitle").Default(""),
		field.String("hero_image_url").Default(""),
		field.String("cta_primary_label").Default(""),
		field.String("cta_secondary_label").Default(""),
		field.String("cta_secondary_href").Default(""),
		field.String("features_title").Default(""),
		field.String("features_subtitle").Default(""),
		field.String("how_title").Default(""),
		field.String("how_body").Default(""),
		field.String("perm_title").Default(""),
		field.String("perm_body").Default(""),
		field.String("deploy_title").Default(""),
		field.String("deploy_body").Default(""),
		field.String("doc_title").Default(""),
		field.String("closing_title").Default(""),
		field.String("closing_description").Default(""),
		field.String("closing_button_label").Default(""),
		field.String("updated_by").Default(""),
	}
}

func (SiteHomeContent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("locale").Unique(),
	}
}
