package bo

import "time"

// SiteHomeContent 是官网首页的版式文案，按语言各一份。
// 版式由前端固定，后台只填这些短文本；因此每个字段都有硬性字数上限，
// 目的是让「随便填」也不会把版式撑坏。
type SiteHomeContent struct {
	Locale string

	// Hero：首屏。主标题不许以站点名开头，避免首屏重复品牌。
	HeroTitle    string
	HeroSubtitle string
	HeroImageURL string
	// 主 CTA 只有文案，目标是系统槽位（注册 / 进入后台），后台改不了。
	CtaPrimaryLabel string
	// 次 CTA 文案与站内地址必须同存同空，地址只允许 / 或 # 开头。
	CtaSecondaryLabel string
	CtaSecondaryHref  string

	FeaturesTitle    string
	FeaturesSubtitle string

	HowTitle string
	HowBody  string

	PermTitle string
	PermBody  string

	DeployTitle string
	DeployBody  string

	DocTitle string

	ClosingTitle       string
	ClosingDescription string
	ClosingButtonLabel string

	UpdatedBy string
	UpdatedAt time.Time
}

// SiteFeature 是首页能力卡片的一项。图标只能是白名单里的键名。
type SiteFeature struct {
	UID         string
	Locale      string
	SortOrder   int
	IconKey     string
	Title       string
	Description string
}

// SiteHome 是一个语言下的首页内容：版式文案 + 能力卡片。
type SiteHome struct {
	Content  SiteHomeContent
	Features []SiteFeature
}
