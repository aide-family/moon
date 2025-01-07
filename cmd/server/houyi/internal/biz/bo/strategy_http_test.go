package bo_test

import (
	"context"
	"testing"

	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/pkg/label"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

func TestStrategyHTTP_Eval(t *testing.T) {
	httpStrategy := &bo.StrategyHTTP{
		StrategyType:          vobj.StrategyTypeHTTP,
		URL:                   "https://baidu.com",
		StatusCode:            "5xx",
		StatusCodeCondition:   vobj.ConditionGT,
		Headers:               map[string]string{"Content-Type": "application/json"},
		Body:                  "",
		Method:                vobj.HTTPMethodGet,
		ResponseTime:          1,
		ResponseTimeCondition: vobj.ConditionGTE,
		Labels:                label.NewLabels(map[string]string{"http": "baidu.com"}),
		Annotations: label.NewAnnotations(map[string]string{
			"summary":     "baidu.com http 探测",
			"description": "baidu.com http 探测, 明细 {{ . }}",
		}),
		ReceiverGroupIDs: nil,
		LabelNotices:     nil,
		TeamID:           1,
		Status:           vobj.StatusEnable,
		Alert:            "baidu http 探测",
		LevelID:          1,
		ID:               1,
	}
	eval, err := httpStrategy.Eval(context.Background())
	if err != nil {
		t.Error(err)
		return
	}

	for indexer, point := range eval {
		bs, _ := types.Marshal(point.Values)
		t.Logf("indexer: %s, point: %+v, meet: ", indexer, string(bs))
		t.Log(httpStrategy.IsCompletelyMeet(point.Values))
	}
}
