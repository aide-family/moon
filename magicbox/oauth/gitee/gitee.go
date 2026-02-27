// Package gitee is the gitee auth package for the goddess service.
package gitee

import (
	"encoding/json"

	"github.com/go-kratos/kratos/v2/transport/http"
	"golang.org/x/oauth2"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/oauth"
)

func init() {
	oauth.RegisterOAuth2LoginFun(config.OAuth2_GITEE, Login)
}

func Login(ctx http.Context, oauthConfig *oauth2.Config) (*oauth.OAuth2User, error) {
	code := ctx.Request().URL.Query().Get("code")
	if code == "" {
		return nil, merr.ErrorInvalidArgument("code is required")
	}
	opts := []oauth2.AuthCodeOption{
		// https://gitee.com/oauth/token?grant_type=authorization_code&code={code}&client_id={client_id}&redirect_uri={redirect_uri}&client_secret={client_secret}
		oauth2.SetAuthURLParam("grant_type", "authorization_code"),
		oauth2.SetAuthURLParam("client_secret", oauthConfig.ClientSecret),
		oauth2.SetAuthURLParam("client_id", oauthConfig.ClientID),
		oauth2.SetAuthURLParam("redirect_uri", oauthConfig.RedirectURL),
		oauth2.SetAuthURLParam("code", code),
	}
	token, err := oauthConfig.Exchange(ctx, code, opts...)
	if err != nil {
		return nil, merr.ErrorInternalServer("exchange token failed").WithCause(err)
	}
	client := oauthConfig.Client(ctx, token)
	resp, err := client.Get("https://gitee.com/api/v5/user")
	if err != nil {
		return nil, merr.ErrorInternalServer("get user info failed").WithCause(err)
	}
	defer resp.Body.Close()
	var user GiteeUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, merr.ErrorInternalServer("decode user info failed").WithCause(err)
	}
	return user.Parse(), nil
}
