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
	}
)

// ToApiSelectV1 告警页面列表转换为api告警页面列表
func (b *StrategyBO) ToApiSelectV1() []*api.AlarmPageSelectV1 {
	return ListToApiAlarmPageSelectV1(b.AlarmPages...)
}

// CategoryInfoToApiSelectV1 分类信息转换为api分类列表
func (b *StrategyBO) CategoryInfoToApiSelectV1() []*api.DictSelectV1 {
	return ListToApiDictSelectV1(b.Categories...)
}

// GetAlarmPages .
func (b *StrategyBO) GetAlarmPages() []*AlarmPageBO {
	if b == nil {
		return nil
	}
	return b.AlarmPages
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
		Labels:       strategyBO.Labels.Map(),
		Annotations:  strategyBO.Annotations.Map(),
		Remark:       strategyBO.Remark,
		Status:       strategyBO.Status.Value(),
		GroupId:      strategyBO.GroupId,
		AlarmLevelId: strategyBO.AlarmLevelId,

		GroupInfo:      strategyBO.GroupInfo.ToApiSelectV1(),
		AlarmLevelInfo: strategyBO.AlarmLevelInfo.ToApiSelectV1(),
		AlarmPageIds:   strategyBO.AlarmPageIds,
		AlarmPageInfo:  strategyBO.ToApiSelectV1(),
		CategoryIds:    strategyBO.CategoryIds,
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
		Category: ListToApiDictSelectV1(b.Categories...),
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
		Labels:       b.Labels,
		Annotations:  b.Annotations,
		AlertLevelID: uint(b.AlarmLevelId),
		Status:       b.Status,
		Remark:       b.Remark,
		AlarmPages: slices.To(b.AlarmPages, func(alarmPageInfo *AlarmPageBO) *model.PromAlarmPage {
			return alarmPageInfo.ToModel()
		}),
		Categories: slices.To(b.Categories, func(dictInfo *DictBO) *model.PromDict {
			return dictInfo.ToModel()
		}),
		AlertLevel: b.AlarmLevelInfo.ToModel(),
		GroupInfo:  b.GroupInfo.ToModel(),
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
		Labels:         m.Labels,
		Annotations:    m.Annotations,
		Status:         m.Status,
		Remark:         m.Remark,
		GroupId:        uint32(m.GroupID),
		GroupInfo:      StrategyGroupModelToBO(m.GroupInfo),
		AlarmLevelId:   uint32(m.AlertLevelID),
		AlarmLevelInfo: DictModelToBO(m.AlertLevel),
		AlarmPageIds: slices.To(m.AlarmPages, func(alarmPageInfo *model.PromAlarmPage) uint32 {
			return uint32(alarmPageInfo.ID)
		}),
		AlarmPages: slices.To(m.AlarmPages, func(dictInfo *model.PromAlarmPage) *AlarmPageBO {
			return AlarmPageModelToBO(dictInfo)
		}),
		CategoryIds: slices.To(m.Categories, func(dictInfo *model.PromDict) uint32 {
			return uint32(dictInfo.ID)
		}),
		Categories: slices.To(m.Categories, func(dictInfo *model.PromDict) *DictBO {
			return DictModelToBO(dictInfo)
		}),
		CreatedAt: m.CreatedAt.Unix(),
		UpdatedAt: m.UpdatedAt.Unix(),
		DeletedAt: int64(m.DeletedAt),
	}
}
