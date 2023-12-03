package bo

import (
	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/valueobj"
)

type (
	AlarmBeenNotifyMemberBO struct {
		ID                uint                 `json:"id"`
		RealtimeAlarmID   uint                 `json:"realtimeAlarmID"`
		Status            valueobj.Status      `json:"status"`
		NotifyTypes       valueobj.NotifyTypes `json:"notifyTypes"`
		MemberId          uint                 `json:"memberId"`
		PromAlarmNotifyID uint                 `json:"promAlarmNotifyID"`
		Msg               string               `json:"msg"`

		CreatedAt int64 `json:"createdAt"`
		UpdatedAt int64 `json:"updatedAt"`
		DeletedAt int64 `json:"deletedAt"`
	}
)

// ToModel 转换为model
func (l *AlarmBeenNotifyMemberBO) ToModel() *model.PromAlarmBeenNotifyMember {
	return &model.PromAlarmBeenNotifyMember{
		BaseModel:         query.BaseModel{ID: l.ID},
		RealtimeAlarmID:   l.RealtimeAlarmID,
		NotifyTypes:       l.NotifyTypes,
		MemberId:          l.MemberId,
		Member:            nil,
		Msg:               l.Msg,
		Status:            l.Status,
		PromAlarmNotifyID: l.PromAlarmNotifyID,
		PromAlarmNotify:   nil,
	}
}

func AlarmBeenNotifyMemberModelToBO(m *model.PromAlarmBeenNotifyMember) *AlarmBeenNotifyMemberBO {
	return &AlarmBeenNotifyMemberBO{
		ID:                m.ID,
		RealtimeAlarmID:   m.RealtimeAlarmID,
		Status:            m.Status,
		NotifyTypes:       m.NotifyTypes,
		MemberId:          m.MemberId,
		PromAlarmNotifyID: m.PromAlarmNotifyID,
		Msg:               m.Msg,
		CreatedAt:         m.CreatedAt.Unix(),
		UpdatedAt:         m.UpdatedAt.Unix(),
		DeletedAt:         int64(m.DeletedAt),
	}
}
