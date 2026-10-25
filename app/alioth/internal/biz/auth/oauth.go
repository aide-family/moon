package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/id"
	"github.com/aide-family/Alioth/pkg/merr"
)

type oauthProfile struct {
	openID        string
	email         string
	emailVerified bool
	name          string
	avatar        string
}

// OAuthLogin 先按第三方 open id 找已绑定账号。没有绑定时，用邮箱并入已有用户，否则新建用户再绑定。
// 第三方没有邮箱时使用本地占位邮箱，避免和真实邮箱冲突。封禁用户仍由 session 拒绝签发令牌。
func (authenticator *Authenticator) OAuthLogin(ctx context.Context, provider, code, redirectURI string) (*bo.Session, error) {
	provider = strings.ToUpper(strings.TrimSpace(provider))
	client, ok := authenticator.oauthClients[provider]
	if !ok || client.id == "" || client.secret == "" {
		return nil, merr.Client("oauth provider is not configured")
	}
	profile, err := authenticator.exchangeOAuth(ctx, provider, client, code, redirectURI)
	if err != nil {
		authenticator.log.Error(fmt.Sprintf("oauth exchange failed for %s", provider))
		return nil, merr.Client("oauth login failed")
	}
	if profile.openID == "" {
		return nil, merr.Client("oauth login failed")
	}
	if linked, err := authenticator.oauth.OAuthAccount(ctx, provider, profile.openID); err == nil {
		return authenticator.session(linked)
	} else if err != nil && !merr.IsNotFound(err) {
		return nil, authenticator.wrap(err)
	}
	email := normalizeEmail(profile.email)
	if email != "" && !profile.emailVerified {
		// 没有验证过的第三方邮箱不能并入已有本地账号，否则等于绕过邮箱所有权。
		if _, err := authenticator.users.UserByEmail(ctx, email); err == nil {
			authenticator.log.Warn(fmt.Sprintf("oauth login rejected: %s email is not verified but already registered", provider))
			return nil, merr.Client("oauth email is not verified")
		}
	}
	if email == "" {
		email = fmt.Sprintf("%s-%s@users.alioth.local", strings.ToLower(provider), profile.openID)
	}
	account, err := authenticator.users.UserByEmail(ctx, email)
	if err != nil {
		if !merr.IsNotFound(err) {
			return nil, authenticator.wrap(err)
		}
		name := profile.name
		if name == "" {
			name = strings.Split(email, "@")[0]
		}
		account, err = authenticator.users.CreateUser(ctx, bo.Account{
			UID:      id.New(),
			Name:     name,
			Nickname: name,
			Email:    email,
			Avatar:   profile.avatar,
			Status:   "ACTIVE",
		})
		if err != nil {
			return nil, authenticator.wrap(err)
		}
	}
	if err := authenticator.oauth.LinkOAuth(ctx, account.UID, provider, profile.openID, email); err != nil {
		return nil, authenticator.wrap(err)
	}
	return authenticator.session(account)
}

func (authenticator *Authenticator) exchangeOAuth(ctx context.Context, provider string, client oauthClient, code, redirectURI string) (oauthProfile, error) {
	switch provider {
	case "GITHUB":
		return authenticator.exchangeGitHub(ctx, client, code, redirectURI)
	case "GITEE":
		return authenticator.exchangeGitee(ctx, client, code, redirectURI)
	case "FEISHU":
		return authenticator.exchangeFeishu(ctx, client, code, redirectURI)
	default:
		return oauthProfile{}, merr.Client("oauth login failed")
	}
}

func (authenticator *Authenticator) exchangeGitHub(ctx context.Context, client oauthClient, code, redirectURI string) (oauthProfile, error) {
	body := url.Values{}
	body.Set("client_id", client.id)
	body.Set("client_secret", client.secret)
	body.Set("code", code)
	body.Set("redirect_uri", redirectURI)
	var token struct {
		AccessToken string `json:"access_token"`
	}
	if err := oauthForm(ctx, http.MethodPost, "https://github.com/login/oauth/access_token", body, "application/json", &token); err != nil {
		return oauthProfile{}, err
	}
	var user struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := oauthJSON(ctx, http.MethodGet, "https://api.github.com/user", token.AccessToken, nil, &user); err != nil {
		return oauthProfile{}, err
	}
	name := user.Name
	if name == "" {
		name = user.Login
	}
	profile := oauthProfile{openID: fmt.Sprint(user.ID), email: user.Email, name: name, avatar: user.AvatarURL}
	// GitHub 的 /user 邮箱不一定验证过，只有 /user/emails 里 primary + verified 的地址才能并入本地账号。
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := oauthJSON(ctx, http.MethodGet, "https://api.github.com/user/emails", token.AccessToken, nil, &emails); err == nil {
		for _, item := range emails {
			if item.Verified && item.Email != "" {
				profile.email = item.Email
				profile.emailVerified = true
				if item.Primary {
					break
				}
			}
		}
	}
	return profile, nil
}

func (authenticator *Authenticator) exchangeGitee(ctx context.Context, client oauthClient, code, redirectURI string) (oauthProfile, error) {
	body := url.Values{}
	body.Set("grant_type", "authorization_code")
	body.Set("client_id", client.id)
	body.Set("client_secret", client.secret)
	body.Set("code", code)
	body.Set("redirect_uri", redirectURI)
	var token struct {
		AccessToken string `json:"access_token"`
	}
	if err := oauthForm(ctx, http.MethodPost, "https://gitee.com/oauth/token", body, "application/json", &token); err != nil {
		return oauthProfile{}, err
	}
	var user struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := oauthJSON(ctx, http.MethodGet, "https://gitee.com/api/v5/user?access_token="+url.QueryEscape(token.AccessToken), "", nil, &user); err != nil {
		return oauthProfile{}, err
	}
	name := user.Name
	if name == "" {
		name = user.Login
	}
	profile := oauthProfile{openID: fmt.Sprint(user.ID), email: user.Email, name: name, avatar: user.AvatarURL}
	// Gitee 的 /user 不带邮箱验证状态。/emails 里 state=confirmed 的才是已验证邮箱，
	// 其中 scope 含 primary 的是主邮箱。接口需要 emails 权限：拿不到就退回 /user 的邮箱，按未验证处理。
	if verified, err := authenticator.giteeVerifiedEmail(ctx, token.AccessToken); err == nil && verified != "" {
		profile.email = verified
		profile.emailVerified = true
	}
	return profile, nil
}

// giteeEmail 是 Gitee 邮箱列表的一项：state=confirmed 表示已验证，scope 含 primary 表示主邮箱。
type giteeEmail struct {
	Email string   `json:"email"`
	State string   `json:"state"`
	Scope []string `json:"scope"`
}

// giteeVerifiedEmail 读取 Gitee 邮箱列表并挑一个已验证邮箱。
// 依赖授权时申请的 emails 权限；Gitee 应用没勾选该权限时接口会拒绝访问，此时按
// “没有已验证邮箱”处理，不阻断登录，由前端引导用户去 Gitee 完成邮箱验证或改用邮箱验证码登录。
func (authenticator *Authenticator) giteeVerifiedEmail(ctx context.Context, accessToken string) (string, error) {
	var emails []giteeEmail
	endpoint := "https://gitee.com/api/v5/emails?access_token=" + url.QueryEscape(accessToken)
	if err := oauthJSON(ctx, http.MethodGet, endpoint, "", nil, &emails); err != nil {
		// 拿不到验证状态时只是无法并入已有账号，登录本身仍然可以继续，记一条警告便于排查权限问题。
		authenticator.log.Warn(fmt.Sprintf("gitee emails unavailable: %v", err))
		return "", err
	}
	verified := pickGiteeVerifiedEmail(emails)
	if verified == "" {
		authenticator.log.Warn(fmt.Sprintf("gitee emails: no confirmed primary email among %d entries", len(emails)))
	}
	return verified, nil
}

// pickGiteeVerifiedEmail 优先返回已验证的主邮箱，没有主邮箱时退回第一个已验证邮箱。
func pickGiteeVerifiedEmail(emails []giteeEmail) string {
	fallback := ""
	for _, item := range emails {
		if item.State != "confirmed" || item.Email == "" {
			continue
		}
		if slices.Contains(item.Scope, "primary") {
			return item.Email
		}
		if fallback == "" {
			fallback = item.Email
		}
	}
	return fallback
}

func (authenticator *Authenticator) exchangeFeishu(ctx context.Context, client oauthClient, code, redirectURI string) (oauthProfile, error) {
	payload := map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     client.id,
		"client_secret": client.secret,
		"code":          code,
		"redirect_uri":  redirectURI,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return oauthProfile{}, merr.Client("oauth login failed")
	}
	var token map[string]any
	if err := oauthJSON(ctx, http.MethodPost, "https://open.feishu.cn/open-apis/authen/v2/oauth/token", "", raw, &token); err != nil {
		return oauthProfile{}, err
	}
	access, _ := token["access_token"].(string)
	if access == "" {
		if data, ok := token["data"].(map[string]any); ok {
			access, _ = data["access_token"].(string)
		}
	}
	if access == "" {
		return oauthProfile{}, merr.Client("oauth login failed")
	}
	var info map[string]any
	if err := oauthJSON(ctx, http.MethodGet, "https://open.feishu.cn/open-apis/authen/v1/user_info", access, nil, &info); err != nil {
		return oauthProfile{}, err
	}
	data, _ := info["data"].(map[string]any)
	if data == nil {
		data = info
	}
	openID, _ := data["open_id"].(string)
	name, _ := data["name"].(string)
	email, _ := data["email"].(string)
	avatar, _ := data["avatar_url"].(string)
	return oauthProfile{openID: openID, email: email, name: name, avatar: avatar}, nil
}

func oauthForm(ctx context.Context, method, endpoint string, body url.Values, accept string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, strings.NewReader(body.Encode()))
	if err != nil {
		return merr.Client("oauth login failed")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", accept)
	return doOAuth(req, dest)
}

func oauthJSON(ctx context.Context, method, endpoint, token string, body []byte, dest any) error {
	var reader io.Reader
	if body != nil {
		reader = strings.NewReader(string(body))
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return merr.Client("oauth login failed")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	return doOAuth(req, dest)
}

func doOAuth(req *http.Request, dest any) error {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return merr.Client("oauth login failed")
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return merr.Client("oauth login failed")
	}
	if resp.StatusCode >= 300 {
		return merr.Client("oauth login failed")
	}
	if dest == nil {
		return nil
	}
	if err := json.Unmarshal(payload, dest); err != nil {
		return merr.Client("oauth login failed")
	}
	return nil
}

// oauthProvider 是一个可以列出授权地址的第三方来源。
type oauthProvider struct {
	key               string
	name              string
	authorizeEndpoint string
}

// oauthProviderCatalog 只包含已经实现授权码流程的来源。飞书仍可登录，但不在来源列表里。
var oauthProviderCatalog = []oauthProvider{
	{key: "GITHUB", name: "GitHub", authorizeEndpoint: "https://github.com/login/oauth/authorize"},
	{key: "GITEE", name: "Gitee", authorizeEndpoint: "https://gitee.com/oauth/authorize"},
}

// authorizeURL 按来源拼出完整授权地址。GitHub 和 Gitee 的查询参数不同：
// GitHub 用 allow_signup 和空格分隔的 scope，Gitee 要 response_type=code。
func (provider oauthProvider) authorizeURL(client oauthClient, redirectURI, state string) string {
	params := url.Values{}
	params.Set("client_id", client.id)
	params.Set("redirect_uri", redirectURI)
	params.Set("state", state)
	switch provider.key {
	case "GITHUB":
		params.Set("scope", "read:user user:email")
		params.Set("allow_signup", "true")
	case "GITEE":
		params.Set("response_type", "code")
		// emails 权限用来读 /api/v5/emails 判断邮箱是否已验证。Gitee 应用必须先在
		// 应用设置里勾选 emails，否则授权页会报“请求范围无效、未知或格式不正确”。
		params.Set("scope", "user_info emails")
	}
	// url.Values 把空格编码成 "+"。Gitee 的授权页只认 %20，用 "+" 会报
	// “请求范围无效、未知或格式不正确”，所以这里统一把空格换成 %20
	// （值里的字面 + 已被编码成 %2B，不会误伤）。
	return provider.authorizeEndpoint + "?" + strings.ReplaceAll(params.Encode(), "+", "%20")
}

// OAuthProviders 只返回客户端已配置的来源：client_id 和 client_secret 都不为空才出现。
// 前端据此决定显示哪些第三方登录按钮。redirectUri 原样写进授权地址，
// 必须是 http(s)，否则第三方会拒绝，或者把用户带到不安全的自定义协议。
func (authenticator *Authenticator) OAuthProviders(redirectURI, state string) ([]bo.OAuthProvider, error) {
	redirectURI = strings.TrimSpace(redirectURI)
	if !strings.HasPrefix(redirectURI, "http://") && !strings.HasPrefix(redirectURI, "https://") {
		return nil, merr.Client("invalid redirect uri")
	}
	providers := make([]bo.OAuthProvider, 0, len(oauthProviderCatalog))
	for _, item := range oauthProviderCatalog {
		client, ok := authenticator.oauthClients[item.key]
		if !ok || client.id == "" || client.secret == "" {
			continue
		}
		providers = append(providers, bo.OAuthProvider{
			Provider:     item.key,
			Name:         item.name,
			AuthorizeURL: item.authorizeURL(client, redirectURI, state),
		})
	}
	return providers, nil
}
