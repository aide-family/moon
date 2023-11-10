package biz

import (
	"time"

	"prometheus-manager/api"
	"prometheus-manager/pkg/util/slices"
)

type (
	StrategyGroupBO struct {
		Id            uint32
		Name          string
		Remark        string
		Status        api.Status
		StrategyCount int64

		CategoryIds []uint32
		Categories  []*DictBO

		CreatedAt int64
		UpdatedAt int64
		DeletedAt int64
	}

	StrategyGroupDO struct {
		Id            uint
		Name          string
		Remark        string
		Status        int32
		StrategyCount int64

		CategoryIds []uint
		Categories  []*DictDO

		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt int64
	}
)

// NewStrategyGroupBO .
func NewStrategyGroupBO(values ...*StrategyGroupBO) IBO[*StrategyGroupBO, *StrategyGroupDO] {
	return NewBO[*StrategyGroupBO, *StrategyGroupDO](
		BOWithValues[*StrategyGroupBO, *StrategyGroupDO](values...),
		BOWithDToB[*StrategyGroupBO, *StrategyGroupDO](strategyGroupDoToBo),
		BOWithBToD[*StrategyGroupBO, *StrategyGroupDO](strategyGroupBoToDo),
	)
}

// NewStrategyGroupDO .
func NewStrategyGroupDO(values ...*StrategyGroupDO) IDO[*StrategyGroupBO, *StrategyGroupDO] {
	return NewDO[*StrategyGroupBO, *StrategyGroupDO](
		DOWithValues[*StrategyGroupBO, *StrategyGroupDO](values...),
		DOWithBToD[*StrategyGroupBO, *StrategyGroupDO](strategyGroupBoToDo),
		DOWithDToB[*StrategyGroupBO, *StrategyGroupDO](strategyGroupDoToBo),
	)
}

// strategyGroupDoToBo .
func strategyGroupDoToBo(d *StrategyGroupDO) *StrategyGroupBO {
	if d == nil {
		return nil
	}
	return &StrategyGroupBO{
		Id:            uint32(d.Id),
		Name:          d.Name,
		Remark:        d.Remark,
		Status:        api.Status(d.Status),
		StrategyCount: d.StrategyCount,
		Categories:    NewDictDO(d.Categories...).BO().List(),
		CategoryIds: slices.To[uint, uint32](d.CategoryIds, func(u uint) uint32 {
			return uint32(u)
		}),
		CreatedAt: d.CreatedAt.Unix(),
		UpdatedAt: d.UpdatedAt.Unix(),
		DeletedAt: d.DeletedAt,
	}
}

// strategyGroupBoToDo .
func strategyGroupBoToDo(b *StrategyGroupBO) *StrategyGroupDO {
	if b == nil {
		return nil
	}
	return &StrategyGroupDO{
		Id:            uint(b.Id),
		Name:          b.Name,
		Remark:        b.Remark,
		Status:        int32(b.Status),
		StrategyCount: b.StrategyCount,
		Categories:    NewDictBO(b.Categories...).DO().List(),
		CategoryIds: slices.To[uint32, uint](b.CategoryIds, func(u uint32) uint {
			return uint(u)
		}),
		CreatedAt: time.Unix(b.CreatedAt, 0),
		UpdatedAt: time.Unix(b.UpdatedAt, 0),
		DeletedAt: b.DeletedAt,
	}
}
