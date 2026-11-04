package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SiteFeature 是官网首页能力卡片的一项，按语言归属。
// 同一语言最多 6 张、sort_order 唯一，由 biz 层在保存时校验并整体替换。
type SiteFeature struct {
	ent.Schema
}

func (SiteFeature) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (SiteFeature) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		field.String("locale").NotEmpty(),
		field.Int("sort_order").Default(0),
		// icon_key 是前端图标白名单里的键名，不是任意图标名。
		field.String("icon_key").Default(""),
		field.String("title").Default(""),
		field.String("description").Default(""),
	}
}

func (SiteFeature) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
		index.Fields("locale", "sort_order"),
	}
}
