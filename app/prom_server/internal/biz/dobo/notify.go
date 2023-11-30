package dobo

import (
	"time"

	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/api"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/util/slices"
)

type (
	NotifyDO struct {
		Id              uint              `json:"id"`
		Name            string            `json:"name"`
		Status          int32             `json:"status"`
		Remark          string            `json:"remark"`
		CreatedAt       time.Time         `json:"createdAt"`
		UpdatedAt       time.Time         `json:"updatedAt"`
		DeletedAt       int64             `json:"deletedAt"`
		ChatGroups      []*ChatGroupDO    `json:"chatGroups"`
		BeNotifyMembers []*NotifyMemberDO `json:"beNotifyMembers"`
	}

	NotifyBO struct {
		Id              uint              `json:"id"`
		Name            string            `json:"name"`
		Status          int32             `json:"status"`
		Remark          string            `json:"remark"`
		CreatedAt       int64             `json:"createdAt"`
		UpdatedAt       int64             `json:"updatedAt"`
		DeletedAt       int64             `json:"deletedAt"`
		ChatGroups      []*ChatGroupBO    `json:"chatGroups"`
		BeNotifyMembers []*NotifyMemberBO `json:"beNotifyMembers"`
	}
)

func NewNotifyBO(values ...*NotifyBO) IBO[*NotifyBO, *NotifyDO] {
	return NewBO[*NotifyBO, *NotifyDO](
		BOWithValues[*NotifyBO, *NotifyDO](values...),
		BOWithDToB[*NotifyBO, *NotifyDO](notifyDoToBo),
		BOWithBToD[*NotifyBO, *NotifyDO](notifyBoToDo),
	)
}

func NewNotifyDO(values ...*NotifyDO) IDO[*NotifyBO, *NotifyDO] {
	return NewDO[*NotifyBO, *NotifyDO](
		DOWithValues[*NotifyBO, *NotifyDO](values...),
		DOWithDToB[*NotifyBO, *NotifyDO](notifyDoToBo),
		DOWithBToD[*NotifyBO, *NotifyDO](notifyBoToDo),
	)
}

func notifyDoToBo(d *NotifyDO) *NotifyBO {
	if d == nil {
		return nil
	}
	return &NotifyBO{
		Id:              d.Id,
		Name:            d.Name,
		Status:          d.Status,
		Remark:          d.Remark,
		CreatedAt:       d.CreatedAt.Unix(),
		UpdatedAt:       d.UpdatedAt.Unix(),
		DeletedAt:       d.DeletedAt,
		ChatGroups:      slices.To(d.ChatGroups, func(d *ChatGroupDO) *ChatGroupBO { return chatGroupDoToBo(d) }),
		BeNotifyMembers: slices.To(d.BeNotifyMembers, func(d *NotifyMemberDO) *NotifyMemberBO { return notifyMemberDoToBo(d) }),
	}
}

func notifyBoToDo(b *NotifyBO) *NotifyDO {
	if b == nil {
		return nil
	}
	return &NotifyDO{
		Id:              b.Id,
		Name:            b.Name,
		Status:          b.Status,
		Remark:          b.Remark,
		CreatedAt:       time.Unix(b.CreatedAt, 0),
		UpdatedAt:       time.Unix(b.UpdatedAt, 0),
		DeletedAt:       b.DeletedAt,
		ChatGroups:      slices.To(b.ChatGroups, func(b *ChatGroupBO) *ChatGroupDO { return chatGroupBoToDo(b) }),
		BeNotifyMembers: slices.To(b.BeNotifyMembers, func(b *NotifyMemberBO) *NotifyMemberDO { return notifyMemberBoToDo(b) }),
	}
}

// ToModel ...
func (d *NotifyDO) ToModel() *model.PromNotify {
	return &model.PromNotify{
		BaseModel:       query.BaseModel{ID: d.Id},
		Name:            d.Name,
		Status:          d.Status,
		Remark:          d.Remark,
		ChatGroups:      slices.To(d.ChatGroups, func(d *ChatGroupDO) *model.PromChatGroup { return d.ToModel() }),
		BeNotifyMembers: slices.To(d.BeNotifyMembers, func(d *NotifyMemberDO) *model.PromNotifyMember { return d.ToModel() }),
	}
}

// ToApi ...
func (d *NotifyBO) ToApi() *api.NotifyV1 {
	if d == nil {
		return nil
	}
	return &api.NotifyV1{
		Id:         uint32(d.Id),
		Name:       d.Name,
		Remark:     d.Remark,
		Status:     api.Status(d.Status),
		Members:    nil,
		ChatGroups: slices.To(d.ChatGroups, func(d *ChatGroupBO) *api.ChatGroupSelectV1 { return d.ToSelectApi() }),
		CreatedAt:  d.CreatedAt,
		UpdatedAt:  d.UpdatedAt,
		DeletedAt:  d.DeletedAt,
	}
}

// ToApiSelectV1 ...
func (d *NotifyBO) ToApiSelectV1() *api.NotifySelectV1 {
	if d == nil {
		return nil
	}
	return &api.NotifySelectV1{
		Value:  uint32(d.Id),
		Label:  d.Name,
		Remark: d.Remark,
		Status: api.Status(d.Status),
	}
}

// NotifyModelToDO ...
func NotifyModelToDO(m *model.PromNotify) *NotifyDO {
	if m == nil {
		return nil
	}
	return &NotifyDO{
		Id:              m.ID,
		Name:            m.Name,
		Status:          m.Status,
		Remark:          m.Remark,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
		DeletedAt:       int64(m.DeletedAt),
		ChatGroups:      slices.To(m.ChatGroups, func(m *model.PromChatGroup) *ChatGroupDO { return ChatGroupModelToDO(m) }),
		BeNotifyMembers: slices.To(m.BeNotifyMembers, func(m *model.PromNotifyMember) *NotifyMemberDO { return NotifyMemberModelToDO(m) }),
	}
}
