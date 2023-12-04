package bo

import (
	"prometheus-manager/api"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/valueobj"
	"prometheus-manager/pkg/util/slices"

	query "github.com/aide-cloud/gorm-normalize"
)

type (
	AlarmRealtimeBO struct {
		ID                   uint                              `json:"id"`
		Instance             string                            `json:"instance"`
		Note                 string                            `json:"note"`
		Level                *DictBO                           `json:"level"`
		EventAt              int64                             `json:"eventAt"`
		Status               valueobj.AlarmStatus              `json:"status"`
		AlarmPages           []*AlarmPageBO                    `json:"alarmPages"`
		AlarmIntervenes      []*AlarmInterveneBO               `json:"alarmIntervenes"`
		BeNotifyMemberDetail []*AlarmBeenNotifyMemberBO        `json:"beNotifyMemberDetail"`
		NotifiedAt           int64                             `json:"notifiedAt"`
		HistoryID            uint                              `json:"historyId"`
		AlarmUpgradeInfo     *AlarmUpgradeBO                   `json:"alarmUpgradeInfo"`
		AlarmSuppressInfo    *AlarmSuppressBO                  `json:"alarmSuppressInfo"`
		StrategyID           uint                              `json:"strategyId"`
		BeNotifiedChatGroups []*PromAlarmBeenNotifyChatGroupBO `json:"beNotifiedChatGroups"`

		CreatedAt int64 `json:"createdAt"`
		UpdatedAt int64 `json:"updatedAt"`
		DeletedAt int64 `json:"deletedAt"`
	}
)

// GetBeNotifyMemberDetail 获取通知成员详情
func (l *AlarmRealtimeBO) GetBeNotifyMemberDetail() []*AlarmBeenNotifyMemberBO {
	if l == nil {
		return nil
	}
	return l.BeNotifyMemberDetail
}

// GetBeNotifiedChatGroups 获取通知群组详情
func (l *AlarmRealtimeBO) GetBeNotifiedChatGroups() []*PromAlarmBeenNotifyChatGroupBO {
	if l == nil {
		return nil
	}
	return l.BeNotifiedChatGroups
}

// GetAlarmIntervenes 获取告警干预详情
func (l *AlarmRealtimeBO) GetAlarmIntervenes() []*AlarmInterveneBO {
	if l == nil {
		return nil
	}
	return l.AlarmIntervenes
}

// GetAlarmUpgradeInfo 获取告警升级详情
func (l *AlarmRealtimeBO) GetAlarmUpgradeInfo() *AlarmUpgradeBO {
	if l == nil {
		return nil
	}
	return l.AlarmUpgradeInfo
}

// GetAlarmSuppressInfo 获取告警抑制详情
func (l *AlarmRealtimeBO) GetAlarmSuppressInfo() *AlarmSuppressBO {
	if l == nil {
		return nil
	}
	return l.AlarmSuppressInfo
}

func (l *AlarmRealtimeBO) ToModel() *model.PromAlarmRealtime {
	if l == nil {
		return nil
	}

	return &model.PromAlarmRealtime{
		BaseModel:         query.BaseModel{ID: l.ID},
		StrategyID:        l.StrategyID,
		Instance:          l.Instance,
		Note:              l.Note,
		Status:            l.Status,
		EventAt:           l.EventAt,
		BeenNotifyMembers: slices.To(l.GetBeNotifyMemberDetail(), func(i *AlarmBeenNotifyMemberBO) *model.PromAlarmBeenNotifyMember { return i.ToModel() }),
		BeenChatGroups:    slices.To(l.GetBeNotifiedChatGroups(), func(i *PromAlarmBeenNotifyChatGroupBO) *model.PromAlarmBeenNotifyChatGroup { return i.ToModel() }),
		NotifiedAt:        l.NotifiedAt,
		HistoryID:         uint32(l.HistoryID),
		AlarmIntervenes:   slices.To(l.GetAlarmIntervenes(), func(i *AlarmInterveneBO) *model.PromAlarmIntervene { return i.ToModel() }),
		AlarmUpgradeInfo:  l.GetAlarmUpgradeInfo().ToModel(),
		AlarmSuppressInfo: l.GetAlarmSuppressInfo().ToModel(),
	}
}

// ToApi 将BO转为API对象
func (l *AlarmRealtimeBO) ToApi() *api.RealtimeAlarmData {
	if l == nil {
		return nil
	}

	return &api.RealtimeAlarmData{}
}

// AlarmRealtimeModelToBO 将model转为BO对象
func AlarmRealtimeModelToBO(m *model.PromAlarmRealtime) *AlarmRealtimeBO {
	if m == nil {
		return nil
	}

	return &AlarmRealtimeBO{
		ID:              m.ID,
		Instance:        m.Instance,
		Note:            m.Note,
		Level:           DictModelToBO(m.GetStrategy().GetAlertLevel()),
		EventAt:         m.EventAt,
		Status:          m.Status,
		AlarmPages:      slices.To(m.GetStrategy().GetAlarmPages(), func(i *model.PromAlarmPage) *AlarmPageBO { return AlarmPageModelToBO(i) }),
		AlarmIntervenes: slices.To(m.GetAlarmIntervenes(), func(i *model.PromAlarmIntervene) *AlarmInterveneBO { return AlarmInterveneModelToBO(i) }),
		BeNotifyMemberDetail: slices.To(m.GetBeenNotifyMembers(), func(i *model.PromAlarmBeenNotifyMember) *AlarmBeenNotifyMemberBO {
			return AlarmBeenNotifyMemberModelToBO(i)
		}),
		NotifiedAt:        m.NotifiedAt,
		HistoryID:         uint(m.HistoryID),
		AlarmUpgradeInfo:  AlarmUpgradeModelToBO(m.GetAlarmUpgradeInfo()),
		AlarmSuppressInfo: AlarmSuppressModelToBO(m.GetAlarmSuppressInfo()),
		StrategyID:        m.StrategyID,
		BeNotifiedChatGroups: slices.To(m.GetBeenChatGroups(), func(i *model.PromAlarmBeenNotifyChatGroup) *PromAlarmBeenNotifyChatGroupBO {
			return PromAlarmBeenNotifyChatGroupModelToBO(i)
		}),
		CreatedAt: m.CreatedAt.Unix(),
		UpdatedAt: m.UpdatedAt.Unix(),
		DeletedAt: int64(m.DeletedAt),
	}
}
