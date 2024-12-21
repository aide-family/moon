package bizmodel_test

import (
	"testing"

	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

func TestStrategyLevel_UnmarshalBinary(t *testing.T) {
	level := &bizmodel.StrategyMetricsLevel{}
	details := `{"id":1,"created_at":"2024-10-26 01:32:38","updated_at":"2024-10-26 01:32:38","deleted_at":0,"creator_id":1,"strategy_id":1,"strategy":{"id":1,"created_at":"2024-10-26 01:32:38","updated_at":"2024-10-26 01:32:38","creator_id":1,"strategy_type":0,"template_id":0,"group_id":1,"deleted_at":0,"template_source":2,"name":"prom-up监控","expr":"up","labels":{"server":"up"},"annotations":{"description":"服务{{ .labels.job }}实例为{{ .labels.instance }}, 当前状态为健康","summary":"服务{{ .labels.job }} 已上线"},"remark":"","status":1,"step":15,"datasource":null,"categories":null,"levels":null,"domain_levels":null,"http_levels":null,"ping_levels":null,"alarm_groups":null,"group":null},"duration":10,"count":1,"sustain_type":1,"interval":0,"condition":1,"threshold":1,"level_id":5,"level":{"id":5,"created_at":"2024-10-25 15:24:50","updated_at":"2024-10-25 15:24:50","deleted_at":0,"creator_id":0,"Name":"五级告警","Value":"5","DictType":3,"ColorType":"warning","CSSClass":"#165DFF","Icon":"","ImageURL":"","Status":1,"LanguageCode":1,"Remark":"","strategy_levels":null},"status":1,"alarm_page":[{"id":7,"created_at":"2024-10-25 15:24:50","updated_at":"2024-10-25 15:24:50","deleted_at":0,"creator_id":0,"Name":"测试告警","Value":"test-alarm-page","DictType":4,"ColorType":"warning","CSSClass":"#165DFF","Icon":"","ImageURL":"","Status":1,"LanguageCode":1,"Remark":"","strategy_levels":null}],"alarm_groups":[],"label_notices":[]}`
	err := level.UnmarshalBinary([]byte(details))
	if err != nil {
		t.Fatal(err)
	}

	t.Log(level)
	t.Log(level.Level)
}
