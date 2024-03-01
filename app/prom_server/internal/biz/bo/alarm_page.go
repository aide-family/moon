package bo

import (
	"encoding/json"

	"prometheus-manager/api"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/vo"
	"prometheus-manager/pkg/util/slices"
)

type (
	AlarmPageBO struct {
		Id        uint32    `json:"id"`
		Name      string    `json:"name"`
		Icon      string    `json:"icon"`
		Color     string    `json:"color"`
		Remark    string    `json:"remark"`
		Status    vo.Status `json:"status"`
		CreatedAt int64     `json:"createdAt"`
		UpdatedAt int64     `json:"updatedAt"`
		DeletedAt int64     `json:"deletedAt"`

		PromStrategies []*StrategyBO `json:"promStrategies"`
	}
)

// String json string
func (b *AlarmPageBO) String() string {
	if b == nil {
		return "{}"
	}
	marshal, err := json.Marshal(b)
	if err != nil {
		return "{}"
	}
	return string(marshal)
}

// GetPromStrategies 策略列表
func (b *AlarmPageBO) GetPromStrategies() []*StrategyBO {
	if b == nil {
		return nil
	}
	return b.PromStrategies
}

// ToApiSelectV1 .
func (b *AlarmPageBO) ToApiSelectV1() *api.AlarmPageSelectV1 {
	if b == nil {
		return nil
	}
	return &api.AlarmPageSelectV1{
		Value:  b.Id,
		Label:  b.Name,
		Icon:   b.Icon,
		Color:  b.Color,
		Status: b.Status.Value(),
		Remark: b.Remark,
	}
}

// ListToApiAlarmPageSelectV1 .
func ListToApiAlarmPageSelectV1(values ...*AlarmPageBO) []*api.AlarmPageSelectV1 {
	return slices.To(values, func(info *AlarmPageBO) *api.AlarmPageSelectV1 {
		return info.ToApiSelectV1()
	})
}

// ToModel .
func (b *AlarmPageBO) ToModel() *do.PromAlarmPage {
	if b == nil {
		return nil
	}
	return &do.PromAlarmPage{
		BaseModel: do.BaseModel{ID: b.Id},
		Name:      b.Name,
		Icon:      b.Icon,
		Color:     b.Color,
		Remark:    b.Remark,
		Status:    b.Status,
	}
}

// AlarmPageModelToBO .
func AlarmPageModelToBO(m *do.PromAlarmPage) *AlarmPageBO {
	if m == nil {
		return nil
	}
	return &AlarmPageBO{
		Id:        m.ID,
		Name:      m.Name,
		Icon:      m.Icon,
		Color:     m.Color,
		Remark:    m.Remark,
		Status:    m.Status,
		CreatedAt: m.CreatedAt.Unix(),
		UpdatedAt: m.UpdatedAt.Unix(),
		DeletedAt: int64(m.DeletedAt),
		PromStrategies: slices.To(m.PromStrategies, func(info *do.PromStrategy) *StrategyBO {
			return StrategyModelToBO(info)
		}),
	}
}

// ToApi .
func (b *AlarmPageBO) ToApi() *api.AlarmPageV1 {
	if b == nil {
		return nil
	}
	return &api.AlarmPageV1{
		Id:        b.Id,
		Name:      b.Name,
		Icon:      b.Icon,
		Color:     b.Color,
		Status:    b.Status.Value(),
		Remark:    b.Remark,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
		DeletedAt: b.DeletedAt,
	}
}
