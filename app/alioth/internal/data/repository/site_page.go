package repository

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// SitePageFilter 是官网页面列表的过滤条件。所有字段都是可选的。
type SitePageFilter struct {
	Page     int32
	PageSize int32
	Keyword  string
	State    string
	Locale   string
}

// SitePageRepo 维护官网页面。页面按 (slug, locale) 唯一，
// 草稿与发布快照分开保存，官网只读发布快照。
type SitePageRepo interface {
	// ListSitePages 按过滤条件分页列出页面，返回当前页、总数与规范化后的页码与页大小。
	ListSitePages(ctx context.Context, filter SitePageFilter) ([]bo.SitePage, int, int, int, error)
	// GetSitePage 按标识读取页面，未删除记录上查找；不存在返回 NotFound。
	GetSitePage(ctx context.Context, uid string) (*bo.SitePage, error)
	// GetSitePageBySlug 按 (slug, locale) 读取页面，用于新建时查重。
	GetSitePageBySlug(ctx context.Context, slug, locale string) (*bo.SitePage, error)
	// CreateSitePage 写入一条草稿状态的页面。
	CreateSitePage(ctx context.Context, in bo.SitePage) (*bo.SitePage, error)
	// UpdateSitePageDraft 覆盖草稿。expectedUpdatedAt 非空时做乐观锁，行已被改动返回 Conflict。
	UpdateSitePageDraft(ctx context.Context, in bo.SitePage, expectedUpdatedAt *time.Time) (*bo.SitePage, error)
	// PublishSitePage 把草稿整份复制成发布快照，并置为已发布。
	PublishSitePage(ctx context.Context, uid, publishedBy string) (*bo.SitePage, error)
	// UnpublishSitePage 把页面置为已下架，发布快照保留。
	UnpublishSitePage(ctx context.Context, uid string) (*bo.SitePage, error)
	// DeleteSitePage 软删页面。
	DeleteSitePage(ctx context.Context, uid string) error
	// UpdateSitePageNav 修改导航展示与顺序。
	UpdateSitePageNav(ctx context.Context, uid string, showInNav bool, sortOrder int) (*bo.SitePage, error)
	// GetPublishedSitePage 读取指定语言下已发布的页面；不存在返回 NotFound。
	GetPublishedSitePage(ctx context.Context, slug, locale string) (*bo.SitePage, error)
	// ListPublishedSiteNav 列出指定语言下已发布且展示在导航里的页面，按导航顺序排列。
	ListPublishedSiteNav(ctx context.Context, locale string) ([]bo.SiteNavItem, error)
}
