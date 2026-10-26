package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/aide-family/Alioth/app/alioth/pkg/actor"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

// SelfOAuthLinks 列出当前账号已经绑定的第三方身份。个人资料页据此渲染已绑定行，
// 只按上下文里的当前账号查询，不接受调用方指定 user_uid。
func (authenticator *Authenticator) SelfOAuthLinks(ctx context.Context) ([]bo.OAuthLink, error) {
	userUID, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	links, err := authenticator.oauth.ListOAuthByUser(ctx, userUID)
	if err != nil {
		return nil, authenticator.wrap(err)
	}
	return links, nil
}

// BindSelfOAuth 用第三方授权码把身份绑定到当前账号。它只写 user_oauth，不签发会话，
// 也不改变当前登录身份。已经绑在别人名下的第三方身份必须拒绝：这里绝不能调用
// MoveOAuth，否则已登录用户会把别人的登录方式抢过来（账号劫持）。
func (authenticator *Authenticator) BindSelfOAuth(ctx context.Context, provider, code, redirectURI string) (*bo.OAuthLink, error) {
	userUID, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	provider = strings.ToUpper(strings.TrimSpace(provider))
	client, ok := authenticator.oauthClients[provider]
	if !ok || client.id == "" || client.secret == "" {
		return nil, merr.Client("oauth provider is not configured")
	}
	profile, err := authenticator.exchangeOAuth(ctx, provider, client, code, redirectURI)
	if err != nil {
		authenticator.log.Error(fmt.Sprintf("oauth exchange failed for %s", provider))
		return nil, merr.Client("oauth bind failed")
	}
	if profile.openID == "" {
		return nil, merr.Client("oauth bind failed")
	}
	// 先按第三方身份查归属：已绑本账号按幂等返回，已绑其他账号返回冲突。
	linked, err := authenticator.oauth.OAuthAccount(ctx, provider, profile.openID)
	switch {
	case err == nil:
		if linked.UID != userUID {
			return nil, merr.Conflict("oauth account is already bound")
		}
		return authenticator.userOAuthLink(ctx, userUID, provider, profile.openID)
	case !merr.IsNotFound(err):
		return nil, authenticator.wrap(err)
	}
	// 还没有绑定行：新建绑定。邮箱按登录链路一致的口径归一化后保存，可能为空。
	if err := authenticator.oauth.LinkOAuth(ctx, userUID, provider, profile.openID, normalizeEmail(profile.email)); err != nil {
		// 撞 (provider, open_id) 唯一索引说明并发请求刚绑好：确认最终归属，
		// 本账号按幂等成功处理，其他账号按冲突处理，不能变成服务端错误。
		linked, queryErr := authenticator.oauth.OAuthAccount(ctx, provider, profile.openID)
		if queryErr != nil {
			return nil, authenticator.wrap(err)
		}
		if linked.UID != userUID {
			return nil, merr.Conflict("oauth account is already bound")
		}
	}
	return authenticator.userOAuthLink(ctx, userUID, provider, profile.openID)
}

// UnbindSelfOAuth 解绑当前账号的一条第三方身份。先按 uid + user_uid 查归属，别人的
// 绑定一律按不存在处理，不泄露绑定是否存在；删除前确认解绑后还有可用的登录方式，
// 避免把账号锁死。删除时仓储仍会同时带 user_uid，防止并发下按 uid 误删。
func (authenticator *Authenticator) UnbindSelfOAuth(ctx context.Context, uid string) error {
	userUID, ok := actor.UserID(ctx)
	if !ok {
		return merr.Unauthenticated("missing token")
	}
	uid = strings.TrimSpace(uid)
	links, err := authenticator.oauth.ListOAuthByUser(ctx, userUID)
	if err != nil {
		return authenticator.wrap(err)
	}
	found := false
	remaining := 0
	for i := range links {
		if links[i].UID == uid {
			found = true
			continue
		}
		remaining++
	}
	if !found {
		// 他人的 uid 和根本不存在的 uid 返回同一个错误，不泄露绑定归属。
		return merr.NotFound("oauth link not found")
	}
	account, err := authenticator.users.UserByUID(ctx, userUID)
	if err != nil {
		return authenticator.wrap(err)
	}
	// 解绑后没有任何其他 OAuth 绑定，邮箱也不可用（空邮箱或历史占位邮箱）时，
	// 这次解绑会让账号再也无法登录，必须拒绝。
	if remaining == 0 && !hasLoginEmail(account.Email) {
		return merr.Conflict("cannot unlink the last login method")
	}
	if err := authenticator.oauth.UnlinkOAuth(ctx, userUID, uid); err != nil {
		if merr.IsNotFound(err) {
			return merr.NotFound("oauth link not found")
		}
		return authenticator.wrap(err)
	}
	return nil
}

// userOAuthLink 取出当前账号下 provider + open_id 精确匹配的绑定。
// OAuthByUserProvider 只按来源返回最近一条，同一来源绑了多个身份时退回完整列表匹配。
func (authenticator *Authenticator) userOAuthLink(ctx context.Context, userUID, provider, openID string) (*bo.OAuthLink, error) {
	link, err := authenticator.oauth.OAuthByUserProvider(ctx, userUID, provider)
	if err == nil && link.OpenID == openID {
		return link, nil
	}
	if err != nil && !merr.IsNotFound(err) {
		return nil, authenticator.wrap(err)
	}
	links, err := authenticator.oauth.ListOAuthByUser(ctx, userUID)
	if err != nil {
		return nil, authenticator.wrap(err)
	}
	for i := range links {
		if links[i].Provider == provider && links[i].OpenID == openID {
			return &links[i], nil
		}
	}
	return nil, merr.Server("request failed")
}

// hasLoginEmail 判断邮箱还能不能用来收验证码登录。历史占位邮箱收不到信，
// 不算可用登录方式。
func hasLoginEmail(email string) bool {
	email = normalizeEmail(email)
	return email != "" && !isPlaceholderEmail(email)
}
