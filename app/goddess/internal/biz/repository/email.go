package repository

import (
	"context"

	"github.com/aide-family/goddess/internal/biz/bo"
)

type Email interface {
	SendEmail(ctx context.Context, req *bo.SendEmailBo) error
}
