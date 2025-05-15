package build

import (
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/pkg/api/palace"
	"github.com/moon-monitor/moon/pkg/api/palace/common"
	"github.com/moon-monitor/moon/pkg/util/slices"
	"github.com/moon-monitor/moon/pkg/util/strutil"
	"github.com/moon-monitor/moon/pkg/util/validate"
)

// ToSaveSMSConfigRequest converts API request to business object
func ToSaveSMSConfigRequest(req *palace.SaveSMSConfigRequest) *bo.SaveSMSConfigRequest {
	if validate.IsNil(req) {
		return nil
	}

	isSetConfig := validate.TextIsNotNull(req.GetAccessKeyId()) &&
		validate.TextIsNotNull(req.GetAccessKeySecret()) &&
		validate.TextIsNotNull(req.GetSignName()) &&
		validate.TextIsNotNull(req.GetEndpoint())
	item := &bo.SaveSMSConfigRequest{
		Config:   nil,
		ID:       req.GetSmsConfigId(),
		Name:     req.GetName(),
		Remark:   req.GetRemark(),
		Status:   vobj.GlobalStatus(req.GetStatus()),
		Provider: vobj.SMSProviderType(req.GetProvider()),
	}
	if isSetConfig {
		item.Config = &do.SMS{
			AccessKeyID:     req.GetAccessKeyId(),
			AccessKeySecret: req.GetAccessKeySecret(),
			SignName:        req.GetSignName(),
			Endpoint:        req.GetEndpoint(),
		}
	}
	return item
}

// ToListSMSConfigRequest converts API request to business object
func ToListSMSConfigRequest(req *palace.GetSMSConfigsRequest) *bo.ListSMSConfigRequest {
	if validate.IsNil(req) {
		return nil
	}
	return &bo.ListSMSConfigRequest{
		PaginationRequest: ToPaginationRequest(req.GetPagination()),
		Keyword:           req.GetKeyword(),
		Status:            vobj.GlobalStatus(req.GetStatus()),
		Provider:          vobj.SMSProviderType(req.GetProvider()),
	}
}

func ToSMSConfigItem(item do.TeamSMSConfig) *palace.SMSConfigItem {
	if validate.IsNil(item) {
		return nil
	}

	config := item.GetSMSConfig()
	if validate.IsNil(config) {
		return nil
	}
	return &palace.SMSConfigItem{
		ProviderType:    common.SMSProviderType(item.GetProviderType().GetValue()),
		AccessKeyId:     strutil.MaskString(config.AccessKeyID, 2, 4),
		AccessKeySecret: strutil.MaskString(config.AccessKeySecret, 2, 4),
		SignName:        strutil.MaskString(config.SignName, 0, 2),
		Endpoint:        config.Endpoint,
		Name:            item.GetName(),
		Remark:          item.GetRemark(),
		SmsConfigId:     item.GetID(),
		Status:          common.GlobalStatus(item.GetStatus().GetValue()),
	}
}

func ToSMSConfigItems(items []do.TeamSMSConfig) []*palace.SMSConfigItem {
	return slices.Map(items, ToSMSConfigItem)
}

// ToSMSConfigReply converts business object to API response
func ToSMSConfigReply(reply *bo.ListSMSConfigListReply) *palace.GetSMSConfigsReply {
	if validate.IsNil(reply) {
		return nil
	}
	return &palace.GetSMSConfigsReply{
		Pagination: ToPaginationReply(reply.PaginationReply),
		Items:      ToSMSConfigItems(reply.Items),
	}
}
