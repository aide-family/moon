package repository

import (
	"context"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
)

type Captcha interface {
	Generate(ctx context.Context) (*bo.Captcha, error)
	Verify(ctx context.Context, req *bo.CaptchaVerify) bool
}
