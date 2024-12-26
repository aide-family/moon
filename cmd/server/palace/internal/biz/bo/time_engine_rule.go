package bo

import (
	"context"

	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	// CreateTimeEngineRuleRequest 创建时间引擎规则请求
	CreateTimeEngineRuleRequest struct {
		Name     string                  `json:"name"`
		Remark   string                  `json:"remark"`
		Status   vobj.Status             `json:"status"`
		Category vobj.TimeEngineRuleType `json:"category"`
		Rule     string                  `json:"rule"`
	}

	// UpdateTimeEngineRuleRequest 更新时间引擎规则请求
	UpdateTimeEngineRuleRequest struct {
		ID       uint32                  `json:"id"`
		Name     string                  `json:"name"`
		Remark   string                  `json:"remark"`
		Status   vobj.Status             `json:"status"`
		Category vobj.TimeEngineRuleType `json:"category"`
		Rule     string                  `json:"rule"`
	}

	// DeleteTimeEngineRuleRequest 删除时间引擎规则请求
	DeleteTimeEngineRuleRequest struct {
		ID uint32 `json:"id"`
	}

	// GetTimeEngineRuleRequest 获取时间引擎规则请求
	GetTimeEngineRuleRequest struct {
		ID uint32 `json:"id"`
	}

	// ListTimeEngineRuleRequest 获取时间引擎规则列表请求
	ListTimeEngineRuleRequest struct {
		Page     types.Pagination        `json:"page"`
		Category vobj.TimeEngineRuleType `json:"category"`
		Status   vobj.Status             `json:"status"`
		Keyword  string                  `json:"keyword"`
	}

	// BatchUpdateTimeEngineRuleStatusRequest 批量更新时间引擎规则状态请求
	BatchUpdateTimeEngineRuleStatusRequest struct {
		IDs    []uint32    `json:"ids"`
		Status vobj.Status `json:"status"`
	}

	// CreateTimeEngineRequest 创建时间引擎请求
	CreateTimeEngineRequest struct {
		Name    string      `json:"name"`
		Remark  string      `json:"remark"`
		Status  vobj.Status `json:"status"`
		RuleIDs []uint32    `json:"rule_ids"`
	}

	// UpdateTimeEngineRequest 更新时间引擎请求
	UpdateTimeEngineRequest struct {
		ID      uint32      `json:"id"`
		Name    string      `json:"name"`
		Remark  string      `json:"remark"`
		Status  vobj.Status `json:"status"`
		RuleIDs []uint32    `json:"rule_ids"`
	}

	// DeleteTimeEngineRequest 删除时间引擎请求
	DeleteTimeEngineRequest struct {
		ID uint32 `json:"id"`
	}

	// GetTimeEngineRequest 获取时间引擎请求
	GetTimeEngineRequest struct {
		ID uint32 `json:"id"`
	}

	// ListTimeEngineRequest 获取时间引擎列表请求
	ListTimeEngineRequest struct {
		Page    types.Pagination `json:"page"`
		Status  vobj.Status      `json:"status"`
		Keyword string           `json:"keyword"`
	}

	// BatchUpdateTimeEngineStatusRequest 批量更新时间引擎状态请求
	BatchUpdateTimeEngineStatusRequest struct {
		IDs    []uint32    `json:"ids"`
		Status vobj.Status `json:"status"`
	}
)

// Do 转换为时间引擎规则
func (r *CreateTimeEngineRuleRequest) Do(ctx context.Context) *bizmodel.TimeEngineRule {
	if r == nil {
		return nil
	}
	do := &bizmodel.TimeEngineRule{
		AllFieldModel: bizmodel.AllFieldModel{
			TeamID: middleware.GetTeamID(ctx),
		},
		Name:     r.Name,
		Remark:   r.Remark,
		Status:   r.Status,
		Category: r.Category,
		Rule:     r.Rule,
	}
	do.WithContext(ctx)
	return do
}

// Do 转换为时间引擎规则
func (r *UpdateTimeEngineRuleRequest) Do(ctx context.Context) *bizmodel.TimeEngineRule {
	if r == nil {
		return nil
	}
	do := &bizmodel.TimeEngineRule{
		AllFieldModel: bizmodel.AllFieldModel{
			AllFieldModel: model.AllFieldModel{ID: r.ID},
			TeamID:        middleware.GetTeamID(ctx),
		},
		Name:     r.Name,
		Remark:   r.Remark,
		Status:   r.Status,
		Category: r.Category,
		Rule:     r.Rule,
	}
	do.WithContext(ctx)
	return do
}

// Do 转换为时间引擎
func (r *CreateTimeEngineRequest) Do(ctx context.Context) *bizmodel.TimeEngine {
	if r == nil {
		return nil
	}
	do := &bizmodel.TimeEngine{
		Name:   r.Name,
		Remark: r.Remark,
		Status: r.Status,
		Rules:  buildRules(ctx, r.RuleIDs),
	}
	do.WithContext(ctx)
	return do
}

// Do 转换为时间引擎
func (r *UpdateTimeEngineRequest) Do(ctx context.Context) *bizmodel.TimeEngine {
	if r == nil {
		return nil
	}
	do := &bizmodel.TimeEngine{
		AllFieldModel: bizmodel.AllFieldModel{
			AllFieldModel: model.AllFieldModel{ID: r.ID},
			TeamID:        middleware.GetTeamID(ctx),
		},
		Name:   r.Name,
		Remark: r.Remark,
		Status: r.Status,
		Rules:  buildRules(ctx, r.RuleIDs),
	}
	do.WithContext(ctx)
	return do
}

// buildRules 构建规则
func buildRules(ctx context.Context, ruleIDs []uint32) []*bizmodel.TimeEngineRule {
	if len(ruleIDs) == 0 {
		return nil
	}
	return types.SliceTo(ruleIDs, func(id uint32) *bizmodel.TimeEngineRule {
		do := &bizmodel.TimeEngineRule{
			AllFieldModel: bizmodel.AllFieldModel{
				AllFieldModel: model.AllFieldModel{ID: id},
				TeamID:        middleware.GetTeamID(ctx),
			},
		}
		do.WithContext(ctx)
		return do
	})
}
