package builder

import (
	"context"

	authorizationapi "github.com/aide-family/moon/api/admin/authorization"
	"github.com/aide-family/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-family/moon/pkg/util/types"
)

type IOauthModuleBuilder interface {
	ToAPI(config *palaceconf.OAuth2) []*authorizationapi.OauthItem
}

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
			Redirect: types.GetAPI(config.GetGithub().GetCallbackUri()),
		})
	}
	if config.Gitee != nil {
		list = append(list, &authorizationapi.OauthItem{
			Icon:     "gitee",
			Label:    "Gitee登录",
			Redirect: types.GetAPI(config.GetGitee().GetCallbackUri()),
		})
	}
	return list
}
