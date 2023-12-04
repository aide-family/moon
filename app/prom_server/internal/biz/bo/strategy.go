package bo

import (
	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/api"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/valueobj"
	"prometheus-manager/pkg/strategy"
	"prometheus-manager/pkg/util/slices"
)

type (
	StrategyBO struct {
		Id             uint32                `json:"id"`
		Alert          string                `json:"alert"`
		Expr           string                `json:"expr"`
		Duration       string                `json:"duration"`
		Labels         *strategy.Labels      `json:"labels"`
		Annotations    *strategy.Annotations `json:"annotations"`
		Status         valueobj.Status       `json:"status"`
		Remark         string                `json:"remark"`
		GroupId        uint32                `json:"groupId"`
		GroupInfo      *StrategyGroupBO      `json:"groupInfo"`
		AlarmLevelId   uint32                `json:"alarmLevelId"`
		AlarmLevelInfo *DictBO               `json:"alarmLevelInfo"`
		AlarmPageIds   []uint32              `json:"alarmPageIds"`
		AlarmPages     []*AlarmPageBO        `json:"alarmPages"`
		CategoryIds    []uint32              `json:"categoryIds"`
		Categories     []*DictBO             `json:"categories"`
		CreatedAt      int64                 `json:"createdAt"`
		UpdatedAt      int64                 `json:"updatedAt"`
		DeletedAt      int64                 `json:"deletedAt"`

		PromNotifies      []*NotifyBO `json:"promNotifies"`
		PromNotifyUpgrade []*NotifyBO `json:"promNotifyUpgrade"`
	}
)

// GetPromNotifies 获取通知信息列表
func (b *StrategyBO) GetPromNotifies() []*NotifyBO {
	if b == nil {
		return nil
	}
	return b.PromNotifies
}

// GetPromNotifyUpgrade 获取告警升级通知信息列表
func (b *StrategyBO) GetPromNotifyUpgrade() []*NotifyBO {
	if b == nil {
		return nil
	}
	return b.PromNotifyUpgrade
}

// GetLabels 获取标签
func (b *StrategyBO) GetLabels() *strategy.Labels {
	if b == nil {
		return nil
	}
	return b.Labels
}

// GetAnnotations 获取注解
func (b *StrategyBO) GetAnnotations() *strategy.Annotations {
	if b == nil {
		return nil
	}
	return b.Annotations
}

// GetGroupInfo 获取分组信息
func (b *StrategyBO) GetGroupInfo() *StrategyGroupBO {
	if b == nil {
		return nil
	}
	return b.GroupInfo
}

// GetAlarmLevelInfo 获取告警级别信息
func (b *StrategyBO) GetAlarmLevelInfo() *DictBO {
	if b == nil {
		return nil
	}
	return b.AlarmLevelInfo
}

// GetAlarmPageIds 获取告警页面ID列表
func (b *StrategyBO) GetAlarmPageIds() []uint32 {
	if b == nil {
		return nil
	}
	return b.AlarmPageIds
}

// GetCategoryIds 获取分类ID列表
func (b *StrategyBO) GetCategoryIds() []uint32 {
	if b == nil {
		return nil
	}
	return b.CategoryIds
}

// GetCategories 获取分类信息列表
func (b *StrategyBO) GetCategories() []*DictBO {
	if b == nil {
		return nil
	}
	return b.Categories
}

// GetAlarmPages .
func (b *StrategyBO) GetAlarmPages() []*AlarmPageBO {
	if b == nil {
		return nil
	}
	return b.AlarmPages
}

// GetAlert .
func (b *StrategyBO) GetAlert() string {
	if b == nil {
		return ""
	}
	return b.Alert
}

// ToApiSelectV1 告警页面列表转换为api告警页面列表
func (b *StrategyBO) ToApiSelectV1() []*api.AlarmPageSelectV1 {
	return ListToApiAlarmPageSelectV1(b.GetAlarmPages()...)
}

// CategoryInfoToApiSelectV1 分类信息转换为api分类列表
func (b *StrategyBO) CategoryInfoToApiSelectV1() []*api.DictSelectV1 {
	return ListToApiDictSelectV1(b.GetCategories()...)
}

// ToApiV1 策略转换为api策略
func (b *StrategyBO) ToApiV1() *api.PromStrategyV1 {
	if b == nil {
		return nil
	}
	strategyBO := b
	return &api.PromStrategyV1{
		Id:           strategyBO.Id,
		Alert:        strategyBO.Alert,
		Expr:         strategyBO.Expr,
		Duration:     strategyBO.Duration,
		Labels:       strategyBO.GetLabels().Map(),
		Annotations:  strategyBO.GetAnnotations().Map(),
		Remark:       strategyBO.Remark,
		Status:       strategyBO.Status.Value(),
		GroupId:      strategyBO.GroupId,
		AlarmLevelId: strategyBO.AlarmLevelId,

		GroupInfo:      strategyBO.GetGroupInfo().ToApiSelectV1(),
		AlarmLevelInfo: strategyBO.GetAlarmLevelInfo().ToApiSelectV1(),
		AlarmPageIds:   strategyBO.GetAlarmPageIds(),
		AlarmPageInfo:  strategyBO.ToApiSelectV1(),
		CategoryIds:    strategyBO.GetCategoryIds(),
		CategoryInfo:   strategyBO.CategoryInfoToApiSelectV1(),
		CreatedAt:      strategyBO.CreatedAt,
		UpdatedAt:      strategyBO.UpdatedAt,
		DeletedAt:      strategyBO.DeletedAt,
	}
}

// ToApiPromStrategySelectV1 策略转换为api策略
func (b *StrategyBO) ToApiPromStrategySelectV1() *api.PromStrategySelectV1 {
	if b == nil {
		return nil
	}

	return &api.PromStrategySelectV1{
		Value:    b.Id,
		Label:    b.Alert,
		Category: ListToApiDictSelectV1(b.GetCategories()...),
		Status:   b.Status.Value(),
	}
}

// ListToApiPromStrategyV1 策略列表转换为api策略列表
func ListToApiPromStrategyV1(values ...*StrategyBO) []*api.PromStrategyV1 {
	list := make([]*api.PromStrategyV1, 0, len(values))
	for _, v := range values {
		list = append(list, v.ToApiV1())
	}
	return list
}

// ListToApiPromStrategySelectV1 策略列表转换为api策略列表
func ListToApiPromStrategySelectV1(values ...*StrategyBO) []*api.PromStrategySelectV1 {
	list := make([]*api.PromStrategySelectV1, 0, len(values))
	for _, v := range values {
		list = append(list, v.ToApiPromStrategySelectV1())
	}
	return list
}

func (b *StrategyBO) ToModel() *model.PromStrategy {
	if b == nil {
		return nil
	}
	return &model.PromStrategy{
		BaseModel: query.BaseModel{
			ID: uint(b.Id),
		},
		GroupID:      uint(b.GroupId),
		Alert:        b.Alert,
		Expr:         b.Expr,
		For:          b.Duration,
		Labels:       b.GetLabels(),
		Annotations:  b.GetAnnotations(),
		AlertLevelID: uint(b.AlarmLevelId),
		Status:       b.Status,
		Remark:       b.Remark,
		AlarmPages: slices.To(b.GetAlarmPages(), func(alarmPageInfo *AlarmPageBO) *model.PromAlarmPage {
			return alarmPageInfo.ToModel()
		}),
		Categories: slices.To(b.GetCategories(), func(dictInfo *DictBO) *model.PromDict {
			return dictInfo.ToModel()
		}),
		AlertLevel: b.GetAlarmLevelInfo().ToModel(),
		GroupInfo:  b.GetGroupInfo().ToModel(),
	}
}

// StrategyModelToBO .
func StrategyModelToBO(m *model.PromStrategy) *StrategyBO {
	if m == nil {
		return nil
	}
	return &StrategyBO{
		Id:             uint32(m.ID),
		Alert:          m.Alert,
		Expr:           m.Expr,
		Duration:       m.For,
		Labels:         m.GetLabels(),
		Annotations:    m.GetAnnotations(),
		Status:         m.Status,
		Remark:         m.Remark,
		GroupId:        uint32(m.GroupID),
		GroupInfo:      StrategyGroupModelToBO(m.GetGroupInfo()),
		AlarmLevelId:   uint32(m.AlertLevelID),
		AlarmLevelInfo: DictModelToBO(m.GetAlertLevel()),
		AlarmPageIds: slices.To(m.GetAlarmPages(), func(alarmPageInfo *model.PromAlarmPage) uint32 {
			return uint32(alarmPageInfo.ID)
		}),
		AlarmPages: slices.To(m.GetAlarmPages(), func(dictInfo *model.PromAlarmPage) *AlarmPageBO {
			return AlarmPageModelToBO(dictInfo)
		}),
		CategoryIds: slices.To(m.GetCategories(), func(dictInfo *model.PromDict) uint32 {
			return uint32(dictInfo.ID)
		}),
		Categories: slices.To(m.GetCategories(), func(dictInfo *model.PromDict) *DictBO {
			return DictModelToBO(dictInfo)
		}),
		CreatedAt: m.CreatedAt.Unix(),
		UpdatedAt: m.UpdatedAt.Unix(),
		DeletedAt: int64(m.DeletedAt),
		PromNotifies: slices.To(m.GetPromNotifies(), func(notifyInfo *model.PromAlarmNotify) *NotifyBO {
			return NotifyModelToBO(notifyInfo)
		}),
		PromNotifyUpgrade: slices.To(m.GetPromNotifyUpgrade(), func(notifyInfo *model.PromAlarmNotify) *NotifyBO {
			return NotifyModelToBO(notifyInfo)
		}),
	}
}
