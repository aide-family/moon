package site

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

// homeSlug 是官网首页的固定 slug。首页是站点入口，不允许下架或删除。
const homeSlug = "home"

var (
	// sitePageSlugPattern 与接口层的校验一致：小写字母、数字与连字符。
	sitePageSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	// siteLocales 是官网支持的语言，顺序即回落顺序。
	siteLocales = []string{"zh", "en"}
)

// Pages 维护官网页面的草稿、发布快照与导航。
// 官网只读发布快照，后台编辑草稿；发布是唯一把草稿复制成快照的动作。
type Pages struct {
	pages  repository.SitePageRepo
	access access.Checker
}

func NewPages(pages repository.SitePageRepo, checker access.Checker) *Pages {
	return &Pages{pages: pages, access: checker}
}

// PublicPage 读取公开页面：只返回已发布快照。
// 请求的语言没有已发布页面时回落到另一种语言，并在响应里标明回落，
// 这样官网不会因为只写了中文就整页 404。
func (pages *Pages) PublicPage(ctx context.Context, slug, locale string) (*bo.PublicSitePage, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	if !sitePageSlugPattern.MatchString(slug) {
		return nil, merr.NotFound("site page not found")
	}
	locale = normalizeLocale(locale)
	fallback := false
	page, err := pages.pages.GetPublishedSitePage(ctx, slug, locale)
	if err != nil {
		if !merr.IsNotFound(err) {
			return nil, pages.access.Wrap(err)
		}
		other := otherLocale(locale)
		page, err = pages.pages.GetPublishedSitePage(ctx, slug, other)
		if err != nil {
			if merr.IsNotFound(err) {
				// 没写过和已下架用同一条文案，不对外泄露草稿是否存在。
				return nil, merr.NotFound("site page not found")
			}
			return nil, pages.access.Wrap(err)
		}
		locale = other
		fallback = true
	}
	return publicSitePageOf(*page, locale, fallback), nil
}

// PublicNav 读取公开导航：已发布且展示在导航里的页面，按导航顺序排列。
func (pages *Pages) PublicNav(ctx context.Context, locale string) ([]bo.SiteNavItem, error) {
	items, err := pages.pages.ListPublishedSiteNav(ctx, normalizeLocale(locale))
	if err != nil {
		return nil, pages.access.Wrap(err)
	}
	return items, nil
}

// List 分页列出页面，需要 system:site:read。
func (pages *Pages) List(ctx context.Context, filter repository.SitePageFilter) ([]bo.SitePage, int, int, int, error) {
	if _, err := pages.access.RequireSystem(ctx, perm.SystemSiteRead); err != nil {
		return nil, 0, 0, 0, err
	}
	items, total, p, s, err := pages.pages.ListSitePages(ctx, filter)
	if err != nil {
		return nil, 0, 0, 0, pages.access.Wrap(err)
	}
	return items, total, p, s, nil
}

// Get 读取页面详情，需要 system:site:read。
func (pages *Pages) Get(ctx context.Context, uid string) (*bo.SitePage, error) {
	if _, err := pages.access.RequireSystem(ctx, perm.SystemSiteRead); err != nil {
		return nil, err
	}
	item, err := pages.pages.GetSitePage(ctx, uid)
	if err != nil {
		return nil, pages.access.Wrap(err)
	}
	return item, nil
}

// Create 新建页面，写入草稿且状态为 DRAFT，需要 system:site:write。
// (slug, locale) 未删除记录上唯一，重复直接 409，不静默改成别的地址。
func (pages *Pages) Create(ctx context.Context, in bo.SitePage) (*bo.SitePage, error) {
	if _, err := pages.access.RequireSystem(ctx, perm.SystemSiteWrite); err != nil {
		return nil, err
	}
	in.Slug = strings.ToLower(strings.TrimSpace(in.Slug))
	in.Locale = normalizeLocale(in.Locale)
	in.Draft = normalizePageContent(in.Draft)
	if !sitePageSlugPattern.MatchString(in.Slug) {
		return nil, merr.Invalid("invalid slug")
	}
	if in.SortOrder < 0 || in.SortOrder > 999 {
		return nil, merr.Invalid("sort order must be between 0 and 999")
	}
	if err := validatePageContent(in.Draft); err != nil {
		return nil, err
	}
	existing, err := pages.pages.GetSitePageBySlug(ctx, in.Slug, in.Locale)
	if err != nil && !merr.IsNotFound(err) {
		return nil, pages.access.Wrap(err)
	}
	if existing != nil {
		return nil, merr.Conflict("site page slug already exists")
	}
	item, err := pages.pages.CreateSitePage(ctx, in)
	if err != nil {
		return nil, pages.access.Wrap(err)
	}
	return item, nil
}

// UpdateDraft 保存草稿，不改变发布状态，需要 system:site:write。
// expectedUpdatedAt 非空时做乐观锁：别人先保存过就返回 409，不静默覆盖。
func (pages *Pages) UpdateDraft(ctx context.Context, in bo.SitePage, expectedUpdatedAt *time.Time) (*bo.SitePage, error) {
	if _, err := pages.access.RequireSystem(ctx, perm.SystemSiteWrite); err != nil {
		return nil, err
	}
	in.Draft = normalizePageContent(in.Draft)
	if strings.TrimSpace(in.Draft.Title) == "" {
		return nil, merr.Invalid("site page title is required")
	}
	if err := validatePageContent(in.Draft); err != nil {
		return nil, err
	}
	item, err := pages.pages.UpdateSitePageDraft(ctx, in, expectedUpdatedAt)
	if err != nil {
		return nil, pages.access.Wrap(err)
	}
	return item, nil
}

// Publish 把草稿复制成发布快照，需要 system:site:write。
// 发布之后官网立即读到新内容；再发布一次就是覆盖旧快照。
func (pages *Pages) Publish(ctx context.Context, uid string) (*bo.SitePage, error) {
	account, err := pages.access.RequireSystem(ctx, perm.SystemSiteWrite)
	if err != nil {
		return nil, err
	}
	current, err := pages.pages.GetSitePage(ctx, uid)
	if err != nil {
		return nil, pages.access.Wrap(err)
	}
	if strings.TrimSpace(current.Draft.Title) == "" {
		return nil, merr.Invalid("site page title is required")
	}
	item, err := pages.pages.PublishSitePage(ctx, uid, account.UID)
	if err != nil {
		return nil, pages.access.Wrap(err)
	}
	return item, nil
}

// Unpublish 把页面置为已下架，需要 system:site:write。发布快照保留，可以重新发布。
// 首页不允许下架：它是站点入口。
func (pages *Pages) Unpublish(ctx context.Context, uid string) (*bo.SitePage, error) {
	if _, err := pages.access.RequireSystem(ctx, perm.SystemSiteWrite); err != nil {
		return nil, err
	}
	current, err := pages.pages.GetSitePage(ctx, uid)
	if err != nil {
		return nil, pages.access.Wrap(err)
	}
	if current.Slug == homeSlug {
		return nil, merr.Invalid("home page cannot be unpublished or deleted")
	}
	item, err := pages.pages.UnpublishSitePage(ctx, uid)
	if err != nil {
		return nil, pages.access.Wrap(err)
	}
	return item, nil
}

// Delete 软删页面，需要 system:site:write。首页不允许删除。
func (pages *Pages) Delete(ctx context.Context, uid string) error {
	if _, err := pages.access.RequireSystem(ctx, perm.SystemSiteWrite); err != nil {
		return err
	}
	current, err := pages.pages.GetSitePage(ctx, uid)
	if err != nil {
		return pages.access.Wrap(err)
	}
	if current.Slug == homeSlug {
		return merr.Invalid("home page cannot be unpublished or deleted")
	}
	if err := pages.pages.DeleteSitePage(ctx, uid); err != nil {
		return pages.access.Wrap(err)
	}
	return nil
}

// UpdateNav 修改导航展示与顺序，需要 system:site:write。
// 同一语言内导航顺序不能重复，否则导航会出现不确定的先后。
func (pages *Pages) UpdateNav(ctx context.Context, uid string, showInNav bool, sortOrder int) (*bo.SitePage, error) {
	if _, err := pages.access.RequireSystem(ctx, perm.SystemSiteWrite); err != nil {
		return nil, err
	}
	if sortOrder < 0 || sortOrder > 999 {
		return nil, merr.Invalid("sort order must be between 0 and 999")
	}
	current, err := pages.pages.GetSitePage(ctx, uid)
	if err != nil {
		return nil, pages.access.Wrap(err)
	}
	items, _, _, _, err := pages.pages.ListSitePages(ctx, repository.SitePageFilter{
		Page: 1, PageSize: 100, Locale: current.Locale,
	})
	if err != nil {
		return nil, pages.access.Wrap(err)
	}
	for _, item := range items {
		if item.UID != uid && item.SortOrder == sortOrder {
			return nil, merr.Conflict("site page sort order already exists")
		}
	}
	updated, err := pages.pages.UpdateSitePageNav(ctx, uid, showInNav, sortOrder)
	if err != nil {
		return nil, pages.access.Wrap(err)
	}
	return updated, nil
}

// normalizeLocale 把语言收进支持范围，缺省按中文处理。
func normalizeLocale(locale string) string {
	locale = strings.ToLower(strings.TrimSpace(locale))
	for _, item := range siteLocales {
		if item == locale {
			return item
		}
	}
	return siteLocales[0]
}

// otherLocale 返回另一种语言，用于内容回落。
func otherLocale(locale string) string {
	for _, item := range siteLocales {
		if item != locale {
			return item
		}
	}
	return locale
}

// normalizePageContent 裁剪标题与摘要：它们会出现在官网页面的可见位置，
// 前后空格会直接显示出来。正文不裁剪，Markdown 的空行有意义。
func normalizePageContent(in bo.SitePageContent) bo.SitePageContent {
	in.Title = strings.TrimSpace(in.Title)
	in.Summary = strings.TrimSpace(in.Summary)
	in.CoverImageURL = strings.TrimSpace(in.CoverImageURL)
	in.SEOTitle = strings.TrimSpace(in.SEOTitle)
	in.SEODescription = strings.TrimSpace(in.SEODescription)
	return in
}

// validatePageContent 校验封面与 SEO 图片地址只接受 http(s)，
// 避免 `javascript:` 之类的 scheme 从后台一路进入公开页面。
func validatePageContent(content bo.SitePageContent) error {
	if content.CoverImageURL != "" && !isHTTPURL(content.CoverImageURL) {
		return merr.Invalid("cover image url is invalid")
	}
	return nil
}

func publicSitePageOf(page bo.SitePage, locale string, fallback bool) *bo.PublicSitePage {
	published := page.Published
	publishedAt := time.Time{}
	if page.PublishedAt != nil {
		publishedAt = *page.PublishedAt
	}
	return &bo.PublicSitePage{
		Slug:           page.Slug,
		Locale:         locale,
		Title:          published.Title,
		Summary:        published.Summary,
		BodyMD:         published.BodyMD,
		CoverImageURL:  published.CoverImageURL,
		SEOTitle:       published.SEOTitle,
		SEODescription: published.SEODescription,
		PublishedAt:    publishedAt,
		LocaleFallback: fallback,
	}
}
