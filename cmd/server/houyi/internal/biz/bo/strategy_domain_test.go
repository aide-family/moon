package bo_test

import (
	"context"
	"testing"

	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

func TestStrategyDomain_Eval(t *testing.T) {
	domainStrategy := &bo.StrategyDomain{
		ReceiverGroupIDs: nil,
		LabelNotices:     nil,
		ID:               1,
		LevelID:          1,
		TeamID:           1,
		Status:           vobj.StatusEnable,
		Alert:            "百度域名证书监控",
		Threshold:        1,
		Labels:           vobj.NewLabels(map[string]string{"domain": "baidu.com"}),
		Annotations: vobj.NewAnnotations(map[string]string{
			"summary":     "百度域名证书监控",
			"description": "百度域名证书监控 明细 {{ . }}",
		}),
		Domain:       "baidu.com",
		Port:         443,
		StrategyType: vobj.StrategyTypeDomainCertificate,
		Condition:    vobj.ConditionGTE,
	}

	eval, err := domainStrategy.Eval(context.Background())
	if err != nil {
		t.Error(err)
		return
	}

	for indexer, point := range eval {
		bs, _ := types.Marshal(point.Values)
		t.Logf("indexer: %s, point: %+v, meet: ", indexer, string(bs))
		t.Log(domainStrategy.IsCompletelyMeet(point.Values))
	}
}

func TestStrategyDomainPort_Eval(t *testing.T) {
	domainStrategy := &bo.StrategyDomain{
		ReceiverGroupIDs: nil,
		LabelNotices:     nil,
		ID:               1,
		LevelID:          1,
		TeamID:           1,
		Status:           vobj.StatusEnable,
		Alert:            "百度域名端口监控",
		Threshold:        1,
		Labels:           vobj.NewLabels(map[string]string{"domain": "baidu.com"}),
		Annotations: vobj.NewAnnotations(map[string]string{
			"summary":     "百度域名端口监控",
			"description": "百度域名端口监控 明细 {{ . }}",
		}),
		Domain:       "baidu.com",
		Port:         1443,
		StrategyType: vobj.StrategyTypeDomainPort,
		Condition:    vobj.ConditionGTE,
	}

	eval, err := domainStrategy.Eval(context.Background())
	if err != nil {
		t.Error(err)
		return
	}

	for indexer, point := range eval {
		bs, _ := types.Marshal(point.Values)
		t.Logf("indexer: %s, point: %+v, meet: ", indexer, string(bs))
		t.Log(domainStrategy.IsCompletelyMeet(point.Values))
	}
}
