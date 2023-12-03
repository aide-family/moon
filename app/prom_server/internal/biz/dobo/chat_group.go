package dobo

import (
	"time"

	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/api"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/valueobj"
)

type (
	ChatGroupDO struct {
		Id        uint               `json:"id"`
		Name      string             `json:"name"`
		Status    int32              `json:"status"`
		Remark    string             `json:"remark"`
		CreatedAt time.Time          `json:"createdAt"`
		UpdatedAt time.Time          `json:"updatedAt"`
		DeletedAt int64              `json:"deletedAt"`
		Hook      string             `json:"hook"`
		NotifyApp valueobj.NotifyApp `json:"notifyApp"`
		HookName  string             `json:"hookName"`
	}

	ChatGroupBO struct {
		Id        uint   `json:"id"`
		Name      string `json:"name"`
		Status    int32  `json:"status"`
		Remark    string `json:"remark"`
		CreatedAt int64  `json:"createdAt"`
		UpdatedAt int64  `json:"updatedAt"`
		DeletedAt int64  `json:"deletedAt"`
		Hook      string `json:"hook"`
		NotifyApp int32  `json:"notifyApp"`
		HookName  string `json:"hookName"`
	}
)

func NewChatGroupDO(values ...*ChatGroupDO) IDO[*ChatGroupBO, *ChatGroupDO] {
	return NewDO[*ChatGroupBO, *ChatGroupDO](
		DOWithValues[*ChatGroupBO, *ChatGroupDO](values...),
		DOWithBToD[*ChatGroupBO, *ChatGroupDO](chatGroupBoToDo),
		DOWithDToB[*ChatGroupBO, *ChatGroupDO](chatGroupDoToBo),
	)
}

func NewChatGroupBO(values ...*ChatGroupBO) IBO[*ChatGroupBO, *ChatGroupDO] {
	return NewBO[*ChatGroupBO, *ChatGroupDO](
		BOWithValues[*ChatGroupBO, *ChatGroupDO](values...),
		BOWithBToD[*ChatGroupBO, *ChatGroupDO](chatGroupBoToDo),
		BOWithDToB[*ChatGroupBO, *ChatGroupDO](chatGroupDoToBo),
	)
}

func chatGroupBoToDo(b *ChatGroupBO) *ChatGroupDO {
	if b == nil {
		return nil
	}
	return &ChatGroupDO{
		Id:        b.Id,
		Name:      b.Name,
		Status:    b.Status,
		Remark:    b.Remark,
		CreatedAt: time.Unix(b.CreatedAt, 0),
		UpdatedAt: time.Unix(b.UpdatedAt, 0),
		DeletedAt: b.DeletedAt,
		Hook:      b.Hook,
		NotifyApp: valueobj.NotifyApp(b.NotifyApp),
		HookName:  b.HookName,
	}
}

func chatGroupDoToBo(d *ChatGroupDO) *ChatGroupBO {
	if d == nil {
		return nil
	}
	return &ChatGroupBO{
		Id:        d.Id,
		Name:      d.Name,
		Status:    d.Status,
		Remark:    d.Remark,
		CreatedAt: d.CreatedAt.Unix(),
		UpdatedAt: d.UpdatedAt.Unix(),
		DeletedAt: d.DeletedAt,
		Hook:      d.Hook,
		NotifyApp: int32(d.NotifyApp),
		HookName:  d.HookName,
	}
}

// ToModel ...
func (d *ChatGroupDO) ToModel() *model.PromAlarmChatGroup {
	if d == nil {
		return nil
	}
	return &model.PromAlarmChatGroup{
		BaseModel: query.BaseModel{ID: d.Id},
		Status:    d.Status,
		Remark:    d.Remark,
		Name:      d.Name,
		Hook:      d.Hook,
		NotifyApp: d.NotifyApp.Value(),
		HookName:  d.HookName,
	}
}

// ToApi ...
func (b *ChatGroupBO) ToApi() *api.ChatGroup {
	if b == nil {
		return nil
	}
	return &api.ChatGroup{
		Id:        uint32(b.Id),
		Name:      b.Name,
		Remark:    b.Remark,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
		Hook:      b.Hook,
		Status:    b.Status,
		App:       b.NotifyApp,
		HookName:  b.HookName,
	}
}

// ToSelectApi ...
func (b *ChatGroupBO) ToSelectApi() *api.ChatGroupSelectV1 {
	if b == nil {
		return nil
	}
	return &api.ChatGroupSelectV1{
		Value:  uint32(b.Id),
		App:    b.NotifyApp,
		Label:  b.HookName,
		Status: b.Status,
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

// ChatGroupModelToDO ...
func ChatGroupModelToDO(m *model.PromAlarmChatGroup) *ChatGroupDO {
	if m == nil {
		return nil
	}
	return &ChatGroupDO{
		Id:        m.ID,
		Name:      m.Name,
		Status:    m.Status,
		Remark:    m.Remark,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: int64(m.DeletedAt),
		Hook:      m.Hook,
		NotifyApp: valueobj.NotifyApp(m.NotifyApp),
		HookName:  m.HookName,
	}
}

// ChatGroupApiToBO ...
func ChatGroupApiToBO(a *api.ChatGroup) *ChatGroupBO {
	if a == nil {
		return nil
	}
	return &ChatGroupBO{
		Id:        uint(a.Id),
		Name:      a.Name,
		Status:    int32(a.Status),
		Remark:    a.Remark,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
		Hook:      a.Hook,
		NotifyApp: int32(a.App),
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
