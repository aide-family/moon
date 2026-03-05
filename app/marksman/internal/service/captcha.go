package service

import (
	"github.com/aide-family/marksman/internal/biz"
)

func NewCaptchaService(captchaBiz *biz.Captcha) *CaptchaService {
	return &CaptchaService{
		Captcha: captchaBiz,
	}
}

type CaptchaService struct {
	*biz.Captcha
}
