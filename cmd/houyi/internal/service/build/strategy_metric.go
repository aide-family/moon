package build

import (
	"strconv"

	"github.com/aide-family/moon/cmd/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/houyi/internal/biz/do"
	"github.com/aide-family/moon/cmd/houyi/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/api/houyi/common"
	"github.com/aide-family/moon/pkg/util/cnst"
	"github.com/aide-family/moon/pkg/util/kv/label"
	"github.com/aide-family/moon/pkg/util/slices"
)

func ToMetricRules(strategyItems ...*common.MetricStrategyItem) []bo.MetricRule {
	if len(strategyItems) == 0 {
		return nil
	}
	rules := make([]bo.MetricRule, 0, len(strategyItems)*5*3)
	for _, strategyItem := range strategyItems {
		if strategyItem == nil {
			continue
		}
		for _, rule := range strategyItem.Rules {
			if rule == nil {
				continue
			}
			datasourceConfigs := strategyItem.GetDatasource()
			for _, datasourceItem := range datasourceConfigs {
				if datasourceItem == nil {
					continue
				}
				annotations := strategyItem.GetAnnotations()
				item := &do.MetricRule{
					TeamId:       strategyItem.GetTeam().GetTeamId(),
					DatasourceId: datasourceItem.GetId(),
					Datasource:   vobj.MetricDatasourceUniqueKey(datasourceItem.GetDriver(), strategyItem.GetTeam().GetTeamId(), datasourceItem.GetId()),
					StrategyId:   strategyItem.GetStrategyId(),
					LevelId:      rule.GetLevelId(),
					Receiver:     rule.GetReceiverRoutes(),
					LabelReceiver: slices.Map(rule.GetLabelNotices(), func(item *common.MetricStrategyItem_LabelNotices) *do.LabelNotices {
						return ToLabelNotice(item)
					}),
					Expr: strategyItem.GetExpr(),
					Labels: label.NewLabel(strategyItem.GetLabels()).Appends(map[string]string{
						cnst.LabelKeyTeamID:       strconv.FormatUint(uint64(strategyItem.GetTeam().GetTeamId()), 10),
						cnst.LabelKeyStrategyID:   strconv.FormatUint(uint64(strategyItem.GetStrategyId()), 10),
						cnst.LabelKeyLevelID:      strconv.FormatUint(uint64(rule.GetLevelId()), 10),
						cnst.LabelKeyDatasourceID: strconv.FormatUint(uint64(datasourceItem.GetId()), 10),
					}),
					Annotations: label.NewAnnotation(annotations[cnst.AnnotationKeySummary], annotations[cnst.AnnotationKeyDescription]),
					Duration:    rule.GetDuration().AsDuration(),
					Count:       rule.GetCount(),
					Values:      rule.GetValues(),
					SampleMode:  rule.GetSampleMode(),
					Condition:   rule.GetCondition(),
					Enable:      rule.GetEnable(),
				}
				rules = append(rules, item)
			}
		}
	}
	return rules
}
