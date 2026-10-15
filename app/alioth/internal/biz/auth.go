package biz

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/pkg/merr"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type smtpConfig struct {
	host     string
	port     int
	username string
	password string
	from     string
}

type oauthClient struct {
	id     string
	secret string
}

type IdentityUsecase struct {
	repo        IdentityRepo
	secret      string
	accessTTL   time.Duration
	refreshTTL  time.Duration
	serviceKeys map[string]struct{}
	smtp        smtpConfig
	oauth       map[string]oauthClient
	log         *log.Helper
}

func NewIdentityUsecase(repo IdentityRepo, auth *conf.Auth, logger log.Logger) (*IdentityUsecase, error) {
	if auth == nil || auth.GetJwtSecret() == "" {
		return nil, fmt.Errorf("jwt secret is empty")
	}
	access := time.Duration(auth.GetAccessTtlSeconds()) * time.Second
	if access <= 0 {
		access = 2 * time.Hour
	}
	refresh := time.Duration(auth.GetRefreshTtlSeconds()) * time.Second
	if refresh <= 0 {
		refresh = 7 * 24 * time.Hour
	}
	keys := map[string]struct{}{}
	for _, item := range auth.GetServiceKeys() {
		if item.GetKey() != "" {
			keys[item.GetKey()] = struct{}{}
		}
	}
	clients := map[string]oauthClient{}
	for _, item := range auth.GetOauth() {
		clients[strings.ToUpper(item.GetProvider())] = oauthClient{id: item.GetClientId(), secret: item.GetClientSecret()}
	}
	mail := smtpConfig{}
	if auth.GetSmtp() != nil {
		mail = smtpConfig{
			host:     auth.GetSmtp().GetHost(),
			port:     int(auth.GetSmtp().GetPort()),
			username: auth.GetSmtp().GetUsername(),
			password: auth.GetSmtp().GetPassword(),
			from:     auth.GetSmtp().GetFrom(),
		}
	}
	return &IdentityUsecase{
		repo:        repo,
		secret:      auth.GetJwtSecret(),
		accessTTL:   access,
		refreshTTL:  refresh,
		serviceKeys: keys,
		smtp:        mail,
		oauth:       clients,
		log:         log.NewHelper(logger),
	}, nil
}

func (uc *IdentityUsecase) Authenticate(ctx context.Context, operation, authorization, namespace string) (context.Context, error) {
	token := bearerToken(authorization)
	if isServiceKey(token) {
		if _, ok := uc.serviceKeys[token]; !ok {
			return nil, merr.Forbidden("service key is not accepted")
		}
		if !publicOperation(operation) {
			return nil, merr.Forbidden("service key cannot call this api")
		}
		return ctx, nil
	}
	if publicOperation(operation) {
		return ctx, nil
	}
	if token == "" {
		return nil, merr.Unauthenticated("missing token")
	}
	uid, kind, err := uc.parseToken(token)
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
	account, err := uc.repo.UserByUID(ctx, uid)
	if err != nil || account.Status != "ACTIVE" {
		return nil, merr.Unauthenticated("user is not active")
	}
	ctx = WithUser(ctx, account.UID)
	if namespaceOptional(operation) {
		if namespace != "" {
			ctx = WithNamespace(ctx, namespace)
		}
		return ctx, nil
	}
	if namespace == "" {
		return nil, merr.Client("X-Namespace is required")
	}
	return WithNamespace(ctx, namespace), nil
}

func (uc *IdentityUsecase) GetCaptcha(ctx context.Context) (string, string, error) {
	code, err := randomCode(4)
	if err != nil {
		return "", "", merr.Server("request failed")
	}
	image, err := drawCaptcha(code)
	if err != nil {
		uc.log.Errorf("draw captcha: %v", err)
		return "", "", merr.Server("request failed")
	}
	id := newUID()
	if err := uc.repo.SaveCaptcha(ctx, id, uc.digest(code), time.Now().Add(5*time.Minute)); err != nil {
		return "", "", uc.wrap(err)
	}
	return id, image, nil
}

func (uc *IdentityUsecase) Register(ctx context.Context, email, password, name, captchaID, captchaCode string) (*Session, error) {
	email = normalizeEmail(email)
	if !strings.Contains(email, "@") || len(password) < 8 || captchaID == "" || strings.TrimSpace(captchaCode) == "" {
		return nil, merr.Client("invalid register request")
	}
	ok, err := uc.repo.ConsumeCaptcha(ctx, captchaID, uc.digest(captchaCode), time.Now())
	if err != nil {
		return nil, uc.wrap(err)
	}
	if !ok {
		return nil, merr.Client("captcha is invalid")
	}
	if _, err := uc.repo.UserByEmail(ctx, email); err == nil {
		return nil, merr.Client("email already exists")
	} else if !errors.Is(err, ErrNotFound) {
		return nil, uc.wrap(err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, merr.Server("request failed")
	}
	if strings.TrimSpace(name) == "" {
		name = strings.Split(email, "@")[0]
	}
	account, err := uc.repo.CreateUser(ctx, Account{
		UID:          newUID(),
		Name:         name,
		Nickname:     name,
		Email:        email,
		Status:       "ACTIVE",
		PasswordHash: string(hash),
	})
	if err != nil {
		return nil, uc.wrap(err)
	}
	return uc.session(account)
}

func (uc *IdentityUsecase) PasswordLogin(ctx context.Context, email, password string) (*Session, error) {
	account, err := uc.repo.UserByEmail(ctx, normalizeEmail(email))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, merr.Unauthenticated("invalid email or password")
		}
		return nil, uc.wrap(err)
	}
	if bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password)) != nil {
		return nil, merr.Unauthenticated("invalid email or password")
	}
	return uc.session(account)
}

func (uc *IdentityUsecase) SendEmailLoginCode(ctx context.Context, email string) error {
	email = normalizeEmail(email)
	if _, err := uc.repo.UserByEmail(ctx, email); err != nil {
		if errors.Is(err, ErrNotFound) {
			return merr.Client("email is not registered")
		}
		return uc.wrap(err)
	}
	code, err := randomDigits(6)
	if err != nil {
		return merr.Server("request failed")
	}
	if err := uc.repo.SaveEmailCode(ctx, email, uc.digest(code), time.Now().Add(10*time.Minute)); err != nil {
		return uc.wrap(err)
	}
	if err := uc.sendMail(email, code); err != nil {
		uc.log.Errorf("smtp send failed")
		return merr.Server("failed to send email")
	}
	return nil
}

func (uc *IdentityUsecase) EmailLogin(ctx context.Context, email, code string) (*Session, error) {
	email = normalizeEmail(email)
	ok, err := uc.repo.ConsumeEmailCode(ctx, email, uc.digest(code), time.Now())
	if err != nil {
		return nil, uc.wrap(err)
	}
	if !ok {
		return nil, merr.Unauthenticated("invalid email or code")
	}
	account, err := uc.repo.UserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, merr.Unauthenticated("invalid email or code")
		}
		return nil, uc.wrap(err)
	}
	return uc.session(account)
}

func (uc *IdentityUsecase) Refresh(ctx context.Context) (string, error) {
	uid, ok := UserID(ctx)
	if !ok {
		return "", merr.Unauthenticated("missing token")
	}
	account, err := uc.repo.UserByUID(ctx, uid)
	if err != nil || account.Status != "ACTIVE" {
		return "", merr.Unauthenticated("user is not active")
	}
	return uc.sign(uid, "access", uc.accessTTL)
}

func (uc *IdentityUsecase) Self(ctx context.Context) (*Account, error) {
	uid, ok := UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	account, err := uc.repo.UserByUID(ctx, uid)
	if err != nil {
		return nil, uc.wrap(err)
	}
	cleaned := publicAccount(*account)
	return &cleaned, nil
}

func (uc *IdentityUsecase) ChangeEmail(ctx context.Context, email string) (*Account, error) {
	uid, ok := UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	email = normalizeEmail(email)
	existing, err := uc.repo.UserByEmail(ctx, email)
	if err == nil && existing.UID != uid {
		return nil, merr.Client("email already exists")
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, uc.wrap(err)
	}
	account, err := uc.repo.UpdateUserProfile(ctx, uid, &email, nil, nil)
	if err != nil {
		return nil, uc.wrap(err)
	}
	cleaned := publicAccount(*account)
	return &cleaned, nil
}

func (uc *IdentityUsecase) ChangeAvatar(ctx context.Context, avatar string) (*Account, error) {
	uid, ok := UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	account, err := uc.repo.UpdateUserProfile(ctx, uid, nil, &avatar, nil)
	if err != nil {
		return nil, uc.wrap(err)
	}
	cleaned := publicAccount(*account)
	return &cleaned, nil
}

func (uc *IdentityUsecase) ChangeRemark(ctx context.Context, remark string) (*Account, error) {
	uid, ok := UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	account, err := uc.repo.UpdateUserProfile(ctx, uid, nil, nil, &remark)
	if err != nil {
		return nil, uc.wrap(err)
	}
	cleaned := publicAccount(*account)
	return &cleaned, nil
}

func (uc *IdentityUsecase) SelfNamespaces(ctx context.Context) ([]Namespace, error) {
	uid, ok := UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	items, err := uc.repo.ListUserNamespaces(ctx, uid)
	return items, uc.wrap(err)
}

func (uc *IdentityUsecase) SelfPermissions(ctx context.Context) ([]string, error) {
	uid, ok := UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	namespace, ok := NamespaceID(ctx)
	if !ok {
		return []string{}, nil
	}
	perms, err := uc.repo.Permissions(ctx, namespace, uid)
	if err != nil {
		return nil, uc.wrap(err)
	}
	names := make([]string, 0, len(perms))
	for name := range perms {
		names = append(names, name)
	}
	return names, nil
}

func (uc *IdentityUsecase) session(account *Account) (*Session, error) {
	if account.Status != "ACTIVE" {
		return nil, merr.Unauthenticated("user is banned")
	}
	access, err := uc.sign(account.UID, "access", uc.accessTTL)
	if err != nil {
		return nil, merr.Server("request failed")
	}
	refresh, err := uc.sign(account.UID, "refresh", uc.refreshTTL)
	if err != nil {
		return nil, merr.Server("request failed")
	}
	return &Session{AccessToken: access, RefreshToken: refresh, User: publicAccount(*account)}, nil
}

func (uc *IdentityUsecase) sign(uid, kind string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"uid":  uid,
		"kind": kind,
		"iat":  now.Unix(),
		"exp":  now.Add(ttl).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uc.secret))
}

func (uc *IdentityUsecase) parseToken(raw string) (string, string, error) {
	token, err := jwt.Parse(raw, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(uc.secret), nil
	})
	if err != nil || !token.Valid {
		return "", "", fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", fmt.Errorf("invalid token")
	}
	uid, _ := claims["uid"].(string)
	kind, _ := claims["kind"].(string)
	if uid == "" || (kind != "access" && kind != "refresh") {
		return "", "", fmt.Errorf("invalid token")
	}
	return uid, kind, nil
}

func (uc *IdentityUsecase) digest(value string) string {
	sum := sha256.Sum256([]byte(uc.secret + ":" + strings.ToLower(strings.TrimSpace(value))))
	return hex.EncodeToString(sum[:])
}

func (uc *IdentityUsecase) sendMail(to, code string) error {
	if uc.smtp.host == "" || uc.smtp.port == 0 || uc.smtp.from == "" {
		return fmt.Errorf("smtp is not configured")
	}
	addr := net.JoinHostPort(uc.smtp.host, strconv.Itoa(uc.smtp.port))
	message := "From: " + uc.smtp.from + "\r\nTo: " + to + "\r\nSubject: Alioth login code\r\n\r\n" + code + "\r\n"
	var auth smtp.Auth
	if uc.smtp.username != "" {
		auth = smtp.PlainAuth("", uc.smtp.username, uc.smtp.password, uc.smtp.host)
	}
	return smtp.SendMail(addr, auth, uc.smtp.from, []string{to}, []byte(message))
}

func (uc *IdentityUsecase) wrap(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrNotFound) {
		return merr.Client("not found")
	}
	if errors.Is(err, ErrConflict) {
		return merr.Client("name already exists")
	}
	if errors.Is(err, ErrInvalid) {
		return merr.Client("invalid request")
	}
	uc.log.Errorf("identity: %v", err)
	return merr.Server("request failed")
}

func publicAccount(account Account) Account {
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

func publicOperation(operation string) bool {
	switch operation {
	case v1.OperationHealthServiceCheck,
		v1.OperationCaptchaServiceGetCaptcha,
		v1.OperationAuthServiceRegister,
		v1.OperationAuthServicePasswordLogin,
		v1.OperationAuthServiceSendEmailLoginCode,
		v1.OperationAuthServiceEmailLogin,
		v1.OperationAuthServiceOAuth2Login:
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
		v1.OperationNamespaceServiceCreateNamespace:
		return true
	default:
		return false
	}
}

func newUID() string {
	var raw [16]byte
	_, _ = rand.Read(raw[:])
	return hex.EncodeToString(raw[:])
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
		return "", err
	}
	for i := range buf {
		buf[i] = alphabet[int(raw[i])%len(alphabet)]
	}
	return string(buf), nil
}
