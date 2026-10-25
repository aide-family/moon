package auth

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	megrezv1 "github.com/aide-family/Alioth/api/megrez/v1"
	merakv1 "github.com/aide-family/Alioth/api/merak/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/user"
	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
	"github.com/aide-family/Alioth/app/alioth/pkg/actor"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

type userStub struct{ repository.UserRepo }

func (userStub) UserByUID(context.Context, string) (*bo.Account, error) {
	return &bo.Account{UID: "user-1", Status: "ACTIVE"}, nil
}

// userRepoStub 用内存账号表代替数据库，记录创建过的用户。
type userRepoStub struct {
	repository.UserRepo
	byEmail map[string]bo.Account
	created []bo.Account
}

func (stub *userRepoStub) UserByEmail(_ context.Context, email string) (*bo.Account, error) {
	account, ok := stub.byEmail[email]
	if !ok {
		return nil, merr.NotFound("not found")
	}
	copied := account
	return &copied, nil
}

func (stub *userRepoStub) CreateUser(_ context.Context, in bo.Account) (*bo.Account, error) {
	stub.created = append(stub.created, in)
	stub.byEmail[in.Email] = in
	copied := in
	return &copied, nil
}

// captchaRepoStub 用内存表保存图形验证码答案摘要，并记录校验和消费次数。
type captchaRepoStub struct {
	repository.CaptchaRepo
	answers      map[string]string
	used         map[string]bool
	checks       int
	consumes     int
	savedExpires time.Time
}

func newCaptchaRepoStub() *captchaRepoStub {
	return &captchaRepoStub{answers: map[string]string{}, used: map[string]bool{}}
}

func (stub *captchaRepoStub) SaveCaptcha(_ context.Context, id, answerHash string, expiresAt time.Time) error {
	stub.answers[id] = answerHash
	stub.savedExpires = expiresAt
	return nil
}

func (stub *captchaRepoStub) CheckCaptcha(_ context.Context, id, answerHash string, _ time.Time) (bool, error) {
	stub.checks++
	if stub.used[id] {
		return false, nil
	}
	return stub.answers[id] == answerHash, nil
}

func (stub *captchaRepoStub) ConsumeCaptcha(_ context.Context, id, answerHash string, _ time.Time) (bool, error) {
	stub.consumes++
	if stub.used[id] || stub.answers[id] != answerHash {
		return false, nil
	}
	stub.used[id] = true
	return true, nil
}

// emailCodeRepoStub 保存邮箱验证码摘要。issuedAt 由测试控制，用来模拟 60 秒节流命中。
type emailCodeRepoStub struct {
	repository.EmailCodeRepo
	codes       map[string]string
	issuedAt    map[string]time.Time
	issuedCalls int
	consumed    int
}

func newEmailCodeRepoStub() *emailCodeRepoStub {
	return &emailCodeRepoStub{codes: map[string]string{}, issuedAt: map[string]time.Time{}}
}

func (stub *emailCodeRepoStub) SaveEmailCode(_ context.Context, email, codeHash string, _ time.Time) error {
	stub.codes[email] = codeHash
	return nil
}

func (stub *emailCodeRepoStub) ConsumeEmailCode(_ context.Context, email, codeHash string, _ time.Time) (bool, error) {
	if stub.codes[email] != codeHash {
		return false, nil
	}
	delete(stub.codes, email)
	stub.consumed++
	return true, nil
}

func (stub *emailCodeRepoStub) LastEmailCodeIssuedAt(_ context.Context, email string) (time.Time, bool, error) {
	stub.issuedCalls++
	issuedAt, ok := stub.issuedAt[email]
	return issuedAt, ok, nil
}

// codeMailerStub 记录发出的验证码邮件，测试不连 SMTP。
type codeMailerStub struct {
	recipients []string
	subjects   []string
	bodies     []string
}

func (stub *codeMailerStub) Send(to, subject, body string) error {
	stub.recipients = append(stub.recipients, to)
	stub.subjects = append(stub.subjects, subject)
	stub.bodies = append(stub.bodies, body)
	return nil
}

type authFixture struct {
	authenticator *Authenticator
	users         *userRepoStub
	captchas      *captchaRepoStub
	emailCodes    *emailCodeRepoStub
	mail          *codeMailerStub
}

func newAuthFixture(t *testing.T) *authFixture {
	t.Helper()
	users := &userRepoStub{byEmail: map[string]bo.Account{}}
	captchas := newCaptchaRepoStub()
	emailCodes := newEmailCodeRepoStub()
	mail := &codeMailerStub{}
	authenticator, err := NewAuthenticator(users, captchas, emailCodes, nil, nil, nil, &conf.Auth{JwtSecret: "test-secret"}, slog.Default())
	if err != nil {
		t.Fatalf("NewAuthenticator: %v", err)
	}
	authenticator.mail = mail
	return &authFixture{authenticator: authenticator, users: users, captchas: captchas, emailCodes: emailCodes, mail: mail}
}

// solveCaptcha 模拟用户解题：把图形验证码答案摘要写进桩仓储。
func (fixture *authFixture) solveCaptcha(t *testing.T, captchaID, answer string) {
	t.Helper()
	if err := fixture.captchas.SaveCaptcha(context.Background(), captchaID, fixture.authenticator.digest(answer), time.Now().Add(captchaTTL)); err != nil {
		t.Fatalf("SaveCaptcha: %v", err)
	}
}

func (fixture *authFixture) addUser(email string) {
	fixture.users.byEmail[email] = bo.Account{UID: "user-1", Email: email, Status: "ACTIVE"}
}

func assertClientMessage(t *testing.T, err error, message string) {
	t.Helper()
	if kerrors.Reason(err) != merr.ReasonClient {
		t.Fatalf("err = %v, want client error", err)
	}
	if got := kerrors.FromError(err).Message; got != message {
		t.Fatalf("message = %q, want %q", got, message)
	}
}

func TestClaimRejectsUserJWT(t *testing.T) {
	authenticator, err := NewAuthenticator(userStub{}, nil, nil, nil, nil, nil, &conf.Auth{
		JwtSecret:   "test-secret",
		ServiceKeys: []*conf.Auth_ServiceKey{{Name: "megrez", Key: "sk-megrez-dev"}},
	}, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	token, err := authenticator.sign("user-1", "access", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	_, err = authenticator.Authenticate(context.Background(), megrezv1.EvalService_ClaimRules_FullMethodName, "Bearer "+token, "")
	if kerrors.Reason(err) != merr.ReasonForbidden {
		t.Fatalf("user jwt claim: %v", err)
	}
	ctx, err := authenticator.Authenticate(context.Background(), megrezv1.EvalService_ReportEval_FullMethodName, "Bearer sk-megrez-dev", "")
	if err != nil {
		t.Fatal(err)
	}
	name, ok := actor.ServiceName(ctx)
	if !ok || name != "megrez" {
		t.Fatalf("service = %q ok=%v", name, ok)
	}
	statusCtx, err := authenticator.Authenticate(context.Background(), megrezv1.EvalService_ReportWorkerStatus_FullMethodName, "Bearer sk-megrez-dev", "")
	if err != nil {
		t.Fatalf("megrez status report: %v", err)
	}
	if name, ok := actor.ServiceName(statusCtx); !ok || name != "megrez" {
		t.Fatalf("status reporter = %q ok=%v", name, ok)
	}
	_, err = authenticator.Authenticate(context.Background(), merakv1.NotifyService_ReportWorkerStatus_FullMethodName, "Bearer sk-megrez-dev", "")
	if kerrors.Reason(err) != merr.ReasonForbidden {
		t.Fatalf("megrez key reached merak heartbeat: %v", err)
	}
	_, err = authenticator.Authenticate(context.Background(), v1.OperationAlertEventServiceListAlertEvents, "Bearer sk-megrez-dev", "ns")
	if kerrors.Reason(err) != merr.ReasonForbidden {
		t.Fatalf("service key list: %v", err)
	}
}

func TestGetCaptcha(t *testing.T) {
	fixture := newAuthFixture(t)
	id, image, kind, err := fixture.authenticator.GetCaptcha(context.Background())
	if err != nil {
		t.Fatalf("GetCaptcha: %v", err)
	}
	if id == "" {
		t.Fatal("图形验证码标识为空")
	}
	if kind != captchaKindNumeric && kind != captchaKindMath {
		t.Fatalf("图形验证码题型 = %q", kind)
	}
	if !strings.HasPrefix(image, "data:image/png;base64,") {
		t.Fatalf("图片不是 data URL: %q", image[:min(len(image), 32)])
	}
	if _, ok := fixture.captchas.answers[id]; !ok {
		t.Fatal("答案摘要没有落库")
	}
	expiresAt := fixture.captchas.savedExpires
	if expiresAt.Before(time.Now().Add(4*time.Minute)) || expiresAt.After(time.Now().Add(6*time.Minute)) {
		t.Fatalf("过期时间 = %v，应为 5 分钟后", expiresAt)
	}
}

func TestOAuthProviders(t *testing.T) {
	const (
		redirectURI = "https://app.example.com/oauth2/callback"
		state       = "state-123456"
	)
	configureBoth := func() *Authenticator {
		fixture := newAuthFixture(t)
		fixture.authenticator.oauthClients = map[string]oauthClient{
			"GITHUB": {id: "github-id", secret: "github-secret"},
			"GITEE":  {id: "gitee-id", secret: "gitee-secret"},
		}
		return fixture.authenticator
	}

	t.Run("lists both configured providers", func(t *testing.T) {
		providers, err := configureBoth().OAuthProviders(redirectURI, state)
		if err != nil {
			t.Fatalf("OAuthProviders: %v", err)
		}
		if len(providers) != 2 {
			t.Fatalf("providers = %d, want 2", len(providers))
		}
		github := providers[0]
		if github.Provider != "GITHUB" || github.Name != "GitHub" {
			t.Fatalf("github = %+v", github)
		}
		if !strings.HasPrefix(github.AuthorizeURL, "https://github.com/login/oauth/authorize?") {
			t.Fatalf("github authorize url = %q", github.AuthorizeURL)
		}
		for _, want := range []string{"client_id=github-id", "state=state-123456", "allow_signup=true", "read%3Auser%20user%3Aemail", "redirect_uri="} {
			if !strings.Contains(github.AuthorizeURL, want) {
				t.Fatalf("github authorize url 缺少 %q: %s", want, github.AuthorizeURL)
			}
		}
		gitee := providers[1]
		if gitee.Provider != "GITEE" || gitee.Name != "Gitee" {
			t.Fatalf("gitee = %+v", gitee)
		}
		if !strings.HasPrefix(gitee.AuthorizeURL, "https://gitee.com/oauth/authorize?") {
			t.Fatalf("gitee authorize url = %q", gitee.AuthorizeURL)
		}
		for _, want := range []string{"client_id=gitee-id", "state=state-123456", "response_type=code", "scope=user_info%20emails"} {
			if !strings.Contains(gitee.AuthorizeURL, want) {
				t.Fatalf("gitee authorize url 缺少 %q: %s", want, gitee.AuthorizeURL)
			}
		}
		// Gitee 授权页不把 "+" 当空格，空格必须是 %20（见 authorizeURL 注释）。
		for _, provider := range []bo.OAuthProvider{github, gitee} {
			if strings.Contains(provider.AuthorizeURL, "+") {
				t.Fatalf("%s authorize url 不应包含 +: %s", provider.Provider, provider.AuthorizeURL)
			}
		}
	})

	t.Run("lists feishu with email scope", func(t *testing.T) {
		// redirect_uri 带空格，用来确认空格被编码成 %20 而不是 url.Values 默认的 "+"。
		const feishuRedirectURI = "https://app.example.com/oauth2/callback?next=some page"
		authenticator := configureBoth()
		authenticator.oauthClients["FEISHU"] = oauthClient{id: "feishu-id", secret: "feishu-secret"}
		providers, err := authenticator.OAuthProviders(feishuRedirectURI, state)
		if err != nil {
			t.Fatalf("OAuthProviders: %v", err)
		}
		if len(providers) != 3 {
			t.Fatalf("providers = %d, want 3", len(providers))
		}
		feishu := providers[2]
		if feishu.Provider != "FEISHU" || feishu.Name != "飞书" {
			t.Fatalf("feishu = %+v", feishu)
		}
		if !strings.HasPrefix(feishu.AuthorizeURL, "https://accounts.feishu.cn/open-apis/authen/v1/authorize?") {
			t.Fatalf("feishu authorize url = %q", feishu.AuthorizeURL)
		}
		for _, want := range []string{
			"client_id=feishu-id",
			"state=state-123456",
			"response_type=code",
			"scope=contact%3Auser.email%3Areadonly",
			"redirect_uri=https%3A%2F%2Fapp.example.com%2Foauth2%2Fcallback%3Fnext%3Dsome%20page",
		} {
			if !strings.Contains(feishu.AuthorizeURL, want) {
				t.Fatalf("feishu authorize url 缺少 %q: %s", want, feishu.AuthorizeURL)
			}
		}
		// 飞书授权页不把 "+" 当空格，空格必须是 %20（见 authorizeURL 注释）。
		if strings.Contains(feishu.AuthorizeURL, "+") {
			t.Fatalf("feishu authorize url 不应包含 +: %s", feishu.AuthorizeURL)
		}
	})

	t.Run("omits feishu without secret", func(t *testing.T) {
		authenticator := configureBoth()
		authenticator.oauthClients["FEISHU"] = oauthClient{id: "feishu-id"}
		providers, err := authenticator.OAuthProviders(redirectURI, state)
		if err != nil {
			t.Fatalf("OAuthProviders: %v", err)
		}
		for _, provider := range providers {
			if provider.Provider == "FEISHU" {
				t.Fatalf("缺少 client_secret 的飞书不应出现在来源列表: %+v", provider)
			}
		}
	})

	t.Run("lists only github when gitee has no secret", func(t *testing.T) {
		authenticator := configureBoth()
		authenticator.oauthClients["GITEE"] = oauthClient{id: "gitee-id"}
		providers, err := authenticator.OAuthProviders(redirectURI, state)
		if err != nil {
			t.Fatalf("OAuthProviders: %v", err)
		}
		if len(providers) != 1 || providers[0].Provider != "GITHUB" {
			t.Fatalf("providers = %+v, want only GITHUB", providers)
		}
	})

	t.Run("returns empty when nothing configured", func(t *testing.T) {
		fixture := newAuthFixture(t)
		providers, err := fixture.authenticator.OAuthProviders(redirectURI, state)
		if err != nil {
			t.Fatalf("OAuthProviders: %v", err)
		}
		if len(providers) != 0 {
			t.Fatalf("providers = %+v, want empty", providers)
		}
	})

	t.Run("rejects non http redirect uri", func(t *testing.T) {
		_, err := configureBoth().OAuthProviders("app://callback", state)
		assertClientMessage(t, err, "invalid redirect uri")
	})
}

// feishuStub 用 httptest 代替飞书 v3 令牌接口和 user_info 接口，记录收到的请求并按预设 JSON 回复。
// 两个端点变量在 Cleanup 里恢复，测试不访问真实飞书。
type feishuStub struct {
	mu               sync.Mutex
	tokenResponse    string
	userInfoResponse string
	tokenMethod      string
	tokenContentType string
	tokenForm        url.Values
	userInfoMethod   string
	userInfoAuth     string
}

func newFeishuStub(t *testing.T, tokenResponse, userInfoResponse string) *feishuStub {
	t.Helper()
	stub := &feishuStub{tokenResponse: tokenResponse, userInfoResponse: userInfoResponse}
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("飞书令牌请求不是合法表单: %v", err)
		}
		stub.recordToken(r.Method, r.Header.Get("Content-Type"), r.PostForm)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(stub.tokenResponse))
	}))
	t.Cleanup(tokenServer.Close)
	userInfoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stub.recordUserInfo(r.Method, r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(stub.userInfoResponse))
	}))
	t.Cleanup(userInfoServer.Close)
	restoreToken, restoreUserInfo := feishuTokenEndpoint, feishuUserInfoEndpoint
	feishuTokenEndpoint, feishuUserInfoEndpoint = tokenServer.URL, userInfoServer.URL
	t.Cleanup(func() {
		feishuTokenEndpoint, feishuUserInfoEndpoint = restoreToken, restoreUserInfo
	})
	return stub
}

func (stub *feishuStub) recordToken(method, contentType string, form url.Values) {
	stub.mu.Lock()
	defer stub.mu.Unlock()
	stub.tokenMethod = method
	stub.tokenContentType = contentType
	stub.tokenForm = form
}

func (stub *feishuStub) recordUserInfo(method, authorization string) {
	stub.mu.Lock()
	defer stub.mu.Unlock()
	stub.userInfoMethod = method
	stub.userInfoAuth = authorization
}

func (stub *feishuStub) tokenRequest() (string, string, url.Values) {
	stub.mu.Lock()
	defer stub.mu.Unlock()
	return stub.tokenMethod, stub.tokenContentType, stub.tokenForm
}

func (stub *feishuStub) userInfoRequest() (string, string) {
	stub.mu.Lock()
	defer stub.mu.Unlock()
	return stub.userInfoMethod, stub.userInfoAuth
}

// TestExchangeFeishu 覆盖飞书 v3 令牌交换和 user_info 读取，端点都指向 httptest，不访问真实飞书。
func TestExchangeFeishu(t *testing.T) {
	const (
		feishuClientID     = "feishu-id"
		feishuClientSecret = "feishu-secret"
		feishuCode         = "auth-code"
		feishuRedirectURI  = "https://app.example.com/oauth2/callback"
		feishuAccessToken  = "u-feishu-token"
		successToken       = `{"code":0,"msg":"success","access_token":"u-feishu-token","token_type":"Bearer"}`
	)
	client := oauthClient{id: feishuClientID, secret: feishuClientSecret}
	exchange := func(t *testing.T) (oauthProfile, error) {
		t.Helper()
		return newAuthFixture(t).authenticator.exchangeFeishu(context.Background(), client, feishuCode, feishuRedirectURI)
	}

	t.Run("exchanges code and reads profile", func(t *testing.T) {
		stub := newFeishuStub(t, successToken, `{"code":0,"msg":"success","data":{"name":"张三","en_name":"Zhang San","avatar_url":"https://avatar.example.com/zhangsan.png","open_id":"ou_zhangsan","union_id":"on_zhangsan","email":"zhangsan@example.com","enterprise_email":"zhangsan@corp.example.com"}}`)
		profile, err := exchange(t)
		if err != nil {
			t.Fatalf("exchangeFeishu: %v", err)
		}
		method, contentType, form := stub.tokenRequest()
		if method != http.MethodPost {
			t.Fatalf("令牌请求方法 = %q, want POST", method)
		}
		if contentType != "application/x-www-form-urlencoded" {
			t.Fatalf("令牌请求 Content-Type = %q", contentType)
		}
		for field, want := range map[string]string{
			"grant_type":    "authorization_code",
			"client_id":     feishuClientID,
			"client_secret": feishuClientSecret,
			"code":          feishuCode,
			"redirect_uri":  feishuRedirectURI,
		} {
			if got := form.Get(field); got != want {
				t.Fatalf("令牌请求 %s = %q, want %q", field, got, want)
			}
		}
		userInfoMethod, authorization := stub.userInfoRequest()
		if userInfoMethod != http.MethodGet {
			t.Fatalf("user_info 请求方法 = %q, want GET", userInfoMethod)
		}
		if authorization != "Bearer "+feishuAccessToken {
			t.Fatalf("user_info Authorization = %q", authorization)
		}
		if profile.openID != "ou_zhangsan" || profile.email != "zhangsan@example.com" || !profile.emailVerified {
			t.Fatalf("profile = %+v", profile)
		}
		if profile.name != "张三" || profile.avatar != "https://avatar.example.com/zhangsan.png" {
			t.Fatalf("profile = %+v", profile)
		}
	})

	t.Run("falls back to enterprise email and english name", func(t *testing.T) {
		newFeishuStub(t, successToken, `{"code":0,"msg":"success","data":{"en_name":"Zhang San","open_id":"ou_zhangsan","email":"","enterprise_email":"zhangsan@corp.example.com"}}`)
		profile, err := exchange(t)
		if err != nil {
			t.Fatalf("exchangeFeishu: %v", err)
		}
		if profile.email != "zhangsan@corp.example.com" || !profile.emailVerified {
			t.Fatalf("email = %q, emailVerified = %v，应退回 enterprise_email 并视为已确认", profile.email, profile.emailVerified)
		}
		if profile.name != "Zhang San" {
			t.Fatalf("name = %q，应退回 en_name", profile.name)
		}
	})

	t.Run("treats empty email as unverified", func(t *testing.T) {
		newFeishuStub(t, successToken, `{"code":0,"msg":"success","data":{"name":"张三","open_id":"ou_zhangsan"}}`)
		profile, err := exchange(t)
		if err != nil {
			t.Fatalf("exchangeFeishu: %v", err)
		}
		if profile.email != "" || profile.emailVerified {
			t.Fatalf("email = %q, emailVerified = %v，没有邮箱时不能视为已验证", profile.email, profile.emailVerified)
		}
	})

	t.Run("rejects token error code", func(t *testing.T) {
		newFeishuStub(t, `{"code":20002,"error":"invalid_grant","error_description":"code is invalid"}`, `{}`)
		_, err := exchange(t)
		assertClientMessage(t, err, "oauth login failed")
	})

	t.Run("rejects empty access token", func(t *testing.T) {
		newFeishuStub(t, `{"code":0,"msg":"success","access_token":""}`, `{}`)
		_, err := exchange(t)
		assertClientMessage(t, err, "oauth login failed")
	})

	t.Run("rejects user info error code", func(t *testing.T) {
		newFeishuStub(t, successToken, `{"code":99991663,"msg":"tenant access token is invalid"}`)
		_, err := exchange(t)
		assertClientMessage(t, err, "oauth login failed")
	})

	t.Run("rejects user info without open id", func(t *testing.T) {
		newFeishuStub(t, successToken, `{"code":0,"msg":"success","data":{"name":"张三"}}`)
		_, err := exchange(t)
		assertClientMessage(t, err, "oauth login failed")
	})
}

// TestPickGiteeVerifiedEmail 覆盖 Gitee 邮箱列表的挑选规则：
// 已验证主邮箱优先，其次任意已验证邮箱，未验证项忽略。
func TestPickGiteeVerifiedEmail(t *testing.T) {
	cases := []struct {
		name   string
		emails []giteeEmail
		want   string
	}{
		{name: "空列表", emails: nil, want: ""},
		{
			name: "优先使用已验证主邮箱",
			emails: []giteeEmail{
				{Email: "other@example.com", State: "confirmed", Scope: []string{"secondary"}},
				{Email: "main@example.com", State: "confirmed", Scope: []string{"primary"}},
			},
			want: "main@example.com",
		},
		{
			name: "没有主邮箱时退回已验证邮箱",
			emails: []giteeEmail{
				{Email: "unverified@example.com", State: "unconfirmed", Scope: []string{"primary"}},
				{Email: "ok@example.com", State: "confirmed", Scope: []string{"secondary"}},
			},
			want: "ok@example.com",
		},
		{
			name: "未验证主邮箱不采用",
			emails: []giteeEmail{
				{Email: "unverified@example.com", State: "unconfirmed", Scope: []string{"primary"}},
			},
			want: "",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := pickGiteeVerifiedEmail(testCase.emails); got != testCase.want {
				t.Fatalf("pickGiteeVerifiedEmail() = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestEmailLogin(t *testing.T) {
	const email = "user@example.com"
	setup := func(t *testing.T) *authFixture {
		fixture := newAuthFixture(t)
		fixture.addUser(email)
		fixture.solveCaptcha(t, "captcha-1", "12")
		fixture.emailCodes.codes[email] = fixture.authenticator.digest("123456")
		return fixture
	}

	t.Run("rejects wrong captcha without consuming email code", func(t *testing.T) {
		fixture := setup(t)
		_, err := fixture.authenticator.EmailLogin(context.Background(), email, "123456", "captcha-1", "99")
		assertClientMessage(t, err, "captcha is invalid")
		if fixture.emailCodes.consumed != 0 {
			t.Fatal("图形验证码未通过时不应消费邮箱验证码")
		}
		if fixture.captchas.used["captcha-1"] {
			t.Fatal("图形验证码未通过时不应消费图形验证码")
		}
	})

	t.Run("rejects wrong email code", func(t *testing.T) {
		fixture := setup(t)
		_, err := fixture.authenticator.EmailLogin(context.Background(), email, "000000", "captcha-1", "12")
		assertClientMessage(t, err, "invalid email or code")
		if fixture.captchas.used["captcha-1"] {
			t.Fatal("邮箱验证码错误时图形验证码应保留，方便用户重试")
		}
	})

	t.Run("logs in with captcha and email code", func(t *testing.T) {
		fixture := setup(t)
		session, err := fixture.authenticator.EmailLogin(context.Background(), email, "123456", "captcha-1", "12")
		if err != nil {
			t.Fatalf("EmailLogin: %v", err)
		}
		if session.AccessToken == "" || session.RefreshToken == "" {
			t.Fatal("登录成功后没有签发令牌")
		}
		if session.User.Email != email {
			t.Fatalf("session user = %+v", session.User)
		}
		if !fixture.captchas.used["captcha-1"] {
			t.Fatal("登录成功后应消费图形验证码")
		}
		if fixture.emailCodes.consumed != 1 {
			t.Fatalf("邮箱验证码消费次数 = %d, want 1", fixture.emailCodes.consumed)
		}
	})
}

func TestSendEmailLoginCode(t *testing.T) {
	const email = "user@example.com"

	t.Run("rejects wrong captcha", func(t *testing.T) {
		fixture := newAuthFixture(t)
		fixture.addUser(email)
		fixture.solveCaptcha(t, "captcha-1", "12")
		err := fixture.authenticator.SendEmailLoginCode(context.Background(), email, "captcha-1", "99")
		assertClientMessage(t, err, "captcha is invalid")
		if len(fixture.mail.subjects) != 0 || len(fixture.emailCodes.codes) != 0 {
			t.Fatal("图形验证码未通过时不能存码或发信")
		}
		if fixture.emailCodes.issuedCalls != 0 {
			t.Fatal("图形验证码未通过时不应进入节流检查")
		}
	})

	t.Run("rejects unregistered email", func(t *testing.T) {
		fixture := newAuthFixture(t)
		fixture.solveCaptcha(t, "captcha-1", "12")
		err := fixture.authenticator.SendEmailLoginCode(context.Background(), "nobody@example.com", "captcha-1", "12")
		assertClientMessage(t, err, "email is not registered")
		if len(fixture.mail.subjects) != 0 {
			t.Fatal("未注册邮箱不应发信")
		}
	})

	t.Run("rejects repeated send within 60 seconds", func(t *testing.T) {
		fixture := newAuthFixture(t)
		fixture.addUser(email)
		fixture.solveCaptcha(t, "captcha-1", "12")
		fixture.emailCodes.issuedAt[email] = time.Now().Add(-10 * time.Second)
		err := fixture.authenticator.SendEmailLoginCode(context.Background(), email, "captcha-1", "12")
		assertClientMessage(t, err, "send code too frequently")
		if len(fixture.mail.subjects) != 0 {
			t.Fatal("节流命中时不应发信")
		}
		metadata := kerrors.FromError(err).Metadata
		seconds, convErr := strconv.Atoi(metadata["retryAfterSeconds"])
		if convErr != nil || seconds < 45 || seconds > 55 {
			t.Fatalf("剩余等待时间 = %q（%v），应约 50 秒", metadata["retryAfterSeconds"], convErr)
		}
	})

	t.Run("allows send after 60 seconds", func(t *testing.T) {
		fixture := newAuthFixture(t)
		fixture.addUser(email)
		fixture.solveCaptcha(t, "captcha-1", "12")
		fixture.emailCodes.issuedAt[email] = time.Now().Add(-61 * time.Second)
		if err := fixture.authenticator.SendEmailLoginCode(context.Background(), email, "captcha-1", "12"); err != nil {
			t.Fatalf("节流窗口结束后应允许发码: %v", err)
		}
		if len(fixture.mail.subjects) != 1 {
			t.Fatalf("发信次数 = %d, want 1", len(fixture.mail.subjects))
		}
	})

	t.Run("sends login code", func(t *testing.T) {
		fixture := newAuthFixture(t)
		fixture.addUser(email)
		fixture.solveCaptcha(t, "captcha-1", "12")
		if err := fixture.authenticator.SendEmailLoginCode(context.Background(), email, "captcha-1", "12"); err != nil {
			t.Fatalf("SendEmailLoginCode: %v", err)
		}
		if len(fixture.mail.subjects) != 1 {
			t.Fatalf("发信次数 = %d, want 1", len(fixture.mail.subjects))
		}
		if !strings.Contains(fixture.mail.subjects[0], "login") {
			t.Fatalf("登录码邮件主题 = %q", fixture.mail.subjects[0])
		}
		if _, ok := fixture.emailCodes.codes[email]; !ok {
			t.Fatal("邮箱验证码没有落库")
		}
		if fixture.emailCodes.issuedCalls != 1 {
			t.Fatalf("节流检查次数 = %d, want 1", fixture.emailCodes.issuedCalls)
		}
	})
}

func TestSendRegisterEmailCode(t *testing.T) {
	const email = "new@example.com"

	t.Run("rejects registered email", func(t *testing.T) {
		fixture := newAuthFixture(t)
		fixture.addUser(email)
		fixture.solveCaptcha(t, "captcha-1", "12")
		err := fixture.authenticator.SendRegisterEmailCode(context.Background(), email, "captcha-1", "12")
		assertClientMessage(t, err, "email already exists")
		if len(fixture.mail.subjects) != 0 {
			t.Fatal("已注册邮箱不应再发注册码")
		}
	})

	t.Run("sends register code", func(t *testing.T) {
		fixture := newAuthFixture(t)
		fixture.solveCaptcha(t, "captcha-1", "12")
		if err := fixture.authenticator.SendRegisterEmailCode(context.Background(), email, "captcha-1", "12"); err != nil {
			t.Fatalf("SendRegisterEmailCode: %v", err)
		}
		if len(fixture.mail.subjects) != 1 {
			t.Fatalf("发信次数 = %d, want 1", len(fixture.mail.subjects))
		}
		if !strings.Contains(fixture.mail.subjects[0], "register") {
			t.Fatalf("注册码邮件主题 = %q", fixture.mail.subjects[0])
		}
	})
}

func TestRegister(t *testing.T) {
	const email = "new@example.com"

	t.Run("rejects invalid username", func(t *testing.T) {
		for _, name := range []string{"abcdefghi", "bad-name", "bad name", "名字!"} {
			fixture := newAuthFixture(t)
			fixture.solveCaptcha(t, "captcha-1", "12")
			fixture.emailCodes.codes[email] = fixture.authenticator.digest("123456")
			_, err := fixture.authenticator.Register(context.Background(), email, "123456", name, "captcha-1", "12")
			assertClientMessage(t, err, "invalid username")
			if len(fixture.users.created) != 0 {
				t.Fatalf("用户名 %q 不合法时不应创建用户", name)
			}
			if fixture.emailCodes.consumed != 0 {
				t.Fatalf("用户名 %q 不合法时不应消费邮箱验证码", name)
			}
		}
	})

	t.Run("rejects invalid email code", func(t *testing.T) {
		fixture := newAuthFixture(t)
		fixture.solveCaptcha(t, "captcha-1", "12")
		fixture.emailCodes.codes[email] = fixture.authenticator.digest("123456")
		_, err := fixture.authenticator.Register(context.Background(), email, "000000", "新用户", "captcha-1", "12")
		assertClientMessage(t, err, "invalid email or code")
		if len(fixture.users.created) != 0 {
			t.Fatal("邮箱验证码错误时不应创建用户")
		}
		if fixture.captchas.used["captcha-1"] {
			t.Fatal("邮箱验证码错误时图形验证码应保留")
		}
	})

	t.Run("rejects registered email", func(t *testing.T) {
		fixture := newAuthFixture(t)
		fixture.addUser(email)
		fixture.solveCaptcha(t, "captcha-1", "12")
		fixture.emailCodes.codes[email] = fixture.authenticator.digest("123456")
		_, err := fixture.authenticator.Register(context.Background(), email, "123456", "newuser", "captcha-1", "12")
		assertClientMessage(t, err, "email already exists")
		if len(fixture.users.created) != 0 {
			t.Fatal("已存在的邮箱不应创建新用户")
		}
	})

	t.Run("creates user without password hash", func(t *testing.T) {
		fixture := newAuthFixture(t)
		fixture.solveCaptcha(t, "captcha-1", "12")
		fixture.emailCodes.codes[email] = fixture.authenticator.digest("123456")
		session, err := fixture.authenticator.Register(context.Background(), email, "123456", "新用户", "captcha-1", "12")
		if err != nil {
			t.Fatalf("Register: %v", err)
		}
		if session.AccessToken == "" || session.User.Email != email {
			t.Fatalf("session = %+v", session)
		}
		if len(fixture.users.created) != 1 {
			t.Fatalf("创建用户次数 = %d, want 1", len(fixture.users.created))
		}
		created := fixture.users.created[0]
		if created.Name != "新用户" {
			t.Fatalf("用户名称 = %q，应为注册时填写的名称", created.Name)
		}
		if !fixture.captchas.used["captcha-1"] {
			t.Fatal("注册成功后应消费图形验证码")
		}
		if fixture.emailCodes.consumed != 1 {
			t.Fatalf("邮箱验证码消费次数 = %d, want 1", fixture.emailCodes.consumed)
		}
	})

	// 没填名称时保留旧行为：用邮箱前缀，但派生结果也要合规。
	t.Run("derives compliant username when blank", func(t *testing.T) {
		const derivedEmail = "john.doe123456@example.com"
		fixture := newAuthFixture(t)
		fixture.solveCaptcha(t, "captcha-1", "12")
		fixture.emailCodes.codes[derivedEmail] = fixture.authenticator.digest("123456")
		if _, err := fixture.authenticator.Register(context.Background(), derivedEmail, "123456", "", "captcha-1", "12"); err != nil {
			t.Fatalf("Register: %v", err)
		}
		if len(fixture.users.created) != 1 {
			t.Fatalf("创建用户次数 = %d, want 1", len(fixture.users.created))
		}
		name := fixture.users.created[0].Name
		if name != "johndoe1" {
			t.Fatalf("派生用户名 = %q，应为去掉特殊字符并截断到 8 位的 johndoe1", name)
		}
		if err := user.ValidateUsername(name); err != nil {
			t.Fatalf("派生用户名不合规: %v", err)
		}
	})
}

// TestRegisterRequestHasNoPassword 确认契约里 password 字段已删除，字段号 2 已 reserved。
func TestRegisterRequestHasNoPassword(t *testing.T) {
	fields := (&v1.RegisterRequest{}).ProtoReflect().Descriptor().Fields()
	if fields.ByName("password") != nil {
		t.Fatal("注册请求仍然有 password 字段")
	}
	if fields.ByNumber(2) != nil {
		t.Fatalf("注册请求的字段号 2 已被 %q 占用，应为 reserved", fields.ByNumber(2).Name())
	}
}

func TestAccountHasNoPasswordHash(t *testing.T) {
	if _, ok := reflect.TypeOf(bo.Account{}).FieldByName("PasswordHash"); ok {
		t.Fatal("Account 仍然保存密码哈希")
	}
}

func TestPublicOperation(t *testing.T) {
	public := []string{
		v1.OperationCaptchaServiceGetCaptcha,
		v1.OperationAuthServiceSendRegisterEmailCode,
		v1.OperationAuthServiceRegister,
		v1.OperationAuthServiceSendEmailLoginCode,
		v1.OperationAuthServiceEmailLogin,
		v1.OperationAuthServiceListOAuth2Providers,
		v1.OperationAuthServiceOAuth2Login,
	}
	for _, operation := range public {
		if !publicOperation(operation) {
			t.Fatalf("%s 应该是公开接口", operation)
		}
	}
	if publicOperation("/alioth.v1.AuthService/PasswordLogin") {
		t.Fatal("PasswordLogin 已删除，不能出现在公开接口白名单里")
	}
}
