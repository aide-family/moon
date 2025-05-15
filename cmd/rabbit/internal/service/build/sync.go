package build

import (
	"github.com/aide-family/moon/cmd/rabbit/internal/biz/bo"
	"github.com/aide-family/moon/cmd/rabbit/internal/biz/do"
	"github.com/aide-family/moon/pkg/api/rabbit/common"
	"github.com/aide-family/moon/pkg/util/slices"
)

func ToSMSConfig(smsItem *common.SMSConfig) (bo.SMSConfig, bool) {
	if smsItem == nil {
		return nil, false
	}
	switch smsItem.Type {
	case common.SMSConfig_ALIYUN:
		aliyunConfig := smsItem.GetAliyun()
		return &do.SMSConfig{
			AccessKeyId:     aliyunConfig.GetAccessKeyId(),
			AccessKeySecret: aliyunConfig.GetAccessKeySecret(),
			Endpoint:        aliyunConfig.GetEndpoint(),
			Name:            aliyunConfig.GetName(),
			SignName:        aliyunConfig.GetSignName(),
			Type:            smsItem.GetType(),
			Enable:          smsItem.GetEnable(),
		}, true
	default:
		return nil, false
	}
}

func ToSMSConfigs(smsConfigs []*common.SMSConfig) []bo.SMSConfig {
	return slices.MapFilter(smsConfigs, func(smsItem *common.SMSConfig) (bo.SMSConfig, bool) {
		return ToSMSConfig(smsItem)
	})
}
