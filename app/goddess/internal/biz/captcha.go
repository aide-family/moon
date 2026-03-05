package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/mojocn/base64Captcha"

	"github.com/aide-family/magicbox/captcha"
	"github.com/aide-family/magicbox/merr"
)

func NewCaptcha() *Captcha {
	return &Captcha{captcha: captcha.Global()}
}

type Captcha struct {
	captcha *base64Captcha.Captcha
}

func (c *Captcha) Generate(ctx context.Context) (string, string, error) {
	id, b64s, answer, err := c.captcha.Generate()
	if err != nil {
		log.Errorf("generate captcha failed: %v", err)
		return "", "", err
	}
	log.Context(ctx).Debugw("msg", "generate captcha success", "id", id, "answer", answer)
	return id, b64s, nil
}

func (c *Captcha) Verify(ctx context.Context, id, answer string) error {
	log.Context(ctx).Debugw("msg", "verify captcha", "id", id, "answer", answer)
	if !c.captcha.Verify(id, answer, true) {
		return merr.ErrorParams("verify captcha failed")
	}
	return nil
}

func (c *Captcha) EmailLoginCode(ctx context.Context) (string, string, error) {
	id, _, code, err := c.captcha.Generate()
	if err != nil {
		log.Errorf("generate email login code failed: %v", err)
		return "", "", err
	}
	log.Context(ctx).Debugw("msg", "generate email login code success", "id", id, "code", code)
	return id, code, nil
}
