package bo

import "time"

// SitePageContent 是一份页面内容。草稿与发布快照共用同一套字段，
// 发布时整份复制，因此两份内容可以逐字段比较出「有没有未发布的改动」。
type SitePageContent struct {
	Title          string
	Summary        string
	BodyMD         string
	CoverImageURL  string
	SEOTitle       string
	SEODescription string
}

// SitePage 是官网页面的一个语言版本，按 (Slug, Locale) 唯一。
type SitePage struct {
	UID       string
	Slug      string
	Locale    string
	State     string
	ShowInNav bool
	SortOrder int
	Draft     SitePageContent
	Published SitePageContent
	// PublishedAt 为空表示从未发布过。
	PublishedAt *time.Time
	PublishedBy string
	UpdatedAt   time.Time
}

// HasUnpublishedChanges 判断草稿是否与发布快照不同。
// 从未发布过时不算「有未发布的改动」，列表上只显示状态徽标。
func (page SitePage) HasUnpublishedChanges() bool {
	if page.PublishedAt == nil {
		return false
	}
	return page.Draft != page.Published
}

// SiteNavItem 是公开导航的一项。
type SiteNavItem struct {
	Slug      string
	Title     string
	SortOrder int
}

// PublicSitePage 是公开接口下发的页面，只含已发布字段。
type PublicSitePage struct {
	Slug           string
	Locale         string
	Title          string
	Summary        string
	BodyMD         string
	CoverImageURL  string
	SEOTitle       string
	SEODescription string
	PublishedAt    time.Time
	// LocaleFallback 表示请求的语言没有已发布页面，返回的是另一种语言的内容。
	LocaleFallback bool
}
