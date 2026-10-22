package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	kerrors "github.com/go-kratos/kratos/v3/errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/wire"
	"golang.org/x/crypto/bcrypt"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	megrezv1 "github.com/aide-family/Alioth/api/megrez/v1"
	merakv1 "github.com/aide-family/Alioth/api/merak/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
	"github.com/aide-family/Alioth/app/alioth/internal/mailer"
	"github.com/aide-family/Alioth/app/alioth/pkg/actor"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/id"
	"github.com/aide-family/Alioth/app/alioth/pkg/perm"
	"github.com/aide-family/Alioth/pkg/merr"
)

var ProviderSet = wire.NewSet(NewAuthenticator)

var phonePattern = regexp.MustCompile(`^\+?[0-9]{6,20}$`)

// AgentTokenLookup 用明文 Agent 令牌解析机器。只给 Connect 用。
type AgentTokenLookup interface {
	AuthenticateToken(ctx context.Context, token string) (*bo.Agent, error)
}

type oauthClient struct {
	id     string
	secret string
}

// Authenticator 负责身份凭证。租户、成员和角色的判断不放在这里。
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
	mail         *mailer.Sender
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
		mail:         mail,
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

// GetCaptcha 生成 4 位验证码。保存的是摘要，答案只出现在返回的图片里，5 分钟后过期。
func (authenticator *Authenticator) GetCaptcha(ctx context.Context) (string, string, error) {
	code, err := randomCode(4)
	if err != nil {
		return "", "", merr.Server("request failed")
	}
	image, err := drawCaptcha(code)
	if err != nil {
		authenticator.log.Error(fmt.Sprintf("draw captcha: %v", err))
		return "", "", merr.Server("request failed")
	}
	captchaID := id.New()
	if err := authenticator.captchas.SaveCaptcha(ctx, captchaID, authenticator.digest(code), time.Now().Add(5*time.Minute)); err != nil {
		return "", "", authenticator.wrap(err)
	}
	return captchaID, image, nil
}

// Register 先消费验证码，再创建用户。验证码只能成功使用一次，密码至少 8 位。
// 邮箱已存在时返回客户端错误，不覆盖原账号。
func (authenticator *Authenticator) Register(ctx context.Context, email, password, name, captchaID, captchaCode string) (*bo.Session, error) {
	email = normalizeEmail(email)
	if !strings.Contains(email, "@") || len(password) < 8 || captchaID == "" || strings.TrimSpace(captchaCode) == "" {
		return nil, merr.Client("invalid register request")
	}
	ok, err := authenticator.captchas.ConsumeCaptcha(ctx, captchaID, authenticator.digest(captchaCode), time.Now())
	if err != nil {
		return nil, authenticator.wrap(err)
	}
	if !ok {
		return nil, merr.Client("captcha is invalid")
	}
	if _, err := authenticator.users.UserByEmail(ctx, email); err == nil {
		return nil, merr.Client("email already exists")
	} else if !merr.IsNotFound(err) {
		return nil, authenticator.wrap(err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, merr.Server("request failed")
	}
	if strings.TrimSpace(name) == "" {
		name = strings.Split(email, "@")[0]
	}
	account, err := authenticator.users.CreateUser(ctx, bo.Account{
		UID:          id.New(),
		Name:         name,
		Nickname:     name,
		Email:        email,
		Status:       "ACTIVE",
		PasswordHash: string(hash),
	})
	if err != nil {
		return nil, authenticator.wrap(err)
	}
	return authenticator.session(account)
}

func (authenticator *Authenticator) PasswordLogin(ctx context.Context, email, password string) (*bo.Session, error) {
	account, err := authenticator.users.UserByEmail(ctx, normalizeEmail(email))
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.Unauthenticated("invalid email or password")
		}
		return nil, authenticator.wrap(err)
	}
	if bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password)) != nil {
		return nil, merr.Unauthenticated("invalid email or password")
	}
	return authenticator.session(account)
}

// SendEmailLoginCode 只给已注册邮箱发码。码保存为摘要，10 分钟内有效，发出去的明文不再落库。
func (authenticator *Authenticator) SendEmailLoginCode(ctx context.Context, email string) error {
	email = normalizeEmail(email)
	if _, err := authenticator.users.UserByEmail(ctx, email); err != nil {
		if merr.IsNotFound(err) {
			return merr.Client("email is not registered")
		}
		return authenticator.wrap(err)
	}
	code, err := randomDigits(6)
	if err != nil {
		return merr.Server("request failed")
	}
	if err := authenticator.emailCodes.SaveEmailCode(ctx, email, authenticator.digest(code), time.Now().Add(10*time.Minute)); err != nil {
		return authenticator.wrap(err)
	}
	if err := authenticator.sendMail(email, code); err != nil {
		return err
	}
	return nil
}

// EmailLogin 消费邮箱验证码。同一封码不能登录两次。
func (authenticator *Authenticator) EmailLogin(ctx context.Context, email, code string) (*bo.Session, error) {
	email = normalizeEmail(email)
	ok, err := authenticator.emailCodes.ConsumeEmailCode(ctx, email, authenticator.digest(code), time.Now())
	if err != nil {
		return nil, authenticator.wrap(err)
	}
	if !ok {
		return nil, merr.Unauthenticated("invalid email or code")
	}
	account, err := authenticator.users.UserByEmail(ctx, email)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.Unauthenticated("invalid email or code")
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
	cleaned := publicAccount(*account)
	return &cleaned, nil
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

func (authenticator *Authenticator) ChangeEmail(ctx context.Context, email string) (*bo.Account, error) {
	uid, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	email = normalizeEmail(email)
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
	cleaned := publicAccount(*account)
	return &cleaned, nil
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
	cleaned := publicAccount(*account)
	return &cleaned, nil
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
	cleaned := publicAccount(*account)
	return &cleaned, nil
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
	cleaned := publicAccount(*account)
	return &cleaned, nil
}

// session 签发访问令牌和刷新令牌。封禁用户不能拿到新会话，即使密码正确。
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
	return &bo.Session{AccessToken: access, RefreshToken: refresh, User: publicAccount(*account)}, nil
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

func (authenticator *Authenticator) sendMail(to, code string) error {
	if authenticator.mail == nil {
		return merr.Server("smtp is not configured")
	}
	return authenticator.mail.Send(to, "Alioth login code", "Your Alioth login code is "+code+"\r\nIt expires in 10 minutes.\r\n")
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

func publicAccount(account bo.Account) bo.Account {
	account.PasswordHash = ""
	return account
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
	case megrezv1.EvalService_ClaimRules_FullMethodName, megrezv1.EvalService_ReportEval_FullMethodName, megrezv1.EvalService_ReportWorkerStatus_FullMethodName:
		return "megrez", true
	case merakv1.NotifyService_ClaimTasks_FullMethodName, merakv1.NotifyService_MarkSending_FullMethodName, merakv1.NotifyService_ReportNotify_FullMethodName, merakv1.NotifyService_ReportWorkerStatus_FullMethodName:
		return "merak", true
	default:
		return "", false
	}
}

func publicOperation(operation string) bool {
	switch operation {
	case v1.OperationHealthServiceCheck,
		v1.OperationCaptchaServiceGetCaptcha,
		v1.OperationAuthServiceRegister,
		v1.OperationAuthServicePasswordLogin,
		v1.OperationAuthServiceSendEmailLoginCode,
		v1.OperationAuthServiceEmailLogin,
		v1.OperationAuthServiceOAuth2Login,
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
		v1.OperationNamespaceServiceCreateNamespace,
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
		v1.OperationWorkerServiceListWorkerInstances:
		return true
	default:
		return false
	}
}

func randomCode(n int) (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	return randomFrom(alphabet, n)
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
