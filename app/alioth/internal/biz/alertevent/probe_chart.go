package alertevent

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	labelsutil "github.com/aide-family/moon/pkg/labels"
	"github.com/aide-family/moon/pkg/merr"
)

// probeChartMetricLabel 是探测图表序列里的观测对象键。序列标签必须是事件标签的
// 子集，前端 filterSeriesByLabels 才能把序列收敛到当前事件。
const (
	probeChartTargetLabel     = "probe_target"
	probeChartDatasourceLabel = "datasource"
)

// probeMatrixBody 是探测图表返回的 Prometheus matrix 兼容结构，前端与指标图表
// 共用同一个解析与渲染路径。
type probeMatrixBody struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string              `json:"resultType"`
		Result     []probeMatrixSeries `json:"result"`
	} `json:"data"`
}

type probeMatrixSeries struct {
	Metric map[string]string `json:"metric"`
	Values [][2]any          `json:"values"`
}

// probeChart 为探测事件生成图表数据：从 probe_results 取该目标在事件区间内的
// 观测值，编码成 matrix。探测事件没有数据源，不能走数据源区间查询。
func (coordinator *Coordinator) probeChart(ctx context.Context, namespaceUID string, event bo.AlertEvent) (*bo.AlertEventChart, error) {
	labels, err := labelsutil.ParseObject(event.Labels)
	if err != nil {
		return nil, merr.Invalid("invalid alert event labels")
	}
	start, end, step := chartWindow(event.StartsAt, chartEnd(event, time.Now()))
	body := probeMatrixBody{Status: "success"}
	body.Data.ResultType = "matrix"
	body.Data.Result = []probeMatrixSeries{}

	chart := &bo.AlertEventChart{
		DatasourceName: labels[probeChartDatasourceLabel],
		Expr:           event.Expr,
		Start:          start,
		End:            end,
		Step:           step,
	}
	targetUID := labels[probeTargetUIDLabel]
	if targetUID == "" {
		encoded, _ := json.Marshal(body)
		chart.Body = string(encoded)
		return chart, nil
	}
	target, err := coordinator.probes.GetProbeTarget(ctx, namespaceUID, targetUID)
	if err != nil {
		// 目标被删除后历史事件仍可查看：返回空图表，不把详情页打成报错。
		encoded, _ := json.Marshal(body)
		chart.Body = string(encoded)
		return chart, nil
	}
	if group, err := coordinator.probes.GetProbeTargetGroup(ctx, namespaceUID, target.GroupUID); err == nil && group != nil {
		chart.DatasourceName = group.Name
	}
	records, err := coordinator.probes.ListProbeResults(ctx, namespaceUID, []string{target.UID}, time.Unix(start, 0), time.Unix(end, 0))
	if err != nil {
		return nil, coordinator.access.Wrap(err)
	}
	values := make([][2]any, 0, len(records))
	for _, record := range records {
		if record.Value == "" {
			continue
		}
		values = append(values, [2]any{record.Slot.UTC().Unix(), record.Value})
	}
	series := probeMatrixSeries{
		Metric: map[string]string{
			probeChartTargetLabel:     target.Spec,
			probeChartDatasourceLabel: chart.DatasourceName,
		},
		Values: values,
	}
	body.Data.Result = append(body.Data.Result, series)
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, merr.Server("request failed")
	}
	chart.Body = string(encoded)
	return chart, nil
}
