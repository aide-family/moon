package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/marksman/internal/biz/bo"
)

type StrategyGroup interface {
	CreateStrategyGroup(ctx context.Context, req *bo.CreateStrategyGroupBo) error
	UpdateStrategyGroup(ctx context.Context, req *bo.UpdateStrategyGroupBo) error
	UpdateStrategyGroupStatus(ctx context.Context, req *bo.UpdateStrategyGroupStatusBo) error
	DeleteStrategyGroup(ctx context.Context, uid snowflake.ID) error
	GetStrategyGroup(ctx context.Context, uid snowflake.ID) (*bo.StrategyGroupItemBo, error)
	ListStrategyGroup(ctx context.Context, req *bo.ListStrategyGroupBo) (*bo.PageResponseBo[*bo.StrategyGroupItemBo], error)
	SelectStrategyGroup(ctx context.Context, req *bo.SelectStrategyGroupBo) (*bo.SelectStrategyGroupBoResult, error)
	StrategyGroupBindReceivers(ctx context.Context, req *bo.StrategyGroupBindReceiversBo) error
}

type Strategy interface {
	CreateStrategy(ctx context.Context, req *bo.CreateStrategyBo) error
	UpdateStrategy(ctx context.Context, req *bo.UpdateStrategyBo) error
	UpdateStrategyStatus(ctx context.Context, req *bo.UpdateStrategyStatusBo) error
	DeleteStrategy(ctx context.Context, uid snowflake.ID) error
	GetStrategy(ctx context.Context, uid snowflake.ID) (*bo.StrategyItemBo, error)
	ListStrategy(ctx context.Context, req *bo.ListStrategyBo) (*bo.PageResponseBo[*bo.StrategyItemBo], error)
}
