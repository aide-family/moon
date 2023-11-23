package dobo

import (
	"encoding/json"
	"time"

	"prometheus-manager/api"
	"prometheus-manager/app/prom_server/internal/biz/valueobj"
)

type (
	AlarmHistoryBO struct {
		Id         uint32               `json:"id"`
		Md5        string               `json:"md5"`
		StrategyId uint32               `json:"strategyId"`
		StrategyBO *StrategyBO          `json:"strategyBO"`
		LevelId    uint32               `json:"levelId"`
		Level      *DictBO              `json:"level"`
		Status     valueobj.AlarmStatus `json:"status"`
		StartAt    int64                `json:"startAt"`
		EndAt      int64                `json:"endAt"`
		Instance   string               `json:"instance"`
		Duration   int64                `json:"duration"`
		Info       *AlertBo             `json:"info"`
		CreatedAt  int64                `json:"createdAt"`
		UpdatedAt  int64                `json:"UpdatedAt"`
	}

	AlarmHistoryDO struct {
		Id         uint        `json:"id"`
		Md5        string      `json:"md5"`
		StrategyId uint        `json:"strategyId"`
		StrategyDO *StrategyDO `json:"strategyDO"`
		LevelId    uint        `json:"levelId"`
		Level      *DictDO     `json:"level"`
		Status     int32       `json:"status"`
		StartAt    int64       `json:"startAt"`
		EndAt      int64       `json:"endAt"`
		Instance   string      `json:"instance"`
		Duration   int64       `json:"duration"`
		Info       string      `json:"info"`
		CreatedAt  time.Time   `json:"createdAt"`
		UpdatedAt  time.Time   `json:"updatedAt"`
	}
)

// NewAlarmHistoryBO .
func NewAlarmHistoryBO(values ...*AlarmHistoryBO) IBO[*AlarmHistoryBO, *AlarmHistoryDO] {
	return NewBO[*AlarmHistoryBO, *AlarmHistoryDO](
		BOWithValues[*AlarmHistoryBO, *AlarmHistoryDO](values...),
		BOWithBToD[*AlarmHistoryBO, *AlarmHistoryDO](alarmHistoryBoToDo),
		BOWithDToB[*AlarmHistoryBO, *AlarmHistoryDO](alarmHistoryDoToBo),
	)
}

// NewAlarmHistoryDO .
func NewAlarmHistoryDO(values ...*AlarmHistoryDO) IDO[*AlarmHistoryBO, *AlarmHistoryDO] {
	return NewDO[*AlarmHistoryBO, *AlarmHistoryDO](
		DOWithValues[*AlarmHistoryBO, *AlarmHistoryDO](values...),
		DOWithBToD[*AlarmHistoryBO, *AlarmHistoryDO](alarmHistoryBoToDo),
		DOWithDToB[*AlarmHistoryBO, *AlarmHistoryDO](alarmHistoryDoToBo),
	)
}

func alarmHistoryBoToDo(b *AlarmHistoryBO) *AlarmHistoryDO {
	if b == nil {
		return nil
	}
	return &AlarmHistoryDO{
		Id:         uint(b.Id),
		Md5:        b.Md5,
		StrategyId: uint(b.StrategyId),
		LevelId:    uint(b.LevelId),
		Status:     int32(b.Status),
		Instance:   b.Instance,
		Duration:   b.Duration,
		StartAt:    b.StartAt,
		EndAt:      b.EndAt,
		Info:       b.Info.String(),
		CreatedAt:  time.Unix(b.CreatedAt, 0),
		UpdatedAt:  time.Unix(b.UpdatedAt, 0),
	}
}

func alarmHistoryDoToBo(d *AlarmHistoryDO) *AlarmHistoryBO {
	if d == nil {
		return nil
	}

	info := &AlertBo{}
	_ = json.Unmarshal([]byte(d.Info), info)

	return &AlarmHistoryBO{
		Id:         uint32(d.Id),
		Md5:        d.Md5,
		StrategyId: uint32(d.StrategyId),
		StrategyBO: NewStrategyDO(d.StrategyDO).BO().First(),
		LevelId:    uint32(d.LevelId),
		Level:      NewDictDO(d.Level).BO().First(),
		Status:     valueobj.AlarmStatus(d.Status),
		StartAt:    d.StartAt,
		EndAt:      d.EndAt,
		Instance:   d.Instance,
		Duration:   d.Duration,
		Info:       info,
		CreatedAt:  d.CreatedAt.Unix(),
		UpdatedAt:  d.UpdatedAt.Unix(),
	}
}

// ToApiAlarmHistory .
func (b *AlarmHistoryBO) ToApiAlarmHistory() *api.AlarmHistoryV1 {
	if b == nil {
		return nil
	}
	return &api.AlarmHistoryV1{
		Id:          b.Id,
		AlarmId:     b.StrategyId,
		AlarmName:   b.StrategyBO.Alert,
		AlarmLevel:  b.Level.ToApiDictSelectV1(),
		AlarmStatus: b.Info.GetStatus(),
		Labels:      b.Info.GetLabels(),
		Annotations: b.Info.GetAnnotations(),
		StartAt:     b.StartAt,
		EndAt:       b.EndAt,
	}
}
