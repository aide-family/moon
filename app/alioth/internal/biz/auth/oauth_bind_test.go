package auth

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	kerrors "github.com/go-kratos/kratos/v3/errors"
	"github.com/golang-jwt/jwt/v5"

	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

const (
	oauthRedirectURI = "https://app.example.com/oauth2/callback"
	// feishuTokenOK 是飞书令牌接口的成功响应，两个 exchange 测试共用。
	feishuTokenOK = `{"code":0,"msg":"success","access_token":"u-feishu-token","token_type":"Bearer"}`
)

// oauthLink 是桩里的一条第三方绑定。
type oauthLink struct {
	userUID  string
	provider string
	openID   string
	email    string
}

// oauthMove 记录一次改绑调用，测试据此断言占位账号的绑定被迁到了邮箱账号。
type oauthMove struct {
	provider    string
	openID      string
	fromUserUID string
	toUserUID   string
	email       string
}

// oauthRepoStub 用内存绑定表代替 user_oauth 表，返回的账号从 userRepoStub 里取。
type oauthRepoStub struct {
	repository.OAuthRepo
	users *userRepoStub
	links []oauthLink
	moves []oauthMove
}

func newOAuthRepoStub(users *userRepoStub) *oauthRepoStub {
	return &oauthRepoStub{users: users}
}

// addLink 预置一条第三方绑定，用来模拟历史占位账号已经绑过钉钉或飞书。
func (stub *oauthRepoStub) addLink(userUID, provider, openID, email string) {
	stub.links = append(stub.links, oauthLink{userUID: userUID, provider: provider, openID: openID, email: email})
}

func (stub *oauthRepoStub) find(provider, openID string) (oauthLink, bool) {
	for _, link := range stub.links {
		if link.provider == provider && link.openID == openID {
			return link, true
		}
	}
	return oauthLink{}, false
}

func (stub *oauthRepoStub) OAuthAccount(_ context.Context, provider, openID string) (*bo.Account, error) {
	link, ok := stub.find(provider, openID)
	if !ok {
		return nil, merr.NotFound("not found")
	}
	account, ok := stub.users.byUID[link.userUID]
	if !ok {
		return nil, merr.NotFound("not found")
	}
	copied := account
	return &copied, nil
}

func (stub *oauthRepoStub) LinkOAuth(_ context.Context, userUID, provider, openID, email string) error {
	if _, exists := stub.find(provider, openID); exists {
		return merr.Conflict("name already exists")
	}
	stub.links = append(stub.links, oauthLink{userUID: userUID, provider: provider, openID: openID, email: email})
	return nil
}

func (stub *oauthRepoStub) MoveOAuth(_ context.Context, provider, openID, fromUserUID, toUserUID, email string) error {
	stub.moves = append(stub.moves, oauthMove{
		provider: provider, openID: openID, fromUserUID: fromUserUID, toUserUID: toUserUID, email: email,
	})
	for i := range stub.links {
		if stub.links[i].provider == provider && stub.links[i].openID == openID && stub.links[i].userUID == fromUserUID {
			stub.links[i].userUID = toUserUID
			stub.links[i].email = email
			return nil
		}
	}
	return merr.NotFound("not found")
}

// configureClients 给 authenticator 配好第三方客户端，令牌和用户信息仍由 httptest 桩提供。
func configureClients(fixture *authFixture, providers ...string) {
	clients := map[string]oauthClient{}
	for _, provider := range providers {
		clients[provider] = oauthClient{id: strings.ToLower(provider) + "-id", secret: strings.ToLower(provider) + "-secret"}
	}
	fixture.authenticator.oauthClients = clients
}

// saveEmailCode 给邮箱预置一条验证码，模拟用户收到邮件后输入。
func saveEmailCode(fixture *authFixture, email, code string) {
	fixture.emailCodes.codes[email] = fixture.authenticator.digest(code)
}

// kerrorsMetadata 取出客户端错误里的元数据，用来断言节流剩余秒数。
func kerrorsMetadata(t *testing.T, err error) map[string]string {
	t.Helper()
	return kerrors.FromError(err).Metadata
}

// tamperJWT 改掉 JWT 载荷里的一个字符，签名随即失效，用来模拟被篡改的待绑定凭证。
func tamperJWT(token string) string {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[1] == "" {
		return token + "x"
	}
	payload := []byte(parts[1])
	if payload[0] == 'e' {
		payload[0] = 'f'
	} else {
		payload[0] = 'e'
	}
	parts[1] = string(payload)
	return strings.Join(parts, ".")
}

// TestOAuthLoginWithoutEmailBindsAccount 覆盖钉钉没有邮箱的完整流程：
// OAuthLogin 只返回待绑定凭证、不创建用户；验证邮箱后建账号并绑定；
// 同一邮箱再绑定飞书身份后仍然只有一个账号、两条绑定。
func TestOAuthLoginWithoutEmailBindsAccount(t *testing.T) {
	const bindEmail = "zhangsan@example.com"
	newDingtalkStub(t, dingtalkReply{
		tokenResponse:    `{"accessToken":"u-dingtalk-token"}`,
		userInfoResponse: `{"nick":"张三","avatarUrl":"https://avatar.example.com/zhangsan.png","openId":"open-zhangsan"}`,
	})
	fixture := newAuthFixture(t)
	configureClients(fixture, "DINGTALK")

	session, binding, err := fixture.authenticator.OAuthLogin(context.Background(), "DINGTALK", "auth-code", oauthRedirectURI)
	if err != nil {
		t.Fatalf("OAuthLogin: %v", err)
	}
	if session != nil {
		t.Fatal("第三方没有可用邮箱时不能签发会话")
	}
	if binding == nil || binding.Token == "" {
		t.Fatal("没有可用邮箱时应返回待绑定凭证")
	}
	if binding.SuggestedEmail != "" {
		t.Fatalf("钉钉没有返回邮箱时 suggested_email 应为空: %q", binding.SuggestedEmail)
	}
	if len(fixture.users.created) != 0 {
		t.Fatal("邮箱验证通过前不应创建用户，也不应再生成占位邮箱账号")
	}
	if len(fixture.oauth.links) != 0 {
		t.Fatal("邮箱验证通过前不应写入第三方绑定")
	}

	fixture.solveCaptcha(t, "captcha-1", "12")
	saveEmailCode(fixture, bindEmail, "123456")
	session, err = fixture.authenticator.BindOAuthEmail(context.Background(), binding.Token, bindEmail, "123456", "captcha-1", "12")
	if err != nil {
		t.Fatalf("BindOAuthEmail: %v", err)
	}
	if session.User.Email != bindEmail {
		t.Fatalf("会话账号邮箱 = %q, want %q", session.User.Email, bindEmail)
	}
	if len(fixture.users.created) != 1 {
		t.Fatalf("创建用户次数 = %d, want 1", len(fixture.users.created))
	}
	created := fixture.users.created[0]
	if created.Name != "张三" || created.Avatar != "https://avatar.example.com/zhangsan.png" {
		t.Fatalf("新账号应沿用第三方昵称和头像: %+v", created)
	}
	link, ok := fixture.oauth.find("DINGTALK", "open-zhangsan")
	if !ok {
		t.Fatal("验证邮箱后应写入第三方绑定")
	}
	if link.userUID != session.User.UID || link.email != bindEmail {
		t.Fatalf("绑定 = %+v, want user %q email %q", link, session.User.UID, bindEmail)
	}

	// 同一邮箱再绑定另一个没有邮箱的第三方身份（飞书）：仍只有一个用户、两条绑定。
	newFeishuStub(t, feishuTokenOK, `{"code":0,"msg":"success","data":{"name":"李四","open_id":"ou_lisi"}}`)
	fixture.authenticator.oauthClients["FEISHU"] = oauthClient{id: "feishu-id", secret: "feishu-secret"}
	_, feishuBinding, err := fixture.authenticator.OAuthLogin(context.Background(), "FEISHU", "auth-code", oauthRedirectURI)
	if err != nil {
		t.Fatalf("OAuthLogin(FEISHU): %v", err)
	}
	if feishuBinding == nil {
		t.Fatal("飞书没有邮箱时应返回待绑定凭证")
	}
	fixture.solveCaptcha(t, "captcha-2", "12")
	saveEmailCode(fixture, bindEmail, "654321")
	second, err := fixture.authenticator.BindOAuthEmail(context.Background(), feishuBinding.Token, bindEmail, "654321", "captcha-2", "12")
	if err != nil {
		t.Fatalf("BindOAuthEmail(FEISHU): %v", err)
	}
	if second.User.UID != session.User.UID {
		t.Fatalf("同一邮箱应进入同一个账号: %q != %q", second.User.UID, session.User.UID)
	}
	if len(fixture.users.created) != 1 {
		t.Fatalf("同一邮箱不应重复建账号，创建次数 = %d", len(fixture.users.created))
	}
	if len(fixture.oauth.links) != 2 {
		t.Fatalf("绑定条数 = %d, want 2", len(fixture.oauth.links))
	}
}

// TestOAuthLoginVerifiedFeishuEmailSignsIn 确认飞书企业邮箱沿用 emailVerified 例外：
// 有可用邮箱时直接签发会话，不再走二次邮箱验证。
func TestOAuthLoginVerifiedFeishuEmailSignsIn(t *testing.T) {
	const email = "zhangsan@corp.example.com"
	newFeishuStub(t, feishuTokenOK, `{"code":0,"msg":"success","data":{"name":"张三","open_id":"ou_zhangsan","enterprise_email":"zhangsan@corp.example.com"}}`)
	fixture := newAuthFixture(t)
	configureClients(fixture, "FEISHU")

	session, binding, err := fixture.authenticator.OAuthLogin(context.Background(), "FEISHU", "auth-code", oauthRedirectURI)
	if err != nil {
		t.Fatalf("OAuthLogin: %v", err)
	}
	if binding != nil {
		t.Fatalf("飞书企业邮箱是可用邮箱，不应要求二次验证: %+v", binding)
	}
	if session == nil || session.User.Email != email {
		t.Fatalf("session = %+v, want email %q", session, email)
	}
	if len(fixture.users.created) != 1 {
		t.Fatalf("创建用户次数 = %d, want 1", len(fixture.users.created))
	}
	if len(fixture.oauth.links) != 1 || fixture.oauth.links[0].userUID != session.User.UID {
		t.Fatalf("绑定 = %+v", fixture.oauth.links)
	}
}

// TestOAuthLoginUnverifiedEmailRequiresBinding 覆盖第三方返回了邮箱但未验证的情况：
// 即使该邮箱已属于另一个本地账号，也不能并入，只回填建议邮箱并返回待绑定凭证，
// 不再返回旧的 "oauth email is not verified" 错误。
func TestOAuthLoginUnverifiedEmailRequiresBinding(t *testing.T) {
	const email = "taken@example.com"
	newDingtalkStub(t, dingtalkReply{
		tokenResponse:    `{"accessToken":"u-dingtalk-token"}`,
		userInfoResponse: `{"nick":"张三","openId":"open-zhangsan","email":"taken@example.com"}`,
	})
	fixture := newAuthFixture(t)
	configureClients(fixture, "DINGTALK")
	fixture.addAccount("user-1", email)

	session, binding, err := fixture.authenticator.OAuthLogin(context.Background(), "DINGTALK", "auth-code", oauthRedirectURI)
	if err != nil {
		t.Fatalf("OAuthLogin: %v", err)
	}
	if session != nil || binding == nil {
		t.Fatalf("未验证邮箱不能直接登录: session=%v binding=%v", session, binding)
	}
	if binding.SuggestedEmail != email {
		t.Fatalf("suggested_email = %q, want %q", binding.SuggestedEmail, email)
	}
	claims, err := fixture.authenticator.parseOAuthBinding(binding.Token)
	if err != nil {
		t.Fatalf("parseOAuthBinding: %v", err)
	}
	if claims.PendingUserUID != "" {
		t.Fatalf("未绑定账号不应带 pending_user_uid: %q", claims.PendingUserUID)
	}
	if len(fixture.oauth.links) != 0 {
		t.Fatal("未验证邮箱不能写入第三方绑定")
	}
}

// TestOAuthLoginPlaceholderAccountRequiresEmail 覆盖历史占位账号：
// 下次第三方登录仍要求补邮箱；补的邮箱没被占用时直接改该账号邮箱，保留 UID，不新建用户。
func TestOAuthLoginPlaceholderAccountRequiresEmail(t *testing.T) {
	const (
		pendingUID    = "pending-1"
		placeholder   = "DINGTALK-OPEN-ZHANGSAN@USERS.ALIOTH.LOCAL"
		verifiedEmail = "real@example.com"
	)
	newDingtalkStub(t, dingtalkReply{
		tokenResponse:    `{"accessToken":"u-dingtalk-token"}`,
		userInfoResponse: `{"nick":"张三","openId":"open-zhangsan"}`,
	})
	fixture := newAuthFixture(t)
	configureClients(fixture, "DINGTALK")
	fixture.addAccount(pendingUID, placeholder)
	fixture.oauth.addLink(pendingUID, "DINGTALK", "open-zhangsan", placeholder)

	session, binding, err := fixture.authenticator.OAuthLogin(context.Background(), "DINGTALK", "auth-code", oauthRedirectURI)
	if err != nil {
		t.Fatalf("OAuthLogin: %v", err)
	}
	if session != nil || binding == nil {
		t.Fatalf("占位邮箱账号必须先补邮箱: session=%v binding=%v", session, binding)
	}
	// 占位邮箱后缀大小写不敏感，且不能作为建议邮箱回填给用户。
	if binding.SuggestedEmail != "" {
		t.Fatalf("占位邮箱不能作为 suggested_email: %q", binding.SuggestedEmail)
	}
	claims, err := fixture.authenticator.parseOAuthBinding(binding.Token)
	if err != nil {
		t.Fatalf("parseOAuthBinding: %v", err)
	}
	if claims.PendingUserUID != pendingUID {
		t.Fatalf("pending_user_uid = %q, want %q", claims.PendingUserUID, pendingUID)
	}

	fixture.solveCaptcha(t, "captcha-1", "12")
	saveEmailCode(fixture, verifiedEmail, "123456")
	session, err = fixture.authenticator.BindOAuthEmail(context.Background(), binding.Token, verifiedEmail, "123456", "captcha-1", "12")
	if err != nil {
		t.Fatalf("BindOAuthEmail: %v", err)
	}
	if session.User.UID != pendingUID || session.User.Email != verifiedEmail {
		t.Fatalf("会话账号 = %+v，应保留占位账号 UID 并更新邮箱", session.User)
	}
	if len(fixture.users.created) != 0 {
		t.Fatal("邮箱没被占用时不应新建用户")
	}
	if len(fixture.users.merged) != 0 {
		t.Fatal("邮箱没被占用时不应合并账号")
	}
	link, ok := fixture.oauth.find("DINGTALK", "open-zhangsan")
	if !ok {
		t.Fatal("绑定行丢失")
	}
	if link.userUID != pendingUID || link.email != verifiedEmail {
		t.Fatalf("绑定 = %+v，应保留原账号并同步验证邮箱", link)
	}
}

// TestOAuthBindMergesPlaceholderIntoEmailAccount 覆盖补的邮箱已被另一个账号占用：
// 调用 MergeAccounts 以邮箱账号为母本全量合并（真库会把绑定一起并过去），
// 再把本次绑定邮箱同步到邮箱账号，会话属于邮箱账号。
// 重复提交同一份待绑定凭证仍应成功，不重复合并、不报服务端错误。
func TestOAuthBindMergesPlaceholderIntoEmailAccount(t *testing.T) {
	const (
		pendingUID = "pending-1"
		targetUID  = "user-1"
		targetMail = "owner@example.com"
	)
	newDingtalkStub(t, dingtalkReply{
		tokenResponse:    `{"accessToken":"u-dingtalk-token"}`,
		userInfoResponse: `{"nick":"张三","openId":"open-zhangsan"}`,
	})
	fixture := newAuthFixture(t)
	configureClients(fixture, "DINGTALK")
	fixture.addAccount(pendingUID, "dingtalk-open-zhangsan@users.alioth.local")
	fixture.addAccount(targetUID, targetMail)
	fixture.oauth.addLink(pendingUID, "DINGTALK", "open-zhangsan", "dingtalk-open-zhangsan@users.alioth.local")

	_, binding, err := fixture.authenticator.OAuthLogin(context.Background(), "DINGTALK", "auth-code", oauthRedirectURI)
	if err != nil {
		t.Fatalf("OAuthLogin: %v", err)
	}
	if binding == nil {
		t.Fatal("占位账号必须补邮箱")
	}
	fixture.solveCaptcha(t, "captcha-1", "12")
	saveEmailCode(fixture, targetMail, "123456")
	session, err := fixture.authenticator.BindOAuthEmail(context.Background(), binding.Token, targetMail, "123456", "captcha-1", "12")
	if err != nil {
		t.Fatalf("BindOAuthEmail: %v", err)
	}
	if session.User.UID != targetUID {
		t.Fatalf("会话账号 UID = %q, want %q", session.User.UID, targetUID)
	}
	if len(fixture.users.merged) != 1 || fixture.users.merged[0] != [2]string{pendingUID, targetUID} {
		t.Fatalf("MergeAccounts 调用 = %v, want [%s %s]", fixture.users.merged, pendingUID, targetUID)
	}
	if len(fixture.users.created) != 0 {
		t.Fatal("邮箱已被占用时不应新建用户")
	}
	link, ok := fixture.oauth.find("DINGTALK", "open-zhangsan")
	if !ok || link.userUID != targetUID || link.email != targetMail {
		t.Fatalf("绑定 = %+v，应改绑到邮箱账号并同步邮箱", link)
	}

	// 重复提交（网络重试/刷新页面）：同一凭证再来一次，邮箱码和图形验证码都是新的。
	fixture.solveCaptcha(t, "captcha-2", "12")
	saveEmailCode(fixture, targetMail, "654321")
	repeated, err := fixture.authenticator.BindOAuthEmail(context.Background(), binding.Token, targetMail, "654321", "captcha-2", "12")
	if err != nil {
		t.Fatalf("重复提交应幂等成功: %v", err)
	}
	if repeated.User.UID != targetUID {
		t.Fatalf("重复提交会话账号 = %q, want %q", repeated.User.UID, targetUID)
	}
	if len(fixture.users.merged) != 1 {
		t.Fatalf("重复提交不应再次合并账号，MergeAccounts 调用 = %v", fixture.users.merged)
	}
	if len(fixture.oauth.links) != 1 {
		t.Fatalf("重复提交不应新增绑定，条数 = %d", len(fixture.oauth.links))
	}
}

// TestBindOAuthEmailRepeatedSubmitIsIdempotent 覆盖新建账号分支的重复提交：
// 第一次已经把 (provider, open_id) 绑到新账号，再跑一次不能再走 LinkOAuth 撞唯一索引。
func TestBindOAuthEmailRepeatedSubmitIsIdempotent(t *testing.T) {
	const bindEmail = "new@example.com"
	newDingtalkStub(t, dingtalkReply{
		tokenResponse:    `{"accessToken":"u-dingtalk-token"}`,
		userInfoResponse: `{"nick":"张三","openId":"open-zhangsan"}`,
	})
	fixture := newAuthFixture(t)
	configureClients(fixture, "DINGTALK")

	_, binding, err := fixture.authenticator.OAuthLogin(context.Background(), "DINGTALK", "auth-code", oauthRedirectURI)
	if err != nil || binding == nil {
		t.Fatalf("OAuthLogin: binding=%v err=%v", binding, err)
	}
	fixture.solveCaptcha(t, "captcha-1", "12")
	saveEmailCode(fixture, bindEmail, "123456")
	first, err := fixture.authenticator.BindOAuthEmail(context.Background(), binding.Token, bindEmail, "123456", "captcha-1", "12")
	if err != nil {
		t.Fatalf("BindOAuthEmail: %v", err)
	}
	fixture.solveCaptcha(t, "captcha-2", "12")
	saveEmailCode(fixture, bindEmail, "654321")
	second, err := fixture.authenticator.BindOAuthEmail(context.Background(), binding.Token, bindEmail, "654321", "captcha-2", "12")
	if err != nil {
		t.Fatalf("重复提交应幂等成功: %v", err)
	}
	if second.User.UID != first.User.UID {
		t.Fatalf("重复提交会话账号 = %q, want %q", second.User.UID, first.User.UID)
	}
	if len(fixture.users.created) != 1 {
		t.Fatalf("重复提交不应重复建账号，创建次数 = %d", len(fixture.users.created))
	}
	if len(fixture.oauth.links) != 1 {
		t.Fatalf("重复提交不应新增绑定，条数 = %d", len(fixture.oauth.links))
	}
}

// TestOAuthBindTokenValidation 覆盖待绑定凭证的校验：非法、被篡改、类型不符（会话令牌）
// 和已过期都必须按凭证失效拒绝，且在验凭证阶段不消耗图形验证码和邮箱验证码。
func TestOAuthBindTokenValidation(t *testing.T) {
	fixture := newAuthFixture(t)
	valid, err := fixture.authenticator.signOAuthBinding(oauthBindClaims{
		Provider: "DINGTALK", OpenID: "open-zhangsan", Name: "张三",
	})
	if err != nil {
		t.Fatalf("signOAuthBinding: %v", err)
	}
	expired := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"kind":     oauthBindKind,
		"provider": "DINGTALK",
		"open_id":  "open-zhangsan",
		"iat":      time.Now().Add(-2 * time.Hour).Unix(),
		"exp":      time.Now().Add(-time.Hour).Unix(),
	})
	expiredToken, err := expired.SignedString([]byte(fixture.authenticator.secret))
	if err != nil {
		t.Fatalf("sign expired token: %v", err)
	}
	accessToken, err := fixture.authenticator.sign("user-1", "access", time.Hour)
	if err != nil {
		t.Fatalf("sign access token: %v", err)
	}
	cases := map[string]string{
		"空凭证":    "",
		"随机字符串":  "not-a-token",
		"被篡改":    tamperJWT(valid),
		"已过期":    expiredToken,
		"会话令牌冒充": accessToken,
	}
	for name, token := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := fixture.authenticator.BindOAuthEmail(context.Background(), token, "user@example.com", "123456", "captcha-1", "12")
			assertClientMessage(t, err, "oauth binding is invalid or expired")
			err = fixture.authenticator.SendOAuthBindEmailCode(context.Background(), token, "user@example.com", "captcha-1", "12")
			assertClientMessage(t, err, "oauth binding is invalid or expired")
			if fixture.captchas.checks != 0 {
				t.Fatal("凭证无效时不应校验图形验证码")
			}
			if fixture.emailCodes.consumed != 0 {
				t.Fatal("凭证无效时不应消费邮箱验证码")
			}
		})
	}

	t.Run("token expires in 15 minutes", func(t *testing.T) {
		parsed, err := jwt.Parse(valid, func(token *jwt.Token) (interface{}, error) {
			return []byte(fixture.authenticator.secret), nil
		})
		if err != nil || !parsed.Valid {
			t.Fatalf("parse valid binding token: %v", err)
		}
		claims := parsed.Claims.(jwt.MapClaims)
		issuedAt, _ := claims["iat"].(float64)
		expiresAt, _ := claims["exp"].(float64)
		if lifetime := time.Duration(expiresAt-issuedAt) * time.Second; lifetime != oauthBindTTL {
			t.Fatalf("待绑定凭证有效期 = %v, want %v", lifetime, oauthBindTTL)
		}
	})
}

// TestBindOAuthEmailValidationOrder 覆盖绑定提交时的失败顺序：
// 图形验证码错误不消费邮箱验证码；邮箱验证码错误不消费图形验证码，两种情况都不应创建用户或绑定。
func TestBindOAuthEmailValidationOrder(t *testing.T) {
	const email = "bind@example.com"
	newBinding := func(t *testing.T) (*authFixture, string) {
		t.Helper()
		fixture := newAuthFixture(t)
		token, err := fixture.authenticator.signOAuthBinding(oauthBindClaims{
			Provider: "DINGTALK", OpenID: "open-zhangsan", Name: "张三",
		})
		if err != nil {
			t.Fatalf("signOAuthBinding: %v", err)
		}
		return fixture, token
	}

	t.Run("rejects wrong captcha without consuming email code", func(t *testing.T) {
		fixture, token := newBinding(t)
		fixture.solveCaptcha(t, "captcha-1", "12")
		saveEmailCode(fixture, email, "123456")
		_, err := fixture.authenticator.BindOAuthEmail(context.Background(), token, email, "123456", "captcha-1", "99")
		assertClientMessage(t, err, "captcha is invalid")
		if fixture.emailCodes.consumed != 0 {
			t.Fatal("图形验证码未通过时不应消费邮箱验证码")
		}
		if len(fixture.users.created) != 0 || len(fixture.oauth.links) != 0 {
			t.Fatal("校验失败时不应创建用户或绑定")
		}
	})

	t.Run("rejects wrong email code without consuming captcha", func(t *testing.T) {
		fixture, token := newBinding(t)
		fixture.solveCaptcha(t, "captcha-1", "12")
		saveEmailCode(fixture, email, "123456")
		_, err := fixture.authenticator.BindOAuthEmail(context.Background(), token, email, "000000", "captcha-1", "12")
		assertClientMessage(t, err, "invalid email or code")
		if fixture.captchas.used["captcha-1"] {
			t.Fatal("邮箱验证码错误时图形验证码应保留，方便用户重试")
		}
		if len(fixture.users.created) != 0 || len(fixture.oauth.links) != 0 {
			t.Fatal("校验失败时不应创建用户或绑定")
		}
	})
}

// TestOAuthLoginLinkedAccountSignsIn 确认已绑定且邮箱不是占位地址的账号直接签发会话。
func TestOAuthLoginLinkedAccountSignsIn(t *testing.T) {
	const (
		userUID = "user-1"
		email   = "linked@example.com"
	)
	newDingtalkStub(t, dingtalkReply{
		tokenResponse:    `{"accessToken":"u-dingtalk-token"}`,
		userInfoResponse: `{"nick":"张三","openId":"open-zhangsan"}`,
	})
	fixture := newAuthFixture(t)
	configureClients(fixture, "DINGTALK")
	fixture.addAccount(userUID, email)
	fixture.oauth.addLink(userUID, "DINGTALK", "open-zhangsan", email)

	session, binding, err := fixture.authenticator.OAuthLogin(context.Background(), "DINGTALK", "auth-code", oauthRedirectURI)
	if err != nil {
		t.Fatalf("OAuthLogin: %v", err)
	}
	if binding != nil {
		t.Fatalf("已绑定且邮箱可用时不应要求二次验证: %+v", binding)
	}
	if session == nil || session.User.UID != userUID {
		t.Fatalf("session = %+v, want uid %q", session, userUID)
	}
}

// TestSendOAuthBindEmailCode 覆盖绑定发码的图形验证码失败、60 秒节流和成功路径。
// 绑定发码不要求邮箱已注册：邮箱空闲时会新建账号。
func TestSendOAuthBindEmailCode(t *testing.T) {
	const email = "bind@example.com"
	newBindingToken := func(t *testing.T, fixture *authFixture) string {
		t.Helper()
		token, err := fixture.authenticator.signOAuthBinding(oauthBindClaims{
			Provider: "DINGTALK", OpenID: "open-zhangsan", Name: "张三",
		})
		if err != nil {
			t.Fatalf("signOAuthBinding: %v", err)
		}
		return token
	}

	t.Run("rejects wrong captcha", func(t *testing.T) {
		fixture := newAuthFixture(t)
		token := newBindingToken(t, fixture)
		fixture.solveCaptcha(t, "captcha-1", "12")
		err := fixture.authenticator.SendOAuthBindEmailCode(context.Background(), token, email, "captcha-1", "99")
		assertClientMessage(t, err, "captcha is invalid")
		if len(fixture.mail.subjects) != 0 || len(fixture.emailCodes.codes) != 0 {
			t.Fatal("图形验证码未通过时不能存码或发信")
		}
		if fixture.emailCodes.issuedCalls != 0 {
			t.Fatal("图形验证码未通过时不应进入节流检查")
		}
	})

	t.Run("rejects repeated send within 60 seconds", func(t *testing.T) {
		fixture := newAuthFixture(t)
		token := newBindingToken(t, fixture)
		fixture.solveCaptcha(t, "captcha-1", "12")
		fixture.emailCodes.issuedAt[email] = time.Now().Add(-10 * time.Second)
		err := fixture.authenticator.SendOAuthBindEmailCode(context.Background(), token, email, "captcha-1", "12")
		assertClientMessage(t, err, "send code too frequently")
		if len(fixture.mail.subjects) != 0 {
			t.Fatal("节流命中时不应发信")
		}
		metadata := kerrorsMetadata(t, err)
		seconds, convErr := strconv.Atoi(metadata["retryAfterSeconds"])
		if convErr != nil || seconds < 45 || seconds > 55 {
			t.Fatalf("剩余等待时间 = %q（%v），应约 50 秒", metadata["retryAfterSeconds"], convErr)
		}
	})

	t.Run("sends bind code without registered email", func(t *testing.T) {
		fixture := newAuthFixture(t)
		token := newBindingToken(t, fixture)
		fixture.solveCaptcha(t, "captcha-1", "12")
		if err := fixture.authenticator.SendOAuthBindEmailCode(context.Background(), token, email, "captcha-1", "12"); err != nil {
			t.Fatalf("SendOAuthBindEmailCode: %v", err)
		}
		if len(fixture.mail.subjects) != 1 {
			t.Fatalf("发信次数 = %d, want 1", len(fixture.mail.subjects))
		}
		if !strings.Contains(fixture.mail.subjects[0], "bind") {
			t.Fatalf("绑定码邮件主题 = %q", fixture.mail.subjects[0])
		}
		if _, ok := fixture.emailCodes.codes[email]; !ok {
			t.Fatal("邮箱验证码没有落库")
		}
		if len(fixture.users.created) != 0 {
			t.Fatal("发码阶段不应创建用户")
		}
	})
}
