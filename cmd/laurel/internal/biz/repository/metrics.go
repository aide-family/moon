package repository

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics interface {
	Metrics(ctx context.Context, target string) (map[string]prometheus.Collector, error)
}
