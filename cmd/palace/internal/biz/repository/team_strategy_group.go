package repository

import (
	"context"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
)

type TeamStrategyGroup interface {
	Create(ctx context.Context, params *bo.SaveTeamStrategyGroupParams) error
	Update(ctx context.Context, params *bo.SaveTeamStrategyGroupParams) error
	Delete(ctx context.Context, id uint32) error
	Get(ctx context.Context, id uint32) (do.StrategyGroup, error)
	List(ctx context.Context, listParams *bo.ListTeamStrategyGroupParams) (*bo.ListTeamStrategyGroupReply, error)
	UpdateStatus(ctx context.Context, params *bo.UpdateTeamStrategyGroupStatusParams) error
}
