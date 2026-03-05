package service

import (
	"context"

	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/oauth"
	"github.com/aide-family/magicbox/safety"

	"github.com/aide-family/goddess/internal/biz"
	"github.com/aide-family/goddess/internal/biz/bo"
	apiv1 "github.com/aide-family/goddess/pkg/api/v1"
)

type AuthService struct {
	apiv1.UnimplementedAuthServiceServer
	loginBiz    *biz.LoginBiz
	emailBiz    *biz.Email
	captchaBiz  *biz.Captcha
	sendedEmail *safety.SyncMap[string, string]
}

func NewAuthService(loginBiz *biz.LoginBiz, emailBiz *biz.Email, captchaBiz *biz.Captcha) *AuthService {
	return &AuthService{
		loginBiz:    loginBiz,
		emailBiz:    emailBiz,
		captchaBiz:  captchaBiz,
		sendedEmail: safety.NewSyncMap(make(map[string]string)),
	}
}

func (s *AuthService) OAuth2Login(ctx context.Context, req *oauth.OAuth2LoginRequest) (*apiv1.LoginReply, error) {
	loginBo := bo.NewOAuth2LoginBo(req)
	token, err := s.loginBiz.Login(ctx, loginBo)
	if err != nil {
		return nil, err
	}
	return &apiv1.LoginReply{Token: token}, nil
}

func (s *AuthService) SendEmailLoginCode(ctx context.Context, req *apiv1.SendEmailLoginCodeRequest) (*apiv1.SendEmailLoginCodeReply, error) {
	if err := s.captchaBiz.Verify(ctx, req.GetCaptchaId(), req.GetCaptchaAnswer()); err != nil {
		return nil, err
	}
	codeID, code, err := s.captchaBiz.EmailLoginCode(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.emailBiz.SendEmailLoginCode(ctx, req.GetEmail(), codeID, code); err != nil {
		return nil, err
	}
	s.sendedEmail.Set(req.GetEmail(), codeID)
	return &apiv1.SendEmailLoginCodeReply{Message: "Email login code sent successfully"}, nil
}

func (s *AuthService) EmailLogin(ctx context.Context, req *apiv1.EmailLoginRequest) (*apiv1.LoginReply, error) {
	codeID, ok := s.sendedEmail.Get(req.GetEmail())
	if !ok {
		return nil, merr.ErrorNotFound("Please send email login code first")
	}
	if err := s.captchaBiz.Verify(ctx, codeID, req.GetCode()); err != nil {
		return nil, err
	}
	token, err := s.loginBiz.EmailLogin(ctx, req.GetEmail())
	if err != nil {
		return nil, err
	}
	return &apiv1.LoginReply{Token: token}, nil
}
