package bo

import "time"

// SiteSettings 是官网（公开站）的站点级配置，全平台只有一份。
// 备案字段留空表示官网页脚不展示该行，不能退化成占位号码。
type SiteSettings struct {
	SiteName           string
	LogoURL            string
	FaviconURL         string
	ContactEmail       string
	RepositoryURL      string
	CopyrightText      string
	FilingICPNumber    string
	FilingICPURL       string
	FilingPoliceNumber string
	FilingPoliceURL    string
	FilingExtraText    string
	UpdatedBy          string
	UpdatedAt          time.Time
}

// LoginMethods 是登录方式的平台级开关。
// 第三方来源的客户端凭据仍来自配置文件，这里的开关决定是否对外暴露与是否允许发起登录。
type LoginMethods struct {
	EmailEnabled  bool
	OAuth2Enabled bool
	UpdatedBy     string
	UpdatedAt     time.Time
}

// SiteConfig 是公开接口下发的站点配置：站点信息、登录方式与首页内容。
// 登录页、注册页与官网共用这一份配置，避免备案号出现第二个来源。
type SiteConfig struct {
	Site  SiteSettings
	Login LoginMethods
	// Home 是当前语言的首页文案与能力卡片；没有配置时字段为空，前端按默认文案渲染。
	Home SiteHome
}
