package bo

import (
	"encoding"
	"encoding/json"
	"time"

	"prometheus-manager/api"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/vo"
	"prometheus-manager/pkg/util/slices"
)

var _ encoding.BinaryMarshaler = (*AlarmRealtimeBO)(nil)
var _ encoding.BinaryUnmarshaler = (*AlarmRealtimeBO)(nil)

type (
	ListRealtimeReq struct {
		Page        Pagination     `json:"page"`
		Keyword     string         `json:"keyword"`
		Status      vo.AlarmStatus `json:"status"`
		AlarmPageId uint32         `json:"alarm_page_id"`
	}
	AlarmRealtimeBO struct {
		ID                   uint32                            `json:"id"`
		Instance             string                            `json:"instance"`
		Note                 string                            `json:"note"`
		LevelId              uint32                            `json:"levelId"`
		Level                *DictBO                           `json:"level"`
		EventAt              int64                             `json:"eventAt"`
		Status               vo.AlarmStatus                    `json:"status"`
		AlarmIntervenes      []*AlarmInterveneBO               `json:"alarmIntervenes"`
		BeNotifyMemberDetail []*AlarmBeenNotifyMemberBO        `json:"beNotifyMemberDetail"`
		NotifiedAt           int64                             `json:"notifiedAt"`
		HistoryID            uint32                            `json:"historyId"`
		AlarmUpgradeInfo     *AlarmUpgradeBO                   `json:"alarmUpgradeInfo"`
		AlarmSuppressInfo    *AlarmSuppressBO                  `json:"alarmSuppressInfo"`
		StrategyID           uint32                            `json:"strategyId"`
		Strategy             *StrategyBO                       `json:"strategy"`
		BeNotifiedChatGroups []*PromAlarmBeenNotifyChatGroupBO `json:"beNotifiedChatGroups"`

		CreatedAt int64 `json:"createdAt"`
		UpdatedAt int64 `json:"updatedAt"`
		DeletedAt int64 `json:"deletedAt"`
	}
)

// String json string
func (l *AlarmRealtimeBO) String() string {
	if l == nil {
		return "{}"
	}
	marshal, err := json.Marshal(l)
	if err != nil {
		return "{}"
	}
	return string(marshal)
}

func (l *AlarmRealtimeBO) Bytes() []byte {
	if l == nil {
		return nil
	}
	bs, _ := json.Marshal(l)
	return bs
}

func (l *AlarmRealtimeBO) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, l)
}

func (l *AlarmRealtimeBO) MarshalBinary() (data []byte, err error) {
	return json.Marshal(l)
}

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

// GetStrategy 获取策略详情
func (l *AlarmRealtimeBO) GetStrategy() *StrategyBO {
	if l == nil {
		return nil
	}
	return l.Strategy
}

func (l *AlarmRealtimeBO) ToModel() *do.PromAlarmRealtime {
	if l == nil {
		return nil
	}

	return &do.PromAlarmRealtime{
		BaseModel:         do.BaseModel{ID: l.ID},
		StrategyID:        l.StrategyID,
		LevelId:           l.LevelId,
		Instance:          l.Instance,
		Note:              l.Note,
		Status:            l.Status,
		EventAt:           l.EventAt,
		BeenNotifyMembers: slices.To(l.GetBeNotifyMemberDetail(), func(i *AlarmBeenNotifyMemberBO) *do.PromAlarmBeenNotifyMember { return i.ToModel() }),
		BeenChatGroups:    slices.To(l.GetBeNotifiedChatGroups(), func(i *PromAlarmBeenNotifyChatGroupBO) *do.PromAlarmBeenNotifyChatGroup { return i.ToModel() }),
		NotifiedAt:        l.NotifiedAt,
		HistoryID:         l.HistoryID,
		AlarmIntervenes:   slices.To(l.GetAlarmIntervenes(), func(i *AlarmInterveneBO) *do.PromAlarmIntervene { return i.ToModel() }),
		AlarmUpgradeInfo:  l.GetAlarmUpgradeInfo().ToModel(),
		AlarmSuppressInfo: l.GetAlarmSuppressInfo().ToModel(),
	}
}

// GetLevel 获取告警等级详情
func (l *AlarmRealtimeBO) GetLevel() *DictBO {
	if l == nil {
		return nil
	}
	return l.Level
}

// ToApi 将BO转为API对象
func (l *AlarmRealtimeBO) ToApi() *api.RealtimeAlarmData {
	if l == nil {
		return nil
	}

	return &api.RealtimeAlarmData{
		Id:                 l.ID,
		Instance:           l.Instance,
		Note:               l.Note,
		LevelId:            l.LevelId,
		EventAt:            l.EventAt,
		Status:             l.Status.Value(),
		IntervenedUser:     slices.To(l.GetAlarmIntervenes(), func(i *AlarmInterveneBO) *api.InterveneInfo { return i.ToApi() }),
		BeenNotifyMembers:  slices.To(l.GetBeNotifyMemberDetail(), func(i *AlarmBeenNotifyMemberBO) *api.BeNotifyMemberDetail { return i.ToApi() }),
		NotifiedAt:         l.NotifiedAt,
		HistoryId:          l.HistoryID,
		UpgradedUser:       l.GetAlarmUpgradeInfo().ToApi(),
		SuppressedUser:     l.GetAlarmSuppressInfo().ToApi(),
		StrategyId:         l.StrategyID,
		NotifiedChatGroups: slices.To(l.GetBeNotifiedChatGroups(), func(i *PromAlarmBeenNotifyChatGroupBO) *api.ChatGroupSelectV1 { return i.ToApi() }),
		CreatedAt:          l.CreatedAt,
		UpdatedAt:          l.UpdatedAt,
		Level:              l.GetLevel().ToApiSelectV1(),
		Strategy:           l.GetStrategy().ToApiV1(),
		Duration:           time.Unix(time.Now().Unix(), 0).Sub(time.Unix(l.EventAt, 0)).Abs().String(),
	}
}

// AlarmRealtimeModelToBO 将model转为BO对象
func AlarmRealtimeModelToBO(m *do.PromAlarmRealtime) *AlarmRealtimeBO {
	if m == nil {
		return nil
	}

	return &AlarmRealtimeBO{
		ID:              m.ID,
		Instance:        m.Instance,
		Note:            m.Note,
		LevelId:         m.LevelId,
		Level:           DictModelToBO(m.GetLevel()),
		EventAt:         m.EventAt,
		Status:          m.Status,
		AlarmIntervenes: slices.To(m.GetAlarmIntervenes(), func(i *do.PromAlarmIntervene) *AlarmInterveneBO { return AlarmInterveneModelToBO(i) }),
		BeNotifyMemberDetail: slices.To(m.GetBeenNotifyMembers(), func(i *do.PromAlarmBeenNotifyMember) *AlarmBeenNotifyMemberBO {
			return AlarmBeenNotifyMemberModelToBO(i)
		}),
		NotifiedAt:        m.NotifiedAt,
		HistoryID:         m.HistoryID,
		AlarmUpgradeInfo:  AlarmUpgradeModelToBO(m.GetAlarmUpgradeInfo()),
		AlarmSuppressInfo: AlarmSuppressModelToBO(m.GetAlarmSuppressInfo()),
		StrategyID:        m.StrategyID,
		Strategy:          StrategyModelToBO(m.GetStrategy()),
		BeNotifiedChatGroups: slices.To(m.GetBeenChatGroups(), func(i *do.PromAlarmBeenNotifyChatGroup) *PromAlarmBeenNotifyChatGroupBO {
			return PromAlarmBeenNotifyChatGroupModelToBO(i)
		}),
		CreatedAt: m.CreatedAt.Unix(),
		UpdatedAt: m.UpdatedAt.Unix(),
		DeletedAt: int64(m.DeletedAt),
	}
}
