package authorization

import (
	"context"

	pb "github.com/aide-cloud/moon/api/admin/authorization"
	"github.com/aide-cloud/moon/cmd/moon/internal/biz"
	"github.com/aide-cloud/moon/cmd/moon/internal/biz/bo"
	"github.com/aide-cloud/moon/pkg/utils/captcha"
)

type Service struct {
	pb.UnimplementedAuthorizationServer

	captchaBiz *biz.CaptchaBiz
}

func NewAuthorizationService(captchaBiz *biz.CaptchaBiz) *Service {
	return &Service{
		captchaBiz: captchaBiz,
	}
}

func (s *Service) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginReply, error) {
	captchaInfo := req.GetCaptcha()
	// 校验验证码
	if err := s.captchaBiz.VerifyCaptcha(ctx, &bo.ValidateCaptchaParams{
		Id:    captchaInfo.GetId(),
		Value: captchaInfo.GetCode(),
	}); err != nil {
		return nil, err
	}
	return &pb.LoginReply{}, nil
}

func (s *Service) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutReply, error) {
	return &pb.LogoutReply{}, nil
}

func (s *Service) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenReply, error) {
	return &pb.RefreshTokenReply{}, nil
}

func (s *Service) Captcha(ctx context.Context, req *pb.CaptchaReq) (*pb.CaptchaReply, error) {
	generateCaptcha, err := s.captchaBiz.GenerateCaptcha(ctx, &bo.GenerateCaptchaParams{
		Type:  captcha.Type(req.GetCaptchaType()),
		Theme: captcha.Theme(req.GetTheme()),
		Size:  []int{int(req.GetHeight()), int(req.GetWidth())},
	})
	if err != nil {
		return nil, err
	}
	return &pb.CaptchaReply{
		Captcha:     generateCaptcha.Base64s,
		CaptchaType: req.GetCaptchaType(),
		Id:          generateCaptcha.Id,
	}, nil
}
