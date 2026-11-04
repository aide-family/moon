package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SiteSetting 是官网（公开站）的站点级配置，全平台只有一行。
// 页脚备案、版权与联系方式从这里下发：改一次，官网、登录页与注册页同时生效，
// 不再把备案号写死在 i18n 文案里。备案字段留空表示不展示，不退化成占位假号。
type SiteSetting struct {
	ent.Schema
}

func (SiteSetting) Mixin() []ent.Mixin {
	return []ent.Mixin{Audit{}}
}

func (SiteSetting) Fields() []ent.Field {
	return []ent.Field{
		field.String("uid").NotEmpty().Immutable(),
		// 站点名称，留空时前端回落为 Alioth。
		field.String("site_name").Default(""),
		// logo 与 favicon 只收 URL，不做文件上传。
		field.String("logo_url").Default(""),
		field.String("favicon_url").Default(""),
		field.String("contact_email").Default(""),
		field.String("repository_url").Default(""),
		// 版权文案，留空时前端按当前年份回落。
		field.String("copyright_text").Default(""),
		// ICP 备案号与查询地址；号码为空时整条不展示。
		field.String("filing_icp_number").Default(""),
		field.String("filing_icp_url").Default(""),
		// 公安备案号与查询地址；号码为空时整条不展示。
		field.String("filing_police_number").Default(""),
		field.String("filing_police_url").Default(""),
		// 备案补充说明，例如备案主体名称。
		field.String("filing_extra_text").Default(""),
		field.String("updated_by").Default(""),
	}
}

func (SiteSetting) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("uid").Unique(),
	}
}
