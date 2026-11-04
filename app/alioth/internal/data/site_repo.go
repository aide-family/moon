package data

import (
	"context"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/authloginsetting"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/sitesetting"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// 站点设置与登录方式开关都是平台级单例，用固定 uid 定位同一行。
const (
	siteSettingsUID = "site"
	loginMethodsUID = "login"
)

func (r *siteRepo) SiteSettings(ctx context.Context) (*bo.SiteSettings, error) {
	row, err := r.client.SiteSetting.Query().Where(sitesetting.UIDEQ(siteSettingsUID)).Only(ctx)
	if err == nil {
		return siteSettingsOf(row), nil
	}
	if !ent.IsNotFound(err) {
		return nil, r.fail(err)
	}
	created, err := r.client.SiteSetting.Create().SetUID(siteSettingsUID).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return siteSettingsOf(created), nil
}

func (r *siteRepo) SaveSiteSettings(ctx context.Context, in bo.SiteSettings) (*bo.SiteSettings, error) {
	if _, err := r.SiteSettings(ctx); err != nil {
		return nil, err
	}
	_, err := r.client.SiteSetting.Update().
		Where(sitesetting.UIDEQ(siteSettingsUID)).
		SetSiteName(in.SiteName).
		SetLogoURL(in.LogoURL).
		SetFaviconURL(in.FaviconURL).
		SetContactEmail(in.ContactEmail).
		SetRepositoryURL(in.RepositoryURL).
		SetCopyrightText(in.CopyrightText).
		SetFilingIcpNumber(in.FilingICPNumber).
		SetFilingIcpURL(in.FilingICPURL).
		SetFilingPoliceNumber(in.FilingPoliceNumber).
		SetFilingPoliceURL(in.FilingPoliceURL).
		SetFilingExtraText(in.FilingExtraText).
		SetUpdatedBy(in.UpdatedBy).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return r.SiteSettings(ctx)
}

func (r *loginMethodRepo) LoginMethods(ctx context.Context) (*bo.LoginMethods, error) {
	row, err := r.client.AuthLoginSetting.Query().Where(authloginsetting.UIDEQ(loginMethodsUID)).Only(ctx)
	if err == nil {
		return loginMethodsOf(row), nil
	}
	if !ent.IsNotFound(err) {
		return nil, r.fail(err)
	}
	created, err := r.client.AuthLoginSetting.Create().SetUID(loginMethodsUID).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return loginMethodsOf(created), nil
}

func (r *loginMethodRepo) SaveLoginMethods(ctx context.Context, in bo.LoginMethods) (*bo.LoginMethods, error) {
	if _, err := r.LoginMethods(ctx); err != nil {
		return nil, err
	}
	_, err := r.client.AuthLoginSetting.Update().
		Where(authloginsetting.UIDEQ(loginMethodsUID)).
		SetEmailEnabled(in.EmailEnabled).
		SetOauth2Enabled(in.OAuth2Enabled).
		SetUpdatedBy(in.UpdatedBy).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return r.LoginMethods(ctx)
}

func siteSettingsOf(row *ent.SiteSetting) *bo.SiteSettings {
	return &bo.SiteSettings{
		SiteName:           row.SiteName,
		LogoURL:            row.LogoURL,
		FaviconURL:         row.FaviconURL,
		ContactEmail:       row.ContactEmail,
		RepositoryURL:      row.RepositoryURL,
		CopyrightText:      row.CopyrightText,
		FilingICPNumber:    row.FilingIcpNumber,
		FilingICPURL:       row.FilingIcpURL,
		FilingPoliceNumber: row.FilingPoliceNumber,
		FilingPoliceURL:    row.FilingPoliceURL,
		FilingExtraText:    row.FilingExtraText,
		UpdatedBy:          row.UpdatedBy,
		UpdatedAt:          row.UpdatedAt,
	}
}

func loginMethodsOf(row *ent.AuthLoginSetting) *bo.LoginMethods {
	return &bo.LoginMethods{
		EmailEnabled:  row.EmailEnabled,
		OAuth2Enabled: row.Oauth2Enabled,
		UpdatedBy:     row.UpdatedBy,
		UpdatedAt:     row.UpdatedAt,
	}
}
