package repository

import (
	"context"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// SiteRepo 维护官网的站点设置。站点设置是平台级单例，不属于任何租户。
type SiteRepo interface {
	// SiteSettings 读取站点设置；还没有保存过时返回默认空值，不返回 NotFound。
	SiteSettings(ctx context.Context) (*bo.SiteSettings, error)
	// SaveSiteSettings 覆盖保存站点设置，并记录修改人。
	SaveSiteSettings(ctx context.Context, in bo.SiteSettings) (*bo.SiteSettings, error)
}

// LoginMethodRepo 维护登录方式的平台级开关。开关与第三方来源的客户端凭据分离：
// 凭据来自配置文件，开关决定是否对外暴露与是否允许发起登录。
type LoginMethodRepo interface {
	// LoginMethods 读取登录方式开关；还没有保存过时两种都按开启返回，不返回 NotFound。
	LoginMethods(ctx context.Context) (*bo.LoginMethods, error)
	// SaveLoginMethods 覆盖保存登录方式开关，并记录修改人。
	SaveLoginMethods(ctx context.Context, in bo.LoginMethods) (*bo.LoginMethods, error)
}
