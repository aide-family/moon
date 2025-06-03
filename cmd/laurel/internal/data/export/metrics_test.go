package export_test

import (
	"context"
	"testing"

	"github.com/aide-family/moon/cmd/laurel/internal/data/export"
	"github.com/go-kratos/kratos/v2/log"
)

func TestNodeExportMetricRepo_Metrics(t *testing.T) {
	repo := export.NewNodeExportMetricRepo(log.GetLogger())
	metrics, err := repo.Metrics(context.Background(), "http://localhost:8000/metrics")
	if err != nil {
		t.Fatalf("failed to get metrics: %v", err)
	}
	t.Logf("metrics: %v", metrics)
}
