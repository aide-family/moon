package data

import (
	"context"
	"strings"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/sitepage"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/app/alioth/pkg/page"
	"github.com/aide-family/moon/pkg/merr"
)

func (r *sitePageRepo) ListSitePages(ctx context.Context, filter repository.SitePageFilter) ([]bo.SitePage, int, int, int, error) {
	p, s := page.Normalize(filter.Page, filter.PageSize)
	query := r.client.SitePage.Query().Where(sitepage.DeletedAtIsNil())
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		query = query.Where(sitepage.Or(
			sitepage.DraftTitleContainsFold(keyword),
			sitepage.SlugContainsFold(keyword),
		))
	}
	if filter.State != "" {
		query = query.Where(sitepage.StateEQ(sitepage.State(filter.State)))
	}
	if filter.Locale != "" {
		query = query.Where(sitepage.LocaleEQ(filter.Locale))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, 0, 0, r.fail(err)
	}
	rows, err := query.
		Order(ent.Desc(sitepage.FieldUpdatedAt)).
		Offset((p - 1) * s).
		Limit(s).
		All(ctx)
	if err != nil {
		return nil, 0, 0, 0, r.fail(err)
	}
	items := make([]bo.SitePage, 0, len(rows))
	for _, row := range rows {
		items = append(items, sitePageOf(row))
	}
	return items, total, p, s, nil
}

func (r *sitePageRepo) GetSitePage(ctx context.Context, uid string) (*bo.SitePage, error) {
	row, err := r.client.SitePage.Query().
		Where(sitepage.UIDEQ(uid), sitepage.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, merr.NotFound("site page not found")
		}
		return nil, r.fail(err)
	}
	page := sitePageOf(row)
	return &page, nil
}

func (r *sitePageRepo) GetSitePageBySlug(ctx context.Context, slug, locale string) (*bo.SitePage, error) {
	row, err := r.client.SitePage.Query().
		Where(sitepage.SlugEQ(slug), sitepage.LocaleEQ(locale), sitepage.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, merr.NotFound("site page not found")
		}
		return nil, r.fail(err)
	}
	page := sitePageOf(row)
	return &page, nil
}

func (r *sitePageRepo) CreateSitePage(ctx context.Context, in bo.SitePage) (*bo.SitePage, error) {
	row, err := r.client.SitePage.Create().
		SetUID(id.New()).
		SetSlug(in.Slug).
		SetLocale(in.Locale).
		SetState(sitepage.StateDRAFT).
		SetShowInNav(in.ShowInNav).
		SetSortOrder(in.SortOrder).
		SetDraftTitle(in.Draft.Title).
		SetDraftSummary(in.Draft.Summary).
		SetDraftBodyMd(in.Draft.BodyMD).
		SetDraftCoverImageURL(in.Draft.CoverImageURL).
		SetDraftSeoTitle(in.Draft.SEOTitle).
		SetDraftSeoDescription(in.Draft.SEODescription).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return r.GetSitePage(ctx, row.UID)
}

func (r *sitePageRepo) UpdateSitePageDraft(ctx context.Context, in bo.SitePage, expectedUpdatedAt *time.Time) (*bo.SitePage, error) {
	update := r.client.SitePage.Update().
		Where(sitepage.UIDEQ(in.UID), sitepage.DeletedAtIsNil())
	if expectedUpdatedAt != nil {
		// 乐观锁：客户端读到的更新时间已经变了，说明别人先保存过，不能静默覆盖。
		update = update.Where(sitepage.UpdatedAtEQ(*expectedUpdatedAt))
	}
	n, err := update.
		SetDraftTitle(in.Draft.Title).
		SetDraftSummary(in.Draft.Summary).
		SetDraftBodyMd(in.Draft.BodyMD).
		SetDraftCoverImageURL(in.Draft.CoverImageURL).
		SetDraftSeoTitle(in.Draft.SEOTitle).
		SetDraftSeoDescription(in.Draft.SEODescription).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		if expectedUpdatedAt != nil {
			return nil, merr.Conflict("site page was modified")
		}
		return nil, merr.NotFound("site page not found")
	}
	return r.GetSitePage(ctx, in.UID)
}

func (r *sitePageRepo) PublishSitePage(ctx context.Context, uid, publishedBy string) (*bo.SitePage, error) {
	page, err := r.GetSitePage(ctx, uid)
	if err != nil {
		return nil, err
	}
	_, err = r.client.SitePage.Update().
		Where(sitepage.UIDEQ(uid), sitepage.DeletedAtIsNil()).
		SetState(sitepage.StatePUBLISHED).
		SetPublishedTitle(page.Draft.Title).
		SetPublishedSummary(page.Draft.Summary).
		SetPublishedBodyMd(page.Draft.BodyMD).
		SetPublishedCoverImageURL(page.Draft.CoverImageURL).
		SetPublishedSeoTitle(page.Draft.SEOTitle).
		SetPublishedSeoDescription(page.Draft.SEODescription).
		SetPublishedAt(time.Now()).
		SetPublishedBy(publishedBy).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return r.GetSitePage(ctx, uid)
}

func (r *sitePageRepo) UnpublishSitePage(ctx context.Context, uid string) (*bo.SitePage, error) {
	n, err := r.client.SitePage.Update().
		Where(sitepage.UIDEQ(uid), sitepage.DeletedAtIsNil()).
		SetState(sitepage.StateARCHIVED).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("site page not found")
	}
	return r.GetSitePage(ctx, uid)
}

func (r *sitePageRepo) DeleteSitePage(ctx context.Context, uid string) error {
	n, err := r.client.SitePage.Update().
		Where(sitepage.UIDEQ(uid), sitepage.DeletedAtIsNil()).
		SetDeletedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("site page not found")
	}
	return nil
}

func (r *sitePageRepo) UpdateSitePageNav(ctx context.Context, uid string, showInNav bool, sortOrder int) (*bo.SitePage, error) {
	n, err := r.client.SitePage.Update().
		Where(sitepage.UIDEQ(uid), sitepage.DeletedAtIsNil()).
		SetShowInNav(showInNav).
		SetSortOrder(sortOrder).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("site page not found")
	}
	return r.GetSitePage(ctx, uid)
}

func (r *sitePageRepo) GetPublishedSitePage(ctx context.Context, slug, locale string) (*bo.SitePage, error) {
	row, err := r.client.SitePage.Query().
		Where(
			sitepage.SlugEQ(slug),
			sitepage.LocaleEQ(locale),
			sitepage.StateEQ(sitepage.StatePUBLISHED),
			sitepage.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, merr.NotFound("site page not found")
		}
		return nil, r.fail(err)
	}
	page := sitePageOf(row)
	return &page, nil
}

func (r *sitePageRepo) ListPublishedSiteNav(ctx context.Context, locale string) ([]bo.SiteNavItem, error) {
	rows, err := r.client.SitePage.Query().
		Where(
			sitepage.LocaleEQ(locale),
			sitepage.StateEQ(sitepage.StatePUBLISHED),
			sitepage.ShowInNavEQ(true),
			sitepage.DeletedAtIsNil(),
		).
		Order(ent.Asc(sitepage.FieldSortOrder), ent.Asc(sitepage.FieldSlug)).
		All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]bo.SiteNavItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, bo.SiteNavItem{
			Slug:      row.Slug,
			Title:     row.PublishedTitle,
			SortOrder: row.SortOrder,
		})
	}
	return items, nil
}

func sitePageOf(row *ent.SitePage) bo.SitePage {
	return bo.SitePage{
		UID:       row.UID,
		Slug:      row.Slug,
		Locale:    row.Locale,
		State:     string(row.State),
		ShowInNav: row.ShowInNav,
		SortOrder: row.SortOrder,
		Draft: bo.SitePageContent{
			Title:          row.DraftTitle,
			Summary:        row.DraftSummary,
			BodyMD:         row.DraftBodyMd,
			CoverImageURL:  row.DraftCoverImageURL,
			SEOTitle:       row.DraftSeoTitle,
			SEODescription: row.DraftSeoDescription,
		},
		Published: bo.SitePageContent{
			Title:          row.PublishedTitle,
			Summary:        row.PublishedSummary,
			BodyMD:         row.PublishedBodyMd,
			CoverImageURL:  row.PublishedCoverImageURL,
			SEOTitle:       row.PublishedSeoTitle,
			SEODescription: row.PublishedSeoDescription,
		},
		PublishedAt: row.PublishedAt,
		PublishedBy: row.PublishedBy,
		UpdatedAt:   row.UpdatedAt,
	}
}
