package bo

import (
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
)

type (
	AlarmBeenNotifyMemberBO struct {
		ID                uint32          `json:"id"`
		RealtimeAlarmID   uint32          `json:"realtimeAlarmID"`
		Status            vobj.Status     `json:"status"`
		NotifyType        vobj.NotifyType `json:"notifyType"`
		MemberId          uint32          `json:"memberId"`
		PromAlarmNotifyID uint32          `json:"promAlarmNotifyID"`
		Msg               string          `json:"msg"`
		Member            *UserBO         `json:"member"`

		CreatedAt int64 `json:"createdAt"`
		UpdatedAt int64 `json:"updatedAt"`
		DeletedAt int64 `json:"deletedAt"`
	}
)

// GetMember 获取用户
func (l *AlarmBeenNotifyMemberBO) GetMember() *UserBO {
	if l == nil {
		return nil
	}
	return l.Member
}

// ToModel 转换为model
func (l *AlarmBeenNotifyMemberBO) ToModel() *do.PromAlarmBeenNotifyMember {
	return &do.PromAlarmBeenNotifyMember{
		BaseModel:         do.BaseModel{ID: l.ID},
		RealtimeAlarmID:   l.RealtimeAlarmID,
		NotifyType:        l.NotifyType,
		MemberId:          l.MemberId,
		Msg:               l.Msg,
		Status:            l.Status,
		PromAlarmNotifyID: l.PromAlarmNotifyID,
	}
}

// ToApi 转换为api
func (l *AlarmBeenNotifyMemberBO) ToApi() *api.BeNotifyMemberDetail {
	return &api.BeNotifyMemberDetail{
		MemberId:   l.MemberId,
		NotifyType: l.NotifyType.Value(),
		User:       l.GetMember().ToApiSelectV1(),
		Status:     l.Status.Value(),
		Id:         l.ID,
	}
}

func AlarmBeenNotifyMemberModelToBO(m *do.PromAlarmBeenNotifyMember) *AlarmBeenNotifyMemberBO {
	if m == nil {
		return nil
	}
	return &AlarmBeenNotifyMemberBO{
		ID:                m.ID,
		RealtimeAlarmID:   m.RealtimeAlarmID,
		Status:            m.Status,
		NotifyType:        m.NotifyType,
		MemberId:          m.MemberId,
		PromAlarmNotifyID: m.PromAlarmNotifyID,
		Msg:               m.Msg,
		CreatedAt:         m.CreatedAt.Unix(),
		UpdatedAt:         m.UpdatedAt.Unix(),
		DeletedAt:         int64(m.DeletedAt),
	}
}
