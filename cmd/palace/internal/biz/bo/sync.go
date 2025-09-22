package bo

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/api/houyi/common"
	rabbitconmmon "github.com/aide-family/moon/pkg/api/rabbit/common"
	"github.com/aide-family/moon/pkg/util/kv"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

type SyncRequest struct {
	Rows ChangedRows      `json:"rows"`
	Type vobj.ChangedType `json:"type"`
}

// ChangedRows is a map of changed rows.
//
//	key: team id
//	value: changed row ids
type ChangedRows map[uint32][]uint32

type ChangedMetricDatasource ChangedRows

type ChangedMetricStrategy ChangedRows

type ChangedNoticeGroup ChangedRows

type ChangedNoticeSMSConfig ChangedRows

type ChangedNoticeEmailConfig ChangedRows

type ChangedNoticeHookConfig ChangedRows

func ToSyncMetricDatasourceItem(item do.DatasourceMetric) *common.MetricDatasourceItem {
	if validate.IsNil(item) {
		return nil
	}
	return &common.MetricDatasourceItem{
		Driver:         common.MetricDatasourceDriver(item.GetDriver().GetValue()),
		Config:         ToSyncMetricDatasourceConfigItem(item),
		Enable:         item.GetStatus().IsEnable() && item.GetDeletedAt() == 0,
		Id:             item.GetID(),
		Name:           item.GetName(),
		ScrapeInterval: durationpb.New(item.GetScrapeInterval()),
		TeamId:         item.GetTeamID(),
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
		Tls:       ToSyncMetricDatasourceTLSItem(item.GetTLS()),
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

func ToSyncMetricDatasourceTLSItem(tls *do.TLS) *common.TLS {
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

func ToSyncMetricStrategyItem(item do.StrategyMetric) *common.MetricStrategyItem {
	if validate.IsNil(item) {
		return nil
	}
	return &common.MetricStrategyItem{
		TeamId:         item.GetTeamID(),
		Datasource:     ToSyncMetricSimpleDatasourceItems(item.GetDatasourceList()),
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

func ToSyncMetricSimpleDatasourceItems(datasource []do.DatasourceMetric) []*common.MetricStrategyItem_MetricDatasourceItem {
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
		Driver:       common.MetricDatasourceDriver(item.GetDriver().GetValue()),
		DatasourceId: item.GetID(),
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

func ToSyncNoticeGroupItems(groupDos []do.NoticeGroup) []*rabbitconmmon.NoticeGroup {
	if validate.IsNil(groupDos) {
		return nil
	}
	return slices.MapFilter(groupDos, func(groupDo do.NoticeGroup) (*rabbitconmmon.NoticeGroup, bool) {
		if validate.IsNil(groupDo) {
			return nil, false
		}
		item := ToSyncNoticeGroupItem(groupDo)
		if validate.IsNil(item) {
			return nil, false
		}
		return item, true
	})
}

func ToSyncNoticeGroupItem(groupDo do.NoticeGroup) *rabbitconmmon.NoticeGroup {
	if validate.IsNil(groupDo) {
		return nil
	}
	return &rabbitconmmon.NoticeGroup{
		Name:            groupDo.GetName(),
		SmsConfigName:   groupDo.GetSMSConfig().GetName(),
		EmailConfigName: groupDo.GetEmailConfig().GetName(),
		HookReceivers: slices.MapFilter(groupDo.GetHooks(), func(hookConfig do.NoticeHook) (string, bool) {
			if validate.IsNil(hookConfig) {
				return "", false
			}
			return hookConfig.GetName(), true
		}),
		SmsReceivers: slices.MapFilter(groupDo.GetNoticeMembers(), func(smsUser do.NoticeMember) (string, bool) {
			if validate.IsNil(smsUser) {
				return "", false
			}
			if !smsUser.GetNoticeType().IsContainsSMS() {
				return "", false
			}
			member := smsUser.GetMember()
			if validate.IsNil(member) || !member.GetStatus().IsNormal() || member.GetDeletedAt() > 0 {
				return "", false
			}
			user := member.GetUser()
			if validate.IsNil(user) || user.GetDeletedAt() > 0 || !user.GetStatus().IsNormal() {
				return "", false
			}
			return user.GetPhone(), true
		}),
		EmailReceivers: slices.MapFilter(groupDo.GetNoticeMembers(), func(emailUser do.NoticeMember) (string, bool) {
			if validate.IsNil(emailUser) {
				return "", false
			}
			if !emailUser.GetNoticeType().IsContainsEmail() {
				return "", false
			}
			member := emailUser.GetMember()
			if validate.IsNil(member) || !member.GetStatus().IsNormal() || member.GetDeletedAt() > 0 {
				return "", false
			}
			user := member.GetUser()
			if validate.IsNil(user) || user.GetDeletedAt() > 0 || !user.GetStatus().IsNormal() {
				return "", false
			}
			return user.GetEmail(), true
		}),
	}
}

func ToSyncSMSConfigItems(smsDos []do.TeamSMSConfig) []*rabbitconmmon.SMSConfig {
	if validate.IsNil(smsDos) {
		return nil
	}
	return slices.MapFilter(smsDos, func(smsDo do.TeamSMSConfig) (*rabbitconmmon.SMSConfig, bool) {
		if validate.IsNil(smsDo) {
			return nil, false
		}
		item := ToSyncSMSConfigItem(smsDo)
		if validate.IsNil(item) {
			return nil, false
		}
		return item, true
	})
}

func ToSyncSMSConfigItem(smsDo do.TeamSMSConfig) *rabbitconmmon.SMSConfig {
	if validate.IsNil(smsDo) {
		return nil
	}
	item := &rabbitconmmon.SMSConfig{
		Type:   rabbitconmmon.SMSConfig_Type(smsDo.GetProviderType().GetValue()),
		Aliyun: nil,
		Enable: smsDo.GetStatus().IsEnable() && smsDo.GetDeletedAt() == 0,
	}
	switch smsDo.GetProviderType() {
	case vobj.SMSProviderTypeAliyun:
		smsConfig := smsDo.GetSMSConfig()
		if validate.IsNil(smsConfig) {
			return nil
		}
		item.Aliyun = &rabbitconmmon.AliyunSMSConfig{
			AccessKeyId:     smsConfig.AccessKeyID,
			AccessKeySecret: smsConfig.AccessKeySecret,
			SignName:        smsConfig.SignName,
			Endpoint:        smsConfig.Endpoint,
			Name:            smsDo.GetName(),
		}
	default:
		return nil
	}
	return item
}

func ToSyncEmailConfigItems(emailDos []do.TeamEmailConfig) []*rabbitconmmon.EmailConfig {
	if validate.IsNil(emailDos) {
		return nil
	}
	return slices.MapFilter(emailDos, func(emailDo do.TeamEmailConfig) (*rabbitconmmon.EmailConfig, bool) {
		if validate.IsNil(emailDo) {
			return nil, false
		}
		item := ToSyncEmailConfigItem(emailDo)
		if validate.IsNil(item) {
			return nil, false
		}
		return item, true
	})
}

func ToSyncEmailConfigItem(emailDo do.TeamEmailConfig) *rabbitconmmon.EmailConfig {
	if validate.IsNil(emailDo) {
		return nil
	}
	emailConfig := emailDo.GetEmailConfig()
	if validate.IsNil(emailConfig) {
		return nil
	}
	return &rabbitconmmon.EmailConfig{
		Enable: emailDo.GetStatus().IsEnable() && emailDo.GetDeletedAt() == 0,
		User:   emailConfig.User,
		Pass:   emailConfig.Pass,
		Host:   emailConfig.Host,
		Port:   emailConfig.Port,
		Name:   emailDo.GetName(),
	}
}

func ToSyncHookConfigItems(hookDos []do.NoticeHook) []*rabbitconmmon.HookConfig {
	if validate.IsNil(hookDos) {
		return nil
	}
	return slices.MapFilter(hookDos, func(hookDo do.NoticeHook) (*rabbitconmmon.HookConfig, bool) {
		if validate.IsNil(hookDo) {
			return nil, false
		}
		item := ToSyncHookConfigItem(hookDo)
		if validate.IsNil(item) {
			return nil, false
		}
		return item, true
	})
}

func ToSyncHookConfigItem(hookDo do.NoticeHook) *rabbitconmmon.HookConfig {
	if validate.IsNil(hookDo) {
		return nil
	}
	return &rabbitconmmon.HookConfig{
		Name:     hookDo.GetName(),
		App:      rabbitconmmon.HookAPP(hookDo.GetApp().GetValue()),
		Url:      hookDo.GetURL(),
		Secret:   hookDo.GetSecret(),
		Token:    hookDo.GetSecret(),
		Username: "",
		Password: "",
		Headers:  ToSyncHookHeadersItem(hookDo.GetHeaders()),
		Enable:   hookDo.GetStatus().IsEnable() && hookDo.GetDeletedAt() == 0,
	}
}

func ToSyncHookHeadersItem(headers []*kv.KV) map[string]string {
	if validate.IsNil(headers) {
		return nil
	}
	headersMap := make(map[string]string)
	for _, header := range headers {
		if validate.IsNil(header) {
			continue
		}
		headersMap[header.Key] = header.Value
	}
	return headersMap
}
