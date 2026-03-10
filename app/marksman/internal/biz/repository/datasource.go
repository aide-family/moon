package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/marksman/internal/biz/bo"
)

type Datasource interface {
	CheckDatasourceNameExist(ctx context.Context, name string, uid ...snowflake.ID) error
	CreateDatasource(ctx context.Context, req *bo.CreateDatasourceBo) (snowflake.ID, error)
	UpdateDatasource(ctx context.Context, req *bo.UpdateDatasourceBo) error
	DeleteDatasource(ctx context.Context, uid snowflake.ID) error
	GetDatasource(ctx context.Context, uid snowflake.ID) (*bo.DatasourceItemBo, error)
	ListDatasource(ctx context.Context, req *bo.ListDatasourceBo) (*bo.PageResponseBo[*bo.DatasourceItemBo], error)
	SelectDatasource(ctx context.Context, req *bo.SelectDatasourceBo) (*bo.SelectDatasourceReplyBo, error)
	// ListAllForProbe returns all enabled-status datasources for health probing (e.g. metrics).
	// No namespace filter; fetches in batches (batchSize) until all enabled datasources are returned.
	ListAllForProbe(ctx context.Context, batchSize int) ([]*bo.DatasourceItemBo, error)
}
