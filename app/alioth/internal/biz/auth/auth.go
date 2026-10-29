package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	kerrors "github.com/go-kratos/kratos/v3/errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/wire"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	megrezv1 "github.com/aide-family/moon/api/megrez/v1"
	merakv1 "github.com/aide-family/moon/api/merak/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/user"
	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/internal/mailer"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

var ProviderSet = wire.NewSet(NewAuthenticator)

var phonePattern = regexp.MustCompile(`^\+?[0-9]{6,20}$`)

const (
	// captchaTTL 图形验证码有效期。题目简单，过期时间短，避免被反复试答案。
	captchaTTL = 5 * time.Minute
	// emailCodeTTL 邮箱验证码有效期。发出去后 10 分钟内可用。
	emailCodeTTL = 10 * time.Minute
	// emailCodeInterval 同一邮箱两次发码的最小间隔，挡住邮件轰炸。
	emailCodeInterval = 60 * time.Second
	// emailCodeMaxAttempts 是单个验证码允许的错误次数；用尽后立即作废，必须重新发码。
	emailCodeMaxAttempts = 5
	// emailCodeFailureWindow 与 emailCodeFailureLimit 是邮箱级锁定：窗口内被作废的码达到上限后，
	// 暂停该邮箱的验证码校验与发码，避免 6 位码被持续枚举。
	emailCodeFailureWindow = 15 * time.Minute
	emailCodeFailureLimit  = 3
	// bindEmailLimit 与 bindEmailWindow 限制同一个第三方绑定凭证可以发码的不同邮箱数量，
	// 避免拿一个 bindToken 给大量地址发码做邮件轰炸。
	bindEmailLimit  = 5
	bindEmailWindow = time.Hour
)

// AgentTokenLookup 用明文 Agent 令牌解析机器。只给 Connect 用。
type AgentTokenLookup interface {
	AuthenticateToken(ctx context.Context, token string) (*bo.Agent, error)
}

type oauthClient struct {
	id     string
	secret string
}

// codeMailer 发送邮箱验证码邮件。生产环境由 *mailer.Sender 实现，测试可以注入桩。
type codeMailer interface {
	Send(to, subject, body string) error
}

// Authenticator 负责身份凭证。租户、成员和角色的判断不放在这里。
// 用户只通过邮箱验证码或第三方登录进入，不使用密码。
type Authenticator struct {
	users        repository.UserRepo
	captchas     repository.CaptchaRepo
	emailCodes   repository.EmailCodeRepo
	oauth        repository.OAuthRepo
	agents       AgentTokenLookup
	secret       string
	accessTTL    time.Duration
	refreshTTL   time.Duration
	serviceKeys  map[string]string
	mail         codeMailer
	oauthClients map[string]oauthClient
	log          *slog.Logger
}

func NewAuthenticator(users repository.UserRepo, captchas repository.CaptchaRepo, emailCodes repository.EmailCodeRepo, oauth repository.OAuthRepo, agents AgentTokenLookup, mail *mailer.Sender, auth *conf.Auth, logger *slog.Logger) (*Authenticator, error) {
	if auth == nil || auth.GetJwtSecret() == "" {
		return nil, merr.Server("jwt secret is empty")
	}
	access := time.Duration(auth.GetAccessTtlSeconds()) * time.Second
	if access <= 0 {
		access = 2 * time.Hour
	}
	refresh := time.Duration(auth.GetRefreshTtlSeconds()) * time.Second
	if refresh <= 0 {
		refresh = 7 * 24 * time.Hour
	}
	keys := map[string]string{}
	for _, item := range auth.GetServiceKeys() {
		if item.GetKey() != "" {
			keys[item.GetKey()] = item.GetName()
		}
	}
	clients := map[string]oauthClient{}
	for _, item := range auth.GetOauth() {
		clients[strings.ToUpper(item.GetProvider())] = oauthClient{id: item.GetClientId(), secret: item.GetClientSecret()}
	}
	// 注意 nil 的 *mailer.Sender 赋给接口后不是 nil，这里显式转换，测试才能用 nil 表示未配置邮件。
	var codes codeMailer
	if mail != nil {
		codes = mail
	}
	return &Authenticator{
		users:        users,
		captchas:     captchas,
		emailCodes:   emailCodes,
		oauth:        oauth,
		agents:       agents,
		secret:       auth.GetJwtSecret(),
		accessTTL:    access,
		refreshTTL:   refresh,
		serviceKeys:  keys,
		mail:         codes,
		oauthClients: clients,
		log:          logger,
	}, nil
}

// Authenticate 决定这次请求以哪种凭证进入后续用例。
// 评测领取和回写只接受名为 megrez 的 service key；投递领取和回写只接受 merak。
// 用户 JWT 在这些接口上拒绝。其他 service key 只能调用健康检查、验证码和登录注册。
// 公开接口不要求令牌。刷新接口只接受 refresh token，其余接口只接受 access token。
// 被封禁的用户在这里拒绝，后面的用例不再重复判断账号是否可登录。
// 个人资料、刷新令牌和创建租户可以不带 X-Namespace；带了则记入上下文。其他接口必须带租户。
func (authenticator *Authenticator) Authenticate(ctx context.Context, operation, authorization, namespace string) (context.Context, error) {
	token := bearerToken(authorization)
	if isServiceKey(token) {
		name, ok := authenticator.serviceKeys[token]
		if !ok {
			return nil, merr.Forbidden("service key is not accepted")
		}
		if required, ok := serviceOperation(operation); ok {
			if name != required {
				return nil, merr.Forbidden("service key cannot call this api")
			}
			return actor.WithService(ctx, name), nil
		}
		if !publicOperation(operation) {
			return nil, merr.Forbidden("service key cannot call this api")
		}
		return ctx, nil
	}
	if isAgentToken(token) {
		if operation != v1.AgentService_Connect_FullMethodName {
			return nil, merr.Forbidden("agent token cannot call this api")
		}
		if authenticator.agents == nil {
			return nil, merr.Unauthenticated("invalid token")
		}
		item, err := authenticator.agents.AuthenticateToken(ctx, token)
		if err != nil {
			return nil, err
		}
		ctx = actor.WithAgentToken(ctx, token)
		return actor.WithNamespace(ctx, item.NamespaceUID), nil
	}
	if operation == v1.AgentService_Connect_FullMethodName {
		return nil, merr.Forbidden("agent token required")
	}
	if _, ok := serviceOperation(operation); ok {
		return nil, merr.Forbidden("service key is required")
	}
	if publicOperation(operation) {
		return ctx, nil
	}
	if token == "" {
		return nil, merr.Unauthenticated("missing token")
	}
	uid, kind, err := authenticator.parseToken(token)
	if err != nil {
		return nil, merr.Unauthenticated("invalid token")
	}
	if operation == v1.OperationSelfServiceRefreshToken {
		if kind != "refresh" {
			return nil, merr.Unauthenticated("refresh token required")
		}
	} else if kind != "access" {
		return nil, merr.Unauthenticated("access token required")
	}
	account, err := authenticator.users.UserByUID(ctx, uid)
	if err != nil || account.Status != "ACTIVE" {
		return nil, merr.Unauthenticated("user is not active")
	}
	ctx = actor.WithUser(ctx, account.UID)
	if namespaceOptional(operation) {
		if namespace != "" {
			ctx = actor.WithNamespace(ctx, namespace)
		}
		return ctx, nil
	}
	if namespace == "" {
		return nil, merr.Client("X-Namespace is required")
	}
	return actor.WithNamespace(ctx, namespace), nil
}

// GetCaptcha 生成一道图形验证码，题型在六位数字码和 100 以内算式之间随机。
// 保存的是答案摘要，题目只出现在返回的图片里，5 分钟后过期。kind 供前端切换输入提示。
func (authenticator *Authenticator) GetCaptcha(ctx context.Context) (captchaID, image, kind string, err error) {
	question, answer, kind, err := newCaptcha()
	if err != nil {
		return "", "", "", authenticator.wrap(err)
	}
	image, err = drawCaptcha(question)
	if err != nil {
		return "", "", "", authenticator.wrap(err)
	}
	captchaID = id.New()
	if err := authenticator.captchas.SaveCaptcha(ctx, captchaID, authenticator.digest(answer), time.Now().Add(captchaTTL)); err != nil {
		return "", "", "", authenticator.wrap(err)
	}
	return captchaID, image, kind, nil
}

// checkCaptcha 校验图形验证码但不消费。发码和提交前都调用它：
// 一次解题可以同时走完发码和提交，但没解题的人既拿不到邮箱码也无法提交。
func (authenticator *Authenticator) checkCaptcha(ctx context.Context, captchaID, captchaCode string) error {
	matched, err := authenticator.captchas.CheckCaptcha(ctx, captchaID, authenticator.digest(captchaCode), time.Now())
	if err != nil {
		return authenticator.wrap(err)
	}
	if !matched {
		return merr.Client("captcha is invalid")
	}
	return nil
}

// consumeCaptcha 在邮箱验证码消费成功之后再消费图形验证码。
// 顺序很重要：邮箱码错误时图形验证码仍有效，用户可以改正后重试，不用重新解题。
func (authenticator *Authenticator) consumeCaptcha(ctx context.Context, captchaID, captchaCode string) error {
	ok, err := authenticator.captchas.ConsumeCaptcha(ctx, captchaID, authenticator.digest(captchaCode), time.Now())
	if err != nil {
		return authenticator.wrap(err)
	}
	if !ok {
		return merr.Client("captcha is invalid")
	}
	return nil
}

// checkEmailCodeInterval 限制同一邮箱 60 秒内只能发一次验证码，返回客户端错误而不是静默丢弃。
// 错误里带上剩余秒数，前端可以据此恢复倒计时，而不是只弹一句“太频繁”。
func (authenticator *Authenticator) checkEmailCodeInterval(ctx context.Context, email string) error {
	issuedAt, ok, err := authenticator.emailCodes.LastEmailCodeIssuedAt(ctx, email)
	if err != nil {
		return authenticator.wrap(err)
	}
	if !ok {
		return nil
	}
	wait := emailCodeInterval - time.Since(issuedAt)
	if wait <= 0 {
		return nil
	}
	seconds := int(math.Ceil(wait.Seconds()))
	if seconds < 1 {
		seconds = 1
	}
	return merr.ClientWithMetadata("send code too frequently", map[string]string{
		"retryAfterSeconds": strconv.Itoa(seconds),
	})
}

// checkEmailCodeLockout 在同一邮箱短时间内多次把验证码试到作废时暂停校验。
// 锁定随失败窗口滑动自动解除，不额外存储状态。
func (authenticator *Authenticator) checkEmailCodeLockout(ctx context.Context, email string) error {
	failures, err := authenticator.emailCodes.RecentExhaustedEmailCodes(ctx, email, time.Now().Add(-emailCodeFailureWindow), emailCodeMaxAttempts)
	if err != nil {
		return authenticator.wrap(err)
	}
	if failures >= emailCodeFailureLimit {
		return merr.Client("too many verification attempts")
	}
	return nil
}

// consumeEmailCode 消费邮箱验证码，并把错误尝试记到当前码上。
// 同一个码达到 emailCodeMaxAttempts 次错误后立即作废，避免在有效期内被无限枚举。
func (authenticator *Authenticator) consumeEmailCode(ctx context.Context, email, code, purpose string) (bool, error) {
	if err := authenticator.checkEmailCodeLockout(ctx, email); err != nil {
		return false, err
	}
	ok, err := authenticator.emailCodes.ConsumeEmailCode(ctx, email, authenticator.digest(code), purpose, time.Now())
	if err != nil {
		return false, authenticator.wrap(err)
	}
	if !ok {
		if err := authenticator.emailCodes.IncrementEmailCodeAttempts(ctx, email, purpose, time.Now(), emailCodeMaxAttempts); err != nil {
			return false, authenticator.wrap(err)
		}
	}
	return ok, nil
}

// SendRegisterEmailCode 给未注册邮箱发注册验证码。
// 顺序是：图形验证码 → 邮箱是否已注册 → 60 秒节流 → 存码并发注册码邮件。
// 先查邮箱占用再节流，已注册的邮箱不会因为反复请求就消耗掉一次发码机会。
func (authenticator *Authenticator) SendRegisterEmailCode(ctx context.Context, email, captchaID, captchaCode string) error {
	email = normalizeEmail(email)
	if !strings.Contains(email, "@") || captchaID == "" || strings.TrimSpace(captchaCode) == "" {
		return merr.Client("invalid request")
	}
	if err := authenticator.checkCaptcha(ctx, captchaID, captchaCode); err != nil {
		return err
	}
	if err := authenticator.checkEmailCodeLockout(ctx, email); err != nil {
		return err
	}
	if _, err := authenticator.users.UserByEmail(ctx, email); err == nil {
		return merr.Client("email already exists")
	} else if !merr.IsNotFound(err) {
		return authenticator.wrap(err)
	}
	if err := authenticator.checkEmailCodeInterval(ctx, email); err != nil {
		return err
	}
	return authenticator.sendEmailCode(ctx, email, "register")
}

// SendEmailLoginCode 给已注册邮箱发登录验证码。
// 顺序是：图形验证码 → 邮箱是否已注册 → 60 秒节流 → 存码并发登录码邮件。
func (authenticator *Authenticator) SendEmailLoginCode(ctx context.Context, email, captchaID, captchaCode string) error {
	email = normalizeEmail(email)
	if !strings.Contains(email, "@") || captchaID == "" || strings.TrimSpace(captchaCode) == "" {
		return merr.Client("invalid request")
	}
	if err := authenticator.checkCaptcha(ctx, captchaID, captchaCode); err != nil {
		return err
	}
	if err := authenticator.checkEmailCodeLockout(ctx, email); err != nil {
		return err
	}
	if _, err := authenticator.users.UserByEmail(ctx, email); err != nil {
		if merr.IsNotFound(err) {
			return merr.Client("email is not registered")
		}
		return authenticator.wrap(err)
	}
	if err := authenticator.checkEmailCodeInterval(ctx, email); err != nil {
		return err
	}
	return authenticator.sendEmailCode(ctx, email, "login")
}

// sendEmailCode 生成并保存一次性邮箱验证码，再按用途发对应主题的邮件。
// kind 同时写入码的 purpose（login/register/bind/change_email），消费时必须匹配。
func (authenticator *Authenticator) sendEmailCode(ctx context.Context, email, kind string) error {
	return authenticator.sendEmailCodeWithBinding(ctx, email, kind, "")
}

// sendEmailCodeWithBinding 同 sendEmailCode，额外记录第三方绑定凭证摘要；
// SendOAuthBindEmailCode 用它限制同一凭证可以发往的不同邮箱数量。
func (authenticator *Authenticator) sendEmailCodeWithBinding(ctx context.Context, email, kind, bindTokenHash string) error {
	code, err := randomDigits(6)
	if err != nil {
		return merr.Server("request failed")
	}
	if err := authenticator.emailCodes.SaveEmailCodeWithBinding(ctx, email, authenticator.digest(code), kind, bindTokenHash, time.Now().Add(emailCodeTTL)); err != nil {
		return authenticator.wrap(err)
	}
	subject := "Alioth " + kind + " code"
	body := "Your Alioth " + kind + " code is " + code + "\r\nIt expires in 10 minutes.\r\n"
	return authenticator.sendCodeMail(email, subject, body)
}

func (authenticator *Authenticator) sendCodeMail(to, subject, body string) error {
	if authenticator.mail == nil {
		return merr.Server("smtp is not configured")
	}
	return authenticator.mail.Send(to, subject, body)
}

// Register 用邮箱验证码注册。顺序是：入参校验 → 图形验证码非消费校验 → 邮箱是否已存在 →
// 消费邮箱验证码 → 消费图形验证码 → 创建用户。注册没有密码，也不写任何密码字段。
// 邮箱码消费成功后才消费图形验证码，保证邮箱码填错时这张图形验证码还能再用一次。
// 用户名必须符合 user.ValidateUsername：仅中英文和数字，最多 8 个字符；没填时沿用旧行为，
// 从邮箱前缀派生一个合法名称。历史上已存在的用户名不做回溯校验，避免破坏老账号。
func (authenticator *Authenticator) Register(ctx context.Context, email, code, name, captchaID, captchaCode string) (*bo.Session, error) {
	email = normalizeEmail(email)
	name = strings.TrimSpace(name)
	if !strings.Contains(email, "@") || strings.TrimSpace(code) == "" || captchaID == "" || strings.TrimSpace(captchaCode) == "" {
		return nil, merr.Client("invalid register request")
	}
	if name == "" {
		name = user.DefaultUsername(email)
	}
	if err := user.ValidateUsername(name); err != nil {
		return nil, err
	}
	if err := authenticator.checkCaptcha(ctx, captchaID, captchaCode); err != nil {
		return nil, err
	}
	if _, err := authenticator.users.UserByEmail(ctx, email); err == nil {
		return nil, merr.Client("email already exists")
	} else if !merr.IsNotFound(err) {
		return nil, authenticator.wrap(err)
	}
	ok, err := authenticator.consumeEmailCode(ctx, email, code, "register")
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, merr.Client("invalid email or code")
	}
	if err := authenticator.consumeCaptcha(ctx, captchaID, captchaCode); err != nil {
		return nil, err
	}
	account, err := authenticator.users.CreateUser(ctx, bo.Account{
		UID:      id.New(),
		Name:     name,
		Nickname: name,
		Email:    email,
		Status:   "ACTIVE",
	})
	if err != nil {
		return nil, authenticator.wrap(err)
	}
	return authenticator.session(account)
}

// EmailLogin 用邮箱验证码登录。顺序是：入参校验 → 图形验证码非消费校验 →
// 消费邮箱验证码 → 消费图形验证码 → 签发会话。同一封邮箱码不能登录两次。
// 封禁用户仍由 session 拒绝签发令牌，这里不重复判断账号状态。
func (authenticator *Authenticator) EmailLogin(ctx context.Context, email, code, captchaID, captchaCode string) (*bo.Session, error) {
	email = normalizeEmail(email)
	if !strings.Contains(email, "@") || strings.TrimSpace(code) == "" || captchaID == "" || strings.TrimSpace(captchaCode) == "" {
		return nil, merr.Client("invalid login request")
	}
	if err := authenticator.checkCaptcha(ctx, captchaID, captchaCode); err != nil {
		return nil, err
	}
	ok, err := authenticator.consumeEmailCode(ctx, email, code, "login")
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, merr.Client("invalid email or code")
	}
	if err := authenticator.consumeCaptcha(ctx, captchaID, captchaCode); err != nil {
		return nil, err
	}
	account, err := authenticator.users.UserByEmail(ctx, email)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.Client("invalid email or code")
		}
		return nil, authenticator.wrap(err)
	}
	return authenticator.session(account)
}

func (authenticator *Authenticator) Refresh(ctx context.Context) (string, error) {
	uid, ok := actor.UserID(ctx)
	if !ok {
		return "", merr.Unauthenticated("missing token")
	}
	account, err := authenticator.users.UserByUID(ctx, uid)
	if err != nil || account.Status != "ACTIVE" {
		return "", merr.Unauthenticated("user is not active")
	}
	return authenticator.sign(uid, "access", authenticator.accessTTL)
}

func (authenticator *Authenticator) Self(ctx context.Context) (*bo.Account, error) {
	uid, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	account, err := authenticator.users.UserByUID(ctx, uid)
	if err != nil {
		return nil, authenticator.wrap(err)
	}
	return account, nil
}

// SelfSystem 返回当前用户的平台角色和平台权限。它和租户权限是两套目录。
func (authenticator *Authenticator) SelfSystem(ctx context.Context) (string, []string, error) {
	uid, ok := actor.UserID(ctx)
	if !ok {
		return "", nil, merr.Unauthenticated("missing token")
	}
	account, err := authenticator.users.UserByUID(ctx, uid)
	if err != nil {
		return "", nil, authenticator.wrap(err)
	}
	return account.SystemRole, perm.SystemPermissions(account.SystemRole), nil
}

// SendSelfChangeEmailCode 给当前用户要换绑的新邮箱发验证码。
// 顺序：登录态 → 邮箱是否已被他人占用 → 图形验证码 → 邮箱级锁定 → 60 秒节流 → 发码。
// 新邮箱可以是尚未注册的地址，验证码本身就是所有权证明。
func (authenticator *Authenticator) SendSelfChangeEmailCode(ctx context.Context, email, captchaID, captchaCode string) error {
	uid, ok := actor.UserID(ctx)
	if !ok {
		return merr.Unauthenticated("missing token")
	}
	email = normalizeEmail(email)
	if !strings.Contains(email, "@") || captchaID == "" || strings.TrimSpace(captchaCode) == "" {
		return merr.Client("invalid request")
	}
	if existing, err := authenticator.users.UserByEmail(ctx, email); err == nil && existing.UID != uid {
		return merr.Client("email already exists")
	} else if err != nil && !merr.IsNotFound(err) {
		return authenticator.wrap(err)
	}
	if err := authenticator.checkCaptcha(ctx, captchaID, captchaCode); err != nil {
		return err
	}
	if err := authenticator.checkEmailCodeLockout(ctx, email); err != nil {
		return err
	}
	if err := authenticator.checkEmailCodeInterval(ctx, email); err != nil {
		return err
	}
	return authenticator.sendEmailCode(ctx, email, "change_email")
}

// ChangeEmail 修改当前用户邮箱。必须先给新邮箱发验证码并提交，验证通过才写入：
// 邮箱是 OAuth 登录的账号身份锚点，未验证的地址不能直接占用，否则他人可以先抢注，
// 再让受害者用第三方登录并入错误账号。顺序与注册一致：图形验证码 → 邮箱码 → 消费图形验证码 → 写库。
func (authenticator *Authenticator) ChangeEmail(ctx context.Context, email, code, captchaID, captchaCode string) (*bo.Account, error) {
	uid, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	email = normalizeEmail(email)
	if !strings.Contains(email, "@") || strings.TrimSpace(code) == "" || captchaID == "" || strings.TrimSpace(captchaCode) == "" {
		return nil, merr.Client("invalid request")
	}
	if err := authenticator.checkCaptcha(ctx, captchaID, captchaCode); err != nil {
		return nil, err
	}
	verified, err := authenticator.consumeEmailCode(ctx, email, code, "change_email")
	if err != nil {
		return nil, err
	}
	if !verified {
		return nil, merr.Client("invalid email or code")
	}
	if err := authenticator.consumeCaptcha(ctx, captchaID, captchaCode); err != nil {
		return nil, err
	}
	existing, err := authenticator.users.UserByEmail(ctx, email)
	if err == nil && existing.UID != uid {
		return nil, merr.Client("email already exists")
	}
	if err != nil && !merr.IsNotFound(err) {
		return nil, authenticator.wrap(err)
	}
	account, err := authenticator.users.UpdateUserProfile(ctx, uid, &email, nil, nil, nil)
	if err != nil {
		return nil, authenticator.wrap(err)
	}
	return account, nil
}

func (authenticator *Authenticator) ChangeAvatar(ctx context.Context, avatar string) (*bo.Account, error) {
	uid, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	account, err := authenticator.users.UpdateUserProfile(ctx, uid, nil, nil, &avatar, nil)
	if err != nil {
		return nil, authenticator.wrap(err)
	}
	return account, nil
}

func (authenticator *Authenticator) ChangeRemark(ctx context.Context, remark string) (*bo.Account, error) {
	uid, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	account, err := authenticator.users.UpdateUserProfile(ctx, uid, nil, nil, nil, &remark)
	if err != nil {
		return nil, authenticator.wrap(err)
	}
	return account, nil
}

func (authenticator *Authenticator) ChangePhone(ctx context.Context, phone string) (*bo.Account, error) {
	uid, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	phone = strings.TrimSpace(phone)
	if phone != "" && !phonePattern.MatchString(phone) {
		return nil, merr.Invalid("invalid phone number")
	}
	account, err := authenticator.users.UpdateUserProfile(ctx, uid, nil, &phone, nil, nil)
	if err != nil {
		return nil, authenticator.wrap(err)
	}
	return account, nil
}

// session 签发访问令牌和刷新令牌。封禁用户不能拿到新会话，即使邮箱验证码正确。
func (authenticator *Authenticator) session(account *bo.Account) (*bo.Session, error) {
	if account.Status != "ACTIVE" {
		return nil, merr.Unauthenticated("user is banned")
	}
	access, err := authenticator.sign(account.UID, "access", authenticator.accessTTL)
	if err != nil {
		return nil, merr.Server("request failed")
	}
	refresh, err := authenticator.sign(account.UID, "refresh", authenticator.refreshTTL)
	if err != nil {
		return nil, merr.Server("request failed")
	}
	return &bo.Session{AccessToken: access, RefreshToken: refresh, User: *account}, nil
}

func (authenticator *Authenticator) sign(uid, kind string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"uid":  uid,
		"kind": kind,
		"iat":  now.Unix(),
		"exp":  now.Add(ttl).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(authenticator.secret))
	if err != nil {
		authenticator.log.Error(fmt.Sprintf("jwt: %v", err))
		return "", merr.Server("request failed")
	}
	return signed, nil
}

func (authenticator *Authenticator) parseToken(raw string) (string, string, error) {
	token, err := jwt.Parse(raw, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, merr.Unauthenticated("invalid token")
		}
		return []byte(authenticator.secret), nil
	})
	if err != nil || !token.Valid {
		return "", "", merr.Unauthenticated("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", merr.Unauthenticated("invalid token")
	}
	uid, _ := claims["uid"].(string)
	kind, _ := claims["kind"].(string)
	if uid == "" || (kind != "access" && kind != "refresh") {
		return "", "", merr.Unauthenticated("invalid token")
	}
	return uid, kind, nil
}

func (authenticator *Authenticator) digest(value string) string {
	sum := sha256.Sum256([]byte(authenticator.secret + ":" + strings.ToLower(strings.TrimSpace(value))))
	return hex.EncodeToString(sum[:])
}

// tokenHash 是第三方绑定凭证的计数键。凭证是签名 JWT、原文不落库，
// 这里只保存加盐摘要，用于统计同一凭证发码的不同邮箱数量。
func (authenticator *Authenticator) tokenHash(value string) string {
	sum := sha256.Sum256([]byte(authenticator.secret + ":token:" + value))
	return hex.EncodeToString(sum[:])
}

func (authenticator *Authenticator) wrap(err error) error {
	if err == nil {
		return nil
	}
	if kerrors.Reason(err) != "" {
		return err
	}
	authenticator.log.Error(fmt.Sprintf("identity: %v", err))
	return merr.Server("request failed")
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func bearerToken(header string) string {
	header = strings.TrimSpace(header)
	if len(header) < 7 || !strings.EqualFold(header[:7], "bearer ") {
		return ""
	}
	return strings.TrimSpace(header[7:])
}

func isServiceKey(token string) bool {
	return strings.HasPrefix(token, "sk-")
}

func isAgentToken(token string) bool {
	return strings.HasPrefix(token, "ag-")
}

func serviceOperation(operation string) (string, bool) {
	switch operation {
	case megrezv1.EvalService_ClaimRules_FullMethodName, megrezv1.EvalService_ReportEval_FullMethodName,
		megrezv1.EvalService_ReportWorkerStatus_FullMethodName,
		megrezv1.EvalService_ClaimProbeTasks_FullMethodName, megrezv1.EvalService_ReportProbeResults_FullMethodName:
		return "megrez", true
	case merakv1.NotifyService_ClaimTasks_FullMethodName, merakv1.NotifyService_MarkSending_FullMethodName, merakv1.NotifyService_ReportNotify_FullMethodName, merakv1.NotifyService_ReportWorkerStatus_FullMethodName:
		return "merak", true
	default:
		return "", false
	}
}

// publicOperation 是无需登录的接口白名单。密码登录已经移除，注册、邮箱发码登录、
// 第三方登录和第三方邮箱绑定都在这里，否则前端拿不到验证码或授权地址。
func publicOperation(operation string) bool {
	switch operation {
	case v1.OperationHealthServiceCheck,
		v1.OperationCaptchaServiceGetCaptcha,
		v1.OperationAuthServiceSendRegisterEmailCode,
		v1.OperationAuthServiceRegister,
		v1.OperationAuthServiceSendEmailLoginCode,
		v1.OperationAuthServiceEmailLogin,
		v1.OperationAuthServiceListOAuth2Providers,
		v1.OperationAuthServiceOAuth2Login,
		v1.OperationAuthServiceSendOAuthBindEmailCode,
		v1.OperationAuthServiceBindOAuthEmail,
		v1.OperationInvitationServiceGetInvitation:
		return true
	default:
		return false
	}
}

func namespaceOptional(operation string) bool {
	switch operation {
	case v1.OperationSelfServiceGetSelfInfo,
		v1.OperationSelfServiceListSelfNamespaces,
		v1.OperationSelfServiceGetSelfPermissions,
		v1.OperationSelfServiceRefreshToken,
		v1.OperationSelfServiceChangeEmail,
		v1.OperationSelfServiceChangeAvatar,
		v1.OperationSelfServiceChangeRemark,
		// 第三方账号关联只要求登录态：个人资料页在未选租户时也要能读取和绑定，
		// 这些接口不做租户级权限判断。
		v1.OperationSelfServiceListSelfOAuthLinks,
		v1.OperationSelfServiceBindSelfOAuth,
		v1.OperationSelfServiceUnbindSelfOAuth,
		v1.OperationNamespaceServiceCreateNamespace,
		// 平台租户列表不读 X-Namespace：系统管理员还没有加入任何租户时，
		// 也要能从租户管理页只读切入，所以不要求租户头。
		v1.OperationNamespaceServiceListSystemNamespaces,
		v1.OperationInvitationServiceListSelfInvitations,
		v1.OperationInvitationServiceAcceptInvitation,
		v1.OperationInvitationServiceDeclineInvitation,
		v1.OperationSystemMessageServiceListSelfMessages,
		v1.OperationSystemMessageServiceGetUnreadMessageCount,
		v1.OperationSystemMessageServiceMarkMessageRead,
		v1.OperationSystemMessageServiceMarkAllMessagesRead,
		v1.OperationUserServiceGetUser,
		v1.OperationUserServiceListUsers,
		v1.OperationUserServiceSelectUsers,
		v1.OperationUserServiceBanUser,
		v1.OperationUserServicePermitUser,
		v1.OperationUserServiceUpdateUserSystemRole,
		v1.OperationWorkerServiceListWorkerInstances,
		v1.OperationWorkerServiceGetWorkerInstance:
		return true
	default:
		return false
	}
}

func randomDigits(n int) (string, error) {
	return randomFrom("0123456789", n)
}

func randomFrom(alphabet string, n int) (string, error) {
	buf := make([]byte, n)
	raw := make([]byte, n)
	if _, err := rand.Read(raw); err != nil {
		return "", merr.Server("request failed")
	}
	for i := range buf {
		buf[i] = alphabet[int(raw[i])%len(alphabet)]
	}
	return string(buf), nil
}
