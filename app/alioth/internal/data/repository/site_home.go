package repository

import (
	"context"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// SiteHomeRepo 维护官网首页的版式文案与能力卡片。两者都按语言归属，
// 首页只读已发布之外没有第二套快照：这里的内容保存即生效。
type SiteHomeRepo interface {
	// HomeContent 读取某个语言的首页文案；没有保存过时返回零值，不返回 NotFound。
	HomeContent(ctx context.Context, locale string) (*bo.SiteHomeContent, error)
	// SaveHomeContent 覆盖保存某个语言的首页文案。
	SaveHomeContent(ctx context.Context, in bo.SiteHomeContent) (*bo.SiteHomeContent, error)
	// SiteFeatures 按排序返回某个语言的能力卡片。
	SiteFeatures(ctx context.Context, locale string) ([]bo.SiteFeature, error)
	// ReplaceSiteFeatures 用给定列表整体替换某个语言的能力卡片，顺序即数组顺序。
	ReplaceSiteFeatures(ctx context.Context, locale string, items []bo.SiteFeature) ([]bo.SiteFeature, error)
}
