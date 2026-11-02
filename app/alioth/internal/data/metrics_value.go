package data

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

// InstantValue 执行一次即时查询并取出第一个采样值。集群概况的四条汇总查询
// 只需要一个数值，不需要把上游的完整 matrix / vector 交给调用方。
//
// 查询成功但没有采样时返回 found=false 且 err=nil：调用方据此区分「本次没有数据」
// 和「查询失败」，前者在汇总里表现为缺失指标，后者表现为部分失败。
func (r *datasourceRepo) InstantValue(ctx context.Context, item bo.Datasource, expr string) (float64, bool, error) {
	if !isMetricsDriver(item.Driver) {
		return 0, false, merr.Invalid("unsupported datasource driver")
	}
	payload, err := r.InstantQuery(ctx, item, expr, 0)
	if err != nil {
		return 0, false, err
	}
	return parsePromInstantValue(payload)
}

// promInstantResponse 只解析即时查询响应里需要的字段。result 的形状随
// resultType 变化：vector 是采样数组，scalar / string 本身就是 [时间戳, 值]。
type promInstantResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string          `json:"resultType"`
		Result     json.RawMessage `json:"result"`
	} `json:"data"`
}

// parsePromInstantValue 从 Prometheus 兼容的即时查询响应里取第一个采样值。
// 支持 vector、scalar 和 string 三种 resultType。
func parsePromInstantValue(payload string) (float64, bool, error) {
	var response promInstantResponse
	if err := json.Unmarshal([]byte(payload), &response); err != nil {
		return 0, false, merr.Server("upstream query failed")
	}
	if response.Status != "success" {
		return 0, false, merr.Server("upstream query failed")
	}
	raw := response.Data.Result
	if len(raw) == 0 || string(raw) == "null" {
		return 0, false, nil
	}
	if response.Data.ResultType == "vector" {
		var samples []struct {
			Value []json.RawMessage `json:"value"`
		}
		if err := json.Unmarshal(raw, &samples); err != nil {
			return 0, false, merr.Server("upstream query failed")
		}
		if len(samples) == 0 {
			return 0, false, nil
		}
		return sampleValue(samples[0].Value)
	}
	var pair []json.RawMessage
	if err := json.Unmarshal(raw, &pair); err != nil {
		return 0, false, merr.Server("upstream query failed")
	}
	return sampleValue(pair)
}

// sampleValue 取 [时间戳, 值] 里的值。Prometheus 用字符串表示浮点数，
// 少数实现返回数字，两种都接受。
func sampleValue(pair []json.RawMessage) (float64, bool, error) {
	if len(pair) < 2 {
		return 0, false, merr.Server("upstream query failed")
	}
	var text string
	if err := json.Unmarshal(pair[1], &text); err == nil {
		value, parseErr := strconv.ParseFloat(text, 64)
		if parseErr != nil {
			return 0, false, merr.Server("upstream query failed")
		}
		return value, true, nil
	}
	var number float64
	if err := json.Unmarshal(pair[1], &number); err != nil {
		return 0, false, merr.Server("upstream query failed")
	}
	return number, true, nil
}
