package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/houyi/internal/biz/bo"
)

type Alert interface {
	Save(ctx context.Context, alerts ...bo.Alert) error
	Get(ctx context.Context, fingerprint string) (bo.Alert, bool)
	GetAll(ctx context.Context) ([]bo.Alert, error)
	Delete(ctx context.Context, fingerprint string) error
}
