package bo

import (
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
)

// Do 转换为时间引擎规则
func (r *CreateTimeEngineRuleRequest) Do() *bizmodel.TimeEngineRule {
	if r == nil {
		return nil
	}
	return &bizmodel.TimeEngineRule{
		Name:     r.Name,
		Remark:   r.Remark,
		Status:   r.Status,
		Category: r.Category,
		Rule:     r.Rule,
	}
}

// Do 转换为时间引擎规则
func (r *UpdateTimeEngineRuleRequest) Do() *bizmodel.TimeEngineRule {
	if r == nil {
		return nil
	}
	return &bizmodel.TimeEngineRule{
		AllFieldModel: model.AllFieldModel{
			ID: uint32(r.ID),
		},
		Name:     r.Name,
		Remark:   r.Remark,
		Status:   r.Status,
		Category: r.Category,
		Rule:     r.Rule,
	}
}
