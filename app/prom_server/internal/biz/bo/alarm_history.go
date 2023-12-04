package bo

import (
	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/api"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/valueobj"
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
)

// ToApiV1 .
func (b *AlarmHistoryBO) ToApiV1() *api.AlarmHistoryV1 {
	if b == nil {
		return nil
	}

	return &api.AlarmHistoryV1{
		Id:          b.Id,
		AlarmId:     b.StrategyId,
		AlarmName:   b.GetStrategyBO().GetAlert(),
		AlarmLevel:  b.GetLevel().ToApiSelectV1(),
		AlarmStatus: b.GetInfo().GetStatus(),
		Labels:      b.GetInfo().ToLabelsMap(),
		Annotations: b.GetInfo().ToAnnotationsMap(),
		StartAt:     b.StartAt,
		EndAt:       b.EndAt,
	}
}

// GetLevel .
func (b *AlarmHistoryBO) GetLevel() *DictBO {
	if b == nil {
		return nil
	}
	return b.Level
}

// NewAlarmRealtimeBO .
func (b *AlarmHistoryBO) NewAlarmRealtimeBO() *AlarmRealtimeBO {
	if b == nil {
		return nil
	}
	return &AlarmRealtimeBO{
		Instance:   b.Instance,
		Note:       b.GetInfo().GetAnnotations().Description(),
		Level:      b.GetLevel(),
		EventAt:    b.StartAt,
		Status:     b.Status,
		AlarmPages: b.GetStrategyBO().GetAlarmPages(),
		HistoryID:  uint(b.Id),
		StrategyID: uint(b.StrategyId),
	}
}

// GetStrategyBO .
func (b *AlarmHistoryBO) GetStrategyBO() *StrategyBO {
	if b == nil {
		return nil
	}
	return b.StrategyBO
}

// GetInfo .
func (b *AlarmHistoryBO) GetInfo() *AlertBo {
	if b == nil {
		return nil
	}
	return b.Info
}

// ToModel .
func (b *AlarmHistoryBO) ToModel() *model.PromAlarmHistory {
	if b == nil {
		return nil
	}
	return &model.PromAlarmHistory{
		BaseModel:  query.BaseModel{ID: uint(b.Id)},
		Instance:   b.Instance,
		Status:     b.Status,
		Info:       b.GetInfo().String(),
		StartAt:    b.StartAt,
		EndAt:      b.EndAt,
		Duration:   b.Duration,
		StrategyID: uint(b.StrategyId),
		LevelID:    uint(b.LevelId),
		Md5:        b.Md5,
		Strategy:   b.GetStrategyBO().ToModel(),
		Level:      b.GetLevel().ToModel(),
	}
}

// AlarmHistoryModelToBO .
func AlarmHistoryModelToBO(m *model.PromAlarmHistory) *AlarmHistoryBO {
	if m == nil {
		return nil
	}
	return &AlarmHistoryBO{
		Id:         uint32(m.ID),
		Md5:        m.Md5,
		StrategyId: uint32(m.StrategyID),
		StrategyBO: StrategyModelToBO(m.GetStrategy()),
		LevelId:    uint32(m.LevelID),
		Level:      DictModelToBO(m.GetLevel()),
		Status:     m.Status,
		StartAt:    m.StartAt,
		EndAt:      m.EndAt,
		Instance:   m.Instance,
		Duration:   m.Duration,
		Info:       StringToAlertBo(m.Info),
		CreatedAt:  m.CreatedAt.Unix(),
		UpdatedAt:  m.UpdatedAt.Unix(),
	}
}
