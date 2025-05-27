package portal_service

import (
	"context"
	nhttp "net/http"

	"github.com/go-kratos/kratos/v2/transport/http"
	"golang.org/x/oauth2"

	"github.com/aide-family/moon/cmd/palace/internal/biz"
	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	palacecommon "github.com/aide-family/moon/pkg/api/palace/common"
	portalapi "github.com/aide-family/moon/pkg/api/palace/portal"
	"github.com/aide-family/moon/pkg/merr"
)

func NewAuthService() *AuthService {
	return &AuthService{}
}

type AuthService struct {
	portalapi.UnimplementedAuthServer
	authBiz    *biz.Auth
	messageBiz *biz.Message
}

func (s *AuthService) Login(ctx context.Context, req *portalapi.LoginRequest) (*portalapi.LoginInfo, error) {
	return &portalapi.LoginInfo{}, nil
}

func (s *AuthService) Register(ctx context.Context, req *portalapi.RegisterRequest) (*portalapi.LoginInfo, error) {
	return &portalapi.LoginInfo{}, nil
}

func (s *AuthService) Logout(ctx context.Context, req *palacecommon.EmptyRequest) (*palacecommon.EmptyReply, error) {
	return &palacecommon.EmptyReply{}, nil
}

func (s *AuthService) GetUserInfo(ctx context.Context, req *palacecommon.EmptyRequest) (*palacecommon.UserBaseItem, error) {
	return &palacecommon.UserBaseItem{}, nil
}

func (s *AuthService) GetCaptcha(ctx context.Context, req *palacecommon.EmptyRequest) (*portalapi.GetCaptchaReply, error) {
	return &portalapi.GetCaptchaReply{}, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, req *portalapi.VerifyEmailRequest) (*palacecommon.EmptyReply, error) {
	return &palacecommon.EmptyReply{}, nil
}

func (s *AuthService) VerifyToken(ctx context.Context, token string) (userDo do.User, err error) {
	return
}

func (s *AuthService) OAuthLogin(app vobj.OAuthAPP) http.HandlerFunc {
	return func(ctx http.Context) error {
		oauthConf, err := s.authBiz.GetOAuthConf(app, vobj.OAuthFromPortal)
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

func (s *AuthService) OAuthLoginCallback(app vobj.OAuthAPP) http.HandlerFunc {
	return func(ctx http.Context) error {
		code := ctx.Query().Get("code")
		if code == "" {
			return merr.ErrorBadRequest("code is empty")
		}
		params := &bo.OAuthLoginParams{
			APP:          app,
			From:         vobj.OAuthFromPortal,
			Code:         code,
			SendEmailFun: s.messageBiz.SendEmail,
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
