package bo

import (
	"encoding/json"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
	"github.com/aide-family/moon/pkg"
	"github.com/aide-family/moon/pkg/util/slices"
)

type (
	ListGroupReq struct {
		Page              Pagination  `json:"page"`
		Keyword           string      `json:"keyword"`
		Status            vobj.Status `json:"status"`
		PreloadCategories bool        `json:"preloadCategories"`
		Ids               []uint32    `json:"ids"`
		CategoryIds       []uint32    `json:"categoryIds"`
	}
	RemoveStrategyGroupBO struct {
		Id uint32 `json:"id"`
	}
	StrategyGroupBO struct {
		Id                  uint32      `json:"id"`
		Name                string      `json:"name"`
		Remark              string      `json:"remark"`
		Status              vobj.Status `json:"status"`
		StrategyCount       int64       `json:"strategyCount"`
		EnableStrategyCount int64       `json:"enableStrategyCount"`
		CategoryIds         []uint32    `json:"categoryIds"`
		Categories          []*DictBO   `json:"categories"`
		CreatedAt           int64       `json:"createdAt"`
		UpdatedAt           int64       `json:"updatedAt"`
		DeletedAt           int64       `json:"deletedAt"`

		PromStrategies []*StrategyBO `json:"promStrategies"`
	}

	ListAllGroupDetailParams struct {
		GroupIds []uint32 `json:"groupIds"`
	}
)

// GetId .
func (b *StrategyGroupBO) GetId() uint32 {
	if b == nil {
		return 0
	}
	return b.Id
}

// String json string
func (b *StrategyGroupBO) String() string {
	if b == nil {
		return "{}"
	}
	marshal, err := json.Marshal(b)
	if err != nil {
		return "{}"
	}
	return string(marshal)
}

// GetPromStrategies 获取策略列表
func (b *StrategyGroupBO) GetPromStrategies() []*StrategyBO {
	if b == nil {
		return nil
	}
	return b.PromStrategies
}

// GetCategoryIds 获取分类ID列表
func (b *StrategyGroupBO) GetCategoryIds() []uint32 {
	if b == nil {
		return nil
	}
	return b.CategoryIds
}

// GetCategories 获取分类列表
func (b *StrategyGroupBO) GetCategories() []*DictBO {
	if b == nil {
		return nil
	}
	return b.Categories
}

// ToApiSelectV1 .
func (b *StrategyGroupBO) ToApiSelectV1() *api.PromGroupSelectV1 {
	if b == nil {
		return nil
	}

	return &api.PromGroupSelectV1{
		Value:    b.Id,
		Label:    b.Name,
		Category: ListToApiDictSelectV1(b.GetCategories()...),
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
		Categories: slices.To(b.GetCategories(), func(t *DictBO) *api.DictSelectV1 {
			return t.ToApiSelectV1()
		}),
		Status:              b.Status.Value(),
		Remark:              b.Remark,
		CreatedAt:           b.CreatedAt,
		UpdatedAt:           b.UpdatedAt,
		DeletedAt:           b.DeletedAt,
		StrategyCount:       b.StrategyCount,
		EnableStrategyCount: b.EnableStrategyCount,
		Strategies: slices.To(b.GetPromStrategies(), func(u *StrategyBO) *api.PromStrategyV1 {
			return u.ToApiV1()
		}),
	}
}

// ToSimpleApi .
func (b *StrategyGroupBO) ToSimpleApi() *api.GroupSimple {
	if b == nil {
		return nil
	}

	return &api.GroupSimple{
		Id:   b.Id,
		Name: b.Name,
		Strategies: slices.To(b.GetPromStrategies(), func(u *StrategyBO) *api.StrategySimple {
			return u.ToSimpleApi()
		}),
	}
}

// ToApiV2 .
func (b *StrategyGroupBO) ToApiV2() *api.EvaluateGroup {
	if b == nil {
		return nil
	}
	return &api.EvaluateGroup{
		Id:   b.Id,
		Name: b.Name,
		Strategies: slices.ToFilter(b.GetPromStrategies(), func(ru *StrategyBO) (*api.EvaluateStrategyItem, bool) {
			if pkg.IsNil(ru) || !ru.Status.IsEnabled() || pkg.IsNil(ru.Endpoint) || ru.GetEndpoint().Endpoint == "" {
				return nil, false
			}
			return ru.ToApiV2(), true
		}),
	}
}

func (b *StrategyGroupBO) ToModel() *do.PromStrategyGroup {
	if b == nil {
		return nil
	}
	return &do.PromStrategyGroup{
		BaseModel: do.BaseModel{
			ID: b.Id,
		},
		Name:                b.Name,
		StrategyCount:       b.StrategyCount,
		EnableStrategyCount: b.EnableStrategyCount,
		Status:              b.Status,
		Remark:              b.Remark,
		PromStrategies: slices.To(b.GetPromStrategies(), func(u *StrategyBO) *do.PromStrategy {
			return u.ToModel()
		}),
		Categories: slices.To(b.GetCategories(), func(u *DictBO) *do.SysDict {
			return u.ToModel()
		}),
	}
}

// StrategyGroupModelToBO .
func StrategyGroupModelToBO(m *do.PromStrategyGroup) *StrategyGroupBO {
	if m == nil {
		return nil
	}

	return &StrategyGroupBO{
		Id:                  m.ID,
		Name:                m.Name,
		Remark:              m.Remark,
		Status:              m.Status,
		StrategyCount:       m.StrategyCount,
		EnableStrategyCount: m.EnableStrategyCount,
		CategoryIds: slices.To(m.GetCategories(), func(u *do.SysDict) uint32 {
			return u.ID
		}),
		Categories: slices.To(m.GetCategories(), func(u *do.SysDict) *DictBO {
			return DictModelToBO(u)
		}),
		CreatedAt: m.CreatedAt.Unix(),
		UpdatedAt: m.UpdatedAt.Unix(),
		DeletedAt: int64(m.DeletedAt),
		PromStrategies: slices.To(m.GetPromStrategies(), func(u *do.PromStrategy) *StrategyBO {
			return StrategyModelToBO(u)
		}),
	}
}
