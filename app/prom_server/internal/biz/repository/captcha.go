package repository

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
)

var _ CaptchaRepo = (*UnimplementedCaptchaRepo)(nil)

type (
	CaptchaRepo interface {
		mustEmbedUnimplemented()
		// CreateCaptcha 创建验证码
		CreateCaptcha(ctx context.Context, captcha *dobo.CaptchaDO) error
		// GetCaptchaById 通过id获取验证码详情
		GetCaptchaById(ctx context.Context, id string) (*dobo.CaptchaDO, error)
	}

	UnimplementedCaptchaRepo struct{}
)

func (UnimplementedCaptchaRepo) mustEmbedUnimplemented() {}

func (UnimplementedCaptchaRepo) CreateCaptcha(_ context.Context, _ *dobo.CaptchaDO) error {
	return status.Errorf(codes.Unimplemented, "method CreateCaptcha not implemented")
}

func (UnimplementedCaptchaRepo) GetCaptchaById(_ context.Context, _ string) (*dobo.CaptchaDO, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCaptchaById not implemented")
}
