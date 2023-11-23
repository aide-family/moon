package dobo

import (
	"time"

	query "github.com/aide-cloud/gorm-normalize"
	"gorm.io/plugin/soft_delete"
	"prometheus-manager/api"
	"prometheus-manager/app/prom_server/internal/biz/valueobj"
	model2 "prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/util/slices"
)

type (
	StrategyGroupBO struct {
		Id            uint32          `json:"id"`
		Name          string          `json:"name"`
		Remark        string          `json:"remark"`
		Status        valueobj.Status `json:"status"`
		StrategyCount int64           `json:"strategyCount"`
		CategoryIds   []uint32        `json:"categoryIds"`
		Categories    []*DictBO       `json:"categories"`
		CreatedAt     int64           `json:"createdAt"`
		UpdatedAt     int64           `json:"updatedAt"`
		DeletedAt     int64           `json:"deletedAt"`
	}

	StrategyGroupDO struct {
		Id            uint      `json:"id"`
		Name          string    `json:"name"`
		Remark        string    `json:"remark"`
		Status        int32     `json:"status"`
		StrategyCount int64     `json:"strategyCount"`
		CategoryIds   []uint    `json:"categoryIds"`
		Categories    []*DictDO `json:"categories"`
		CreatedAt     time.Time `json:"createdAt"`
		UpdatedAt     time.Time `json:"updatedAt"`
		DeletedAt     int64     `json:"deletedAt"`
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
		Status:        valueobj.Status(d.Status),
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

// ToApiPromGroupSelectV1 .
func (b *StrategyGroupBO) ToApiPromGroupSelectV1() *api.PromGroupSelectV1 {
	if b == nil {
		return nil
	}

	return &api.PromGroupSelectV1{
		Value:    b.Id,
		Label:    b.Name,
		Category: ListToApiDictSelectV1(b.Categories...),
		Status:   api.Status(b.Status),
		Remark:   b.Remark,
	}
}

// ToApiPromPromGroup .
func (b *StrategyGroupBO) ToApiPromPromGroup() *api.PromGroup {
	if b == nil {
		return nil
	}

	return &api.PromGroup{
		Id:   b.Id,
		Name: b.Name,
		Categories: slices.To(b.Categories, func(t *DictBO) *api.DictSelectV1 {
			return t.ToApiDictSelectV1()
		}),
		Status:        api.Status(b.Status),
		Remark:        b.Remark,
		CreatedAt:     b.CreatedAt,
		UpdatedAt:     b.UpdatedAt,
		DeletedAt:     b.DeletedAt,
		StrategyCount: b.StrategyCount,
	}
}

// StrategyGroupModelToDO .
func StrategyGroupModelToDO(m *model2.PromGroup) *StrategyGroupDO {
	if m == nil {
		return nil
	}
	return &StrategyGroupDO{
		Id:            0,
		Name:          m.Name,
		Remark:        m.Remark,
		Status:        m.Status,
		StrategyCount: m.StrategyCount,
		CategoryIds: slices.To(m.Categories, func(t *model2.PromDict) uint {
			if t == nil {
				return 0
			}
			return t.ID
		}),
		Categories: slices.To(m.Categories, func(t *model2.PromDict) *DictDO {
			if t == nil {
				return nil
			}
			return DictModelToDO(t)
		}),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: int64(m.DeletedAt),
	}
}

// StrategyGroupDOToModel .
func StrategyGroupDOToModel(d *StrategyGroupDO) *model2.PromGroup {
	if d == nil {
		return nil
	}
	return &model2.PromGroup{
		BaseModel: query.BaseModel{
			ID:        d.Id,
			CreatedAt: d.CreatedAt,
			UpdatedAt: d.UpdatedAt,
			DeletedAt: soft_delete.DeletedAt(d.DeletedAt),
		},
		Name:           d.Name,
		StrategyCount:  d.StrategyCount,
		Status:         d.Status,
		Remark:         d.Remark,
		PromStrategies: nil,
		Categories: slices.To(d.Categories, func(u *DictDO) *model2.PromDict {
			if u == nil {
				return nil
			}
			return DictDOToModel(u)
		}),
	}
}
