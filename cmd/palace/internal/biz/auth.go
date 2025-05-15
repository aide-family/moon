package biz

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/oauth2"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/repository"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/cmd/palace/internal/conf"
	"github.com/moon-monitor/moon/cmd/palace/internal/helper/middleware"
	"github.com/moon-monitor/moon/cmd/palace/internal/helper/permission"
	"github.com/moon-monitor/moon/pkg/merr"
	"github.com/moon-monitor/moon/pkg/util/hash"
	"github.com/moon-monitor/moon/pkg/util/password"
	"github.com/moon-monitor/moon/pkg/util/safety"
	"github.com/moon-monitor/moon/pkg/util/validate"
)

func buildOAuthConf(c *conf.Auth_OAuth2) *safety.Map[vobj.OAuthAPP, *oauth2.Config] {
	oauth2Map := safety.NewMap[vobj.OAuthAPP, *oauth2.Config]()
	for _, config := range c.GetConfigs() {
		oauth2Map.Set(vobj.OAuthAPP(config.GetApp()), &oauth2.Config{
			ClientID:     config.GetClientId(),
			ClientSecret: config.GetClientSecret(),
			Endpoint: oauth2.Endpoint{
				AuthURL:  config.GetAuthUrl(),
				TokenURL: config.GetTokenUrl(),
			},
			RedirectURL: config.GetCallbackUri(),
			Scopes:      config.GetScopes(),
		})
	}
	return oauth2Map
}

func NewAuthBiz(
	bc *conf.Bootstrap,
	userRepo repository.User,
	captchaRepo repository.Captcha,
	cacheRepo repository.Cache,
	oauthRepo repository.OAuth,
	transaction repository.Transaction,
	logger log.Logger,
) *AuthBiz {
	return &AuthBiz{
		bc:           bc,
		redirectURL:  bc.GetAuth().GetOauth2().GetRedirectUri(),
		oauthConfigs: buildOAuthConf(bc.GetAuth().GetOauth2()),
		userRepo:     userRepo,
		captchaRepo:  captchaRepo,
		cacheRepo:    cacheRepo,
		oauthRepo:    oauthRepo,
		transaction:  transaction,
		helper:       log.NewHelper(log.With(logger, "module", "biz.auth")),
	}
}

type AuthBiz struct {
	bc           *conf.Bootstrap
	redirectURL  string
	oauthConfigs *safety.Map[vobj.OAuthAPP, *oauth2.Config]

	userRepo    repository.User
	captchaRepo repository.Captcha
	cacheRepo   repository.Cache
	oauthRepo   repository.OAuth
	transaction repository.Transaction
	helper      *log.Helper
}

// GetCaptcha get image captchaRepo
func (a *AuthBiz) GetCaptcha(ctx context.Context) (*bo.Captcha, error) {
	return a.captchaRepo.Generate(ctx)
}

// VerifyCaptcha Captcha
func (a *AuthBiz) VerifyCaptcha(ctx context.Context, req *bo.CaptchaVerify) error {
	verify := a.captchaRepo.Verify(ctx, req)
	if !verify {
		return merr.ErrorCaptchaError("captcha err").WithMetadata(map[string]string{
			"captcha.answer": "The verification code is incorrect. Please retrieve a new one and try again.",
		})
	}
	return nil
}

// Logout token logout
func (a *AuthBiz) Logout(ctx context.Context, token string) error {
	return a.cacheRepo.BanToken(ctx, token)
}

// VerifyToken verify token
func (a *AuthBiz) VerifyToken(ctx context.Context, token string) error {
	if err := a.cacheRepo.VerifyToken(ctx, token); err != nil {
		return err
	}
	userID, ok := permission.GetUserIDByContext(ctx)
	if !ok {
		return merr.ErrorInvalidToken("token is invalid")
	}
	userDo, err := a.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if !userDo.GetStatus().IsNormal() {
		return merr.ErrorUserForbidden("user forbidden")
	}
	return nil
}

// LoginByPassword login by password
func (a *AuthBiz) LoginByPassword(ctx context.Context, req *bo.LoginByPassword) (*bo.LoginSign, error) {
	user, err := a.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, merr.ErrorPasswordError("password error").WithCause(err)
	}
	if !user.ValidatePassword(req.Password) {
		return nil, merr.ErrorPasswordError("password error")
	}
	return a.login(user)
}

// RefreshToken refresh token
func (a *AuthBiz) RefreshToken(ctx context.Context, req *bo.RefreshToken) (*bo.LoginSign, error) {
	if err := a.VerifyToken(ctx, req.Token); err != nil {
		return nil, err
	}
	userDo, err := a.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := a.cacheRepo.BanToken(ctx, req.Token); err != nil {
			a.helper.WithContext(ctx).Errorf("refresh token unban err: %v", err)
		}
	}()
	return a.login(userDo)
}

func (a *AuthBiz) login(userDo do.User) (*bo.LoginSign, error) {
	base := &middleware.JwtBaseInfo{
		UserID:   userDo.GetID(),
		Username: userDo.GetUsername(),
		Nickname: userDo.GetNickname(),
		Avatar:   userDo.GetAvatar(),
		Gender:   userDo.GetGender(),
	}
	token, err := middleware.NewJwtClaims(a.bc.GetAuth().GetJwt(), base).GetToken()
	if err != nil {
		return nil, err
	}
	return &bo.LoginSign{
		Base:           base,
		ExpiredSeconds: int64(a.bc.GetAuth().GetJwt().GetExpire().AsDuration().Seconds()),
		Token:          token,
	}, nil
}

// GetOAuthConf 获取oauth配置
func (a *AuthBiz) GetOAuthConf(provider vobj.OAuthAPP) (*oauth2.Config, error) {
	config, ok := a.oauthConfigs.Get(provider)
	if !ok {
		return nil, merr.ErrorInternalServerError("not support oauth provider")
	}
	return config, nil
}

func (a *AuthBiz) OAuthLogin(ctx context.Context, req *bo.OAuthLoginParams) (string, error) {
	switch req.APP {
	case vobj.OAuthAPPGithub:
		return a.githubLogin(ctx, req.Code, req.SendEmailFun)
	case vobj.OAuthAPPGitee:
		return a.giteeLogin(ctx, req.Code, req.SendEmailFun)
	default:
		return "", merr.ErrorInternalServerError("not support oauth provider")
	}
}

func (a *AuthBiz) githubLogin(ctx context.Context, code string, sendEmailFunc bo.SendEmailFun) (string, error) {
	githubOAuthConf, err := a.GetOAuthConf(vobj.OAuthAPPGithub)
	if err != nil {
		return "", err
	}

	token, err := githubOAuthConf.Exchange(ctx, code)
	if err != nil {
		return "", err
	}
	// 使用token来获取用户信息
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	userResp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return "", err
	}
	body := userResp.Body
	defer body.Close()
	var userInfo bo.GithubUser
	if err := json.NewDecoder(body).Decode(&userInfo); err != nil {
		return "", err
	}

	return a.oauthLogin(ctx, &userInfo, sendEmailFunc)
}

func (a *AuthBiz) giteeLogin(ctx context.Context, code string, sendEmailFunc bo.SendEmailFun) (string, error) {
	giteeOAuthConf, err := a.GetOAuthConf(vobj.OAuthAPPGitee)
	if err != nil {
		return "", err
	}
	opts := []oauth2.AuthCodeOption{
		// https://gitee.com/oauth/token?grant_type=authorization_code&code={code}&client_id={client_id}&redirect_uri={redirect_uri}&client_secret={client_secret}
		oauth2.SetAuthURLParam("grant_type", "authorization_code"),
		oauth2.SetAuthURLParam("client_secret", giteeOAuthConf.ClientSecret),
		oauth2.SetAuthURLParam("client_id", giteeOAuthConf.ClientID),
		oauth2.SetAuthURLParam("redirect_uri", giteeOAuthConf.RedirectURL),
		oauth2.SetAuthURLParam("code", code),
	}
	token, err := giteeOAuthConf.Exchange(context.Background(), code, opts...)
	if err != nil {
		return "", err
	}
	// 使用token来获取用户信息
	client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))

	resp, err := client.Get("https://gitee.com/api/v5/user")
	if err != nil {
		return "", err
	}
	body := resp.Body
	defer body.Close()
	var userInfo bo.GiteeUser
	if err := json.NewDecoder(body).Decode(&userInfo); err != nil {
		return "", err
	}

	return a.oauthLogin(ctx, &userInfo, sendEmailFunc)
}

func (a *AuthBiz) oauthUserFirstOrCreate(ctx context.Context, userInfo bo.IOAuthUser, sendEmail bo.SendEmailFun) (do.UserOAuth, error) {
	oauthUserDoExist := true
	oauthUserDo, err := a.oauthRepo.FindByOpenID(ctx, userInfo.GetOpenID(), userInfo.GetAPP())
	if err != nil {
		if !merr.IsUserNotFound(err) {
			return nil, err
		}
		oauthUserDoExist = false
	}
	if validate.IsNil(oauthUserDo) || oauthUserDo.GetUserID() == 0 {
		userDo, err := a.userRepo.FindByEmail(ctx, userInfo.GetEmail())
		if err != nil {
			if !merr.IsUserNotFound(err) {
				return nil, err
			}
		}
		if validate.IsNotNil(userDo) {
			userInfo.WithUserID(userDo.GetID())
		}
	}

	err = a.transaction.MainExec(ctx, func(ctx context.Context) error {
		if !oauthUserDoExist {
			oauthUserDo, err = a.oauthRepo.Create(ctx, userInfo)
			if err != nil {
				return err
			}
		}
		if validate.IsNil(oauthUserDo.GetUser()) {
			// 创建用户
			userDo, err := a.userRepo.CreateUserWithOAuthUser(ctx, userInfo, sendEmail)
			if err != nil {
				return err
			}
			oauthUserDo.SetUser(userDo)
			oauthUserDo, err = a.oauthRepo.SetUser(ctx, oauthUserDo)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return oauthUserDo, nil
}

func (a *AuthBiz) oauthLogin(ctx context.Context, userInfo bo.IOAuthUser, sendEmail bo.SendEmailFun) (string, error) {
	oauthUserDo, err := a.oauthUserFirstOrCreate(ctx, userInfo, sendEmail)
	if err != nil {
		return "", err
	}

	userDo := oauthUserDo.GetUser()
	if validate.IsNil(userDo) || validate.CheckEmail(userDo.GetEmail()) != nil {
		oauthParams := &bo.OAuthLoginParams{
			APP:    userInfo.GetAPP(),
			OpenID: oauthUserDo.GetOpenID(),
			Token:  hash.MD5(password.GenerateRandomPassword(64)),
		}
		if err := a.cacheRepo.CacheVerifyOAuthToken(ctx, oauthParams); err != nil {
			return "", err
		}
		redirect := fmt.Sprintf("%s?open_id=%s&app=%d&token=%s#/oauth/register/email", a.redirectURL, oauthParams.OpenID, userInfo.GetAPP(), oauthParams.Token)
		return redirect, nil
	}

	loginSign, err := a.login(userDo)
	if err != nil {
		return "", err
	}
	redirect := fmt.Sprintf("%s?token=%s", a.redirectURL, loginSign.Token)
	return redirect, nil
}

// OAuthLoginWithEmail oauth2 set email login
func (a *AuthBiz) OAuthLoginWithEmail(ctx context.Context, oauthParams *bo.OAuthLoginParams) (*bo.LoginSign, error) {
	if err := a.cacheRepo.VerifyEmailCode(ctx, oauthParams.Email, oauthParams.Code); err != nil {
		return nil, err
	}

	if err := a.cacheRepo.VerifyOAuthToken(ctx, oauthParams); err != nil {
		return nil, err
	}

	oauthUserDo, err := a.oauthRepo.FindByOpenID(ctx, oauthParams.OpenID, oauthParams.APP)
	if err != nil {
		return nil, err
	}
	userDo := oauthUserDo.GetUser()
	if validate.IsNil(userDo) {
		return nil, merr.ErrorUnauthorized("oauth unauthorized")
	}
	if userDo.GetEmail() == oauthParams.Email {
		return a.login(userDo)
	}

	userDo.SetEmail(oauthParams.Email)
	user, err := a.userRepo.SetEmail(ctx, userDo, oauthParams.SendEmailFun)
	if err != nil {
		return nil, err
	}
	return a.login(user)
}

// VerifyEmail verify email
func (a *AuthBiz) VerifyEmail(ctx context.Context, req *bo.VerifyEmailParams) error {
	sendEmailParams, err := a.cacheRepo.SendVerifyEmailCode(ctx, req.Email)
	if err != nil {
		return err
	}
	return req.SendEmailFun(ctx, sendEmailParams)
}

// LoginWithEmail 邮箱登录
func (a *AuthBiz) LoginWithEmail(ctx context.Context, req *bo.LoginWithEmailParams) (*bo.LoginSign, error) {
	if err := a.cacheRepo.VerifyEmailCode(ctx, req.GetEmail(), req.Code); err != nil {
		return nil, err
	}
	userDo, err := a.userRepo.FindByEmail(ctx, req.GetEmail())
	if err == nil {
		return a.login(userDo)
	}
	userDo = req.User
	userDo, err = a.userRepo.Create(ctx, userDo, req.SendEmailFun)
	if err != nil {
		return nil, err
	}
	return a.login(userDo)
}

// GetFilingInformation get filing information
func (a *AuthBiz) GetFilingInformation(_ context.Context) (*bo.FilingInformation, error) {
	filing := a.bc.GetFiling()
	filingInfo := &bo.FilingInformation{
		URL:         filing.GetUrl(),
		Information: filing.GetInformation(),
	}
	return filingInfo, nil
}

func (a *AuthBiz) ReplaceUserRole(ctx context.Context, req *bo.ReplaceUserRoleReq) error {
	return nil
}

func (a *AuthBiz) ReplaceMemberRole(ctx context.Context, req *bo.ReplaceMemberRoleReq) error {
	return nil
}
