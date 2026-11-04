package site

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

// siteFeatureIcons 是能力卡片可用的图标白名单。
// 后台只能选这里的键名：图标由前端映射到具体组件，不能让后台传任意名字。
var siteFeatureIcons = map[string]struct{}{
	"activity": {}, "bell": {}, "shield": {}, "database": {},
	"globe": {}, "boxes": {}, "terminal": {}, "gauge": {},
	"server": {}, "users": {}, "git-branch": {}, "layers": {},
}

// 首页文案里禁止出现的营销词。文案面向工程师，要能验证；这些词与未证实的数字
// 一起属于品牌红线，进入保存校验而不是只写在文档里。
var siteCopyBannedWords = []string{
	"赋能", "闭环", "一站式", "极致", "颠覆", "军工级", "金融级",
	"毫秒级", "海量", "深度打通", "重新定义",
}

const maxSiteFeatures = 6

// Home 读取某个语言的首页内容，需要 system:site:read。
func (settings *Settings) Home(ctx context.Context, locale string) (*bo.SiteHome, error) {
	if _, err := settings.access.RequireSystem(ctx, perm.SystemSiteRead); err != nil {
		return nil, err
	}
	return settings.loadHome(ctx, normalizeLocale(locale))
}

// PublicHome 读取公开首页内容，不需要登录。
func (settings *Settings) PublicHome(ctx context.Context, locale string) (*bo.SiteHome, error) {
	return settings.loadHome(ctx, normalizeLocale(locale))
}

func (settings *Settings) loadHome(ctx context.Context, locale string) (*bo.SiteHome, error) {
	content, err := settings.homes.HomeContent(ctx, locale)
	if err != nil {
		return nil, settings.access.Wrap(err)
	}
	features, err := settings.homes.SiteFeatures(ctx, locale)
	if err != nil {
		return nil, settings.access.Wrap(err)
	}
	return &bo.SiteHome{Content: *content, Features: features}, nil
}

// UpdateHome 保存某个语言的首页内容，需要 system:site:write。
// 能力卡片整体替换：版式固定 6 张上限，顺序即数组顺序。
func (settings *Settings) UpdateHome(ctx context.Context, locale string, content bo.SiteHomeContent, features []bo.SiteFeature) (*bo.SiteHome, error) {
	account, err := settings.access.RequireSystem(ctx, perm.SystemSiteWrite)
	if err != nil {
		return nil, err
	}
	locale = normalizeLocale(locale)
	content = normalizeHomeContent(content)
	content.Locale = locale
	content.UpdatedBy = account.UID
	if err := validateHomeContent(content); err != nil {
		return nil, err
	}
	if len(features) > maxSiteFeatures {
		return nil, merr.Invalid("too many feature cards")
	}
	normalized := make([]bo.SiteFeature, 0, len(features))
	for index, item := range features {
		item.Locale = locale
		item.SortOrder = index
		item.IconKey = strings.ToLower(strings.TrimSpace(item.IconKey))
		item.Title = strings.TrimSpace(item.Title)
		item.Description = strings.TrimSpace(item.Description)
		if _, ok := siteFeatureIcons[item.IconKey]; !ok {
			return nil, merr.Invalid("unsupported feature icon")
		}
		if utf8.RuneCountInString(item.Title) > 16 || utf8.RuneCountInString(item.Description) > 48 {
			return nil, merr.Invalid("site home copy is too long")
		}
		if containsBannedWording(item.Title) || containsBannedWording(item.Description) {
			return nil, merr.Invalid("site copy contains disallowed wording")
		}
		normalized = append(normalized, item)
	}
	savedContent, err := settings.homes.SaveHomeContent(ctx, content)
	if err != nil {
		return nil, settings.access.Wrap(err)
	}
	savedFeatures, err := settings.homes.ReplaceSiteFeatures(ctx, locale, normalized)
	if err != nil {
		return nil, settings.access.Wrap(err)
	}
	return &bo.SiteHome{Content: *savedContent, Features: savedFeatures}, nil
}

func normalizeHomeContent(in bo.SiteHomeContent) bo.SiteHomeContent {
	in.HeroTitle = strings.TrimSpace(in.HeroTitle)
	in.HeroSubtitle = strings.TrimSpace(in.HeroSubtitle)
	in.HeroImageURL = strings.TrimSpace(in.HeroImageURL)
	in.CtaPrimaryLabel = strings.TrimSpace(in.CtaPrimaryLabel)
	in.CtaSecondaryLabel = strings.TrimSpace(in.CtaSecondaryLabel)
	in.CtaSecondaryHref = strings.TrimSpace(in.CtaSecondaryHref)
	in.FeaturesTitle = strings.TrimSpace(in.FeaturesTitle)
	in.FeaturesSubtitle = strings.TrimSpace(in.FeaturesSubtitle)
	in.HowTitle = strings.TrimSpace(in.HowTitle)
	in.HowBody = strings.TrimSpace(in.HowBody)
	in.PermTitle = strings.TrimSpace(in.PermTitle)
	in.PermBody = strings.TrimSpace(in.PermBody)
	in.DeployTitle = strings.TrimSpace(in.DeployTitle)
	in.DeployBody = strings.TrimSpace(in.DeployBody)
	in.DocTitle = strings.TrimSpace(in.DocTitle)
	in.ClosingTitle = strings.TrimSpace(in.ClosingTitle)
	in.ClosingDescription = strings.TrimSpace(in.ClosingDescription)
	in.ClosingButtonLabel = strings.TrimSpace(in.ClosingButtonLabel)
	return in
}

// validateHomeContent 把版式的字数预算落成校验。字段为空表示「用前端默认文案」，
// 所以长度是上限而不是必填；但地址与 CTA 组合必须自洽。
func validateHomeContent(in bo.SiteHomeContent) error {
	limits := []struct {
		value string
		max   int
	}{
		{in.HeroTitle, 16},
		{in.HeroSubtitle, 72},
		{in.CtaPrimaryLabel, 12},
		{in.CtaSecondaryLabel, 12},
		{in.CtaSecondaryHref, 128},
		{in.FeaturesTitle, 16},
		{in.FeaturesSubtitle, 48},
		{in.HowTitle, 16},
		{in.HowBody, 80},
		{in.PermTitle, 16},
		{in.PermBody, 80},
		{in.DeployTitle, 16},
		{in.DeployBody, 80},
		{in.DocTitle, 16},
		{in.ClosingTitle, 20},
		{in.ClosingDescription, 48},
		{in.ClosingButtonLabel, 12},
	}
	for _, item := range limits {
		if utf8.RuneCountInString(item.value) > item.max {
			return merr.Invalid("site home copy is too long")
		}
	}
	if in.HeroImageURL != "" && !isHTTPURL(in.HeroImageURL) {
		return merr.Invalid("invalid hero image url")
	}
	// 次 CTA 只能指向站内锚点或站内路径：首页不是放外链的地方。
	if in.CtaSecondaryHref != "" {
		if !strings.HasPrefix(in.CtaSecondaryHref, "/") && !strings.HasPrefix(in.CtaSecondaryHref, "#") {
			return merr.Invalid("invalid secondary cta link")
		}
	}
	if (in.CtaSecondaryLabel == "") != (in.CtaSecondaryHref == "") {
		return merr.Invalid("secondary cta label and link must be set together")
	}
	if containsBannedWording(in.HeroTitle) || containsBannedWording(in.HeroSubtitle) ||
		containsBannedWording(in.ClosingTitle) || containsBannedWording(in.ClosingDescription) {
		return merr.Invalid("site copy contains disallowed wording")
	}
	return nil
}

func containsBannedWording(value string) bool {
	for _, word := range siteCopyBannedWords {
		if strings.Contains(value, word) {
			return true
		}
	}
	return false
}
