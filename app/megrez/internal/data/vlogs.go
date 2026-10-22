package data

import (
	"context"
	"strconv"
	"time"

	"github.com/aide-family/Alioth/app/megrez/internal/data/repository"
	"github.com/aide-family/Alioth/app/megrez/pkg/bo"
	"github.com/aide-family/Alioth/pkg/upstream"
)

// VictoriaLogs 的 stats 查询接口返回 Prometheus 兼容的 vector / matrix，
// 时间桶由 step 参数决定，规则表达式里不需要再写 `_time:<step>`。
const (
	vlogsStatsRangePath   = "/select/logsql/stats_query_range"
	vlogsStatsInstantPath = "/select/logsql/stats_query"
)

// vlogsDefaultStep 是默认的时间桶秒数，与评测调度周期一致。
const vlogsDefaultStep = 15

type vlogsQuery struct{}

// QueryRange 执行 VictoriaLogs 的 LogsQL stats 区间查询。
//
// 规则表达式必须是 `| stats ...` 查询，例如
// `level:error | stats by (app, level) count() as value`。
// `/select/logsql/stats_query_range` 按 step 切桶并返回 Prometheus matrix；
// 原始日志响应不是 matrix，非数值 stats 点也会被 decodeMatrixStrict 拒绝，
// 不会被当成空结果恢复事件。
func (vlogsQuery) QueryRange(ctx context.Context, item bo.RuleSnapshot, start, end, step int64) ([]bo.Sample, error) {
	if step <= 0 {
		step = vlogsDefaultStep
	}
	target, err := upstream.Resolve(item.Endpoint, vlogsStatsRangePath)
	if err != nil {
		return nil, err
	}
	query := target.Query()
	query.Set("query", item.Expr)
	query.Set("start", strconv.FormatInt(start, 10))
	query.Set("end", strconv.FormatInt(end, 10))
	query.Set("step", strconv.FormatInt(step, 10)+"s")
	target.RawQuery = query.Encode()
	payload, err := requestDatasource(ctx, item, target.String())
	if err != nil {
		return nil, err
	}
	samples, err := decodeMatrixStrict(payload)
	if err != nil {
		return nil, err
	}
	for index := range samples {
		samples[index].Step = step
	}
	return samples, nil
}

// QueryInstant 为告警模板的 query 函数提供 VictoriaLogs 瞬时 stats 查询。
// 取查询时刻之前的一个步长，`/select/logsql/stats_query` 返回 Prometheus vector。
func (vlogsQuery) QueryInstant(ctx context.Context, item bo.RuleSnapshot, expr string, at time.Time) ([]bo.Sample, error) {
	target, err := upstream.Resolve(item.Endpoint, vlogsStatsInstantPath)
	if err != nil {
		return nil, err
	}
	end := at.UTC().Unix()
	query := target.Query()
	query.Set("query", expr)
	query.Set("time", strconv.FormatInt(end, 10))
	query.Set("start", strconv.FormatInt(end-vlogsDefaultStep, 10))
	query.Set("end", strconv.FormatInt(end, 10))
	target.RawQuery = query.Encode()
	payload, err := requestDatasource(ctx, item, target.String())
	if err != nil {
		return nil, err
	}
	return decodeVector(payload)
}

var _ repository.RangeQuery = vlogsQuery{}
var _ repository.InstantQuery = vlogsQuery{}
