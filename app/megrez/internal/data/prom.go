package data

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aide-family/Alioth/app/megrez/internal/data/repository"
	"github.com/aide-family/Alioth/app/megrez/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
	"github.com/aide-family/Alioth/pkg/upstream"
)

var datasourceQueryClient = &http.Client{Timeout: 10 * time.Second}

type promQuery struct{}

func NewRangeQuery() repository.RangeQuery {
	return rangeQuery{}
}

type rangeQuery struct{}

func (q rangeQuery) QueryRange(ctx context.Context, item bo.RuleSnapshot, start, end, step int64) ([]bo.Sample, error) {
	switch item.Driver {
	case "LOGS_LOKI":
		return lokiQuery{}.QueryRange(ctx, item, start, end, step)
	case "LOGS_VLOGS":
		return vlogsQuery{}.QueryRange(ctx, item, start, end, step)
	case "METRICS_PROMETHEUS", "METRICS_VICTORIA_METRICS":
		return promQuery{}.QueryRange(ctx, item, start, end, step)
	default:
		return nil, merr.Invalid("unsupported datasource driver")
	}
}

func (q rangeQuery) QueryInstant(ctx context.Context, item bo.RuleSnapshot, expr string, at time.Time) ([]bo.Sample, error) {
	switch item.Driver {
	case "LOGS_LOKI":
		return lokiQuery{}.QueryInstant(ctx, item, expr, at)
	case "LOGS_VLOGS":
		return vlogsQuery{}.QueryInstant(ctx, item, expr, at)
	case "METRICS_PROMETHEUS", "METRICS_VICTORIA_METRICS":
		return promQuery{}.QueryInstant(ctx, item, expr, at)
	default:
		return nil, merr.Invalid("unsupported datasource driver")
	}
}

func (p promQuery) QueryRange(ctx context.Context, item bo.RuleSnapshot, start, end, step int64) ([]bo.Sample, error) {
	target, err := upstream.Resolve(item.Endpoint, "/api/v1/query_range")
	if err != nil {
		return nil, err
	}
	query := target.Query()
	query.Set("query", item.Expr)
	query.Set("start", strconv.FormatInt(start, 10))
	query.Set("end", strconv.FormatInt(end, 10))
	query.Set("step", strconv.FormatInt(step, 10))
	target.RawQuery = query.Encode()
	payload, err := requestDatasource(ctx, item, target.String())
	if err != nil {
		return nil, err
	}
	return decodeMatrix(payload)
}

func (p promQuery) QueryInstant(ctx context.Context, item bo.RuleSnapshot, expr string, at time.Time) ([]bo.Sample, error) {
	target, err := upstream.Resolve(item.Endpoint, "/api/v1/query")
	if err != nil {
		return nil, err
	}
	query := target.Query()
	query.Set("query", expr)
	query.Set("time", strconv.FormatFloat(float64(at.UnixNano())/float64(time.Second), 'f', -1, 64))
	target.RawQuery = query.Encode()
	payload, err := requestDatasource(ctx, item, target.String())
	if err != nil {
		return nil, err
	}
	return decodeVector(payload)
}

func requestDatasource(ctx context.Context, item bo.RuleSnapshot, target string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, merr.Server("upstream query failed")
	}
	if item.Username != "" {
		req.SetBasicAuth(item.Username, item.Password)
	} else if item.Secret != "" {
		req.Header.Set("Authorization", "Bearer "+item.Secret)
	}
	resp, err := datasourceQueryClient.Do(req)
	if err != nil {
		slog.Default().Error(fmt.Sprintf("rule %s query: %v", item.UID, err))
		return nil, merr.Server("upstream query failed")
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		slog.Default().Error(fmt.Sprintf("rule %s read: %v", item.UID, err))
		return nil, merr.Server("upstream query failed")
	}
	if resp.StatusCode >= 300 {
		slog.Default().Error(fmt.Sprintf("rule %s status %d", item.UID, resp.StatusCode))
		return nil, merr.Server("upstream query failed")
	}
	return payload, nil
}

var _ repository.InstantQuery = promQuery{}
var _ repository.RangeQuery = promQuery{}

func decodeMatrix(payload []byte) ([]bo.Sample, error) {
	return decodeMatrixPoints(payload, false)
}

// decodeMatrixStrict 用于 VictoriaLogs 的 stats 结果：采样点必须全部能解析成
// 数值。stats 也可以输出字符串（例如 values(level)），静默跳过会得到空采样，
// 从而把仍在触发的事件错误恢复。
func decodeMatrixStrict(payload []byte) ([]bo.Sample, error) {
	return decodeMatrixPoints(payload, true)
}

func decodeMatrixPoints(payload []byte, strict bool) ([]bo.Sample, error) {
	var body struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Values [][]any           `json:"values"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &body); err != nil || body.Status != "success" || body.Data.ResultType != "matrix" {
		return nil, merr.Server("upstream query failed")
	}
	samples := make([]bo.Sample, 0, len(body.Data.Result))
	for _, series := range body.Data.Result {
		points := make([]bo.Point, 0, len(series.Values))
		for _, pair := range series.Values {
			point, ok := decodePoint(pair)
			if !ok {
				if strict {
					return nil, merr.Server("upstream query failed")
				}
				continue
			}
			if strict && (math.IsNaN(point.Value) || math.IsInf(point.Value, 0)) {
				return nil, merr.Server("upstream query failed")
			}
			points = append(points, point)
		}
		if strict && len(points) == 0 {
			return nil, merr.Server("upstream query failed")
		}
		samples = append(samples, bo.Sample{Labels: series.Metric, Points: points})
	}
	return samples, nil
}

func decodeVector(payload []byte) ([]bo.Sample, error) {
	var body struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Value  []any             `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &body); err != nil || body.Status != "success" || body.Data.ResultType != "vector" {
		return nil, merr.Server("upstream query failed")
	}
	samples := make([]bo.Sample, 0, len(body.Data.Result))
	for _, series := range body.Data.Result {
		point, ok := decodePoint(series.Value)
		if !ok {
			continue
		}
		samples = append(samples, bo.Sample{Labels: series.Metric, Points: []bo.Point{point}})
	}
	return samples, nil
}

func decodePoint(pair []any) (bo.Point, bool) {
	if len(pair) != 2 {
		return bo.Point{}, false
	}
	at, ok := decodeTimestamp(pair[0])
	if !ok {
		return bo.Point{}, false
	}
	value, raw, ok := decodeSampleValue(pair[1])
	if !ok {
		return bo.Point{}, false
	}
	return bo.Point{At: time.Unix(0, int64(at*float64(time.Second))).UTC(), Value: value, Raw: raw}, true
}

// decodeTimestamp 兼容不同上游/代理的时间戳写法：
// Prometheus 与 VictoriaLogs 返回秒级数字，但部分反向代理会把数字编码成
// 字符串，也有实现返回毫秒/微秒/纳秒。统一折算成 Unix 秒，避免时间戳类型
// 不匹配时整条序列被丢弃、告警永远不触发。
func decodeTimestamp(value any) (float64, bool) {
	var seconds float64
	switch typed := value.(type) {
	case float64:
		seconds = typed
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		if err != nil {
			return 0, false
		}
		seconds = parsed
	default:
		return 0, false
	}
	if math.IsNaN(seconds) || math.IsInf(seconds, 0) {
		return 0, false
	}
	// 阈值按数量级判断：1e11 秒已是公元 5138 年，正常秒级时间戳不会超过它。
	switch magnitude := math.Abs(seconds); {
	case magnitude >= 1e17:
		return seconds / 1e9, true
	case magnitude >= 1e14:
		return seconds / 1e6, true
	case magnitude >= 1e11:
		return seconds / 1e3, true
	default:
		return seconds, true
	}
}

// decodeSampleValue 同时接受字符串和 JSON 数字。VictoriaLogs 官方输出统一
// 是字符串，但代理或兼容实现可能返回数字；只认字符串会让 stats 结果变成
// 空采样，告警既不触发也不会报错。
func decodeSampleValue(value any) (float64, string, bool) {
	switch typed := value.(type) {
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		if err != nil {
			return 0, "", false
		}
		return parsed, typed, true
	case float64:
		return typed, strconv.FormatFloat(typed, 'f', -1, 64), true
	case int:
		return float64(typed), strconv.Itoa(typed), true
	case int64:
		return float64(typed), strconv.FormatInt(typed, 10), true
	default:
		return 0, "", false
	}
}
