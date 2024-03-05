package bo

import (
	"encoding/json"

	"prometheus-manager/api"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

type (
	ListHistoryRequest struct {
		Curr    int32
		Size    int32
		Keyword string
		StartAt int64
		EndAt   int64
	}
	AlarmHistoryBO struct {
		Id         uint32         `json:"id"`
		Md5        string         `json:"md5"`
		StrategyId uint32         `json:"strategyId"`
		StrategyBO *StrategyBO    `json:"strategyBO"`
		LevelId    uint32         `json:"levelId"`
		Level      *DictBO        `json:"level"`
		Status     vo.AlarmStatus `json:"status"`
		StartsAt   int64          `json:"startAt"`
		EndsAt     int64          `json:"endAt"`
		Instance   string         `json:"instance"`
		Duration   int64          `json:"duration"`
		Info       *AlertBo       `json:"info"`
		CreatedAt  int64          `json:"createdAt"`
		UpdatedAt  int64          `json:"UpdatedAt"`
	}
)

// String json string
func (b *AlarmHistoryBO) String() string {
	if b == nil {
		return "{}"
	}
	marshal, err := json.Marshal(b)
	if err != nil {
		return "{}"
	}
	return string(marshal)
}

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
		StartAt:     b.StartsAt,
		EndAt:       b.EndsAt,
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
	// TODO 先简单实现, 后面替换成可视化强的, 能表达所有annotations的数据格式
	note := ""
	for k, v := range b.GetInfo().GetAnnotations().Map() {
		note += k + ": " + v + ";\n"
	}
	status := b.Status
	if b.EndsAt > 0 {
		status = vo.AlarmStatusResolved
	}
	return &AlarmRealtimeBO{
		Instance:   b.Instance,
		Note:       note,
		LevelId:    b.LevelId,
		Level:      b.GetLevel(),
		EventAt:    b.StartsAt,
		Status:     status,
		HistoryID:  b.Id,
		StrategyID: b.StrategyId,
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
func (b *AlarmHistoryBO) ToModel() *do.PromAlarmHistory {
	if b == nil {
		return nil
	}
	return &do.PromAlarmHistory{
		BaseModel:  do.BaseModel{ID: b.Id},
		Instance:   b.Instance,
		Status:     b.Status,
		Info:       b.GetInfo().String(),
		StartAt:    b.StartsAt,
		EndAt:      b.EndsAt,
		Duration:   b.Duration,
		StrategyID: b.StrategyId,
		LevelID:    b.LevelId,
		Md5:        b.Md5,
		Strategy:   b.GetStrategyBO().ToModel(),
		Level:      b.GetLevel().ToModel(),
	}
}

// AlarmHistoryModelToBO .
func AlarmHistoryModelToBO(m *do.PromAlarmHistory) *AlarmHistoryBO {
	if m == nil {
		return nil
	}
	return &AlarmHistoryBO{
		Id:         m.ID,
		Md5:        m.Md5,
		StrategyId: m.StrategyID,
		StrategyBO: StrategyModelToBO(m.GetStrategy()),
		LevelId:    m.LevelID,
		Level:      DictModelToBO(m.GetLevel()),
		Status:     m.Status,
		StartsAt:   m.StartAt,
		EndsAt:     m.EndAt,
		Instance:   m.Instance,
		Duration:   m.Duration,
		Info:       StringToAlertBo(m.Info),
		CreatedAt:  m.CreatedAt.Unix(),
		UpdatedAt:  m.UpdatedAt.Unix(),
	}
}
