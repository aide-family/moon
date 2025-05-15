package repository

import (
	"context"

	"github.com/moon-monitor/moon/cmd/houyi/internal/biz/bo"
)

type Judge interface {
	Metric(ctx context.Context, data *bo.MetricJudgeRequest) ([]bo.Alert, error)
}
