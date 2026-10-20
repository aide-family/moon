package data

import (
	"context"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

// QueryLogs 按日志驱动分发查询。
// Loki 使用 LogQL 区间查询，Elasticsearch 使用 _search；其他驱动返回明确错误。
func (r *datasourceRepo) QueryLogs(ctx context.Context, item bo.Datasource, query string, start, end int64, limit int, direction string) (string, error) {
	switch item.Driver {
	case "LOGS_LOKI":
		return r.queryLokiLogs(ctx, item, query, start, end, limit, direction)
	case "LOGS_ELASTICSEARCH":
		return r.queryElasticsearchLogs(ctx, item, query, start, end, limit, direction)
	default:
		return "", merr.Invalid("unsupported datasource driver")
	}
}
