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
		AlarmName:   b.StrategyBO.Alert,
		AlarmLevel:  b.Level.ToApiSelectV1(),
		AlarmStatus: b.Info.GetStatus(),
		Labels:      b.Info.GetLabels(),
		Annotations: b.Info.GetAnnotations(),
		StartAt:     b.StartAt,
		EndAt:       b.EndAt,
	}
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
		Info:       b.Info.String(),
		StartAt:    b.StartAt,
		EndAt:      b.EndAt,
		Duration:   b.Duration,
		StrategyID: uint(b.StrategyId),
		LevelID:    uint(b.LevelId),
		Md5:        b.Md5,
	}
}

// AlarmHistoryModelToBO .
func AlarmHistoryModelToBO(m *model.PromAlarmHistory) *AlarmHistoryBO {
	if m == nil {
		return nil
	}
	return &AlarmHistoryBO{
		Id:       uint32(m.ID),
		Instance: m.Instance,
		Status:   m.Status,
		// TODO Info:       AlertModelToBO(m.Info),
		//Info:       AlertModelToBO(m.Info),
		StartAt:    m.StartAt,
		EndAt:      m.EndAt,
		Duration:   m.Duration,
		StrategyId: uint32(m.StrategyID),
		LevelId:    uint32(m.LevelID),
		Md5:        m.Md5,
	}
}
