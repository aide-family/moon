package bo

import (
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
)

type NotifyTemplateBO struct {
	Id         uint32                  `json:"id"`
	StrategyID uint32                  `json:"strategyID"`
	NotifyType vobj.NotifyTemplateType `json:"notifyType"`
	Content    string                  `json:"content"`
}

type (
	NotifyTemplateCreateBO struct {
		Content    string                  `json:"content"`
		StrategyID uint32                  `json:"strategyID"`
		NotifyType vobj.NotifyTemplateType `json:"notifyType"`
	}

	NotifyTemplateUpdateBo struct {
		Id         uint32                  `json:"id"`
		Content    string                  `json:"content"`
		StrategyID uint32                  `json:"strategyID"`
		NotifyType vobj.NotifyTemplateType `json:"notifyType"`
	}

	NotifyTemplateListBo struct {
		Page       Pagination `json:"page"`
		StrategyId uint32     `json:"strategyId"`
	}
)

// NotifyTemplateModelToBO 转换
func NotifyTemplateModelToBO(m *do.PromStrategyNotifyTemplate) *NotifyTemplateBO {
	if m == nil {
		return nil
	}
	return &NotifyTemplateBO{
		Id:         m.ID,
		StrategyID: m.StrategyID,
		NotifyType: m.NotifyType,
		Content:    m.Content,
	}
}

// ToApi NotifyTemplateBO to api
func (b *NotifyTemplateBO) ToApi() *api.NotifyTemplateItem {
	if b == nil {
		return nil
	}

	return &api.NotifyTemplateItem{
		Id:         b.Id,
		Content:    b.Content,
		StrategyId: b.StrategyID,
		NotifyType: int32(b.NotifyType),
	}
}

// ToModel NotifyTemplateBO to model
func (b *NotifyTemplateBO) ToModel() *do.PromStrategyNotifyTemplate {
	if b == nil {
		return nil
	}

	return &do.PromStrategyNotifyTemplate{
		BaseModel: do.BaseModel{
			ID: b.Id,
		},
		StrategyID: b.StrategyID,
		NotifyType: b.NotifyType,
		Content:    b.Content,
	}
}
