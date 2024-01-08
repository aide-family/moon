package bo

import (
	"prometheus-manager/api"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/vo"
	"prometheus-manager/pkg/util/slices"
)

type (
	AlarmBeenNotifyMemberBO struct {
		ID                uint32         `json:"id"`
		RealtimeAlarmID   uint32         `json:"realtimeAlarmID"`
		Status            vo.Status      `json:"status"`
		NotifyTypes       vo.NotifyTypes `json:"notifyTypes"`
		MemberId          uint32         `json:"memberId"`
		PromAlarmNotifyID uint32         `json:"promAlarmNotifyID"`
		Msg               string         `json:"msg"`
		Member            *UserBO        `json:"member"`

		CreatedAt int64 `json:"createdAt"`
		UpdatedAt int64 `json:"updatedAt"`
		DeletedAt int64 `json:"deletedAt"`
	}
)

// GetNotifyTypes .
func (l *AlarmBeenNotifyMemberBO) GetNotifyTypes() vo.NotifyTypes {
	if l == nil {
		return vo.NotifyTypes{}
	}
	return l.NotifyTypes
}

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
		NotifyTypes:       l.GetNotifyTypes(),
		MemberId:          l.MemberId,
		Msg:               l.Msg,
		Status:            l.Status,
		PromAlarmNotifyID: l.PromAlarmNotifyID,
	}
}

// ToApi 转换为api
func (l *AlarmBeenNotifyMemberBO) ToApi() *api.BeNotifyMemberDetail {
	return &api.BeNotifyMemberDetail{
		MemberId:    l.MemberId,
		NotifyTypes: slices.To(l.GetNotifyTypes(), func(i vo.NotifyType) int32 { return int32(i) }),
		User:        l.GetMember().ToApiSelectV1(),
		Status:      l.Status.Value(),
		Id:          l.ID,
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
		NotifyTypes:       m.GetNotifyTypes(),
		MemberId:          m.MemberId,
		PromAlarmNotifyID: m.PromAlarmNotifyID,
		Msg:               m.Msg,
		CreatedAt:         m.CreatedAt.Unix(),
		UpdatedAt:         m.UpdatedAt.Unix(),
		DeletedAt:         int64(m.DeletedAt),
	}
}
