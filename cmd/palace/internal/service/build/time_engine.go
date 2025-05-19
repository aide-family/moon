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

// ToSaveTimeEngineRequest 转换保存时间引擎请求
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

// ToDeleteTimeEngineRequest 转换删除时间引擎请求
func ToDeleteTimeEngineRequest(req *api.DeleteTimeEngineRequest) *bo.DeleteTimeEngineRequest {
	if req == nil {
		return nil
	}
	return &bo.DeleteTimeEngineRequest{
		TimeEngineId: req.TimeEngineId,
	}
}

// ToGetTimeEngineRequest 转换获取时间引擎详情请求
func ToGetTimeEngineRequest(req *api.GetTimeEngineRequest) *bo.GetTimeEngineRequest {
	if req == nil {
		return nil
	}
	return &bo.GetTimeEngineRequest{
		TimeEngineId: req.TimeEngineId,
	}
}

// ToListTimeEngineRequest 转换获取时间引擎列表请求
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

// ToSelectTimeEngineRequest 转换获取时间引擎列表请求
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

// ToSaveTimeEngineRuleRequest 转换保存时间引擎规则请求
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

// ToDeleteTimeEngineRuleRequest 转换删除时间引擎规则请求
func ToDeleteTimeEngineRuleRequest(req *api.DeleteTimeEngineRuleRequest) *bo.DeleteTimeEngineRuleRequest {
	if req == nil {
		return nil
	}
	return &bo.DeleteTimeEngineRuleRequest{
		TimeEngineRuleId: req.TimeEngineRuleId,
	}
}

// ToGetTimeEngineRuleRequest 转换获取时间引擎规则详情请求
func ToGetTimeEngineRuleRequest(req *api.GetTimeEngineRuleRequest) *bo.GetTimeEngineRuleRequest {
	if req == nil {
		return nil
	}
	return &bo.GetTimeEngineRuleRequest{
		TimeEngineRuleId: req.TimeEngineRuleId,
	}
}

// ToListTimeEngineRuleRequest 转换获取时间引擎规则列表请求
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

// ToSelectTimeEngineRuleRequest 转换获取时间引擎规则列表请求
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

// ToTimeEngineItem 转换时间引擎详情
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

// ToTimeEngineItemRules 转换时间引擎规则列表
func ToTimeEngineItemRules(rules []do.TimeEngineRule) []*common.TimeEngineItemRule {
	return slices.Map(rules, ToTimeEngineItemRule)
}

// ToTimeEngineItemRule 转换时间引擎规则
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

// ToTimeEngineItemEngines 转换时间引擎规则引擎列表
func ToTimeEngineItemEngines(engines []do.TimeEngine) []*common.TimeEngineItem {
	return slices.Map(engines, ToTimeEngineItem)
}

// ToListTimeEngineReply 转换时间引擎列表
func ToListTimeEngineReply(reply *bo.ListTimeEngineReply) *api.ListTimeEngineReply {
	if validate.IsNil(reply) {
		return nil
	}
	return &api.ListTimeEngineReply{
		Pagination: ToPaginationReply(reply.PaginationReply),
		Items:      slices.Map(reply.Items, ToTimeEngineItem),
	}
}

// ToListTimeEngineRuleReply 转换时间引擎规则列表
func ToListTimeEngineRuleReply(reply *bo.ListTimeEngineRuleReply) *api.ListTimeEngineRuleReply {
	if validate.IsNil(reply) {
		return nil
	}
	return &api.ListTimeEngineRuleReply{
		Pagination: ToPaginationReply(reply.PaginationReply),
		Items:      slices.Map(reply.Items, ToTimeEngineItemRule),
	}
}

// ToUpdateTimeEngineStatusRequest 转换更新时间引擎状态请求
func ToUpdateTimeEngineStatusRequest(req *api.UpdateTimeEngineStatusRequest) *bo.UpdateTimeEngineStatusRequest {
	if validate.IsNil(req) {
		return nil
	}
	return &bo.UpdateTimeEngineStatusRequest{
		TimeEngineIds: req.GetTimeEngineIds(),
		Status:        vobj.GlobalStatus(req.GetStatus()),
	}
}

// ToUpdateTimeEngineRuleStatusRequest 转换更新时间引擎规则状态请求
func ToUpdateTimeEngineRuleStatusRequest(req *api.UpdateTimeEngineRuleStatusRequest) *bo.UpdateTimeEngineRuleStatusRequest {
	if validate.IsNil(req) {
		return nil
	}
	return &bo.UpdateTimeEngineRuleStatusRequest{
		TimeEngineRuleIds: req.GetTimeEngineRuleIds(),
		Status:            vobj.GlobalStatus(req.GetStatus()),
	}
}
