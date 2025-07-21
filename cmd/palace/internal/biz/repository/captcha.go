package repository

import (
	"context"
	"github.com/aide-family/moon/pkg/util/captcha"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
)

type Captcha interface {
	Generate(ctx context.Context) (*captcha.GenResult, error)
	Verify(ctx context.Context, req *bo.CaptchaVerify) bool
}
