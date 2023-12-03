package bo

import (
	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/pkg/helper/model"
)

type (
	AlarmInterveneBO struct {
		ID              uint   `json:"id"`
		RealtimeAlarmID uint   `json:"realtimeAlarmID"`
		UserID          uint   `json:"userID"`
		IntervenedAt    int64  `json:"intervenedAt"`
		Remark          string `json:"remark"`

		CreatedAt int64 `json:"createdAt"`
		UpdatedAt int64 `json:"updatedAt"`
		DeletedAt int64 `json:"deletedAt"`
	}
)

// ToModel ...
func (l *AlarmInterveneBO) ToModel() *model.PromAlarmIntervene {
	if l == nil {
		return nil
	}
	return &model.PromAlarmIntervene{
		BaseModel:       query.BaseModel{ID: l.ID},
		RealtimeAlarmID: l.RealtimeAlarmID,
		UserID:          l.UserID,
		IntervenedAt:    l.IntervenedAt,
		Remark:          l.Remark,
	}
}

// AlarmInterveneModelToBO ...
func AlarmInterveneModelToBO(l *model.PromAlarmIntervene) *AlarmInterveneBO {
	if l == nil {
		return nil
	}
	return &AlarmInterveneBO{
		ID:              l.ID,
		RealtimeAlarmID: l.RealtimeAlarmID,
		UserID:          l.UserID,
		IntervenedAt:    l.IntervenedAt,
		Remark:          l.Remark,
		CreatedAt:       l.CreatedAt.Unix(),
		UpdatedAt:       l.UpdatedAt.Unix(),
		DeletedAt:       int64(l.DeletedAt),
	}
}
