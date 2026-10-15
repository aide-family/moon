package biz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aide-family/Alioth/pkg/merr"
)

type oauthProfile struct {
	openID string
	email  string
	name   string
	avatar string
}

func (uc *IdentityUsecase) OAuthLogin(ctx context.Context, provider, code, redirectURI string) (*Session, error) {
	provider = strings.ToUpper(strings.TrimSpace(provider))
	client, ok := uc.oauth[provider]
	if !ok || client.id == "" || client.secret == "" {
		return nil, merr.Client("oauth provider is not configured")
	}
	profile, err := exchangeOAuth(ctx, provider, client, code, redirectURI)
	if err != nil {
		uc.log.Errorf("oauth exchange failed for %s", provider)
		return nil, merr.Client("oauth login failed")
	}
	if profile.openID == "" {
		return nil, merr.Client("oauth login failed")
	}
	if linked, err := uc.repo.OAuthAccount(ctx, provider, profile.openID); err == nil {
		return uc.session(linked)
	} else if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, uc.wrap(err)
	}
	email := normalizeEmail(profile.email)
	if email == "" {
		email = fmt.Sprintf("%s-%s@users.alioth.local", strings.ToLower(provider), profile.openID)
	}
	account, err := uc.repo.UserByEmail(ctx, email)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return nil, uc.wrap(err)
		}
		name := profile.name
		if name == "" {
			name = strings.Split(email, "@")[0]
		}
		account, err = uc.repo.CreateUser(ctx, Account{
			UID:      newUID(),
			Name:     name,
			Nickname: name,
			Email:    email,
			Avatar:   profile.avatar,
			Status:   "ACTIVE",
		})
		if err != nil {
			return nil, uc.wrap(err)
		}
	}
	if err := uc.repo.LinkOAuth(ctx, account.UID, provider, profile.openID, email); err != nil {
		return nil, uc.wrap(err)
	}
	return uc.session(account)
}

func exchangeOAuth(ctx context.Context, provider string, client oauthClient, code, redirectURI string) (oauthProfile, error) {
	switch provider {
	case "GITHUB":
		return exchangeGitHub(ctx, client, code, redirectURI)
	case "GITEE":
		return exchangeGitee(ctx, client, code, redirectURI)
	case "FEISHU":
		return exchangeFeishu(ctx, client, code, redirectURI)
	default:
		return oauthProfile{}, fmt.Errorf("unknown provider")
	}
}

func exchangeGitHub(ctx context.Context, client oauthClient, code, redirectURI string) (oauthProfile, error) {
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
	return oauthProfile{openID: fmt.Sprint(user.ID), email: user.Email, name: name, avatar: user.AvatarURL}, nil
}

func exchangeGitee(ctx context.Context, client oauthClient, code, redirectURI string) (oauthProfile, error) {
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
	return oauthProfile{openID: fmt.Sprint(user.ID), email: user.Email, name: name, avatar: user.AvatarURL}, nil
}

func exchangeFeishu(ctx context.Context, client oauthClient, code, redirectURI string) (oauthProfile, error) {
	payload := map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     client.id,
		"client_secret": client.secret,
		"code":          code,
		"redirect_uri":  redirectURI,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return oauthProfile{}, err
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
		return oauthProfile{}, fmt.Errorf("missing access token")
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
		return err
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
		return err
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
		return err
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("oauth status %d", resp.StatusCode)
	}
	if dest == nil {
		return nil
	}
	return json.Unmarshal(payload, dest)
}
