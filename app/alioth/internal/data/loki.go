package data

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

// lokiQueryRangePath 是 Loki 的日志区间查询接口。
// endpoint 按数据源根地址解析，和指标查询一样由服务拼接路径。
const lokiQueryRangePath = "/loki/api/v1/query_range"

// queryLokiLogs 向 Loki 查询一段时间的日志。
// 调用方传入 Unix 秒，Loki 接收纳秒时间戳；时间窗口和条数由用例层收敛。
func (r *datasourceRepo) queryLokiLogs(ctx context.Context, item bo.Datasource, query string, start, end int64, limit int, direction string) (string, error) {
	values := url.Values{}
	values.Set("query", query)
	values.Set("start", strconv.FormatInt(time.Unix(start, 0).UnixNano(), 10))
	values.Set("end", strconv.FormatInt(time.Unix(end, 0).UnixNano(), 10))
	values.Set("limit", strconv.Itoa(limit))
	values.Set("direction", direction)
	return r.callUpstream(ctx, item, http.MethodGet, lokiQueryRangePath+"?"+values.Encode(), "")
}
