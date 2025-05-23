package main

import (
	"fmt"
	"net/http"

	"github.com/aide-family/moon/pkg/api/common"
	"github.com/aide-family/moon/pkg/util/pointer"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func ParseMetricsFromEndpoint(url string) ([]*common.MetricItem, error) {
	// 1. 获取指标数据
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// 2. 解析 Prometheus 文本格式
	var parser expfmt.TextParser
	metricFamilies, err := parser.TextToMetricFamilies(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse metrics: %v", err)
	}

	// 3. 转换为 protobuf 结构
	var result []*common.MetricItem
	for name, mf := range metricFamilies {
		// 为每个 Metric 创建一个 MetricItem（而不是每个 MetricFamily）
		for _, metric := range mf.GetMetric() {
			metricItem := &common.MetricItem{
				Name: name,
				Help: mf.GetHelp(),
				Type: mf.GetType().String(),
				Unit: mf.GetUnit(),
			}

			switch pointer.Get(mf.Type) {
			case io_prometheus_client.MetricType_COUNTER:
				metricItem.Value = metric.GetCounter().GetValue()
			case io_prometheus_client.MetricType_GAUGE:
				metricItem.Value = metric.GetGauge().GetValue()
			case io_prometheus_client.MetricType_SUMMARY:
				metricItem.Value = metric.GetSummary().GetSampleSum()
			case io_prometheus_client.MetricType_HISTOGRAM:
				metricItem.Value = float64(metric.GetHistogram().GetSampleCount())
			case io_prometheus_client.MetricType_UNTYPED:
				metricItem.Value = metric.GetUntyped().GetValue()
			}

			// 处理标签
			for _, label := range metric.GetLabel() {
				labelItem := &common.MetricItem_Label{
					Key:    label.GetName(),
					Values: []string{label.GetValue()},
				}
				metricItem.Labels = append(metricItem.Labels, labelItem)
			}

			result = append(result, metricItem)
		}
	}

	return result, nil
}

func main() {
	metrics, err := ParseMetricsFromEndpoint("http://localhost:8000/metrics")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, metric := range metrics {
		fmt.Printf("Metric: %s\n", metric.Name)
		fmt.Printf("  Help: %s\n", metric.Help)
		fmt.Printf("  Type: %s\n", metric.Type)
		fmt.Printf("  Value: %f\n", metric.Value)
		if metric.Unit != "" {
			fmt.Printf("  Unit: %s\n", metric.Unit)
		}
		for _, label := range metric.Labels {
			fmt.Printf("  Label: %s = %v\n", label.Key, label.Values)
		}
		fmt.Println()
	}
}
