package biz

import (
	"github.com/aide-family/rabbit/internal/biz/repository"
)

func NewCaptcha(captchaRepo repository.Captcha) *Captcha {
	return &Captcha{
		Captcha: captchaRepo,
	}
}

type Captcha struct {
	repository.Captcha
}
