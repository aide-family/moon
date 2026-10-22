package data

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

// vlogsQueryPath 是 VictoriaLogs 的 LogsQL 查询接口。
// endpoint 按数据源根地址解析，支持集群版或反向代理的路径前缀。
const vlogsQueryPath = "/select/logsql/query"

// vlogsStatsRangePath 是 VictoriaLogs 的 stats 区间查询接口，
// 返回 Prometheus matrix，供告警规则评测和详情图表共用。
const vlogsStatsRangePath = "/select/logsql/stats_query_range"

// queryVlogsLogs 向 VictoriaLogs 查询一段时间的日志。
//
// VictoriaLogs 的 limit 参数返回最近 N 条日志，正好对应“从新到旧”；
// 需要“从旧到新”时在 LogsQL 末尾追加 sort 和 limit 管道，先按时间升序再截断。
func (r *datasourceRepo) queryVlogsLogs(ctx context.Context, item bo.Datasource, query string, start, end int64, limit int, direction string) (string, error) {
	values := url.Values{}
	values.Set("query", query)
	values.Set("start", strconv.FormatInt(start, 10))
	values.Set("end", strconv.FormatInt(end, 10))
	if direction == "forward" {
		values.Set("query", query+" | sort by (_time) | limit "+strconv.Itoa(limit))
	} else {
		values.Set("limit", strconv.Itoa(limit))
	}
	return r.callUpstream(ctx, item, http.MethodGet, vlogsQueryPath+"?"+values.Encode(), "")
}
