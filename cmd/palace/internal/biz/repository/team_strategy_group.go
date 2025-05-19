package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
)

type TeamStrategyGroup interface {
	Create(ctx context.Context, params *bo.SaveTeamStrategyGroupParams) error
	Update(ctx context.Context, params *bo.SaveTeamStrategyGroupParams) error
	Delete(ctx context.Context, id uint32) error
	Get(ctx context.Context, id uint32) (do.StrategyGroup, error)
	List(ctx context.Context, listParams *bo.ListTeamStrategyGroupParams) (*bo.ListTeamStrategyGroupReply, error)
	Select(ctx context.Context, selectParams *bo.SelectTeamStrategyGroupRequest) (*bo.SelectTeamStrategyGroupReply, error)
	UpdateStatus(ctx context.Context, params *bo.UpdateTeamStrategyGroupStatusParams) error
}
