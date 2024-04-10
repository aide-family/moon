package authservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/aide-family/moon/api/perrors"
	"github.com/aide-family/moon/api/server/auth"
	"github.com/aide-family/moon/app/prom_server/internal/biz"
	"github.com/aide-family/moon/pkg/helper/middler"
	"github.com/aide-family/moon/pkg/helper/prom"
	"github.com/aide-family/moon/pkg/util/captcha"
	"github.com/aide-family/moon/pkg/util/password"
)

type AuthService struct {
	auth.UnimplementedAuthServer
	log *log.Helper

	userBiz    *biz.UserBiz
	captchaBiz *biz.CaptchaBiz
}

func NewAuthService(userBiz *biz.UserBiz, captchaBiz *biz.CaptchaBiz, logger log.Logger) *AuthService {
	return &AuthService{
		log:        log.NewHelper(log.With(logger, "module", "service.auth")),
		userBiz:    userBiz,
		captchaBiz: captchaBiz,
	}
}

func (s *AuthService) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginReply, error) {
	// 认证传递的code, 前端需要校验code的合法性
	if err := s.captchaBiz.VerifyCaptcha(ctx, req.GetCaptchaId(), req.GetCode()); err != nil {
		return nil, err
	}
	pwd := req.GetPassword()
	// 解密前端传递的密码, 拒绝明文传输
	dePwd, err := password.DecryptPassword(pwd, password.DefaultIv)
	if err != nil {
		return nil, perrors.ErrorInvalidParams("密码不规范")
	}
	// 颁发token, 时间建议设置为半天以内
	userBO, token, err := s.userBiz.LoginByUsernameAndPassword(ctx, req.GetUsername(), dePwd)
	if err != nil {
		s.log.Warnf("LoginByUsernameAndPassword error: %v", err)
		return nil, perrors.ErrorUnknown("颁发token失败")
	}
	prom.UPMemberCounter.WithLabelValues("prom-server").Inc()
	return &auth.LoginReply{
		Token: token,
		User:  userBO.ToApiV1(),
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, _ *auth.LogoutRequest) (*auth.LogoutReply, error) {
	authClaims, ok := middler.GetAuthClaims(ctx)
	if !ok {
		return nil, middler.ErrTokenInvalid
	}
	// 记录token md5然后存储到redis
	if err := s.userBiz.Logout(ctx, authClaims); err != nil {
		return nil, err
	}
	prom.UPMemberCounter.WithLabelValues("prom-server").Sub(1)
	return &auth.LogoutReply{
		UserId: authClaims.ID,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *auth.RefreshTokenRequest) (*auth.RefreshTokenReply, error) {
	authClaims, ok := middler.GetAuthClaims(ctx)
	if !ok {
		return nil, middler.ErrTokenInvalid
	}

	userBO, token, err := s.userBiz.RefreshToken(ctx, authClaims, req.GetRoleId())
	if err != nil {
		return nil, err
	}
	return &auth.RefreshTokenReply{
		Token: token,
		User:  userBO.ToApiV1(),
	}, nil
}

func (s *AuthService) Captcha(ctx context.Context, req *auth.CaptchaRequest) (*auth.CaptchaReply, error) {
	generateCaptcha, err := s.captchaBiz.GenerateCaptcha(ctx, captcha.Type(req.GetCaptchaType()), captcha.Theme(req.GetTheme()), int(req.GetX()), int(req.GetY()))
	if err != nil {
		return nil, err
	}
	return &auth.CaptchaReply{
		Captcha:   generateCaptcha.Image,
		CaptchaId: generateCaptcha.Id,
	}, nil
}
