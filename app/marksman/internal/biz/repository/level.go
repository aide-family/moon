package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/marksman/internal/biz/bo"
)

type Level interface {
	CreateLevel(ctx context.Context, req *bo.CreateLevelBo) error
	UpdateLevel(ctx context.Context, req *bo.UpdateLevelBo) error
	UpdateLevelStatus(ctx context.Context, req *bo.UpdateLevelStatusBo) error
	DeleteLevel(ctx context.Context, uid snowflake.ID) error
	GetLevel(ctx context.Context, uid snowflake.ID) (*bo.LevelItemBo, error)
	ListLevel(ctx context.Context, req *bo.ListLevelBo) (*bo.PageResponseBo[*bo.LevelItemBo], error)
	SelectLevel(ctx context.Context, req *bo.SelectLevelBo) (*bo.SelectLevelBoResult, error)
}
