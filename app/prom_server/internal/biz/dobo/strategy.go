package dobo

import (
	"time"

	query "github.com/aide-cloud/gorm-normalize"
	"gorm.io/plugin/soft_delete"
	"prometheus-manager/api"
	model2 "prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/strategy"
	"prometheus-manager/pkg/util/slices"
)

type (
	StrategyBO struct {
		Id          uint32
		Alert       string
		Expr        string
		Duration    string
		Labels      strategy.Labels
		Annotations strategy.Annotations
		Status      api.Status
		Remark      string

		GroupId   uint32
		GroupInfo *StrategyGroupBO

		AlarmLevelId   uint32
		AlarmLevelInfo *DictBO

		AlarmPageIds []uint32
		AlarmPages   []*AlarmPageBO

		CategoryIds []uint32
		Categories  []*DictBO

		CreatedAt int64
		UpdatedAt int64
		DeletedAt int64
	}

	StrategyDO struct {
		Id          uint
		Alert       string
		Expr        string
		Duration    string
		Labels      string
		Annotations string
		Status      int32
		Remark      string

		GroupId   uint
		GroupInfo *StrategyGroupDO

		AlarmLevelId   uint
		AlarmLevelInfo *DictDO

		AlarmPageIds []uint
		AlarmPages   []*AlarmPageDO

		CategoryIds []uint
		Categories  []*DictDO

		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt int64
	}
)

// NewStrategyBO 创建策略业务对象
func NewStrategyBO(values ...*StrategyBO) IBO[*StrategyBO, *StrategyDO] {
	return NewBO[*StrategyBO, *StrategyDO](
		BOWithValues[*StrategyBO, *StrategyDO](values...),
		BOWithDToB[*StrategyBO, *StrategyDO](strategyDoToBo),
		BOWithBToD[*StrategyBO, *StrategyDO](strategyBoToDo),
	)
}

// NewStrategyDO 创建策略数据对象
func NewStrategyDO(values ...*StrategyDO) IDO[*StrategyBO, *StrategyDO] {
	return NewDO[*StrategyBO, *StrategyDO](
		DOWithValues[*StrategyBO, *StrategyDO](values...),
		DOWithBToD[*StrategyBO, *StrategyDO](strategyBoToDo),
		DOWithDToB[*StrategyBO, *StrategyDO](strategyDoToBo),
	)
}

// strategyDoToBo 策略数据对象转换为策略业务对象
func strategyDoToBo(d *StrategyDO) *StrategyBO {
	if d == nil {
		return nil
	}
	return &StrategyBO{
		Id:          uint32(d.Id),
		Alert:       d.Alert,
		Expr:        d.Expr,
		Duration:    d.Duration,
		Labels:      strategy.ToLabels(d.Labels),
		Annotations: strategy.ToAnnotations(d.Annotations),
		Status:      api.Status(d.Status),
		Remark:      d.Remark,

		GroupId:   uint32(d.GroupId),
		GroupInfo: NewStrategyGroupDO(d.GroupInfo).BO().First(),

		AlarmLevelId:   uint32(d.AlarmLevelId),
		AlarmLevelInfo: dictDoToBo(d.AlarmLevelInfo),

		AlarmPageIds: slices.To[uint, uint32](d.AlarmPageIds, func(u uint) uint32 {
			return uint32(u)
		}),
		AlarmPages: NewAlarmPageDO(d.AlarmPages...).BO().List(),

		CategoryIds: slices.To[uint, uint32](d.CategoryIds, func(u uint) uint32 {
			return uint32(u)
		}),
		Categories: NewDictDO(d.Categories...).BO().List(),

		CreatedAt: d.CreatedAt.Unix(),
		UpdatedAt: d.UpdatedAt.Unix(),
		DeletedAt: d.DeletedAt,
	}
}

// strategyBoToDo 策略业务对象转换为策略数据对象
func strategyBoToDo(b *StrategyBO) *StrategyDO {
	if b == nil {
		return nil
	}
	return &StrategyDO{
		Id:          uint(b.Id),
		Alert:       b.Alert,
		Expr:        b.Expr,
		Duration:    b.Duration,
		Labels:      b.Labels.String(),
		Annotations: b.Annotations.String(),
		Status:      int32(b.Status),
		Remark:      b.Remark,

		GroupId:   uint(b.GroupId),
		GroupInfo: NewStrategyGroupBO(b.GroupInfo).DO().First(),

		AlarmLevelId:   uint(b.AlarmLevelId),
		AlarmLevelInfo: dictBoToDo(b.AlarmLevelInfo),

		AlarmPageIds: slices.To[uint32, uint](b.AlarmPageIds, func(u uint32) uint {
			return uint(u)
		}),
		AlarmPages: NewAlarmPageBO(b.AlarmPages...).DO().List(),

		CategoryIds: slices.To[uint32, uint](b.CategoryIds, func(u uint32) uint {
			return uint(u)
		}),
		Categories: NewDictBO(b.Categories...).DO().List(),

		CreatedAt: time.Unix(b.CreatedAt, 0),
		UpdatedAt: time.Unix(b.UpdatedAt, 0),
		DeletedAt: b.DeletedAt,
	}
}

// AlarmPagesToApiAlarmPageSelectV1 告警页面列表转换为api告警页面列表
func (s *StrategyBO) AlarmPagesToApiAlarmPageSelectV1() []*api.AlarmPageSelectV1 {
	return ListToApiAlarmPageSelectV1(s.AlarmPages...)
}

// CategoryInfoToApiDictSelectV1 分类信息转换为api分类列表
func (s *StrategyBO) CategoryInfoToApiDictSelectV1() []*api.DictSelectV1 {
	return ListToApiDictSelectV1(s.Categories...)
}

// ToApiPromStrategyV1 策略转换为api策略
func (s *StrategyBO) ToApiPromStrategyV1() *api.PromStrategyV1 {
	if s == nil {
		return nil
	}
	strategyBO := s
	return &api.PromStrategyV1{
		Id:           strategyBO.Id,
		Alert:        strategyBO.Alert,
		Expr:         strategyBO.Expr,
		Duration:     strategyBO.Duration,
		Labels:       strategyBO.Labels,
		Annotations:  strategyBO.Annotations,
		Remark:       strategyBO.Remark,
		Status:       strategyBO.Status,
		GroupId:      strategyBO.GroupId,
		AlarmLevelId: strategyBO.AlarmLevelId,

		GroupInfo:      strategyBO.GroupInfo.ToApiPromGroupSelectV1(),
		AlarmLevelInfo: strategyBO.AlarmLevelInfo.ToApiDictSelectV1(),
		AlarmPageIds:   strategyBO.AlarmPageIds,
		AlarmPageInfo:  strategyBO.AlarmPagesToApiAlarmPageSelectV1(),
		CategoryIds:    strategyBO.CategoryIds,
		CategoryInfo:   strategyBO.CategoryInfoToApiDictSelectV1(),
		CreatedAt:      strategyBO.CreatedAt,
		UpdatedAt:      strategyBO.UpdatedAt,
		DeletedAt:      strategyBO.DeletedAt,
	}
}

// ToApiPromStrategySelectV1 策略转换为api策略
func (s *StrategyBO) ToApiPromStrategySelectV1() *api.PromStrategySelectV1 {
	if s == nil {
		return nil
	}

	return &api.PromStrategySelectV1{
		Value:    s.Id,
		Label:    s.Alert,
		Category: ListToApiDictSelectV1(s.Categories...),
		Status:   s.Status,
	}
}

// ListToApiPromStrategyV1 策略列表转换为api策略列表
func ListToApiPromStrategyV1(values ...*StrategyBO) []*api.PromStrategyV1 {
	list := make([]*api.PromStrategyV1, 0, len(values))
	for _, v := range values {
		list = append(list, v.ToApiPromStrategyV1())
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

// StrategyModelToDO .
func StrategyModelToDO(m *model2.PromStrategy) *StrategyDO {
	if m == nil {
		return nil
	}
	return &StrategyDO{
		Id:             m.ID,
		Alert:          m.Alert,
		Expr:           m.Alert,
		Duration:       m.For,
		Labels:         m.Labels,
		Annotations:    m.Annotations,
		Status:         m.Status,
		Remark:         m.Remark,
		GroupId:        m.GroupID,
		GroupInfo:      StrategyGroupModelToDO(m.GroupInfo),
		AlarmLevelId:   m.AlertLevelID,
		AlarmLevelInfo: DictModelToDO(m.AlertLevel),
		AlarmPageIds: slices.To(m.AlarmPages, func(i *model2.PromAlarmPage) uint {
			if i == nil {
				return 0
			}
			return i.ID
		}),
		AlarmPages: slices.To(m.AlarmPages, func(i *model2.PromAlarmPage) *AlarmPageDO {
			if i == nil {
				return nil
			}
			return PageModelToDO(i)
		}),
		CategoryIds: slices.To(m.Categories, func(i *model2.PromDict) uint {
			if i == nil {
				return 0
			}
			return i.ID
		}),
		Categories: slices.To(m.Categories, func(i *model2.PromDict) *DictDO {
			if i == nil {
				return nil
			}
			return DictModelToDO(i)
		}),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: int64(m.DeletedAt),
	}
}

// StrategyDOTOModel .
func StrategyDOTOModel(d *StrategyDO) *model2.PromStrategy {
	if d == nil {
		return nil
	}
	return &model2.PromStrategy{
		BaseModel: query.BaseModel{
			ID:        d.Id,
			DeletedAt: soft_delete.DeletedAt(d.DeletedAt),
			UpdatedAt: d.UpdatedAt,
			CreatedAt: d.CreatedAt,
		},
		GroupID:      d.GroupId,
		Alert:        d.Alert,
		Expr:         d.Expr,
		For:          d.Duration,
		Labels:       d.Labels,
		Annotations:  d.Annotations,
		AlertLevelID: d.AlarmLevelId,
		Status:       d.Status,
		Remark:       d.Remark,
		AlarmPages: slices.To(d.AlarmPages, func(alarmPageInfo *AlarmPageDO) *model2.PromAlarmPage {
			if alarmPageInfo == nil {
				return nil
			}
			return PageDOToModel(alarmPageInfo)
		}),
		Categories: slices.To(d.Categories, func(dictInfo *DictDO) *model2.PromDict {
			if dictInfo == nil {
				return nil
			}
			return DictDOToModel(dictInfo)
		}),
		AlertLevel: DictDOToModel(d.AlarmLevelInfo),
		GroupInfo:  StrategyGroupDOToModel(d.GroupInfo),
	}
}
