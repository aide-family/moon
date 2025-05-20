package build

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	api "github.com/aide-family/moon/pkg/api/palace"
	"github.com/aide-family/moon/pkg/api/palace/common"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/timex"
	"github.com/aide-family/moon/pkg/util/validate"
)

// ToSaveTimeEngineRequest converts save time engine request
func ToSaveTimeEngineRequest(req *api.SaveTimeEngineRequest) *bo.SaveTimeEngineRequest {
	if req == nil {
		return nil
	}
	return &bo.SaveTimeEngineRequest{
		TimeEngineId: req.TimeEngineId,
		Name:         req.Name,
		Remark:       req.Remark,
		RuleIds:      req.RuleIds,
	}
}

// ToDeleteTimeEngineRequest converts delete time engine request
func ToDeleteTimeEngineRequest(req *api.DeleteTimeEngineRequest) *bo.DeleteTimeEngineRequest {
	if req == nil {
		return nil
	}
	return &bo.DeleteTimeEngineRequest{
		TimeEngineId: req.TimeEngineId,
	}
}

// ToGetTimeEngineRequest converts get time engine details request
func ToGetTimeEngineRequest(req *api.GetTimeEngineRequest) *bo.GetTimeEngineRequest {
	if req == nil {
		return nil
	}
	return &bo.GetTimeEngineRequest{
		TimeEngineId: req.TimeEngineId,
	}
}

// ToListTimeEngineRequest converts get time engine list request
func ToListTimeEngineRequest(req *api.ListTimeEngineRequest) *bo.ListTimeEngineRequest {
	if req == nil {
		return nil
	}
	return &bo.ListTimeEngineRequest{
		PaginationRequest: ToPaginationRequest(req.Pagination),
		Status:            vobj.GlobalStatus(req.Status),
		Keyword:           req.Keyword,
	}
}

// ToSelectTimeEngineRequest converts get time engine list request
func ToSelectTimeEngineRequest(req *api.SelectTimeEngineRequest) *bo.SelectTimeEngineRequest {
	if req == nil {
		return nil
	}
	return &bo.SelectTimeEngineRequest{
		PaginationRequest: ToPaginationRequest(req.Pagination),
		Status:            vobj.GlobalStatus(req.Status),
		Keyword:           req.Keyword,
	}
}

// ToSaveTimeEngineRuleRequest converts save time engine rule request
func ToSaveTimeEngineRuleRequest(req *api.SaveTimeEngineRuleRequest) *bo.SaveTimeEngineRuleRequest {
	if req == nil {
		return nil
	}
	return &bo.SaveTimeEngineRuleRequest{
		TimeEngineRuleId: req.TimeEngineRuleId,
		Name:             req.Name,
		Remark:           req.Remark,
		Type:             vobj.TimeEngineRuleType(req.Type),
		Rules:            slices.Map(req.RuleIds, func(v uint32) int { return int(v) }),
	}
}

// ToDeleteTimeEngineRuleRequest converts delete time engine rule request
func ToDeleteTimeEngineRuleRequest(req *api.DeleteTimeEngineRuleRequest) *bo.DeleteTimeEngineRuleRequest {
	if req == nil {
		return nil
	}
	return &bo.DeleteTimeEngineRuleRequest{
		TimeEngineRuleId: req.TimeEngineRuleId,
	}
}

// ToGetTimeEngineRuleRequest converts get time engine rule details request
func ToGetTimeEngineRuleRequest(req *api.GetTimeEngineRuleRequest) *bo.GetTimeEngineRuleRequest {
	if req == nil {
		return nil
	}
	return &bo.GetTimeEngineRuleRequest{
		TimeEngineRuleId: req.TimeEngineRuleId,
	}
}

// ToListTimeEngineRuleRequest converts get time engine rule list request
func ToListTimeEngineRuleRequest(req *api.ListTimeEngineRuleRequest) *bo.ListTimeEngineRuleRequest {
	if req == nil {
		return nil
	}
	return &bo.ListTimeEngineRuleRequest{
		PaginationRequest: ToPaginationRequest(req.Pagination),
		Status:            vobj.GlobalStatus(req.Status),
		Keyword:           req.Keyword,
		Types:             slices.Map(req.Types, func(t common.TimeEngineRuleType) vobj.TimeEngineRuleType { return vobj.TimeEngineRuleType(t) }),
	}
}

// ToSelectTimeEngineRuleRequest converts get time engine rule list request
func ToSelectTimeEngineRuleRequest(req *api.SelectTimeEngineRuleRequest) *bo.SelectTimeEngineRuleRequest {
	if req == nil {
		return nil
	}
	return &bo.SelectTimeEngineRuleRequest{
		PaginationRequest: ToPaginationRequest(req.Pagination),
		Status:            vobj.GlobalStatus(req.Status),
		Keyword:           req.Keyword,
		Types:             slices.Map(req.Types, func(t common.TimeEngineRuleType) vobj.TimeEngineRuleType { return vobj.TimeEngineRuleType(t) }),
	}
}

// ToTimeEngineItem converts time engine details
func ToTimeEngineItem(item do.TimeEngine) *common.TimeEngineItem {
	if validate.IsNil(item) {
		return nil
	}
	return &common.TimeEngineItem{
		TimeEngineId: item.GetID(),
		Name:         item.GetName(),
		Remark:       item.GetRemark(),
		Status:       common.GlobalStatus(item.GetStatus()),
		CreatedAt:    timex.Format(item.GetCreatedAt()),
		UpdatedAt:    timex.Format(item.GetUpdatedAt()),
		Rules:        ToTimeEngineItemRules(item.GetRules()),
		Creator:      ToUserBaseItem(item.GetCreator()),
	}
}

// ToTimeEngineItemRules converts time engine rule list
func ToTimeEngineItemRules(rules []do.TimeEngineRule) []*common.TimeEngineItemRule {
	return slices.Map(rules, ToTimeEngineItemRule)
}

// ToTimeEngineItemRule converts time engine rule
func ToTimeEngineItemRule(rule do.TimeEngineRule) *common.TimeEngineItemRule {
	if validate.IsNil(rule) {
		return nil
	}
	return &common.TimeEngineItemRule{
		TimeEngineRuleId: rule.GetID(),
		Name:             rule.GetName(),
		Remark:           rule.GetRemark(),
		Status:           common.GlobalStatus(rule.GetStatus()),
		Engines:          ToTimeEngineItemEngines(rule.GetTimeEngines()),
		Type:             common.TimeEngineRuleType(rule.GetType()),
		Rules:            slices.Map(rule.GetRules(), func(v int) int64 { return int64(v) }),
		CreatedAt:        timex.Format(rule.GetCreatedAt()),
		UpdatedAt:        timex.Format(rule.GetUpdatedAt()),
		Creator:          ToUserBaseItem(rule.GetCreator()),
	}
}

// ToTimeEngineItemEngines converts time engine rule engine list
func ToTimeEngineItemEngines(engines []do.TimeEngine) []*common.TimeEngineItem {
	return slices.Map(engines, ToTimeEngineItem)
}

// ToListTimeEngineReply converts time engine list
func ToListTimeEngineReply(reply *bo.ListTimeEngineReply) *api.ListTimeEngineReply {
	if validate.IsNil(reply) {
		return nil
	}
	return &api.ListTimeEngineReply{
		Pagination: ToPaginationReply(reply.PaginationReply),
		Items:      slices.Map(reply.Items, ToTimeEngineItem),
	}
}

// ToListTimeEngineRuleReply converts time engine rule list
func ToListTimeEngineRuleReply(reply *bo.ListTimeEngineRuleReply) *api.ListTimeEngineRuleReply {
	if validate.IsNil(reply) {
		return nil
	}
	return &api.ListTimeEngineRuleReply{
		Pagination: ToPaginationReply(reply.PaginationReply),
		Items:      slices.Map(reply.Items, ToTimeEngineItemRule),
	}
}

// ToUpdateTimeEngineStatusRequest converts update time engine status request
func ToUpdateTimeEngineStatusRequest(req *api.UpdateTimeEngineStatusRequest) *bo.UpdateTimeEngineStatusRequest {
	if validate.IsNil(req) {
		return nil
	}
	return &bo.UpdateTimeEngineStatusRequest{
		TimeEngineIds: req.GetTimeEngineIds(),
		Status:        vobj.GlobalStatus(req.GetStatus()),
	}
}

// ToUpdateTimeEngineRuleStatusRequest converts update time engine rule status request
func ToUpdateTimeEngineRuleStatusRequest(req *api.UpdateTimeEngineRuleStatusRequest) *bo.UpdateTimeEngineRuleStatusRequest {
	if validate.IsNil(req) {
		return nil
	}
	return &bo.UpdateTimeEngineRuleStatusRequest{
		TimeEngineRuleIds: req.GetTimeEngineRuleIds(),
		Status:            vobj.GlobalStatus(req.GetStatus()),
	}
}
