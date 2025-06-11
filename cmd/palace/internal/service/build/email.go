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

// ToSaveEmailConfigRequest converts proto request to business object
func ToSaveEmailConfigRequest(req *palace.SaveEmailConfigRequest) *bo.SaveEmailConfigRequest {
	if req == nil {
		return nil
	}

	isSetConfig := validate.TextIsNotNull(req.GetUser()) &&
		validate.TextIsNotNull(req.GetPass()) &&
		validate.TextIsNotNull(req.GetHost()) &&
		validate.TextIsNotNull(req.GetName()) &&
		req.GetPort() > 0

	item := &bo.SaveEmailConfigRequest{
		Config: nil,
		ID:     req.GetEmailConfigId(),
		Name:   req.GetName(),
		Remark: req.GetRemark(),
		Status: vobj.GlobalStatus(req.GetStatus()),
	}
	if isSetConfig {
		item.Config = &do.Email{
			User: req.GetUser(),
			Pass: req.GetPass(),
			Host: req.GetHost(),
			Port: req.GetPort(),
			Name: req.GetName(),
		}
	}
	return item
}

func ToEmailConfigItem(config do.TeamEmailConfig) *palace.EmailConfigItem {
	if validate.IsNil(config) {
		return nil
	}

	return &palace.EmailConfigItem{
		User:          strutil.MaskEmail(config.GetUser()),
		Pass:          strutil.MaskString(config.GetPass(), 0, 4),
		Host:          config.GetHost(),
		Port:          config.GetPort(),
		Status:        common.GlobalStatus(config.GetStatus().GetValue()),
		Name:          config.GetName(),
		Remark:        config.GetRemark(),
		EmailConfigId: config.GetID(),
	}
}

func ToEmailConfigItems(configs []do.TeamEmailConfig) []*palace.EmailConfigItem {
	return slices.Map(configs, ToEmailConfigItem)
}

// ToEmailConfigReply converts business object to proto reply
func ToEmailConfigReply(configs *bo.ListEmailConfigListReply) *palace.GetEmailConfigsReply {
	if validate.IsNil(configs) {
		return &palace.GetEmailConfigsReply{}
	}

	return &palace.GetEmailConfigsReply{
		Items:      ToEmailConfigItems(configs.Items),
		Pagination: ToPaginationReply(configs.PaginationReply),
	}
}

// ToListEmailConfigRequest converts proto request to business object
func ToListEmailConfigRequest(req *palace.GetEmailConfigsRequest) *bo.ListEmailConfigRequest {
	if req == nil {
		return nil
	}

	return &bo.ListEmailConfigRequest{
		PaginationRequest: ToPaginationRequest(req.GetPagination()),
		Keyword:           req.GetKeyword(),
		Status:            vobj.GlobalStatus(req.GetStatus()),
	}
}
