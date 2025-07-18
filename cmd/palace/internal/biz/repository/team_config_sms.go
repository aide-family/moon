package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
)

type TeamSMSConfig interface {
	Get(ctx context.Context, id uint32) (do.TeamSMSConfig, error)
	List(ctx context.Context, req *bo.ListSMSConfigRequest) (*bo.ListSMSConfigListReply, error)
	Create(ctx context.Context, config bo.TeamSMSConfig) (uint32, error)
	Update(ctx context.Context, config bo.TeamSMSConfig) error
	FindByIds(ctx context.Context, ids []uint32) ([]do.TeamSMSConfig, error)
	CheckNameUnique(ctx context.Context, name string, configID uint32) error
}
