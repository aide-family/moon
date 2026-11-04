package service

import (
	"context"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/site"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// SiteService 把官网配置接口转成 biz 入参，不做权限判断。
type SiteService struct {
	v1.UnimplementedSiteServiceServer
	settings *site.Settings
}

func NewSiteService(settings *site.Settings) *SiteService {
	return &SiteService{settings: settings}
}

func (s *SiteService) GetSiteConfig(ctx context.Context, _ *v1.GetSiteConfigRequest) (*v1.GetSiteConfigResponse, error) {
	config, err := s.settings.Public(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetSiteConfigResponse{Site: siteSettingsItem(config.Site), LoginMethods: loginMethodsItem(config.Login)}, nil
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
