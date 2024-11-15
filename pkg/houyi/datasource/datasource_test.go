package datasource

import (
	"context"
	"testing"
	"time"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestMetricEval(t *testing.T) {
	dConfig := &api.Datasource{
		Category:    api.DatasourceType(vobj.DatasourceTypeMetrics),
		StorageType: api.StorageType(vobj.StorageTypeVictoriametrics),
		Config:      nil,
		Endpoint:    "https://prometheus.aide-cloud.cn/",
		Id:          1,
	}
	d, err := NewDatasource(dConfig).Metric()
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ts := time.Now()
	defer func() {
		t.Log(time.Since(ts))
	}()
	MetricEval(d, d, d, d, d, d, d, d, d, d, d, d, d, d)(ctx, "up", types.NewDuration(durationpb.New(10*time.Minute)))
}
