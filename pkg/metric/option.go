package metric

import (
	"context"
	"strconv"
)

type MetricServerOption func(*metricServer)

func WithServerHandler(handler Handler) MetricServerOption {
	return func(server *metricServer) {
		server.handlers = append(server.handlers, handler)
	}
}

func defaultServerHandler(_ context.Context, req Request) {
	labels := []string{
		req.GetKind(),
		req.GetOperation(),
		strconv.Itoa(int(req.GetCode())),
		req.GetReason(),
		req.GetServer(),
	}
	RequestTotalMetric.WithLabelValues(labels...).Inc()
	RequestLatencyMetric.WithLabelValues(labels...).Observe(float64(req.GetLatency().Milliseconds()))
}
