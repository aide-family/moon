package auth

import (
	"context"
	"fmt"

	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

// IOAuthUser OAuth用户接口
type IOAuthUser interface {
	fmt.Stringer
	GetOAuthID() uint32
	GetEmail() string
	GetRemark() string
	GetUsername() string
	GetNickname() string
	GetAvatar() string
	GetAPP() vobj.OAuthAPP
}

// OauthLoginParams OAuth登录参数
type OauthLoginParams struct {
	Code    string `json:"code"`
	Email   string `json:"email"`
	OAuthID uint32 `json:"oAuthID"`
	Token   string `json:"token"`
}

// VerifyToken 校验token是否过期
func (o *OauthLoginParams) VerifyToken(ctx context.Context, cacher cache.ICacher) error {
	exist, err := cacher.Client().Exists(ctx, o.GetTokenKey()).Result()
	if err != nil {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	if exist == 0 {
		return merr.ErrorI18nUnauthorized(ctx).WithMetadata(map[string]string{
			"exist": "false",
		})
	}
	return nil
}

// WaitVerifyToken 等待token验证
func (o *OauthLoginParams) WaitVerifyToken(ctx context.Context, cacher cache.ICacher) error {
	return cacher.Client().Set(ctx, o.GetTokenKey(), o.Token, 0).Err()
}

// GetTokenKey 返回token的key
func (o *OauthLoginParams) GetTokenKey() string {
	return fmt.Sprintf("oauth:%d:%s", o.OAuthID, o.Token)
}

// NewOAuthRowData 创建OAuth用户
func NewOAuthRowData(app vobj.OAuthAPP, row string) (IOAuthUser, error) {
	switch app {
	case vobj.OAuthAPPGithub:
		var githubUser GithubUser
		err := types.Unmarshal([]byte(row), &githubUser)
		return &githubUser, err
	case vobj.OAuthAPPGitee:
		var giteeUser GiteeUser
		err := types.Unmarshal([]byte(row), &giteeUser)
		return &giteeUser, err
	default:
		return nil, merr.ErrorI18nNotificationSystemError(context.Background())
	}
}
