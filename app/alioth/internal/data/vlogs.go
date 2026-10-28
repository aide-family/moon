package data

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
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
// stats 查询的结果里没有 _time 字段，不能再追加 sort by (_time)，
// 否则查询会被 VictoriaLogs 拒绝；这类查询统一交给 limit 参数截断。
func (r *datasourceRepo) queryVlogsLogs(ctx context.Context, item bo.Datasource, query string, start, end int64, limit int, direction string) (string, error) {
	values := url.Values{}
	values.Set("query", query)
	values.Set("start", strconv.FormatInt(start, 10))
	values.Set("end", strconv.FormatInt(end, 10))
	if direction == "forward" && !hasVlogsStatsPipe(query) {
		values.Set("query", query+" | sort by (_time) | limit "+strconv.Itoa(limit))
	} else {
		values.Set("limit", strconv.Itoa(limit))
	}
	return r.callUpstream(ctx, item, http.MethodGet, vlogsQueryPath+"?"+values.Encode(), "")
}

// vlogsStatsFuncs 是 stats 管道的聚合函数。stats 管道可以省略 stats 关键字，
// 直接写 `| count() as value`，这里一并识别。
var vlogsStatsFuncs = []string{
	"count", "count_uniq", "count_empty", "sum", "sum_len", "avg",
	"min", "max", "median", "quantile", "stddev", "rate", "rate_sum",
	"histogram", "values", "uniq_values", "any", "row_any", "row_min",
	"row_max", "json_values", "json_values_sorted", "json_values_topk",
}

func hasVlogsStatsPipe(query string) bool {
	lower := strings.ToLower(query)
	for offset := 0; offset < len(lower); {
		index := strings.IndexByte(lower[offset:], '|')
		if index < 0 {
			return false
		}
		index += offset
		rest := strings.TrimLeft(lower[index+1:], " \t\r\n")
		if strings.HasPrefix(rest, "stats") &&
			(len(rest) == len("stats") || !isVlogsWordByte(rest[len("stats")])) {
			return true
		}
		for _, fn := range vlogsStatsFuncs {
			if !strings.HasPrefix(rest, fn) {
				continue
			}
			after := strings.TrimLeft(rest[len(fn):], " \t\r\n")
			if strings.HasPrefix(after, "(") {
				return true
			}
		}
		offset = index + 1
	}
	return false
}

func isVlogsWordByte(char byte) bool {
	return char == '_' || char == '-' || char == '.' ||
		(char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')
}
