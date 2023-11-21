package dobo

import (
	"time"

	"prometheus-manager/api"
	"prometheus-manager/pkg/helper/model"
)

type (
	AlarmPageBO struct {
		Id     uint32
		Name   string
		Icon   string
		Color  string
		Remark string
		Status api.Status

		CreatedAt int64
		UpdatedAt int64
		DeletedAt int64
	}

	AlarmPageDO struct {
		Id     uint
		Name   string
		Icon   string
		Color  string
		Remark string
		Status int32

		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt int64
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
		Status:    api.Status(d.Status),
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

// ToApiAlarmPageSelectV1 .
func (b *AlarmPageBO) ToApiAlarmPageSelectV1() *api.AlarmPageSelectV1 {
	return &api.AlarmPageSelectV1{
		Value:  b.Id,
		Label:  b.Name,
		Icon:   b.Icon,
		Color:  b.Color,
		Status: b.Status,
		Remark: b.Remark,
	}
}

// ListToApiAlarmPageSelectV1 .
func ListToApiAlarmPageSelectV1(values ...*AlarmPageBO) []*api.AlarmPageSelectV1 {
	var list []*api.AlarmPageSelectV1
	for _, v := range values {
		list = append(list, v.ToApiAlarmPageSelectV1())
	}
	return list
}

// PageDOToModel .
func PageDOToModel(do *AlarmPageDO) *model.PromAlarmPage {
	if do == nil {
		return nil
	}
	return &model.PromAlarmPage{
		Name:   do.Name,
		Remark: do.Remark,
		Icon:   do.Icon,
		Color:  do.Color,
		Status: do.Status,
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
