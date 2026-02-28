package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/marksman/internal/biz/bo"
)

type Datasource interface {
	CreateDatasource(ctx context.Context, req *bo.CreateDatasourceBo) error
	UpdateDatasource(ctx context.Context, req *bo.UpdateDatasourceBo) error
	DeleteDatasource(ctx context.Context, uid snowflake.ID) error
	GetDatasource(ctx context.Context, uid snowflake.ID) (*bo.DatasourceItemBo, error)
	ListDatasource(ctx context.Context, req *bo.ListDatasourceBo) (*bo.PageResponseBo[*bo.DatasourceItemBo], error)
}
