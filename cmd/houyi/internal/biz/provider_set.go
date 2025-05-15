package biz

import (
	"github.com/google/wire"
	"github.com/moon-monitor/moon/cmd/houyi/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/houyi/internal/biz/do"
	"github.com/moon-monitor/moon/pkg/util/slices"
)

// ProviderSetBiz set biz dependency
var ProviderSetBiz = wire.NewSet(
	NewHealthBiz,
	NewRegisterBiz,
	NewConfig,
	NewMetric,
	NewEventBus,
	NewAlert,
)

type MetricDatasourceQueryReplyOption func(*bo.MetricDatasourceQueryReply)

func WithMetricDatasourceQueryReplyResults(results []*do.MetricQueryReply) MetricDatasourceQueryReplyOption {
	return func(reply *bo.MetricDatasourceQueryReply) {
		reply.Results = slices.Map(results, func(v *do.MetricQueryReply) *bo.MetricQueryResult {
			return &bo.MetricQueryResult{
				Metric: v.Labels,
				Value: &bo.MetricQueryValue{
					Value:     v.Value.Value,
					Timestamp: v.Value.Timestamp,
				},
			}
		})
	}
}

func WithMetricDatasourceQueryRangeReply(results []*do.MetricQueryRangeReply) MetricDatasourceQueryReplyOption {
	return func(reply *bo.MetricDatasourceQueryReply) {
		reply.Results = slices.Map(results, func(result *do.MetricQueryRangeReply) *bo.MetricQueryResult {
			return &bo.MetricQueryResult{
				Metric: result.Labels,
				Value:  nil,
				Values: slices.Map(result.Values, func(value *do.MetricQueryValue) *bo.MetricQueryValue {
					return &bo.MetricQueryValue{
						Value:     value.Value,
						Timestamp: value.Timestamp,
					}
				}),
			}
		})
	}
}

func NewMetricDatasourceQueryReply(opts ...MetricDatasourceQueryReplyOption) *bo.MetricDatasourceQueryReply {
	reply := &bo.MetricDatasourceQueryReply{}
	for _, opt := range opts {
		opt(reply)
	}
	return reply
}
