// Package alerttmpl 是告警摘要、描述和通知模板共用的模板数据与展开逻辑。
//
// 告警规则模板在 megrez 评测时展开，通知模板在 merak 投递时展开。两边使用
// 同一套字段，因此 Prometheus 的 $labels、$value、.Labels、.Value、{{ . }}
// 写法在告警内容和通知正文里含义一致。
package alerttmpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/common/model"
	promlabels "github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql"
	promtemplate "github.com/prometheus/prometheus/template"
)

// Defs 注入 Prometheus 告警模板约定的小写快捷变量。
const Defs = "{{$labels := .Labels}}{{$externalLabels := .ExternalLabels}}{{$externalURL := .ExternalURL}}{{$value := .Value}}"

// Data 是模板数据。大写字段是 Prometheus 原生写法，小写字段是 Alioth 消息
// 字段：labels 是事件里保存的标签 JSON 字符串，Labels 是解析后的标签 map。
type Data map[string]any

// String 让 {{ . }} 得到采样值，与 {{ $value }} 保持一致。
// Prometheus 的 AlertTemplateData 没有这个方法，{{ . }} 会打印整个结构体；
// Alioth 的摘要和描述模板约定 . 就是采样值，这里不跟随 Prometheus。
func (d Data) String() string {
	if value, ok := d["Value"]; ok && value != nil {
		return fmt.Sprint(value)
	}
	return ""
}

// New 构造模板数据。extra 用于追加事件字段，不会覆盖保留字段。
func New(labels map[string]string, labelsJSON string, value float64, externalURL string, extra map[string]any) Data {
	if labels == nil {
		labels = map[string]string{}
	}
	externalLabels := map[string]string{}
	data := Data{
		"Labels": labels, "ExternalLabels": externalLabels, "ExternalURL": externalURL, "Value": value,
		"labels": labelsJSON, "externalLabels": externalLabels, "externalURL": externalURL, "value": value,
	}
	for key, item := range extra {
		if _, exists := data[key]; !exists {
			data[key] = item
		}
	}
	return data
}

// FromParams 把任务参数补成完整数据。历史任务只有小写字段，这里补齐结构化的
// 标签、采样值和外部地址，避免模板在缺字段时报错。
func FromParams(params map[string]any) Data {
	data := Data{}
	for key, item := range params {
		data[key] = item
	}
	labelsJSON, _ := data["labels"].(string)
	labels := map[string]any{}
	if labelsJSON != "" {
		_ = json.Unmarshal([]byte(labelsJSON), &labels)
	}
	if existing, ok := data["Labels"]; !ok || existing == nil {
		data["Labels"] = labels
	}
	externalLabels := map[string]string{}
	if existing, ok := data["ExternalLabels"]; !ok || existing == nil {
		data["ExternalLabels"] = externalLabels
	}
	if existing, ok := data["externalLabels"]; !ok || existing == nil {
		data["externalLabels"] = externalLabels
	}
	externalURL, _ := data["externalURL"].(string)
	if externalURL == "" {
		externalURL, _ = data["ExternalURL"].(string)
	}
	if existing, ok := data["externalURL"]; !ok || existing == nil {
		data["externalURL"] = externalURL
	}
	if existing, ok := data["ExternalURL"]; !ok || existing == nil {
		data["ExternalURL"] = externalURL
	}
	if existing, ok := data["Value"]; !ok || existing == nil {
		data["Value"] = toFloat(data["value"])
	}
	if existing, ok := data["value"]; !ok || existing == nil {
		data["value"] = data["Value"]
	}
	return data
}

func toFloat(value any) float64 {
	switch item := value.(type) {
	case float64:
		return item
	case float32:
		return float64(item)
	case int:
		return float64(item)
	case int64:
		return float64(item)
	case json.Number:
		n, _ := item.Float64()
		return n
	case string:
		n, _ := strconv.ParseFloat(strings.TrimSpace(item), 64)
		return n
	default:
		return 0
	}
}

// Sample 是一次模板查询返回的一条序列。
type Sample struct {
	Labels map[string]string
	Value  float64
}

// QueryFunc 执行 query 模板函数需要的即时查询。
type QueryFunc func(ctx context.Context, expr string, at time.Time) ([]Sample, error)

// Render 按 Prometheus 告警模板的语义展开 raw。
func Render(ctx context.Context, raw string, data Data, name string, at time.Time, query QueryFunc, externalURL *url.URL) (string, error) {
	var queryFunc promtemplate.QueryFunc
	if query != nil {
		queryFunc = func(ctx context.Context, expr string, queryAt time.Time) (promql.Vector, error) {
			samples, err := query(ctx, expr, queryAt)
			if err != nil {
				return nil, err
			}
			vector := make(promql.Vector, 0, len(samples))
			for _, item := range samples {
				vector = append(vector, promql.Sample{Metric: promlabels.FromMap(item.Labels), F: item.Value})
			}
			return vector, nil
		}
	}
	if externalURL == nil {
		externalURL = &url.URL{}
	}
	// 外部地址用于 $externalURL 和 graphLink 一类函数，不能带出数据源凭据。
	externalURL.User = nil
	return promtemplate.NewTemplateExpander(
		ctx, Defs+raw, name, data,
		model.Time(at.UnixMilli()), queryFunc, externalURL, nil,
	).Expand()
}

// Validate 只做语法校验，供保存前拦截写坏的模板。
func Validate(raw, name string) error {
	return promtemplate.NewTemplateExpander(
		context.Background(), Defs+raw, name, nil, 0, nil, &url.URL{}, nil,
	).ParseTest()
}
