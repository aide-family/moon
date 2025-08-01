package service

import (
	"context"
	nhttp "net/http"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
	"golang.org/x/oauth2"

	"github.com/aide-family/moon/cmd/palace/internal/biz"
	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/system"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/conf"
	"github.com/aide-family/moon/cmd/palace/internal/helper/permission"
	"github.com/aide-family/moon/cmd/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/api/palace"
	"github.com/aide-family/moon/pkg/api/palace/common"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/strutil"
)

func NewAuthService(
	bc *conf.Bootstrap,
	authBiz *biz.Auth,
	permissionBiz *biz.Permission,
	menuBiz *biz.Menu,
	messageBiz *biz.Message,
	logger log.Logger,
) *AuthService {
	return &AuthService{
		authBiz:       authBiz,
		permissionBiz: permissionBiz,
		menuBiz:       menuBiz,
		messageBiz:    messageBiz,
		oauth2List:    builderOAuth2List(bc.GetAuth().GetOauth2()),
		helper:        log.NewHelper(log.With(logger, "module", "service.auth")),
	}
}

type AuthService struct {
	palace.UnimplementedAuthServer
	authBiz       *biz.Auth
	permissionBiz *biz.Permission
	menuBiz       *biz.Menu
	messageBiz    *biz.Message
	oauth2List    []*palace.OAuth2ListReply_OAuthItem
	helper        *log.Helper
}

func builderOAuth2List(oauth2 *conf.Auth_OAuth2) []*palace.OAuth2ListReply_OAuthItem {
	if !oauth2.GetEnable() {
		return nil
	}
	list := oauth2.GetConfigs()
	oauthList := make([]*palace.OAuth2ListReply_OAuthItem, 0, len(list))
	for _, oauth := range list {
		app := vobj.OAuthAPP(oauth.GetApp())
		if !app.Exist() || app.IsUnknown() {
			continue
		}
		oauthList = append(oauthList, &palace.OAuth2ListReply_OAuthItem{
			Icon:     app.String(),
			Label:    strutil.Title(app.String(), "login"),
			Redirect: oauth.GetLoginUrl(),
		})
	}
	return oauthList
}

func login(loginSign *bo.LoginSign, err error) (*palace.LoginReply, error) {
	if err != nil {
		return nil, err
	}
	return build.LoginReply(loginSign), nil
}

func (s *AuthService) GetCaptcha(ctx context.Context, _ *common.EmptyRequest) (*palace.GetCaptchaReply, error) {
	captchaBo, err := s.authBiz.GetCaptcha(ctx)
	if err != nil {
		return nil, err
	}
	return &palace.GetCaptchaReply{
		CaptchaId:      captchaBo.ID,
		CaptchaImg:     captchaBo.B64s,
		ExpiredSeconds: int32(captchaBo.ExpiredSeconds),
	}, nil
}

func (s *AuthService) LoginByPassword(ctx context.Context, req *palace.LoginByPasswordRequest) (*palace.LoginReply, error) {
	captchaReq := req.GetCaptcha()
	captchaVerify := &bo.CaptchaVerify{
		ID:     captchaReq.GetCaptchaId(),
		Answer: captchaReq.GetAnswer(),
		Clear:  true,
	}

	if err := s.authBiz.VerifyCaptcha(ctx, captchaVerify); err != nil {
		return nil, err
	}
	loginReq := &bo.LoginByPassword{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}
	return login(s.authBiz.LoginByPassword(ctx, loginReq))
}

func (s *AuthService) Logout(ctx context.Context, req *palace.LogoutRequest) (*palace.LogoutReply, error) {
	token, ok := permission.GetTokenByContext(ctx)
	if !ok {
		return nil, merr.ErrorUnauthorized("token error")
	}
	if err := s.authBiz.Logout(ctx, token); err != nil {
		return nil, err
	}
	return &palace.LogoutReply{Redirect: req.GetRedirect()}, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, req *palace.VerifyEmailRequest) (*palace.VerifyEmailReply, error) {
	captchaReq := req.GetCaptcha()
	captchaVerify := &bo.CaptchaVerify{
		ID:     captchaReq.GetCaptchaId(),
		Answer: captchaReq.GetAnswer(),
		Clear:  true,
	}

	if err := s.authBiz.VerifyCaptcha(ctx, captchaVerify); err != nil {
		return nil, err
	}
	params := &bo.VerifyEmailParams{
		Email:        req.GetEmail(),
		SendEmailFun: s.messageBiz.SendEmail,
	}
	if err := s.authBiz.VerifyEmail(ctx, params); err != nil {
		return nil, err
	}
	return &palace.VerifyEmailReply{ExpiredSeconds: int32(5 * time.Minute.Seconds())}, nil
}

func (s *AuthService) LoginByEmail(ctx context.Context, req *palace.LoginByEmailRequest) (*palace.LoginReply, error) {
	userDo := &system.User{
		BaseModel: do.BaseModel{},
		Username:  req.GetUsername(),
		Nickname:  req.GetNickname(),
		Email:     req.GetEmail(),
		Remark:    req.GetRemark(),
		Gender:    vobj.Gender(req.GetGender()),
		Position:  vobj.PositionUser,
		Status:    vobj.UserStatusNormal,
	}
	params := &bo.LoginWithEmailParams{
		Code:         req.GetCode(),
		User:         userDo,
		SendEmailFun: s.messageBiz.SendEmail,
	}
	return login(s.authBiz.LoginWithEmail(ctx, params))
}

func (s *AuthService) OAuthLoginByEmail(ctx context.Context, req *palace.OAuthLoginByEmailRequest) (*palace.LoginReply, error) {
	oauthParams := &bo.OAuthLoginParams{
		APP:          vobj.OAuthAPP(req.GetApp()),
		Code:         req.GetCode(),
		Email:        req.GetEmail(),
		OpenID:       req.GetOpenId(),
		Token:        req.GetToken(),
		SendEmailFun: s.messageBiz.SendEmail,
	}
	return login(s.authBiz.OAuthLoginWithEmail(ctx, oauthParams))
}

func (s *AuthService) VerifyToken(ctx context.Context, token string) (userDo do.User, err error) {
	return s.authBiz.VerifyToken(ctx, token)
}

func (s *AuthService) VerifyPermission(ctx context.Context) error {
	return s.permissionBiz.VerifyPermission(ctx)
}

func (s *AuthService) RefreshToken(ctx context.Context, _ *common.EmptyRequest) (*palace.LoginReply, error) {
	token, ok := permission.GetTokenByContext(ctx)
	if !ok {
		return nil, merr.ErrorUnauthorized("token error")
	}
	userID, ok := permission.GetUserIDByContext(ctx)
	if !ok {
		return nil, merr.ErrorUnauthorized("token error")
	}
	refreshReq := &bo.RefreshToken{
		Token:  token,
		UserID: userID,
	}
	return login(s.authBiz.RefreshToken(ctx, refreshReq))
}

func (s *AuthService) OAuth2List(_ context.Context, _ *common.EmptyRequest) (*palace.OAuth2ListReply, error) {
	return &palace.OAuth2ListReply{Items: s.oauth2List}, nil
}

func (s *AuthService) GetFilingInformation(ctx context.Context, _ *common.EmptyRequest) (*palace.GetFilingInformationReply, error) {
	filingInfo, err := s.authBiz.GetFilingInformation(ctx)
	if err != nil {
		return nil, err
	}
	return &palace.GetFilingInformationReply{
		Url:               filingInfo.URL,
		FilingInformation: filingInfo.Information,
	}, nil
}

func (s *AuthService) GetSelfMenuTree(ctx context.Context, _ *common.EmptyRequest) (*palace.GetSelfMenuTreeReply, error) {
	menus, err := s.menuBiz.SelfMenus(ctx)
	if err != nil {
		return nil, err
	}
	menus = slices.MapFilter(menus, func(menu do.Menu) (do.Menu, bool) {
		return menu, menu.GetMenuCategory().IsMenu() && menu.GetStatus().IsEnable()
	})
	return &palace.GetSelfMenuTreeReply{
		Items: build.ToMenuTree(menus),
	}, nil
}

func (s *AuthService) ReplaceUserRole(ctx context.Context, req *palace.ReplaceUserRoleRequest) (*common.EmptyReply, error) {
	updateReq := &bo.ReplaceUserRoleReq{
		UserID: req.GetUserId(),
		Roles:  req.GetRoleIds(),
	}
	if err := s.authBiz.ReplaceUserRole(ctx, updateReq); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (s *AuthService) ReplaceMemberRole(ctx context.Context, req *palace.ReplaceMemberRoleRequest) (*common.EmptyReply, error) {
	updateReq := &bo.ReplaceMemberRoleReq{
		MemberID: req.GetMemberId(),
		Roles:    req.GetRoleIds(),
	}
	if err := s.authBiz.ReplaceMemberRole(ctx, updateReq); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

// OAuthLogin oauth login
func (s *AuthService) OAuthLogin(app vobj.OAuthAPP) http.HandlerFunc {
	return func(ctx http.Context) error {
		oauthConf, err := s.authBiz.GetOAuthConf(app, vobj.OAuthFromAdmin)
		if err != nil {
			return err
		}
		// Redirect to the specified URL
		url := oauthConf.AuthCodeURL("state", oauth2.AccessTypeOnline)
		req := ctx.Request()
		resp := ctx.Response()
		resp.Header().Set("Location", url)
		resp.WriteHeader(nhttp.StatusTemporaryRedirect)
		ctx.Reset(resp, req)
		return nil
	}
}

// OAuthLoginCallback oauth callback
func (s *AuthService) OAuthLoginCallback(app vobj.OAuthAPP) http.HandlerFunc {
	return func(ctx http.Context) error {
		code := ctx.Query().Get("code")
		if code == "" {
			return merr.ErrorBadRequest("code is empty")
		}
		params := &bo.OAuthLoginParams{
			APP:          app,
			From:         vobj.OAuthFromAdmin,
			Code:         code,
			SendEmailFun: s.messageBiz.SendEmail,
			Email:        "",
			OpenID:       "",
			Token:        "",
		}
		loginRedirect, err := s.authBiz.OAuthLogin(ctx, params)
		if err != nil {
			return err
		}
		// Redirect to the specified URL
		req := ctx.Request()
		resp := ctx.Response()

		resp.Header().Set("Location", loginRedirect)
		resp.WriteHeader(nhttp.StatusTemporaryRedirect)
		ctx.Reset(resp, req)
		return nil
	}
}
