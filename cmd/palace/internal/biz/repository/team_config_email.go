package repository

import (
	"context"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
)

type TeamEmailConfig interface {
	Get(ctx context.Context, id uint32) (do.TeamEmailConfig, error)
	List(ctx context.Context, req *bo.ListEmailConfigRequest) (*bo.ListEmailConfigListReply, error)
	Create(ctx context.Context, config bo.TeamEmailConfig) error
	Update(ctx context.Context, config bo.TeamEmailConfig) error
}
