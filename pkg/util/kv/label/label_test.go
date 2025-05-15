package label_test

import (
	"testing"

	"github.com/moon-monitor/moon/pkg/util/cnst"
	"github.com/moon-monitor/moon/pkg/util/kv/label"
)

func TestNewLabel(t *testing.T) {
	//t.Errorf("NewLabel() = %v, want %v", got, tt.want)
	labels := label.NewLabel(map[string]string{
		cnst.LabelKeyLevelID:      "1",
		cnst.LabelKeyTeamID:       "1",
		cnst.LabelKeyDatasourceID: "1",
		cnst.LabelKeyStrategyID:   "1",
	})
	t.Log(labels)
	levelId := labels.GetLevelId()
	if levelId != 1 {
		t.Errorf("GetLevelId() = %v, want %v", levelId, 1)
	}
	teamId := labels.GetTeamId()
	if teamId != 1 {
		t.Errorf("GetTeamId() = %v, want %v", teamId, 1)
	}
	datasourceId := labels.GetDatasourceId()
	if datasourceId != 1 {
		t.Errorf("GetDatasourceId() = %v, want %v", datasourceId, 1)
	}
	strategyId := labels.GetStrategyId()
	if strategyId != 1 {
		t.Errorf("GetStrategyId() = %v, want %v", strategyId, 1)
	}
	labels.SetLevelId(2)
	if labels.GetLevelId() != 2 {
		t.Errorf("SetLevelId() = %v, want %v", labels.GetLevelId(), 2)
	}
	labels.SetTeamId(2)
	if labels.GetTeamId() != 2 {
		t.Errorf("SetTeamId() = %v, want %v", labels.GetTeamId(), 2)
	}
	labels.SetDatasourceId(2)
	if labels.GetDatasourceId() != 2 {
		t.Errorf("SetDatasourceId() = %v, want %v", labels.GetDatasourceId(), 2)
	}
	labels.SetStrategyId(2)
	if labels.GetStrategyId() != 2 {
		t.Errorf("SetStrategyId() = %v, want %v", labels.GetStrategyId(), 2)
	}
}
