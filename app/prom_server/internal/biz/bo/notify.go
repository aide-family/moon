package bo

import (
	"encoding/json"

	"prometheus-manager/api"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/vobj"
	"prometheus-manager/pkg/util/slices"
)

type (
	ListNotifyRequest struct {
		Page    Pagination  `json:"page"`
		Keyword string      `json:"keyword"`
		Status  vobj.Status `json:"status"`
	}
	NotifyBO struct {
		Id                 uint32                 `json:"id"`
		Name               string                 `json:"name"`
		Status             vobj.Status            `json:"status"`
		Remark             string                 `json:"remark"`
		CreatedAt          int64                  `json:"createdAt"`
		UpdatedAt          int64                  `json:"updatedAt"`
		DeletedAt          int64                  `json:"deletedAt"`
		ChatGroups         []*ChatGroupBO         `json:"chatGroups"`
		BeNotifyMembers    []*NotifyMemberBO      `json:"beNotifyMembers"`
		ExternalNotifyObjs []*ExternalNotifyObjBO `json:"externalNotifyObjs"`
	}
)

// String json string
func (d *NotifyBO) String() string {
	if d == nil {
		return "{}"
	}
	marshal, err := json.Marshal(d)
	if err != nil {
		return "{}"
	}
	return string(marshal)
}

// GetChatGroups 获取通知的群组
func (d *NotifyBO) GetChatGroups() []*ChatGroupBO {
	if d == nil {
		return nil
	}
	return d.ChatGroups
}

// GetExternalNotifyObjs 获取通知的外部体系通知对象
func (d *NotifyBO) GetExternalNotifyObjs() []*ExternalNotifyObjBO {
	if d == nil {
		return nil
	}
	return d.ExternalNotifyObjs
}

// GetBeNotifyMembers 获取通知的成员
func (d *NotifyBO) GetBeNotifyMembers() []*NotifyMemberBO {
	if d == nil {
		return nil
	}
	return d.BeNotifyMembers
}

// ToModel ...
func (d *NotifyBO) ToModel() *do.PromAlarmNotify {
	return &do.PromAlarmNotify{
		BaseModel:          do.BaseModel{ID: d.Id},
		Name:               d.Name,
		Status:             d.Status,
		Remark:             d.Remark,
		ChatGroups:         slices.To(d.GetChatGroups(), func(d *ChatGroupBO) *do.PromAlarmChatGroup { return d.ToModel() }),
		BeNotifyMembers:    slices.To(d.GetBeNotifyMembers(), func(d *NotifyMemberBO) *do.PromAlarmNotifyMember { return d.ToModel() }),
		ExternalNotifyObjs: slices.To(d.GetExternalNotifyObjs(), func(d *ExternalNotifyObjBO) *do.ExternalNotifyObj { return d.ToModel() }),
	}
}

// ToApi ...
func (d *NotifyBO) ToApi() *api.NotifyV1 {
	if d == nil {
		return nil
	}
	return &api.NotifyV1{
		Id:                 d.Id,
		Name:               d.Name,
		Remark:             d.Remark,
		Status:             d.Status.Value(),
		Members:            slices.To(d.GetBeNotifyMembers(), func(d *NotifyMemberBO) *api.BeNotifyMemberDetail { return d.ToApi() }),
		ChatGroups:         slices.To(d.GetChatGroups(), func(d *ChatGroupBO) *api.ChatGroupSelectV1 { return d.ToSelectApi() }),
		CreatedAt:          d.CreatedAt,
		UpdatedAt:          d.UpdatedAt,
		DeletedAt:          d.DeletedAt,
		ExternalNotifyObjs: slices.To(d.GetExternalNotifyObjs(), func(d *ExternalNotifyObjBO) *api.ExternalNotifyObj { return d.ToApi() }),
	}
}

// ToApiSelectV1 ...
func (d *NotifyBO) ToApiSelectV1() *api.NotifySelectV1 {
	if d == nil {
		return nil
	}
	return &api.NotifySelectV1{
		Value:  d.Id,
		Label:  d.Name,
		Remark: d.Remark,
		Status: d.Status.Value(),
	}
}

// NotifyModelToBO ...
func NotifyModelToBO(m *do.PromAlarmNotify) *NotifyBO {
	if m == nil {
		return nil
	}
	return &NotifyBO{
		Id:                 m.ID,
		Name:               m.Name,
		Status:             m.Status,
		Remark:             m.Remark,
		CreatedAt:          m.CreatedAt.Unix(),
		UpdatedAt:          m.UpdatedAt.Unix(),
		DeletedAt:          int64(m.DeletedAt),
		ChatGroups:         slices.To(m.GetChatGroups(), func(m *do.PromAlarmChatGroup) *ChatGroupBO { return ChatGroupModelToBO(m) }),
		BeNotifyMembers:    slices.To(m.GetBeNotifyMembers(), func(m *do.PromAlarmNotifyMember) *NotifyMemberBO { return NotifyMemberModelToBO(m) }),
		ExternalNotifyObjs: slices.To(m.GetExternalNotifyObjs(), func(m *do.ExternalNotifyObj) *ExternalNotifyObjBO { return ExternalNotifyObjModelToBO(m) }),
	}
}
