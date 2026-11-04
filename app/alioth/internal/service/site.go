package service

import (
	"context"
	"strings"
	"time"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/site"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/oauthproviders"
)

// SiteService 把官网配置接口转成 biz 入参，不做权限判断。
type SiteService struct {
	v1.UnimplementedSiteServiceServer
	settings *site.Settings
	pages    *site.Pages
}

func NewSiteService(settings *site.Settings, pages *site.Pages) *SiteService {
	return &SiteService{settings: settings, pages: pages}
}

func (s *SiteService) GetSiteConfig(ctx context.Context, req *v1.GetSiteConfigRequest) (*v1.GetSiteConfigResponse, error) {
	config, err := s.settings.Public(ctx, req.GetLocale())
	if err != nil {
		return nil, err
	}
	features := make([]*v1.SiteFeatureItem, 0, len(config.Home.Features))
	for _, item := range config.Home.Features {
		features = append(features, siteFeatureItem(item))
	}
	return &v1.GetSiteConfigResponse{
		Site:         siteSettingsItem(config.Site),
		LoginMethods: loginMethodsItem(config.Login),
		Home:         siteHomeContentItem(config.Home.Content),
		Features:     features,
	}, nil
}

func (s *SiteService) GetSiteSettings(ctx context.Context, _ *v1.GetSiteSettingsRequest) (*v1.GetSiteSettingsResponse, error) {
	item, err := s.settings.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetSiteSettingsResponse{Site: siteSettingsItem(*item)}, nil
}

func (s *SiteService) UpdateSiteSettings(ctx context.Context, req *v1.UpdateSiteSettingsRequest) (*v1.UpdateSiteSettingsResponse, error) {
	item, err := s.settings.Update(ctx, bo.SiteSettings{
		SiteName:           req.GetSiteName(),
		LogoURL:            req.GetLogoUrl(),
		FaviconURL:         req.GetFaviconUrl(),
		ContactEmail:       req.GetContactEmail(),
		RepositoryURL:      req.GetRepositoryUrl(),
		CopyrightText:      req.GetCopyrightText(),
		FilingICPNumber:    req.GetFilingIcpNumber(),
		FilingICPURL:       req.GetFilingIcpUrl(),
		FilingPoliceNumber: req.GetFilingPoliceNumber(),
		FilingPoliceURL:    req.GetFilingPoliceUrl(),
		FilingExtraText:    req.GetFilingExtraText(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateSiteSettingsResponse{Site: siteSettingsItem(*item)}, nil
}

func (s *SiteService) GetLoginMethods(ctx context.Context, _ *v1.GetLoginMethodsRequest) (*v1.GetLoginMethodsResponse, error) {
	item, err := s.settings.GetLoginMethods(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetLoginMethodsResponse{LoginMethods: loginMethodsItem(*item)}, nil
}

func (s *SiteService) UpdateLoginMethods(ctx context.Context, req *v1.UpdateLoginMethodsRequest) (*v1.UpdateLoginMethodsResponse, error) {
	item, err := s.settings.UpdateLoginMethods(ctx, bo.LoginMethods{
		EmailEnabled:  req.GetEmailEnabled(),
		OAuth2Enabled: req.GetOauth2Enabled(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateLoginMethodsResponse{LoginMethods: loginMethodsItem(*item)}, nil
}

func siteSettingsItem(item bo.SiteSettings) *v1.SiteSettingsItem {
	return &v1.SiteSettingsItem{
		SiteName:           item.SiteName,
		LogoUrl:            item.LogoURL,
		FaviconUrl:         item.FaviconURL,
		ContactEmail:       item.ContactEmail,
		RepositoryUrl:      item.RepositoryURL,
		CopyrightText:      item.CopyrightText,
		FilingIcpNumber:    item.FilingICPNumber,
		FilingIcpUrl:       item.FilingICPURL,
		FilingPoliceNumber: item.FilingPoliceNumber,
		FilingPoliceUrl:    item.FilingPoliceURL,
		FilingExtraText:    item.FilingExtraText,
		UpdatedBy:          item.UpdatedBy,
		UpdatedAt:          formatTime(item.UpdatedAt),
	}
}

func loginMethodsItem(item bo.LoginMethods) *v1.LoginMethodsItem {
	return &v1.LoginMethodsItem{
		EmailEnabled:  item.EmailEnabled,
		Oauth2Enabled: item.OAuth2Enabled,
		UpdatedBy:     item.UpdatedBy,
		UpdatedAt:     formatTime(item.UpdatedAt),
	}
}

func (s *SiteService) GetPublicSitePage(ctx context.Context, req *v1.GetPublicSitePageRequest) (*v1.GetPublicSitePageResponse, error) {
	item, err := s.pages.PublicPage(ctx, req.GetSlug(), req.GetLocale())
	if err != nil {
		return nil, err
	}
	return &v1.GetPublicSitePageResponse{Page: publicSitePageItem(*item)}, nil
}

func (s *SiteService) ListPublicSiteNav(ctx context.Context, req *v1.ListPublicSiteNavRequest) (*v1.ListPublicSiteNavResponse, error) {
	items, err := s.pages.PublicNav(ctx, req.GetLocale())
	if err != nil {
		return nil, err
	}
	out := make([]*v1.SiteNavItem, 0, len(items))
	for _, item := range items {
		out = append(out, &v1.SiteNavItem{Slug: item.Slug, Title: item.Title, SortOrder: int32(item.SortOrder)})
	}
	return &v1.ListPublicSiteNavResponse{Items: out}, nil
}

func (s *SiteService) ListSitePages(ctx context.Context, req *v1.ListSitePagesRequest) (*v1.ListSitePagesResponse, error) {
	items, total, p, size, err := s.pages.List(ctx, repository.SitePageFilter{
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
		Keyword:  req.GetKeyword(),
		State:    req.GetState(),
		Locale:   req.GetLocale(),
	})
	if err != nil {
		return nil, err
	}
	out := make([]*v1.SitePageItem, 0, len(items))
	for _, item := range items {
		out = append(out, sitePageItem(item))
	}
	return &v1.ListSitePagesResponse{Items: out, Total: int32(total), Page: int32(p), PageSize: int32(size)}, nil
}

func (s *SiteService) GetSitePage(ctx context.Context, req *v1.GetSitePageRequest) (*v1.GetSitePageResponse, error) {
	item, err := s.pages.Get(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetSitePageResponse{Page: sitePageDetail(*item)}, nil
}

func (s *SiteService) CreateSitePage(ctx context.Context, req *v1.CreateSitePageRequest) (*v1.CreateSitePageResponse, error) {
	item, err := s.pages.Create(ctx, bo.SitePage{
		Slug:   req.GetSlug(),
		Locale: req.GetLocale(),
		Draft: bo.SitePageContent{
			Title:          req.GetTitle(),
			Summary:        req.GetSummary(),
			BodyMD:         req.GetBodyMd(),
			CoverImageURL:  req.GetCoverImageUrl(),
			SEOTitle:       req.GetSeoTitle(),
			SEODescription: req.GetSeoDescription(),
		},
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateSitePageResponse{Page: sitePageDetail(*item)}, nil
}

func (s *SiteService) UpdateSitePage(ctx context.Context, req *v1.UpdateSitePageRequest) (*v1.UpdateSitePageResponse, error) {
	expected, err := parseOptionalTime(req.GetUpdatedAt())
	if err != nil {
		return nil, err
	}
	item, err := s.pages.UpdateDraft(ctx, bo.SitePage{
		UID: req.GetUid(),
		Draft: bo.SitePageContent{
			Title:          req.GetTitle(),
			Summary:        req.GetSummary(),
			BodyMD:         req.GetBodyMd(),
			CoverImageURL:  req.GetCoverImageUrl(),
			SEOTitle:       req.GetSeoTitle(),
			SEODescription: req.GetSeoDescription(),
		},
	}, expected)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateSitePageResponse{Page: sitePageDetail(*item)}, nil
}

func (s *SiteService) PublishSitePage(ctx context.Context, req *v1.PublishSitePageRequest) (*v1.PublishSitePageResponse, error) {
	item, err := s.pages.Publish(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.PublishSitePageResponse{Page: sitePageDetail(*item)}, nil
}

func (s *SiteService) UnpublishSitePage(ctx context.Context, req *v1.UnpublishSitePageRequest) (*v1.UnpublishSitePageResponse, error) {
	item, err := s.pages.Unpublish(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.UnpublishSitePageResponse{Page: sitePageDetail(*item)}, nil
}

func (s *SiteService) DeleteSitePage(ctx context.Context, req *v1.DeleteSitePageRequest) (*v1.DeleteSitePageResponse, error) {
	if err := s.pages.Delete(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteSitePageResponse{}, nil
}

func (s *SiteService) UpdateSitePageNav(ctx context.Context, req *v1.UpdateSitePageNavRequest) (*v1.UpdateSitePageNavResponse, error) {
	item, err := s.pages.UpdateNav(ctx, req.GetUid(), req.GetShowInNav(), int(req.GetSortOrder()))
	if err != nil {
		return nil, err
	}
	return &v1.UpdateSitePageNavResponse{Page: sitePageDetail(*item)}, nil
}

// parseOptionalTime 解析客户端带回来的更新时间；留空表示不做乐观锁检查。
// 解析失败按非法请求处理，不能悄悄当成「不做检查」，那会让并发覆盖变成默认行为。
func parseOptionalTime(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, merr.Invalid("invalid updated at")
	}
	return &parsed, nil
}

func sitePageItem(item bo.SitePage) *v1.SitePageItem {
	return &v1.SitePageItem{
		Uid:                   item.UID,
		Slug:                  item.Slug,
		Locale:                item.Locale,
		State:                 item.State,
		ShowInNav:             item.ShowInNav,
		SortOrder:             int32(item.SortOrder),
		Title:                 item.Draft.Title,
		Summary:               item.Draft.Summary,
		HasUnpublishedChanges: item.HasUnpublishedChanges(),
		UpdatedAt:             formatTime(item.UpdatedAt),
		PublishedAt:           formatNullableTime(item.PublishedAt),
	}
}

func sitePageContent(content bo.SitePageContent) *v1.SitePageContent {
	return &v1.SitePageContent{
		Title:          content.Title,
		Summary:        content.Summary,
		BodyMd:         content.BodyMD,
		CoverImageUrl:  content.CoverImageURL,
		SeoTitle:       content.SEOTitle,
		SeoDescription: content.SEODescription,
	}
}

func sitePageDetail(item bo.SitePage) *v1.SitePageDetail {
	return &v1.SitePageDetail{
		Item:        sitePageItem(item),
		Draft:       sitePageContent(item.Draft),
		Published:   sitePageContent(item.Published),
		PublishedBy: item.PublishedBy,
	}
}

func publicSitePageItem(item bo.PublicSitePage) *v1.PublicSitePage {
	return &v1.PublicSitePage{
		Slug:           item.Slug,
		Locale:         item.Locale,
		Title:          item.Title,
		Summary:        item.Summary,
		BodyMd:         item.BodyMD,
		CoverImageUrl:  item.CoverImageURL,
		SeoTitle:       item.SEOTitle,
		SeoDescription: item.SEODescription,
		PublishedAt:    formatTime(item.PublishedAt),
		LocaleFallback: item.LocaleFallback,
	}
}

func formatNullableTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return formatTime(*value)
}

func (s *SiteService) GetEmailSettings(ctx context.Context, _ *v1.GetEmailSettingsRequest) (*v1.GetEmailSettingsResponse, error) {
	item, err := s.settings.EmailSettings(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetEmailSettingsResponse{Email: emailSettingsItem(*item)}, nil
}

func (s *SiteService) UpdateEmailSettings(ctx context.Context, req *v1.UpdateEmailSettingsRequest) (*v1.UpdateEmailSettingsResponse, error) {
	item, err := s.settings.UpdateEmailSettings(ctx, bo.EmailSettings{
		Enabled:  req.GetEnabled(),
		Host:     req.GetHost(),
		Port:     int(req.GetPort()),
		Username: req.GetUsername(),
		Password: req.GetPassword(),
		From:     req.GetFrom(),
	}, req.GetClearPassword())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateEmailSettingsResponse{Email: emailSettingsItem(*item)}, nil
}

func (s *SiteService) ListOAuthProviderSettings(ctx context.Context, _ *v1.ListOAuthProviderSettingsRequest) (*v1.ListOAuthProviderSettingsResponse, error) {
	items, err := s.settings.OAuthProviders(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*v1.OAuthProviderSettingItem, 0, len(items))
	for _, item := range items {
		out = append(out, oauthProviderSettingItem(item))
	}
	return &v1.ListOAuthProviderSettingsResponse{Items: out}, nil
}

func (s *SiteService) UpdateOAuthProviderSetting(ctx context.Context, req *v1.UpdateOAuthProviderSettingRequest) (*v1.UpdateOAuthProviderSettingResponse, error) {
	item, err := s.settings.UpdateOAuthProvider(ctx, bo.OAuthProviderSetting{
		Provider:     req.GetProvider(),
		Enabled:      req.GetEnabled(),
		ClientID:     req.GetClientId(),
		ClientSecret: req.GetClientSecret(),
	}, req.GetClearSecret())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateOAuthProviderSettingResponse{Item: oauthProviderSettingItem(*item)}, nil
}

func emailSettingsItem(item bo.EmailSettings) *v1.EmailSettingsItem {
	return &v1.EmailSettingsItem{
		Enabled:     item.Enabled,
		Host:        item.Host,
		Port:        int32(item.Port),
		Username:    item.Username,
		From:        item.From,
		PasswordSet: item.PasswordSet,
		UpdatedBy:   item.UpdatedBy,
		UpdatedAt:   formatTime(item.UpdatedAt),
		Source:      item.Source,
	}
}

func oauthProviderSettingItem(item bo.OAuthProviderSetting) *v1.OAuthProviderSettingItem {
	name := item.Provider
	if provider, ok := oauthproviders.Lookup(item.Provider); ok {
		name = provider.Name
	}
	return &v1.OAuthProviderSettingItem{
		Provider:  item.Provider,
		Name:      name,
		Enabled:   item.Enabled,
		ClientId:  item.ClientID,
		SecretSet: item.SecretSet,
		UpdatedBy: item.UpdatedBy,
		UpdatedAt: formatTime(item.UpdatedAt),
		Source:    item.Source,
	}
}

func (s *SiteService) GetSiteHome(ctx context.Context, req *v1.GetSiteHomeRequest) (*v1.GetSiteHomeResponse, error) {
	home, err := s.settings.Home(ctx, req.GetLocale())
	if err != nil {
		return nil, err
	}
	return &v1.GetSiteHomeResponse{
		Home:     siteHomeContentItem(home.Content),
		Features: siteFeatureItems(home.Features),
	}, nil
}

func (s *SiteService) UpdateSiteHome(ctx context.Context, req *v1.UpdateSiteHomeRequest) (*v1.UpdateSiteHomeResponse, error) {
	home := req.GetHome()
	content := bo.SiteHomeContent{}
	if home != nil {
		content = bo.SiteHomeContent{
			HeroTitle:          home.GetHeroTitle(),
			HeroSubtitle:       home.GetHeroSubtitle(),
			HeroImageURL:       home.GetHeroImageUrl(),
			CtaPrimaryLabel:    home.GetCtaPrimaryLabel(),
			CtaSecondaryLabel:  home.GetCtaSecondaryLabel(),
			CtaSecondaryHref:   home.GetCtaSecondaryHref(),
			FeaturesTitle:      home.GetFeaturesTitle(),
			FeaturesSubtitle:   home.GetFeaturesSubtitle(),
			HowTitle:           home.GetHowTitle(),
			HowBody:            home.GetHowBody(),
			PermTitle:          home.GetPermTitle(),
			PermBody:           home.GetPermBody(),
			DeployTitle:        home.GetDeployTitle(),
			DeployBody:         home.GetDeployBody(),
			DocTitle:           home.GetDocTitle(),
			ClosingTitle:       home.GetClosingTitle(),
			ClosingDescription: home.GetClosingDescription(),
			ClosingButtonLabel: home.GetClosingButtonLabel(),
		}
	}
	features := make([]bo.SiteFeature, 0, len(req.GetFeatures()))
	for _, item := range req.GetFeatures() {
		features = append(features, bo.SiteFeature{
			IconKey:     item.GetIconKey(),
			Title:       item.GetTitle(),
			Description: item.GetDescription(),
		})
	}
	saved, err := s.settings.UpdateHome(ctx, req.GetLocale(), content, features)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateSiteHomeResponse{
		Home:     siteHomeContentItem(saved.Content),
		Features: siteFeatureItems(saved.Features),
	}, nil
}

func siteHomeContentItem(item bo.SiteHomeContent) *v1.SiteHomeContentItem {
	return &v1.SiteHomeContentItem{
		Locale:             item.Locale,
		HeroTitle:          item.HeroTitle,
		HeroSubtitle:       item.HeroSubtitle,
		HeroImageUrl:       item.HeroImageURL,
		CtaPrimaryLabel:    item.CtaPrimaryLabel,
		CtaSecondaryLabel:  item.CtaSecondaryLabel,
		CtaSecondaryHref:   item.CtaSecondaryHref,
		FeaturesTitle:      item.FeaturesTitle,
		FeaturesSubtitle:   item.FeaturesSubtitle,
		HowTitle:           item.HowTitle,
		HowBody:            item.HowBody,
		PermTitle:          item.PermTitle,
		PermBody:           item.PermBody,
		DeployTitle:        item.DeployTitle,
		DeployBody:         item.DeployBody,
		DocTitle:           item.DocTitle,
		ClosingTitle:       item.ClosingTitle,
		ClosingDescription: item.ClosingDescription,
		ClosingButtonLabel: item.ClosingButtonLabel,
		UpdatedBy:          item.UpdatedBy,
		UpdatedAt:          formatTime(item.UpdatedAt),
	}
}

func siteFeatureItem(item bo.SiteFeature) *v1.SiteFeatureItem {
	return &v1.SiteFeatureItem{
		Uid:         item.UID,
		IconKey:     item.IconKey,
		Title:       item.Title,
		Description: item.Description,
		SortOrder:   int32(item.SortOrder),
	}
}

func siteFeatureItems(items []bo.SiteFeature) []*v1.SiteFeatureItem {
	out := make([]*v1.SiteFeatureItem, 0, len(items))
	for _, item := range items {
		out = append(out, siteFeatureItem(item))
	}
	return out
}
