package bo

import (
	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/api"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/valueobj"
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
)

// ToApiSelectV1 .
func (b *StrategyGroupBO) ToApiSelectV1() *api.PromGroupSelectV1 {
	if b == nil {
		return nil
	}

	return &api.PromGroupSelectV1{
		Value:    b.Id,
		Label:    b.Name,
		Category: ListToApiDictSelectV1(b.Categories...),
		Status:   b.Status.Value(),
		Remark:   b.Remark,
	}
}

// ToApiV1 .
func (b *StrategyGroupBO) ToApiV1() *api.PromGroup {
	if b == nil {
		return nil
	}

	return &api.PromGroup{
		Id:   b.Id,
		Name: b.Name,
		Categories: slices.To(b.Categories, func(t *DictBO) *api.DictSelectV1 {
			return t.ToApiSelectV1()
		}),
		Status:        b.Status.Value(),
		Remark:        b.Remark,
		CreatedAt:     b.CreatedAt,
		UpdatedAt:     b.UpdatedAt,
		DeletedAt:     b.DeletedAt,
		StrategyCount: b.StrategyCount,
	}
}

func (b *StrategyGroupBO) ToModel() *model.PromStrategyGroup {
	if b == nil {
		return nil
	}
	return &model.PromStrategyGroup{
		BaseModel: query.BaseModel{
			ID: uint(b.Id),
		},
		Name:           b.Name,
		StrategyCount:  b.StrategyCount,
		Status:         b.Status,
		Remark:         b.Remark,
		PromStrategies: nil,
		Categories: slices.To(b.Categories, func(u *DictBO) *model.PromDict {
			return u.ToModel()
		}),
	}
}

// StrategyGroupModelToBO .
func StrategyGroupModelToBO(m *model.PromStrategyGroup) *StrategyGroupBO {
	if m == nil {
		return nil
	}

	return &StrategyGroupBO{
		Id:            uint32(m.ID),
		Name:          m.Name,
		Remark:        m.Remark,
		Status:        m.Status,
		StrategyCount: m.StrategyCount,
		CategoryIds: slices.To(m.Categories, func(u *model.PromDict) uint32 {
			return uint32(u.ID)
		}),
		Categories: slices.To(m.Categories, func(u *model.PromDict) *DictBO {
			return DictModelToBO(u)
		}),
		CreatedAt: m.CreatedAt.Unix(),
		UpdatedAt: m.UpdatedAt.Unix(),
		DeletedAt: int64(m.DeletedAt),
	}
}
