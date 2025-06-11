package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
)

type TeamEmailConfig interface {
	Get(ctx context.Context, id uint32) (do.TeamEmailConfig, error)
	List(ctx context.Context, req *bo.ListEmailConfigRequest) (*bo.ListEmailConfigListReply, error)
	Create(ctx context.Context, config bo.TeamEmailConfig) (uint32, error)
	Update(ctx context.Context, config bo.TeamEmailConfig) error
	FindByIds(ctx context.Context, ids []uint32) ([]do.TeamEmailConfig, error)
	CheckNameUnique(ctx context.Context, name string, configID uint32) error
}
