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
	at, ok := pair[0].(float64)
	raw, okRaw := pair[1].(string)
	if !ok || !okRaw {
		return bo.Point{}, false
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return bo.Point{}, false
	}
	return bo.Point{At: time.Unix(0, int64(at*float64(time.Second))).UTC(), Value: value, Raw: raw}, true
}
