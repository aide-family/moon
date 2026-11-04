package data

import (
	"context"
	"strings"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/sitefeature"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/sitehomecontent"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
)

func (r *siteHomeRepo) HomeContent(ctx context.Context, locale string) (*bo.SiteHomeContent, error) {
	row, err := r.client.SiteHomeContent.Query().
		Where(sitehomecontent.LocaleEQ(locale)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return &bo.SiteHomeContent{Locale: locale}, nil
		}
		return nil, r.fail(err)
	}
	return siteHomeContentOf(row), nil
}

func (r *siteHomeRepo) SaveHomeContent(ctx context.Context, in bo.SiteHomeContent) (*bo.SiteHomeContent, error) {
	existing, err := r.client.SiteHomeContent.Query().
		Where(sitehomecontent.LocaleEQ(in.Locale)).
		Only(ctx)
	switch {
	case ent.IsNotFound(err):
		_, err = r.client.SiteHomeContent.Create().
			SetUID(id.New()).
			SetLocale(in.Locale).
			SetHeroTitle(in.HeroTitle).
			SetHeroSubtitle(in.HeroSubtitle).
			SetHeroImageURL(in.HeroImageURL).
			SetCtaPrimaryLabel(in.CtaPrimaryLabel).
			SetCtaSecondaryLabel(in.CtaSecondaryLabel).
			SetCtaSecondaryHref(in.CtaSecondaryHref).
			SetFeaturesTitle(in.FeaturesTitle).
			SetFeaturesSubtitle(in.FeaturesSubtitle).
			SetHowTitle(in.HowTitle).
			SetHowBody(in.HowBody).
			SetPermTitle(in.PermTitle).
			SetPermBody(in.PermBody).
			SetDeployTitle(in.DeployTitle).
			SetDeployBody(in.DeployBody).
			SetDocTitle(in.DocTitle).
			SetClosingTitle(in.ClosingTitle).
			SetClosingDescription(in.ClosingDescription).
			SetClosingButtonLabel(in.ClosingButtonLabel).
			SetUpdatedBy(in.UpdatedBy).
			Save(ctx)
	case err == nil:
		_, err = r.client.SiteHomeContent.Update().
			Where(sitehomecontent.ID(existing.ID)).
			SetHeroTitle(in.HeroTitle).
			SetHeroSubtitle(in.HeroSubtitle).
			SetHeroImageURL(in.HeroImageURL).
			SetCtaPrimaryLabel(in.CtaPrimaryLabel).
			SetCtaSecondaryLabel(in.CtaSecondaryLabel).
			SetCtaSecondaryHref(in.CtaSecondaryHref).
			SetFeaturesTitle(in.FeaturesTitle).
			SetFeaturesSubtitle(in.FeaturesSubtitle).
			SetHowTitle(in.HowTitle).
			SetHowBody(in.HowBody).
			SetPermTitle(in.PermTitle).
			SetPermBody(in.PermBody).
			SetDeployTitle(in.DeployTitle).
			SetDeployBody(in.DeployBody).
			SetDocTitle(in.DocTitle).
			SetClosingTitle(in.ClosingTitle).
			SetClosingDescription(in.ClosingDescription).
			SetClosingButtonLabel(in.ClosingButtonLabel).
			SetUpdatedBy(in.UpdatedBy).
			Save(ctx)
	}
	if err != nil {
		return nil, r.fail(err)
	}
	return r.HomeContent(ctx, in.Locale)
}

func (r *siteHomeRepo) SiteFeatures(ctx context.Context, locale string) ([]bo.SiteFeature, error) {
	rows, err := r.client.SiteFeature.Query().
		Where(sitefeature.LocaleEQ(locale)).
		Order(ent.Asc(sitefeature.FieldSortOrder)).
		All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]bo.SiteFeature, 0, len(rows))
	for _, row := range rows {
		items = append(items, siteFeatureOf(row))
	}
	return items, nil
}

// ReplaceSiteFeatures 在一个事务里删掉旧列表再写入新列表：
// 能力卡片是整块版式的一部分，不能出现「删了一半」的中间态。
func (r *siteHomeRepo) ReplaceSiteFeatures(ctx context.Context, locale string, items []bo.SiteFeature) ([]bo.SiteFeature, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if _, err := tx.SiteFeature.Delete().Where(sitefeature.LocaleEQ(locale)).Exec(ctx); err != nil {
		_ = tx.Rollback()
		return nil, r.fail(err)
	}
	for index, item := range items {
		if _, err := tx.SiteFeature.Create().
			SetUID(id.New()).
			SetLocale(locale).
			SetSortOrder(index).
			SetIconKey(item.IconKey).
			SetTitle(item.Title).
			SetDescription(item.Description).
			Save(ctx); err != nil {
			_ = tx.Rollback()
			return nil, r.fail(err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.SiteFeatures(ctx, locale)
}

func siteHomeContentOf(row *ent.SiteHomeContent) *bo.SiteHomeContent {
	return &bo.SiteHomeContent{
		Locale:             row.Locale,
		HeroTitle:          row.HeroTitle,
		HeroSubtitle:       row.HeroSubtitle,
		HeroImageURL:       row.HeroImageURL,
		CtaPrimaryLabel:    row.CtaPrimaryLabel,
		CtaSecondaryLabel:  row.CtaSecondaryLabel,
		CtaSecondaryHref:   row.CtaSecondaryHref,
		FeaturesTitle:      row.FeaturesTitle,
		FeaturesSubtitle:   row.FeaturesSubtitle,
		HowTitle:           row.HowTitle,
		HowBody:            row.HowBody,
		PermTitle:          row.PermTitle,
		PermBody:           row.PermBody,
		DeployTitle:        row.DeployTitle,
		DeployBody:         row.DeployBody,
		DocTitle:           row.DocTitle,
		ClosingTitle:       row.ClosingTitle,
		ClosingDescription: row.ClosingDescription,
		ClosingButtonLabel: row.ClosingButtonLabel,
		UpdatedBy:          row.UpdatedBy,
		UpdatedAt:          row.UpdatedAt,
	}
}

func siteFeatureOf(row *ent.SiteFeature) bo.SiteFeature {
	return bo.SiteFeature{
		UID:         row.UID,
		Locale:      row.Locale,
		SortOrder:   row.SortOrder,
		IconKey:     strings.TrimSpace(row.IconKey),
		Title:       row.Title,
		Description: row.Description,
	}
}
