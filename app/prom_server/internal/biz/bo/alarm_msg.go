package bo

import (
	"prometheus-manager/app/prom_server/internal/biz/vobj"
)

type AlarmMsgBo struct {
	AlarmStatus  vobj.AlarmStatus `json:"alarmStatus"`
	AlarmInfo    *AlertBo         `json:"alarmInfo"`
	StartsAt     int64            `json:"startAt"`
	EndsAt       int64            `json:"endAt"`
	StrategyBO   *StrategyBO      `json:"strategyBO"`
	PromNotifies []*NotifyBO      `json:"promNotifies"`
}
