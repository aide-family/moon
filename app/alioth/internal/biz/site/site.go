package site

import (
	"context"
	"net/mail"
	"net/url"
	"strings"

	"github.com/google/wire"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

var ProviderSet = wire.NewSet(NewSettings, NewPages)

// Settings 维护官网的站点设置与登录方式开关。两者都是平台级单例：
// 站点设置决定官网页脚展示什么，登录方式开关决定登录页对外暴露哪些入口。
// 备案字段留空是合法状态，表示官网页脚不展示该行；填了就必须能作为链接打开。
type Settings struct {
	sites  repository.SiteRepo
	logins repository.LoginMethodRepo
	access access.Checker
}

func NewSettings(sites repository.SiteRepo, logins repository.LoginMethodRepo, checker access.Checker) *Settings {
	return &Settings{sites: sites, logins: logins, access: checker}
}

// Public 返回官网与登录页共用的公开配置，不需要登录与租户头。
// 登录页用其中的 loginMethods 决定是否渲染第三方入口，用 site 渲染页脚备案。
func (settings *Settings) Public(ctx context.Context) (*bo.SiteConfig, error) {
	siteSettings, err := settings.sites.SiteSettings(ctx)
	if err != nil {
		return nil, settings.access.Wrap(err)
	}
	logins, err := settings.logins.LoginMethods(ctx)
	if err != nil {
		return nil, settings.access.Wrap(err)
	}
	return &bo.SiteConfig{Site: *siteSettings, Login: *logins}, nil
}

// Get 读取站点设置，需要 system:site:read。
func (settings *Settings) Get(ctx context.Context) (*bo.SiteSettings, error) {
	if _, err := settings.access.RequireSystem(ctx, perm.SystemSiteRead); err != nil {
		return nil, err
	}
	item, err := settings.sites.SiteSettings(ctx)
	if err != nil {
		return nil, settings.access.Wrap(err)
	}
	return item, nil
}

// Update 保存站点设置，需要 system:site:write。全量覆盖，先裁剪空白再校验。
func (settings *Settings) Update(ctx context.Context, in bo.SiteSettings) (*bo.SiteSettings, error) {
	account, err := settings.access.RequireSystem(ctx, perm.SystemSiteWrite)
	if err != nil {
		return nil, err
	}
	in = normalizeSiteSettings(in)
	if err := validateSiteSettings(in); err != nil {
		return nil, err
	}
	in.UpdatedBy = account.UID
	item, err := settings.sites.SaveSiteSettings(ctx, in)
	if err != nil {
		return nil, settings.access.Wrap(err)
	}
	return item, nil
}

// GetLoginMethods 读取登录方式开关，需要 system:login:read。
func (settings *Settings) GetLoginMethods(ctx context.Context) (*bo.LoginMethods, error) {
	if _, err := settings.access.RequireSystem(ctx, perm.SystemLoginRead); err != nil {
		return nil, err
	}
	item, err := settings.logins.LoginMethods(ctx)
	if err != nil {
		return nil, settings.access.Wrap(err)
	}
	return item, nil
}

// UpdateLoginMethods 保存登录方式开关，需要 system:login:write。
// 两种方式同时关闭等于把所有账号锁在门外，直接拒绝，不落库。
func (settings *Settings) UpdateLoginMethods(ctx context.Context, in bo.LoginMethods) (*bo.LoginMethods, error) {
	account, err := settings.access.RequireSystem(ctx, perm.SystemLoginWrite)
	if err != nil {
		return nil, err
	}
	if !in.EmailEnabled && !in.OAuth2Enabled {
		return nil, merr.Invalid("at least one login method must stay enabled")
	}
	in.UpdatedBy = account.UID
	item, err := settings.logins.SaveLoginMethods(ctx, in)
	if err != nil {
		return nil, settings.access.Wrap(err)
	}
	return item, nil
}

// normalizeSiteSettings 裁剪所有文本字段：备案号前后的空格会带进页脚，必须在这里消掉。
func normalizeSiteSettings(in bo.SiteSettings) bo.SiteSettings {
	in.SiteName = strings.TrimSpace(in.SiteName)
	in.LogoURL = strings.TrimSpace(in.LogoURL)
	in.FaviconURL = strings.TrimSpace(in.FaviconURL)
	in.ContactEmail = strings.TrimSpace(in.ContactEmail)
	in.RepositoryURL = strings.TrimSpace(in.RepositoryURL)
	in.CopyrightText = strings.TrimSpace(in.CopyrightText)
	in.FilingICPNumber = strings.TrimSpace(in.FilingICPNumber)
	in.FilingICPURL = strings.TrimSpace(in.FilingICPURL)
	in.FilingPoliceNumber = strings.TrimSpace(in.FilingPoliceNumber)
	in.FilingPoliceURL = strings.TrimSpace(in.FilingPoliceURL)
	in.FilingExtraText = strings.TrimSpace(in.FilingExtraText)
	return in
}

// validateSiteSettings 校验站点设置。长度由接口层限制，这里只管格式：
// 邮箱要能解析，地址只接受 http(s)——`javascript:` 之类的 scheme 不能变成可点链接。
func validateSiteSettings(in bo.SiteSettings) error {
	if in.ContactEmail != "" {
		if _, err := mail.ParseAddress(in.ContactEmail); err != nil {
			return merr.Invalid("contact email is invalid")
		}
	}
	urls := []struct {
		name  string
		value string
	}{
		{"logo url", in.LogoURL},
		{"favicon url", in.FaviconURL},
		{"repository url", in.RepositoryURL},
		{"filing icp url", in.FilingICPURL},
		{"filing police url", in.FilingPoliceURL},
	}
	for _, item := range urls {
		if item.value == "" {
			continue
		}
		if !isHTTPURL(item.value) {
			return merr.Invalid(item.name + " is invalid")
		}
	}
	return nil
}

// isHTTPURL 判断地址是否是带主机的 http(s) 链接。
// 后台配置最终会渲染成公开页面上的 <a href> 或 <img src>，
// 其它 scheme（`javascript:`、`data:`）不能放行。
func isHTTPURL(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	return parsed.Host != ""
}
