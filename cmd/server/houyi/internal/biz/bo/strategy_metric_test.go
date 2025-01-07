package bo

import (
	"context"
	"testing"
	"time"

	"github.com/aide-family/moon/pkg/label"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestStrategyMetric_Eval(t *testing.T) {
	expr := `100 - (avg(irate(node_cpu_seconds_total{mode="idle",}[5m])) by (instance) * 100)`
	strategy := &StrategyMetric{
		ReceiverGroupIDs: nil,
		LabelNotices:     nil,
		ID:               1,
		LevelID:          1,
		Alert:            "CPU usage is too high",
		Expr:             expr,
		For:              types.NewDuration(durationpb.New(time.Second * 600)),
		Count:            1,
		SustainType:      vobj.SustainMax,
		Labels:           label.NewLabels(map[string]string{"xx": "test"}),
		Annotations:      label.NewAnnotations(map[string]string{"xx": "test"}),
		Datasource: []*Datasource{{
			Category:    vobj.DatasourceTypeMetrics,
			StorageType: vobj.StorageTypeVictoriametrics,
			Config:      "{}",
			Endpoint:    "https://prometheus.aide-cloud.cn/",
			ID:          1,
		}},
		Status:    vobj.StatusEnable,
		Condition: vobj.ConditionGT,
		Threshold: 10,
		TeamID:    1,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	evalPoints, err := strategy.Eval(ctx)
	if err != nil {
		log.Warnw("method", "Eval", "error", err)
		return
	}
	for _, point := range evalPoints {
		extJSON, ok := strategy.IsCompletelyMeet(point.Values)
		if !ok {
			log.Warnw("method", "IsCompletelyMeet", "error", "not meet")
			continue
		}
		log.Infow("method", "IsCompletelyMeet", "extJSON", extJSON)
	}
}
