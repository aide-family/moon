package biz

import (
	"time"

	"prometheus-manager/api"
	"prometheus-manager/pkg/alert"
	"prometheus-manager/pkg/util/slices"
)

type (
	StrategyBO struct {
		Id          uint32
		Alert       string
		Expr        string
		Duration    int64
		Labels      alert.Labels
		Annotations alert.Annotations
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
		Duration    int64
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

		CreateAt  time.Time
		UpdateAt  time.Time
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
		Labels:      alert.ToLabels(d.Labels),
		Annotations: alert.ToAnnotations(d.Annotations),
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

		CreatedAt: d.CreateAt.Unix(),
		UpdatedAt: d.UpdateAt.Unix(),
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
		Labels:      alert.KV(b.Labels).String(),
		Annotations: alert.KV(b.Annotations).String(),
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

		CreateAt:  time.Unix(b.CreatedAt, 0),
		UpdateAt:  time.Unix(b.UpdatedAt, 0),
		DeletedAt: b.DeletedAt,
	}
}
