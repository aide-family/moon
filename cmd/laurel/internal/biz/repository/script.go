package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/laurel/internal/biz/bo"
)

type Script interface {
	GetScripts(ctx context.Context) ([]*bo.TaskScript, error)
}
