package dobo

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
		Status               valueobj.Status                   `json:"status"`
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

func (l *AlarmRealtimeBO) ToModel() *model.PromAlarmRealtime {
	if l == nil {
		return nil
	}

	return &model.PromAlarmRealtime{
		BaseModel:         query.BaseModel{ID: l.ID},
		StrategyID:        l.StrategyID,
		Instance:          l.Instance,
		Note:              l.Note,
		Status:            l.Status.Value(),
		EventAt:           l.EventAt,
		BeenNotifyMembers: slices.To(l.BeNotifyMemberDetail, func(i *AlarmBeenNotifyMemberBO) *model.PromAlarmBeenNotifyMember { return i.ToModel() }),
		BeenChatGroups:    slices.To(l.BeNotifiedChatGroups, func(i *PromAlarmBeenNotifyChatGroupBO) *model.PromAlarmBeenNotifyChatGroup { return i.ToModel() }),
		NotifiedAt:        l.NotifiedAt,
		HistoryID:         uint32(l.HistoryID),
		AlarmIntervenes:   slices.To(l.AlarmIntervenes, func(i *AlarmInterveneBO) *model.PromAlarmIntervene { return i.ToModel() }),
		AlarmUpgradeInfo:  l.AlarmUpgradeInfo.ToModel(),
		AlarmSuppressInfo: l.AlarmSuppressInfo.ToModel(),
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
		Status:          valueobj.Status(m.Status),
		AlarmPages:      slices.To(m.GetStrategy().GetAlarmPages(), func(i *model.PromAlarmPage) *AlarmPageBO { return AlarmPageModelToBO(i) }),
		AlarmIntervenes: slices.To(m.AlarmIntervenes, func(i *model.PromAlarmIntervene) *AlarmInterveneBO { return AlarmInterveneModelToBO(i) }),
		BeNotifyMemberDetail: slices.To(m.BeenNotifyMembers, func(i *model.PromAlarmBeenNotifyMember) *AlarmBeenNotifyMemberBO {
			return AlarmBeenNotifyMemberModelToBO(i)
		}),
		NotifiedAt:        m.NotifiedAt,
		HistoryID:         uint(m.HistoryID),
		AlarmUpgradeInfo:  AlarmUpgradeModelToBO(m.AlarmUpgradeInfo),
		AlarmSuppressInfo: AlarmSuppressModelToBO(m.AlarmSuppressInfo),
		StrategyID:        m.StrategyID,
		BeNotifiedChatGroups: slices.To(m.BeenChatGroups, func(i *model.PromAlarmBeenNotifyChatGroup) *PromAlarmBeenNotifyChatGroupBO {
			return PromAlarmBeenNotifyChatGroupModelToBO(i)
		}),
		CreatedAt: m.CreatedAt.Unix(),
		UpdatedAt: m.UpdatedAt.Unix(),
		DeletedAt: int64(m.DeletedAt),
	}
}
