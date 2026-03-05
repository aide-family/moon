package service

import (
	"github.com/aide-family/rabbit/internal/biz"
)

func NewCaptchaService(captchaBiz *biz.Captcha) (*CaptchaService, error) {
	return &CaptchaService{Captcha: captchaBiz}, nil
}

type CaptchaService struct {
	*biz.Captcha
}
