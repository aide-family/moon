package bo

import (
	query "github.com/aide-cloud/gorm-normalize"

	"prometheus-manager/api"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/valueobj"
)

type (
	ChatGroupBO struct {
		Id        uint32             `json:"id"`
		Name      string             `json:"name"`
		Status    valueobj.Status    `json:"status"`
		Remark    string             `json:"remark"`
		CreatedAt int64              `json:"createdAt"`
		UpdatedAt int64              `json:"updatedAt"`
		DeletedAt int64              `json:"deletedAt"`
		Hook      string             `json:"hook"`
		NotifyApp valueobj.NotifyApp `json:"notifyApp"`
		HookName  string             `json:"hookName"`
	}
)

// ToApi ...
func (b *ChatGroupBO) ToApi() *api.ChatGroup {
	if b == nil {
		return nil
	}
	return &api.ChatGroup{
		Id:        b.Id,
		Name:      b.Name,
		Remark:    b.Remark,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
		Hook:      b.Hook,
		Status:    b.Status.Value(),
		App:       b.NotifyApp.Value(),
		HookName:  b.HookName,
	}
}

// ToSelectApi ...
func (b *ChatGroupBO) ToSelectApi() *api.ChatGroupSelectV1 {
	if b == nil {
		return nil
	}
	return &api.ChatGroupSelectV1{
		Value:  b.Id,
		App:    b.NotifyApp.Value(),
		Label:  b.HookName,
		Status: b.Status.Value(),
	}
}

func (b *ChatGroupBO) ToModel() *model.PromAlarmChatGroup {
	if b == nil {
		return nil
	}
	return &model.PromAlarmChatGroup{
		BaseModel: query.BaseModel{ID: b.Id},
		Status:    b.Status,
		Remark:    b.Remark,
		Name:      b.Name,
		Hook:      b.Hook,
		NotifyApp: b.NotifyApp,
		HookName:  b.HookName,
	}
}

// ChatGroupApiToBO ...
func ChatGroupApiToBO(a *api.ChatGroup) *ChatGroupBO {
	if a == nil {
		return nil
	}
	return &ChatGroupBO{
		Id:        a.Id,
		Name:      a.Name,
		Status:    valueobj.Status(a.Status),
		Remark:    a.Remark,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
		Hook:      a.Hook,
		NotifyApp: valueobj.NotifyApp(a.App),
		HookName:  a.HookName,
	}
}

// ChatGroupModelToBO .
func ChatGroupModelToBO(m *model.PromAlarmChatGroup) *ChatGroupBO {
	if m == nil {
		return nil
	}
	return &ChatGroupBO{
		Id:        m.ID,
		Name:      m.Name,
		Status:    m.Status,
		Remark:    m.Remark,
		CreatedAt: m.CreatedAt.Unix(),
		UpdatedAt: m.UpdatedAt.Unix(),
		DeletedAt: int64(m.DeletedAt),
		Hook:      m.Hook,
		NotifyApp: m.NotifyApp,
		HookName:  m.HookName,
	}
}
