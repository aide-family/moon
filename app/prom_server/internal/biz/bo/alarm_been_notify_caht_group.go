package bo

import (
	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/valueobj"
)

type (
	PromAlarmBeenNotifyChatGroupBO struct {
		ID                uint            `json:"id"`
		Status            valueobj.Status `json:"status"`
		Msg               string          `json:"msg"`
		PromAlarmNotifyID uint            `json:"promAlarmNotifyID"`
		RealtimeAlarmID   uint            `json:"realtimeAlarmID"`
		ChatGroup         *ChatGroupBO    `json:"chatGroup"`
		ChatGroupId       uint            `json:"chatGroupId"`

		CreatedAt int64 `json:"createdAt"`
		UpdatedAt int64 `json:"updatedAt"`
		DeletedAt int64 `json:"deletedAt"`
	}
)

// GetChatGroup .
func (l *PromAlarmBeenNotifyChatGroupBO) GetChatGroup() *ChatGroupBO {
	if l.ChatGroup == nil {
		return nil
	}
	return l.ChatGroup
}

// ToModel 转换为模型
func (l *PromAlarmBeenNotifyChatGroupBO) ToModel() *model.PromAlarmBeenNotifyChatGroup {
	return &model.PromAlarmBeenNotifyChatGroup{
		BaseModel:         query.BaseModel{ID: l.ID},
		RealtimeAlarmID:   l.RealtimeAlarmID,
		ChatGroup:         l.GetChatGroup().ToModel(),
		ChatGroupId:       l.ChatGroupId,
		Status:            l.Status,
		Msg:               l.Msg,
		PromAlarmNotifyID: l.PromAlarmNotifyID,
	}
}

// PromAlarmBeenNotifyChatGroupModelToBO 转换为业务对象
func PromAlarmBeenNotifyChatGroupModelToBO(m *model.PromAlarmBeenNotifyChatGroup) *PromAlarmBeenNotifyChatGroupBO {
	return &PromAlarmBeenNotifyChatGroupBO{
		ID:                m.ID,
		Status:            m.Status,
		Msg:               m.Msg,
		PromAlarmNotifyID: m.PromAlarmNotifyID,
		RealtimeAlarmID:   m.RealtimeAlarmID,
		ChatGroup:         ChatGroupModelToBO(m.GetChatGroup()),
		ChatGroupId:       m.ChatGroupId,
		CreatedAt:         m.CreatedAt.Unix(),
		UpdatedAt:         m.UpdatedAt.Unix(),
		DeletedAt:         int64(m.DeletedAt),
	}
}
