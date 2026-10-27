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
	"regexp"
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

// varNamePattern 与规则变量的名字校验保持一致：模板快捷变量只能是标识符，
// 不能带引号、短横线或中文。defs 会拼接进模板文本，因此这里再过滤一次。
var varNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]{0,31}$`)

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

// NewData 构造模板数据。extra 用于追加事件字段，不会覆盖保留字段。
func NewData(labels map[string]string, labelsJSON string, value float64, externalURL string, extra map[string]any) Data {
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
	return render(ctx, "", raw, data, name, at, query, externalURL)
}

// RenderWithVars 额外把规则声明的自定义变量注入模板：在固定的 Defs 之后追加
// {{$<name> := index .Vars "<name>"}}，摘要和描述里就能写 {{ $disk_free }}。
// 变量名来自规则配置，非法名字会被跳过，避免破坏模板文本；没有声明变量时
// 行为与 Render 完全一致。Vars/VarLabels 由调用方放进 Data，变量求值失败时
// 这里补空 map，让 {{ $disk_free }} 渲染成 <no value> 而不是执行报错。
func RenderWithVars(ctx context.Context, raw string, data Data, name string, at time.Time, query QueryFunc, externalURL *url.URL, vars []string) (string, error) {
	if len(vars) == 0 {
		return Render(ctx, raw, data, name, at, query, externalURL)
	}
	return render(ctx, varDefs(vars), raw, ensureVarData(data), name, at, query, externalURL)
}

// render 是 Render 和 RenderWithVars 的公共实现。defs 只包含额外注入的变量
// 定义，始终排在固定 Defs 之后。
func render(ctx context.Context, defs string, raw string, data Data, name string, at time.Time, query QueryFunc, externalURL *url.URL) (string, error) {
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
		ctx, Defs+defs+raw, name, data,
		model.Time(at.UnixMilli()), queryFunc, externalURL, nil,
	).Expand()
}

// Validate 只做语法校验，供保存前拦截写坏的模板。
func Validate(raw, name string) error {
	return promtemplate.NewTemplateExpander(
		context.Background(), Defs+raw, name, nil, 0, nil, &url.URL{}, nil,
	).ParseTest()
}

// ValidateWithVars 与 Validate 相同，只是先把规则声明的变量注入模板。
// 保存带 {{ $disk_free }} 的摘要时必须走这里，否则会被判为未定义变量。
func ValidateWithVars(raw, name string, vars []string) error {
	return promtemplate.NewTemplateExpander(
		context.Background(), Defs+varDefs(vars)+raw, name, nil, 0, nil, &url.URL{}, nil,
	).ParseTest()
}

// varDefs 为合法且不重复的变量名生成快捷变量定义。
func varDefs(vars []string) string {
	seen := make(map[string]struct{}, len(vars))
	var builder strings.Builder
	for _, raw := range vars {
		name := strings.TrimSpace(raw)
		if !varNamePattern.MatchString(name) {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		builder.WriteString("{{$")
		builder.WriteString(name)
		builder.WriteString(` := index .Vars "`)
		builder.WriteString(name)
		builder.WriteString(`"}}`)
	}
	return builder.String()
}

// ensureVarData 保证声明了变量时 Data 里有非 nil 的 Vars 和 VarLabels。
// eval 在变量查询失败时不会写对应键，模板里的 index .Vars 碰到 nil map 会
// 执行失败，因此统一在这里兜底。data 为 nil 时返回一份新的空数据。
func ensureVarData(data Data) Data {
	if data == nil {
		data = Data{}
	}
	if data["Vars"] == nil {
		data["Vars"] = map[string]any{}
	}
	if data["VarLabels"] == nil {
		data["VarLabels"] = map[string]map[string]string{}
	}
	return data
}
