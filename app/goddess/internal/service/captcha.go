package service

import (
	"context"

	"github.com/aide-family/goddess/internal/biz"
	apiv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// NewCaptchaService creates a new CaptchaService.
func NewCaptchaService(captcha *biz.Captcha) *CaptchaService {
	return &CaptchaService{
		captcha: captcha,
	}
}

// CaptchaService implements v1.CaptchaServiceServer for generating graphical captcha.
type CaptchaService struct {
	apiv1.UnimplementedCaptchaServer
	captcha *biz.Captcha
}

// GetCaptcha generates a new captcha and returns id and base64 image.
func (s *CaptchaService) GetCaptcha(ctx context.Context, req *apiv1.GetCaptchaRequest) (*apiv1.GetCaptchaReply, error) {
	id, b64s, err := s.captcha.Generate(ctx)
	if err != nil {
		return nil, err
	}
	return &apiv1.GetCaptchaReply{
		CaptchaId:   id,
		CaptchaB64S: b64s,
	}, nil
}
