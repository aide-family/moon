package dobo

import (
	"time"

	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/api"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/valueobj"
	"prometheus-manager/pkg/util/slices"
)

type (
	AlarmPageBO struct {
		Id        uint32          `json:"id"`
		Name      string          `json:"name"`
		Icon      string          `json:"icon"`
		Color     string          `json:"color"`
		Remark    string          `json:"remark"`
		Status    valueobj.Status `json:"status"`
		CreatedAt int64           `json:"createdAt"`
		UpdatedAt int64           `json:"updatedAt"`
		DeletedAt int64           `json:"deletedAt"`
	}

	AlarmPageDO struct {
		Id        uint      `json:"id"`
		Name      string    `json:"name"`
		Icon      string    `json:"icon"`
		Color     string    `json:"color"`
		Remark    string    `json:"remark"`
		Status    int32     `json:"status"`
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
		DeletedAt int64     `json:"deletedAt"`
	}
)

// NewAlarmPageBO .
func NewAlarmPageBO(values ...*AlarmPageBO) IBO[*AlarmPageBO, *AlarmPageDO] {
	return NewBO[*AlarmPageBO, *AlarmPageDO](
		BOWithValues[*AlarmPageBO, *AlarmPageDO](values...),
		BOWithDToB[*AlarmPageBO, *AlarmPageDO](alarmPageDoToBo),
		BOWithBToD[*AlarmPageBO, *AlarmPageDO](alarmPageBoToDo),
	)
}

// NewAlarmPageDO .
func NewAlarmPageDO(values ...*AlarmPageDO) IDO[*AlarmPageBO, *AlarmPageDO] {
	return NewDO[*AlarmPageBO, *AlarmPageDO](
		DOWithValues[*AlarmPageBO, *AlarmPageDO](values...),
		DOWithBToD[*AlarmPageBO, *AlarmPageDO](alarmPageBoToDo),
		DOWithDToB[*AlarmPageBO, *AlarmPageDO](alarmPageDoToBo),
	)
}

// alarmPageDoToBo .
func alarmPageDoToBo(d *AlarmPageDO) *AlarmPageBO {
	if d == nil {
		return nil
	}
	return &AlarmPageBO{
		Id:        uint32(d.Id),
		Name:      d.Name,
		Icon:      d.Icon,
		Color:     d.Color,
		Remark:    d.Remark,
		Status:    valueobj.Status(d.Status),
		CreatedAt: d.CreatedAt.Unix(),
		UpdatedAt: d.UpdatedAt.Unix(),
		DeletedAt: d.DeletedAt,
	}
}

// alarmPageBoToDo .
func alarmPageBoToDo(b *AlarmPageBO) *AlarmPageDO {
	if b == nil {
		return nil
	}
	return &AlarmPageDO{
		Id:        uint(b.Id),
		Name:      b.Name,
		Icon:      b.Icon,
		Color:     b.Color,
		Remark:    b.Remark,
		Status:    int32(b.Status),
		CreatedAt: time.Unix(b.CreatedAt, 0),
		UpdatedAt: time.Unix(b.UpdatedAt, 0),
		DeletedAt: b.DeletedAt,
	}
}

// ToSelectV1 .
func (b *AlarmPageBO) ToSelectV1() *api.AlarmPageSelectV1 {
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
		return info.ToSelectV1()
	})
}

// ToModel .
func (b *AlarmPageDO) ToModel() *model.PromAlarmPage {
	if b == nil {
		return nil
	}
	return &model.PromAlarmPage{
		BaseModel: query.BaseModel{ID: b.Id},
		Name:      b.Name,
		Icon:      b.Icon,
		Color:     b.Color,
		Remark:    b.Remark,
		Status:    b.Status,
	}
}

// PageModelToDO .
func PageModelToDO(m *model.PromAlarmPage) *AlarmPageDO {
	if m == nil {
		return nil
	}
	return &AlarmPageDO{
		Name:   m.Name,
		Remark: m.Remark,
		Icon:   m.Icon,
		Color:  m.Color,
		Status: m.Status,
	}
}
