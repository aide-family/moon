package bo

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/pkg/api/houyi/common"
	"github.com/aide-family/moon/pkg/util/kv"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

// ChangedRows is a map of changed rows.
//
//	key: team id
//	value: changed row ids
type ChangedRows map[uint32][]uint32

type ChangedMetricDatasource ChangedRows

type ChangedMetricStrategy ChangedRows

func ToSyncMetricDatasourceItem(item do.DatasourceMetric, teamDo do.Team) *common.MetricDatasourceItem {
	if validate.IsNil(item) || validate.IsNil(teamDo) {
		return nil
	}
	return &common.MetricDatasourceItem{
		Team:           ToSyncTeamItem(teamDo),
		Driver:         common.MetricDatasourceDriver(item.GetDriver().GetValue()),
		Config:         ToSyncMetricDatasourceConfigItem(item),
		Enable:         item.GetStatus().IsEnable() && item.GetDeletedAt() == 0,
		Id:             item.GetID(),
		Name:           item.GetName(),
		ScrapeInterval: durationpb.New(item.GetScrapeInterval()),
	}
}

func ToSyncMetricDatasourceConfigItem(item do.DatasourceMetric) *common.MetricDatasourceItem_Config {
	if validate.IsNil(item) {
		return nil
	}
	return &common.MetricDatasourceItem_Config{
		Endpoint:  item.GetEndpoint(),
		BasicAuth: ToSyncMetricDatasourceBasicAuthItem(item.GetBasicAuth()),
		Headers:   ToSyncMetricDatasourceHeadersItem(item.GetHeaders()),
		Ca:        item.GetCA(),
		Tls:       ToSyncMetricDatasourceTlsItem(item.GetTLS()),
		Method:    common.DatasourceQueryMethod(item.GetQueryMethod().GetValue()),
	}
}

func ToSyncMetricDatasourceBasicAuthItem(basicAuth *do.BasicAuth) *common.BasicAuth {
	if validate.IsNil(basicAuth) {
		return nil
	}
	return &common.BasicAuth{
		Username: basicAuth.GetUsername(),
		Password: basicAuth.GetPassword(),
	}
}

func ToSyncMetricDatasourceHeadersItem(headers []*kv.KV) []*common.KeyValueItem {
	if validate.IsNil(headers) {
		return nil
	}
	return slices.MapFilter(headers, func(header *kv.KV) (*common.KeyValueItem, bool) {
		if validate.IsNil(header) {
			return nil, false
		}
		return &common.KeyValueItem{
			Key:   header.Key,
			Value: header.Value,
		}, true
	})
}

func ToSyncMetricDatasourceTlsItem(tls *do.TLS) *common.TLS {
	if validate.IsNil(tls) {
		return nil
	}
	return &common.TLS{
		ServerName: tls.GetServerName(),
		ClientCert: tls.GetClientCert(),
		ClientKey:  tls.GetClientKey(),
		SkipVerify: tls.GetSkipVerify(),
	}
}

func ToSyncTeamItem(teamDo do.Team) *common.TeamItem {
	if validate.IsNil(teamDo) {
		return nil
	}
	return &common.TeamItem{
		TeamId: teamDo.GetID(),
		Uuid:   teamDo.GetUUID().String(),
	}
}

func ToSyncMetricStrategyItem(item do.StrategyMetric, teamDo do.Team) *common.MetricStrategyItem {
	if validate.IsNil(item) || validate.IsNil(teamDo) {
		return nil
	}
	return &common.MetricStrategyItem{
		Team:           ToSyncTeamItem(teamDo),
		Datasource:     ToSyncMetricSimpleDatasourceItems(item.GetDatasourceList(), teamDo),
		Name:           item.GetStrategy().GetName(),
		Expr:           item.GetExpr(),
		ReceiverRoutes: ToSyncReceiverRoutesItems(item.GetStrategy().GetNotices()),
		Labels:         item.GetLabels().ToMap(),
		Annotations:    item.GetAnnotations().ToMap(),
		StrategyId:     item.GetStrategyID(),
		Rules:          ToSyncMetricRuleItems(item.GetRules()),
	}
}

func ToSyncMetricRuleItems(rules []do.StrategyMetricRule) []*common.MetricStrategyItem_MetricRuleItem {
	if validate.IsNil(rules) {
		return nil
	}
	return slices.MapFilter(rules, func(rule do.StrategyMetricRule) (*common.MetricStrategyItem_MetricRuleItem, bool) {
		if validate.IsNil(rule) {
			return nil, false
		}
		return ToSyncMetricRuleItem(rule), true
	})
}

func ToSyncMetricRuleItem(rule do.StrategyMetricRule) *common.MetricStrategyItem_MetricRuleItem {
	if validate.IsNil(rule) {
		return nil
	}
	return &common.MetricStrategyItem_MetricRuleItem{
		StrategyId:     rule.GetStrategyID(),
		LevelId:        rule.GetID(),
		SampleMode:     common.SampleMode(rule.GetSampleMode().GetValue()),
		Count:          rule.GetTotal(),
		Condition:      common.MetricStrategyItem_Condition(rule.GetCondition().GetValue()),
		Values:         rule.GetValues(),
		ReceiverRoutes: ToSyncReceiverRoutesItems(rule.GetNotices()),
		LabelNotices:   ToSyncMetricRuleLabelNoticeItems(rule.GetLabelNotices()),
		Duration:       durationpb.New(time.Duration(rule.GetDuration()) * time.Second),
		Enable:         rule.GetStatus().IsEnable() && rule.GetDeletedAt() == 0,
	}
}

func ToSyncMetricSimpleDatasourceItems(datasource []do.DatasourceMetric, teamDo do.Team) []*common.MetricStrategyItem_MetricDatasourceItem {
	if validate.IsNil(datasource) {
		return nil
	}
	return slices.MapFilter(datasource, func(item do.DatasourceMetric) (*common.MetricStrategyItem_MetricDatasourceItem, bool) {
		syncItem := ToSyncMetricSimpleDatasourceItem(item)
		return syncItem, validate.IsNotNil(syncItem)
	})
}

func ToSyncMetricSimpleDatasourceItem(item do.DatasourceMetric) *common.MetricStrategyItem_MetricDatasourceItem {
	if validate.IsNil(item) {
		return nil
	}
	return &common.MetricStrategyItem_MetricDatasourceItem{
		Driver: common.MetricDatasourceDriver(item.GetDriver().GetValue()),
		Id:     item.GetID(),
	}
}

func ToSyncReceiverRoutesItems(notices []do.NoticeGroup) []string {
	if validate.IsNil(notices) {
		return nil
	}
	return slices.MapFilter(notices, func(noticeGroup do.NoticeGroup) (string, bool) {
		if validate.IsNil(noticeGroup) {
			return "", false
		}
		return fmt.Sprintf("%d:%d", noticeGroup.GetTeamID(), noticeGroup.GetID()), true
	})
}

func ToSyncMetricRuleLabelNoticeItems(labelNotices []do.StrategyMetricRuleLabelNotice) []*common.MetricStrategyItem_LabelNotices {
	if validate.IsNil(labelNotices) {
		return nil
	}
	return slices.MapFilter(labelNotices, func(labelNotice do.StrategyMetricRuleLabelNotice) (*common.MetricStrategyItem_LabelNotices, bool) {
		if validate.IsNil(labelNotice) {
			return nil, false
		}
		return ToSyncMetricRuleLabelNoticeItem(labelNotice), true
	})
}

func ToSyncMetricRuleLabelNoticeItem(labelNotice do.StrategyMetricRuleLabelNotice) *common.MetricStrategyItem_LabelNotices {
	if validate.IsNil(labelNotice) {
		return nil
	}
	return &common.MetricStrategyItem_LabelNotices{
		Key:            labelNotice.GetLabelKey(),
		Value:          labelNotice.GetLabelValue(),
		ReceiverRoutes: ToSyncReceiverRoutesItems(labelNotice.GetNotices()),
	}
}
