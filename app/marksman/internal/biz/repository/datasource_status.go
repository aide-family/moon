package repository

import (
	"context"

	"github.com/aide-family/marksman/internal/biz/bo"
)

// DatasourceStatusQuerier queries the main time-series DB for marksman_datasource_status.
type DatasourceStatusQuerier interface {
	QueryDatasourceStatus(ctx context.Context, req *bo.GetDatasourceStatusRequest) ([]*bo.DatasourceStatusSeriesBo, error)
}
