package data

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/aide-family/Alioth/app/megrez/internal/data/repository"
	"github.com/aide-family/Alioth/app/megrez/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

var promClient = &http.Client{Timeout: 10 * time.Second}

type promQuery struct{}

func NewRangeQuery() repository.RangeQuery {
	return promQuery{}
}

func (promQuery) QueryRange(ctx context.Context, item bo.RuleSnapshot, start, end, step int64) ([]bo.Sample, error) {
	base, err := url.Parse(item.Endpoint)
	if err != nil || base.Host == "" || (base.Scheme != "http" && base.Scheme != "https") {
		return nil, merr.Invalid("invalid request")
	}
	target := base.ResolveReference(&url.URL{Path: "/api/v1/query_range"})
	if target.Host != base.Host {
		return nil, merr.Invalid("invalid request")
	}
	query := target.Query()
	query.Set("query", item.Expr)
	query.Set("start", strconv.FormatInt(start, 10))
	query.Set("end", strconv.FormatInt(end, 10))
	query.Set("step", strconv.FormatInt(step, 10))
	target.RawQuery = query.Encode()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, merr.Server("upstream query failed")
	}
	if item.Username != "" {
		req.SetBasicAuth(item.Username, item.Password)
	} else if item.Secret != "" {
		req.Header.Set("Authorization", "Bearer "+item.Secret)
	}
	resp, err := promClient.Do(req)
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
	return decodeMatrix(payload)
}

func decodeMatrix(payload []byte) ([]bo.Sample, error) {
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
			if len(pair) != 2 {
				continue
			}
			at, ok := pair[0].(float64)
			raw, okRaw := pair[1].(string)
			if !ok || !okRaw {
				continue
			}
			value, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				continue
			}
			points = append(points, bo.Point{At: time.Unix(int64(at), 0).UTC(), Value: value, Raw: raw})
		}
		samples = append(samples, bo.Sample{Labels: series.Metric, Points: points})
	}
	return samples, nil
}
