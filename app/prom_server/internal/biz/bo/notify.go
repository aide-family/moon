package bo

import (
	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/api"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/util/slices"
)

type (
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

// ToModel ...
func (d *NotifyBO) ToModel() *model.PromAlarmNotify {
	return &model.PromAlarmNotify{
		BaseModel:       query.BaseModel{ID: d.Id},
		Name:            d.Name,
		Status:          d.Status,
		Remark:          d.Remark,
		ChatGroups:      slices.To(d.ChatGroups, func(d *ChatGroupBO) *model.PromAlarmChatGroup { return d.ToModel() }),
		BeNotifyMembers: slices.To(d.BeNotifyMembers, func(d *NotifyMemberBO) *model.PromAlarmNotifyMember { return d.ToModel() }),
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
		Status:     d.Status,
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
		Status: d.Status,
	}
}

// NotifyModelToBO ...
func NotifyModelToBO(m *model.PromAlarmNotify) *NotifyBO {
	if m == nil {
		return nil
	}
	return &NotifyBO{
		Id:              m.ID,
		Name:            m.Name,
		Status:          m.Status,
		Remark:          m.Remark,
		CreatedAt:       m.CreatedAt.Unix(),
		UpdatedAt:       m.UpdatedAt.Unix(),
		DeletedAt:       int64(m.DeletedAt),
		ChatGroups:      slices.To(m.ChatGroups, func(m *model.PromAlarmChatGroup) *ChatGroupBO { return ChatGroupModelToBO(m) }),
		BeNotifyMembers: slices.To(m.BeNotifyMembers, func(m *model.PromAlarmNotifyMember) *NotifyMemberBO { return NotifyMemberModelToBO(m) }),
	}
}
