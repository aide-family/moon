// Package github is the github auth package for the goddess service.
package github

import (
	"encoding/json"

	"github.com/go-kratos/kratos/v2/transport/http"
	"golang.org/x/oauth2"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/oauth"
)

func init() {
	oauth.RegisterOAuth2LoginFun(config.OAuth2_GITHUB, Login)
}

func Login(ctx http.Context, oauthConfig *oauth2.Config) (*oauth.OAuth2User, error) {
	code := ctx.Request().URL.Query().Get("code")
	if code == "" {
		return nil, merr.ErrorInvalidArgument("code is required")
	}
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, merr.ErrorInternalServer("exchange token failed").WithCause(err)
	}
	client := oauthConfig.Client(ctx, token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, merr.ErrorInternalServer("get user info failed").WithCause(err)
	}
	defer resp.Body.Close()
	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, merr.ErrorInternalServer("decode user info failed").WithCause(err)
	}
	return user.Parse(), nil
}
