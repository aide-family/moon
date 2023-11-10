package biz

import (
	"time"

	"prometheus-manager/api"
	"prometheus-manager/pkg/alert"
)

type (
	AlarmHistoryBO struct {
		Id         uint32
		Md5        string
		StrategyId uint32
		LevelId    uint32
		Status     api.AlarmStatus

		StartAt int64
		EndAt   int64

		Instance string
		Duration int64

		Info *alert.Alert

		CreatedAt int64
		UpdatedAt int64
	}

	AlarmHistoryDO struct {
		Id         uint
		Md5        string
		StrategyId uint
		LevelId    uint
		Status     int32

		StartAt int64
		EndAt   int64

		Instance string
		Duration int64

		Info string

		CreatedAt time.Time
		UpdatedAt time.Time
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
		Info:       string(b.Info.Byte()),
		CreatedAt:  time.Unix(b.CreatedAt, 0),
		UpdatedAt:  time.Unix(b.UpdatedAt, 0),
	}
}

func alarmHistoryDoToBo(d *AlarmHistoryDO) *AlarmHistoryBO {
	if d == nil {
		return nil
	}
	return &AlarmHistoryBO{
		Id:         uint32(d.Id),
		Md5:        d.Md5,
		StrategyId: uint32(d.StrategyId),
		LevelId:    uint32(d.LevelId),
		Status:     api.AlarmStatus(d.Status),
		Instance:   d.Instance,
		Duration:   d.Duration,
		StartAt:    d.StartAt,
		EndAt:      d.EndAt,
		Info:       alert.NewAlertByString(d.Info),
		CreatedAt:  d.CreatedAt.Unix(),
		UpdatedAt:  d.UpdatedAt.Unix(),
	}
}
