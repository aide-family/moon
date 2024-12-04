package builder

import (
	"context"

	authorizationapi "github.com/aide-family/moon/api/admin/authorization"
	"github.com/aide-family/moon/cmd/server/palace/internal/palaceconf"
)

// IOauthModuleBuilder 构建oauth模块
type IOauthModuleBuilder interface {
	ToAPI(config *palaceconf.OAuth2) []*authorizationapi.OauthItem
}

// NewOauthModuleBuilder 创建oauth模块构建器
func NewOauthModuleBuilder(ctx context.Context) IOauthModuleBuilder {
	return &oauthModuleBuilder{ctx: ctx}
}

type oauthModuleBuilder struct {
	ctx context.Context
}

func (b *oauthModuleBuilder) ToAPI(config *palaceconf.OAuth2) []*authorizationapi.OauthItem {
	if config == nil {
		return nil
	}
	list := make([]*authorizationapi.OauthItem, 0, 2)
	if config.Github != nil {
		list = append(list, &authorizationapi.OauthItem{
			Icon:     "github",
			Label:    "Github登录",
			Redirect: config.GetGithub().GetAuthorizeUri(),
		})
	}
	if config.Gitee != nil {
		list = append(list, &authorizationapi.OauthItem{
			Icon:     "gitee",
			Label:    "Gitee登录",
			Redirect: config.GetGitee().GetAuthorizeUri(),
		})
	}
	return list
}
