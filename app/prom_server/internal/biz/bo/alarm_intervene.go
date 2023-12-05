package bo

import (
	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/api"
	"prometheus-manager/pkg/helper/model"
)

type (
	AlarmInterveneBO struct {
		ID              uint32  `json:"id"`
		RealtimeAlarmID uint32  `json:"realtimeAlarmID"`
		UserID          uint32  `json:"userID"`
		IntervenedAt    int64   `json:"intervenedAt"`
		Remark          string  `json:"remark"`
		IntervenedUser  *UserBO `json:"intervenedUser"`

		CreatedAt int64 `json:"createdAt"`
		UpdatedAt int64 `json:"updatedAt"`
		DeletedAt int64 `json:"deletedAt"`
	}
)

// GetIntervenedUser 获取干预这条信息的用户
func (l *AlarmInterveneBO) GetIntervenedUser() *UserBO {
	if l == nil {
		return nil
	}
	return l.IntervenedUser
}

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

// ToApi ...
func (l *AlarmInterveneBO) ToApi() *api.InterveneInfo {
	if l == nil {
		return nil
	}
	return &api.InterveneInfo{
		IntervenedUser: l.GetIntervenedUser().ToApiSelectV1(),
		IntervenedAt:   l.IntervenedAt,
		Remark:         l.Remark,
		Id:             l.ID,
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
