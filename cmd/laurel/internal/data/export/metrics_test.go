package export_test

import (
	"context"
	"testing"

	"github.com/aide-family/moon/cmd/laurel/internal/data/export"
)

func TestNodeExportMetricRepo_Metrics(t *testing.T) {
	repo := export.NewNodeExportMetricRepo()
	metrics, err := repo.Metrics(context.Background(), "http://localhost:8000/metrics")
	if err != nil {
		t.Fatalf("failed to get metrics: %v", err)
	}
	t.Logf("metrics: %v", metrics)
}
